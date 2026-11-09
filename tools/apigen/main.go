// Command apigen scaffolds new pkg/foreman resource files from a Foreman
// apidoc JSON document - the same machine-readable API description Foreman
// serves at <server>/apidoc/v2.json, and the same one python-foreman's
// client.py (https://github.com/david-caro/python-foreman) parses at
// runtime to build its bindings dynamically via exec(). Go can't do that
// trick in a type-safe way, so apigen does the equivalent job once, ahead
// of time: parse the JSON, emit real .go source per resource.
//
// It only generates resources that look genuinely new - anything whose
// apidoc path already overlaps an existing `*EndpointPrefix` constant in
// the output directory is skipped, so running this against your whole
// Foreman instance's apidoc never clobbers or duplicates a resource
// written by hand in pkg/foreman.
//
// Generated code is a starting point, not a finished resource: review it,
// rename fields to match this package's conventions where it guessed
// wrong, and fold in the things apidoc can't tell us - custom JSON framing
// quirks, nested sub-resources, taxonomy handling. See README.md.
//
// Usage:
//
//	go run ./tools/apigen -input <path-or-foreman-url> [-out pkg/foreman] [-only res1,res2] [-force]
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "apigen:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("apigen", flag.ContinueOnError)
	input := fs.String("input", "", "path to an apidoc JSON file, or a Foreman base URL to fetch /apidoc/v2.json from (required)")
	out := fs.String("out", "pkg/foreman", "directory to write generated <resource>_generated.go files into, and to scan for already-implemented resources")
	only := fs.String("only", "", "comma-separated apidoc resource names to generate (default: every new one found)")
	force := fs.String("force", "", "comma-separated apidoc resource names to generate even if they look already implemented")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *input == "" {
		fs.Usage()
		return fmt.Errorf("-input is required")
	}

	doc, err := loadApidoc(*input)
	if err != nil {
		return err
	}

	existing, err := existingEndpointPrefixes(*out)
	if err != nil {
		return err
	}

	onlySet := toSet(*only)
	forceSet := toSet(*force)

	names := make([]string, 0, len(doc.Docs.Resources))
	for name := range doc.Docs.Resources {
		names = append(names, name)
	}
	sort.Strings(names)

	// One bad resource (a description shaped in a way this tool doesn't
	// yet handle, say) must not block every other resource in the same
	// apidoc from generating - collect failures and keep going, instead
	// of aborting the whole batch on the first one.
	var generated, skipped, failed []string
	for _, name := range names {
		if len(onlySet) > 0 && !onlySet[name] {
			continue
		}

		r, ok := buildResource(name, doc.Docs.Resources[name])
		if !ok {
			continue
		}

		if !forceSet[name] && alreadyImplemented(r, existing) {
			skipped = append(skipped, name)
			continue
		}

		src, err := renderResource(r)
		if err != nil {
			failed = append(failed, fmt.Sprintf("%s: %v", name, err))
			continue
		}

		outPath := filepath.Join(*out, r.PluralName+"_generated.go")
		if err := os.WriteFile(outPath, src, 0o644); err != nil { // #nosec G306 -- generated Go source, meant to be world-readable like any other .go file
			failed = append(failed, fmt.Sprintf("%s: writing %s: %v", name, outPath, err))
			continue
		}
		generated = append(generated, outPath)
	}

	for _, p := range generated {
		fmt.Println("generated:", p)
	}
	for _, s := range skipped {
		fmt.Println("skipped (already implemented):", s)
	}
	for _, f := range failed {
		fmt.Fprintln(os.Stderr, "failed:", f)
	}
	if len(generated) == 0 && len(failed) == 0 {
		fmt.Println("nothing to generate")
	}
	if len(failed) > 0 {
		return fmt.Errorf("%d resource(s) failed to generate (see above); %d generated successfully", len(failed), len(generated))
	}
	return nil
}

func toSet(csv string) map[string]bool {
	set := map[string]bool{}
	for _, s := range strings.Split(csv, ",") {
		if s = strings.TrimSpace(s); s != "" {
			set[s] = true
		}
	}
	return set
}

var endpointPrefixConstRe = regexp.MustCompile(`\w+EndpointPrefix\s*=\s*"([^"]*)"`)

// existingEndpointPrefixes scans every non-generated, non-test .go file in
// dir for `XxxEndpointPrefix = "..."` constants, returning the set of
// their string values. That's the same signal every hand-written resource
// already exposes, so it's what alreadyImplemented checks new resources
// against.
func existingEndpointPrefixes(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading %s: %w", dir, err)
	}

	var prefixes []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") ||
			strings.HasSuffix(name, "_test.go") || strings.HasSuffix(name, "_generated.go") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, name)) // #nosec G304 -- dir is the -out flag; name comes from listing that same dir
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", name, err)
		}
		for _, m := range endpointPrefixConstRe.FindAllStringSubmatch(string(b), -1) {
			prefixes = append(prefixes, m[1])
		}
	}
	return prefixes, nil
}

// alreadyImplemented reports whether r's endpoint looks covered by an
// existing hand-written resource already - a loose substring check in
// both directions, since existing prefixes come in several shapes
// ("hosts", "/common_parameters", "katello/products",
// "smart_class_parameters/%d/override_values").
func alreadyImplemented(r resource, existing []string) bool {
	candidates := []string{r.EndpointPrefix, r.PluralName}
	for _, existingPrefix := range existing {
		if existingPrefix == "" {
			continue
		}
		for _, c := range candidates {
			if c == "" {
				continue
			}
			if strings.Contains(existingPrefix, c) || strings.Contains(c, existingPrefix) {
				return true
			}
		}
	}
	return false
}
