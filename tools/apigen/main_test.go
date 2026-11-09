package main

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Runs the whole pipeline (load -> build -> render -> write) against the
// sample fixture, into a scratch directory seeded with one fake
// already-implemented resource - proving both halves end to end: a new
// resource (bookmarks) gets generated, and one that looks already covered
// (architectures) gets skipped, without ever touching the real
// pkg/foreman package.
func TestRun_GeneratesNewSkipsExisting(t *testing.T) {
	dir := t.TempDir()

	// Stand in for pkg/foreman's existing hand-written resources: just
	// enough for the EndpointPrefix scan to find a match for
	// "architectures".
	existing := "package foreman\n\nconst ArchitectureEndpointPrefix = \"architectures\"\n"
	if err := os.WriteFile(filepath.Join(dir, "architecture.go"), []byte(existing), 0o644); err != nil {
		t.Fatalf("seeding existing resource file: %v", err)
	}

	if err := run([]string{"-input", "testdata/apidoc_sample.json", "-out", dir}); err != nil {
		t.Fatalf("run: %v", err)
	}

	bookmarksPath := filepath.Join(dir, "bookmarks_generated.go")
	src, err := os.ReadFile(bookmarksPath)
	if err != nil {
		t.Fatalf("expected %s to be generated: %v", bookmarksPath, err)
	}

	if _, err := os.Stat(filepath.Join(dir, "architectures_generated.go")); err == nil {
		t.Fatalf("architectures should have been skipped as already implemented, but a file was generated for it")
	}

	// The generated source must actually be valid Go, not just
	// template output that happens to look right.
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, bookmarksPath, src, parser.AllErrors)
	if err != nil {
		t.Fatalf("generated bookmarks file is not valid Go: %v", err)
	}
	if file.Name.Name != "foreman" {
		t.Fatalf("expected package foreman, got %q", file.Name.Name)
	}

	for _, want := range []string{
		"type ForemanBookmark struct",
		"func (c *Client) CreateBookmark(",
		"func (c *Client) ReadBookmark(",
		"func (c *Client) UpdateBookmark(",
		"func (c *Client) DeleteBookmark(",
		"func (c *Client) QueryBookmark(",
		`const BookmarkEndpointPrefix = "bookmarks"`,
	} {
		if !strings.Contains(string(src), want) {
			t.Errorf("expected generated source to contain %q", want)
		}
	}

	// "name" and "id" are provided by the embedded ForemanObject already
	// and must not be redeclared.
	if strings.Contains(string(src), `Name string`) {
		t.Errorf("generated struct should not redeclare Name - ForemanObject already provides it")
	}
}

// Reproduces a real failure seen against a live Foreman server: apipie
// descriptions routinely contain literal HTML ("<p>...</p>") and embedded
// newlines, which - written verbatim into a "//" comment - broke the
// generated source outright ("expected '}', found '<'"). Also covers a
// field name starting with a digit ("2fa_enabled"), which would otherwise
// produce an invalid Go identifier.
func TestRun_SanitizesHTMLDescriptions(t *testing.T) {
	dir := t.TempDir()

	if err := run([]string{"-input", "testdata/apidoc_html_sample.json", "-out", dir}); err != nil {
		t.Fatalf("run: %v", err)
	}

	path := filepath.Join(dir, "auth_source_externals_generated.go")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected %s to be generated: %v", path, err)
	}

	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, path, src, parser.AllErrors); err != nil {
		t.Fatalf("generated file is not valid Go: %v\n%s", err, src)
	}

	if strings.Contains(string(src), "<p>") || strings.Contains(string(src), "</p>") {
		t.Errorf("expected HTML tags to be stripped from comments, found some in:\n%s", src)
	}
	if !strings.Contains(string(src), "X2faEnabled") {
		t.Errorf("expected the digit-leading field \"2fa_enabled\" to become X2faEnabled")
	}
}

func TestBuildResource_SingularizesAndInfersTypes(t *testing.T) {
	doc, err := loadApidoc("testdata/apidoc_sample.json")
	if err != nil {
		t.Fatalf("loadApidoc: %v", err)
	}

	r, ok := buildResource("bookmarks", doc.Docs.Resources["bookmarks"])
	if !ok {
		t.Fatalf("expected bookmarks to build a resource")
	}
	if r.GoName != "Bookmark" {
		t.Errorf("GoName = %q, want Bookmark", r.GoName)
	}
	if r.EndpointPrefix != "bookmarks" {
		t.Errorf("EndpointPrefix = %q, want bookmarks", r.EndpointPrefix)
	}
	if !r.HasCreate || !r.HasRead || !r.HasUpdate || !r.HasDelete || !r.HasQuery {
		t.Errorf("expected all five CRUD operations, got %+v", r)
	}

	byName := map[string]field{}
	for _, f := range r.Fields {
		byName[f.JSONName] = f
	}
	if byName["public"].GoType != "bool" {
		t.Errorf("public field should infer bool from its validator, got %s", byName["public"].GoType)
	}
	if byName["controller"].GoType != "string" {
		t.Errorf("controller field should default to string, got %s", byName["controller"].GoType)
	}
	if _, present := byName["name"]; present {
		t.Errorf("\"name\" should be filtered out - ForemanObject already provides it")
	}
}

// Reproduces the shape of a real Foreman host create param list: a
// "_id" field with no explicit validator (medium_id), a boolean field
// apidoc's validator text doesn't flag as boolean (managed), a plural
// "_ids" field (config_group_ids), and two levels of nested params
// (compute_attributes, and interfaces_attributes whose own entries each
// carry their own compute_attributes). Verifies the type-inference
// fallbacks and that nested params generate real struct types instead of
// collapsing to "string".
func TestBuildResource_NestedParamsAndIDHeuristics(t *testing.T) {
	doc, err := loadApidoc("testdata/apidoc_nested_sample.json")
	if err != nil {
		t.Fatalf("loadApidoc: %v", err)
	}

	r, ok := buildResource("hosts", doc.Docs.Resources["hosts"])
	if !ok {
		t.Fatalf("expected hosts to build a resource")
	}

	byName := map[string]field{}
	for _, f := range r.Fields {
		byName[f.JSONName] = f
	}

	if got := byName["medium_id"].GoType; got != "int" {
		t.Errorf("medium_id should infer int from its _id suffix, got %s", got)
	}
	if got := byName["managed"].GoType; got != "bool" {
		t.Errorf("managed should infer bool from the known-boolean-name fallback, got %s", got)
	}
	if got := byName["config_group_ids"].GoType; got != "[]int" {
		t.Errorf("config_group_ids should infer []int from its _ids suffix, got %s", got)
	}
	if got := byName["compute_attributes"].GoType; got != "*HostComputeAttributes" {
		t.Errorf("optional compute_attributes should generate a pointer-to-nested HostComputeAttributes type, got %s", got)
	}
	if got := byName["interfaces_attributes"].GoType; got != "[]HostInterfacesAttributes" {
		t.Errorf("interfaces_attributes should generate a nested, sliced HostInterfacesAttributes type, got %s", got)
	}

	nestedByName := map[string]nestedType{}
	for _, nt := range r.NestedTypes {
		nestedByName[nt.Name] = nt
	}
	computeAttrs, ok := nestedByName["HostComputeAttributes"]
	if !ok {
		t.Fatalf("expected a HostComputeAttributes nested type, got %+v", r.NestedTypes)
	}
	computeAttrsByName := map[string]field{}
	for _, f := range computeAttrs.Fields {
		computeAttrsByName[f.JSONName] = f
	}
	if got := computeAttrsByName["cpus"].GoType; got != "int" {
		t.Errorf("compute_attributes.cpus should infer int from its validator, got %s", got)
	}

	ifaces, ok := nestedByName["HostInterfacesAttributes"]
	if !ok {
		t.Fatalf("expected a HostInterfacesAttributes nested type, got %+v", r.NestedTypes)
	}
	ifacesByName := map[string]field{}
	for _, f := range ifaces.Fields {
		ifacesByName[f.JSONName] = f
	}
	if _, present := ifacesByName["name"]; !present {
		t.Errorf("nested interfaces_attributes should keep its own \"name\" field - it doesn't embed ForemanObject")
	}
	if got := ifacesByName["compute_attributes"].GoType; got != "*HostInterfacesAttributesComputeAttributes" {
		t.Errorf("interfaces_attributes.compute_attributes should generate its own pointer-to-nested type, got %s", got)
	}
	if _, present := nestedByName["HostInterfacesAttributesComputeAttributes"]; !present {
		t.Errorf("expected a second-level nested type HostInterfacesAttributesComputeAttributes, got %+v", r.NestedTypes)
	}

	// The whole thing must still render as valid, formatted Go.
	src, err := renderResource(r)
	if err != nil {
		t.Fatalf("renderResource: %v", err)
	}
	if !strings.Contains(string(src), "type HostComputeAttributes struct") {
		t.Errorf("expected rendered source to declare HostComputeAttributes, got:\n%s", src)
	}
	if !strings.Contains(string(src), "type HostInterfacesAttributes struct") {
		t.Errorf("expected rendered source to declare HostInterfacesAttributes, got:\n%s", src)
	}
}

// Reproduces a real bug found in a generated file: a sub-resource nested
// under a parent by ID (e.g. "/api/hosts/:host_id/parameters") had its
// EndpointPrefix truncated to just "hosts" - the old endpointPrefixFrom
// cut the URL at the *first* "/:" unconditionally, discarding everything
// after it, including "/parameters". That pointed every generated CRUD
// method at the host endpoint instead of its parameters. Verifies the
// fix: a non-trailing path parameter becomes a "%d" placeholder (keeping
// what follows it) and its name is captured as a leading argument on
// every generated method, while a *trailing* "/:id" (the resource's own
// id, already appended separately by the templates) is still dropped.
func TestBuildResource_NestedParentResource(t *testing.T) {
	doc, err := loadApidoc("testdata/apidoc_nested_parent_sample.json")
	if err != nil {
		t.Fatalf("loadApidoc: %v", err)
	}

	r, ok := buildResource("parameters", doc.Docs.Resources["parameters"])
	if !ok {
		t.Fatalf("expected parameters to build a resource")
	}

	if r.EndpointPrefix != "hosts/%d/parameters" {
		t.Errorf(`EndpointPrefix = %q, want "hosts/%%d/parameters"`, r.EndpointPrefix)
	}
	if len(r.ParentParams) != 1 || r.ParentParams[0] != "hostId" {
		t.Errorf("ParentParams = %v, want [\"hostId\"]", r.ParentParams)
	}

	src, err := renderResource(r)
	if err != nil {
		t.Fatalf("renderResource: %v", err)
	}
	for _, want := range []string{
		`func (c *Client) CreateParameter(ctx context.Context, hostId int, obj *ForemanParameter) (*ForemanParameter, error) {`,
		`func (c *Client) DeleteParameter(ctx context.Context, hostId int, id int) error {`,
		`c.Do(ctx, http.MethodPost, fmt.Sprintf(ParameterEndpointPrefix, hostId), nil, body, out)`,
		`fmt.Sprintf("%s/%d", fmt.Sprintf(ParameterEndpointPrefix, hostId), id)`,
		`func (c *Client) QueryParameter(ctx context.Context, hostId int, search string) (*List[ForemanParameter], error) {`,
	} {
		if !strings.Contains(string(src), want) {
			t.Errorf("expected generated source to contain %q, got:\n%s", want, src)
		}
	}
}

// Reproduces a real gap found after regenerating against a live Foreman
// server: a host's own "compute_attributes" has no nested apidoc schema
// at all (it's provider-dependent - VMware wants cluster/path/volumes,
// libvirt wants something else, and apidoc documents neither), so it
// used to fall through goTypeFor's default and render as a plain
// "string" field - silently wrong, since Foreman expects a JSON object
// there, not a string literal. Verifies the fix: a "*_attributes"-named
// param with no nested .Params becomes "map[string]interface{}" instead.
func TestGoTypeForAttributesBag(t *testing.T) {
	cases := []struct {
		name      string
		validator string
		want      string
	}{
		{"compute_attributes", "", "map[string]interface{}"},
		{"interfaces_attributes", "", "map[string]interface{}"},
		{"controller", "", "string"}, // unaffected: no "_attributes" suffix
		{"location_id", "", "int"},   // unaffected: _id rule still wins
		{"enabled", "", "bool"},      // unaffected: known-boolean fallback still wins
	}
	for _, c := range cases {
		if got := goTypeForAttributesBag(c.name, c.validator); got != c.want {
			t.Errorf("goTypeForAttributesBag(%q, %q) = %q, want %q", c.name, c.validator, got, c.want)
		}
	}
}

// Reproduces a real failure seen against a live Foreman server: CreateHost
// failed to parse its own response with "json: cannot unmarshal array into
// Go struct field ForemanHost.capabilities of type string" - a host's
// "capabilities" comes back as e.g. ["build","image"], but apidoc gives
// this param no description or validator text at all, so there's no
// signal for goTypeFor's text-based heuristics to catch - it needs the
// knownStringArrayNames fallback.
func TestGoTypeFor_KnownStringArrayNames(t *testing.T) {
	if got := goTypeFor("capabilities", ""); got != "[]string" {
		t.Errorf(`goTypeFor("capabilities", "") = %q, want "[]string"`, got)
	}
}

func TestAlreadyImplemented(t *testing.T) {
	existing := []string{"architectures", "katello/products", "smart_class_parameters/%d/override_values"}

	cases := []struct {
		resource resource
		want     bool
	}{
		{resource{EndpointPrefix: "architectures", PluralName: "architectures"}, true},
		{resource{EndpointPrefix: "bookmarks", PluralName: "bookmarks"}, false},
		{resource{EndpointPrefix: "products", PluralName: "products"}, true}, // substring of katello/products
	}
	for _, c := range cases {
		if got := alreadyImplemented(c.resource, existing); got != c.want {
			t.Errorf("alreadyImplemented(%+v) = %v, want %v", c.resource, got, c.want)
		}
	}
}

func TestSingularize(t *testing.T) {
	cases := map[string]string{
		"bookmarks":  "bookmark",
		"categories": "category",
		"classes":    "class",
		"boxes":      "box",
		"watches":    "watch",
		"dishes":     "dish",
		"data":       "data", // irregular - intentionally not handled, documents the limitation
	}
	for plural, want := range cases {
		if got := singularize(plural); got != want {
			t.Errorf("singularize(%q) = %q, want %q", plural, got, want)
		}
	}
}
