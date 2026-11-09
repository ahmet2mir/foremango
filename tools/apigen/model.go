// Turns the raw apidoc JSON model into something codegen.go can render:
// one resource struct's fields, typed as best as a free-text apipie
// validator description lets us guess, plus which of the five CRUD
// operations the apidoc actually documents.
package main

import (
	"html"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type field struct {
	GoName    string // e.g. "Controller"
	JSONName  string // e.g. "controller"
	GoType    string // "string", "int", "bool", a nested type name, or "[]" + any of those
	Comment   string
	OmitEmpty bool
}

// nestedType is an auxiliary struct generated for a param that apidoc
// itself describes with nested params (e.g. a host's "compute_attributes"
// or "interfaces_attributes") rather than a single scalar value. codegen.go
// renders one Go struct per nestedType, alongside the resource's own.
type nestedType struct {
	Name   string // e.g. "HostComputeAttributes" - unique within the resource
	Fields []field
}

type resource struct {
	PluralName     string // apipie resource name, e.g. "bookmarks"
	SingularName   string // e.g. "bookmark" - the Client.Payload resource name and JSON wrapper key
	GoName         string // PascalCase singular, e.g. "Bookmark"
	EndpointPrefix string // e.g. "bookmarks", or "hosts/%d/parameters" when ParentParams is non-empty
	// Go parameter names (e.g. "hostId") for each parent-resource path
	// segment in EndpointPrefix's "%d" placeholders, in order - set when
	// this resource is nested under another (e.g. a host's parameters,
	// "/api/hosts/:host_id/parameters"). Every generated CRUD method takes
	// these as its leading arguments.
	ParentParams []string
	Description  string
	Fields       []field
	NestedTypes  []nestedType
	HasCreate    bool
	HasRead      bool
	HasUpdate    bool
	HasDelete    bool
	HasQuery     bool
}

// buildResource extracts everything codegen.go needs for one apidoc
// resource. Returns ok=false if the resource has none of the five
// standard apipie REST actions (index/show/create/update/destroy) - some
// apidoc resources are pure RPC-style actions with nothing CRUD-shaped to
// generate.
func buildResource(name string, doc resourceDoc) (resource, bool) {
	r := resource{
		PluralName:   name,
		SingularName: singularize(name),
		Description:  sanitizeText(doc.FullDescription, 120),
	}
	r.GoName = pascalCase(r.SingularName)

	var createParams, updateParams []paramDoc
	for _, m := range doc.Methods {
		switch m.Name {
		case "index":
			r.HasQuery = true
			r.setEndpointPrefixFrom(m)
		case "show":
			r.HasRead = true
			r.setEndpointPrefixFrom(m)
		case "create":
			r.HasCreate = true
			createParams = m.Params
			r.setEndpointPrefixFrom(m)
		case "update":
			r.HasUpdate = true
			updateParams = m.Params
			r.setEndpointPrefixFrom(m)
		case "destroy":
			r.HasDelete = true
			r.setEndpointPrefixFrom(m)
		}
	}

	if !r.HasCreate && !r.HasRead && !r.HasUpdate && !r.HasDelete && !r.HasQuery {
		return resource{}, false
	}

	fields := fieldParams(createParams, r.SingularName)
	if len(fields) == 0 {
		fields = fieldParams(updateParams, r.SingularName)
	}
	seenParam := map[string]bool{}
	seenGoName := map[string]bool{}
	for _, p := range fields {
		// "id" and "name" are already provided by the embedded
		// ForemanObject - redeclaring them would just shadow it.
		if p.Name == "id" || p.Name == "name" || seenParam[p.Name] {
			continue
		}
		seenParam[p.Name] = true

		goName := pascalCase(p.Name)
		if seenGoName[goName] {
			// Two different apidoc param names collapsed to the same Go
			// identifier (pascalCase only distinguishes on separators) -
			// keep the field, under a name that won't collide, rather
			// than silently drop it or emit a duplicate struct field.
			for n := 2; seenGoName[goName]; n++ {
				goName = pascalCase(p.Name) + strconv.Itoa(n)
			}
		}
		seenGoName[goName] = true

		goType := goTypeForAttributesBag(p.Name, p.Validator)
		if len(p.Params) > 0 {
			// apidoc describes this param's own shape (e.g. a host's
			// "compute_attributes" or "interfaces_attributes") rather than
			// just a free-text validator - generate a real nested struct
			// for it instead of falling back to "string".
			typeName := r.GoName + goName
			nestedFields := buildNestedFields(typeName, p.Params, &r.NestedTypes)
			r.NestedTypes = append(r.NestedTypes, nestedType{Name: typeName, Fields: nestedFields})
			switch {
			case looksLikeArray(p):
				goType = "[]" + typeName
			case !p.Required:
				// encoding/json's "omitempty" only treats a nil *pointer*
				// as empty for struct-typed fields, never a zero-value
				// struct - without this, an optional nested object would
				// always marshal as "{}" even when the caller never set it.
				goType = "*" + typeName
			default:
				goType = typeName
			}
		}

		r.Fields = append(r.Fields, field{
			GoName:    goName,
			JSONName:  p.Name,
			GoType:    goType,
			Comment:   sanitizeText(p.Description, 100),
			OmitEmpty: !p.Required,
		})
	}

	return r, true
}

// buildNestedFields turns one param's own nested .Params into a field list,
// recursively generating further nestedTypes (appended to *nested) for any
// grand-children that are themselves nested (e.g. a host's
// interfaces_attributes each carry their own compute_attributes). Unlike
// the top-level resource fields, "id"/"name" are kept here - nested structs
// don't embed ForemanObject, so a field named "name" (e.g. an interface's
// own name) is a real field, not a collision.
func buildNestedFields(typeNamePrefix string, params []paramDoc, nested *[]nestedType) []field {
	var fields []field
	seenGoName := map[string]bool{}
	for _, p := range params {
		goName := pascalCase(p.Name)
		if seenGoName[goName] {
			for n := 2; seenGoName[goName]; n++ {
				goName = pascalCase(p.Name) + strconv.Itoa(n)
			}
		}
		seenGoName[goName] = true

		goType := goTypeForAttributesBag(p.Name, p.Validator)
		if len(p.Params) > 0 {
			childTypeName := typeNamePrefix + goName
			childFields := buildNestedFields(childTypeName, p.Params, nested)
			*nested = append(*nested, nestedType{Name: childTypeName, Fields: childFields})
			switch {
			case looksLikeArray(p):
				goType = "[]" + childTypeName
			case !p.Required:
				goType = "*" + childTypeName
			default:
				goType = childTypeName
			}
		}

		fields = append(fields, field{
			GoName:    goName,
			JSONName:  p.Name,
			GoType:    goType,
			Comment:   sanitizeText(p.Description, 100),
			OmitEmpty: !p.Required,
		})
	}
	return fields
}

// looksLikeArray guesses whether a nested param represents a list of
// objects (e.g. a host's "interfaces_attributes") rather than a single
// nested object (e.g. its "compute_attributes"). apidoc has no explicit
// array flag for params, so this goes on the only signal available: apipie
// descriptions for list-shaped params conventionally say so in prose
// ("Array of ...", "array of hashes", etc).
func looksLikeArray(p paramDoc) bool {
	s := strings.ToLower(p.Description + " " + p.Validator)
	return strings.Contains(s, "array")
}

// fieldParams returns the field list a create/update method actually
// describes. apipie nests the writable fields under a param named after
// the resource (e.g. "bookmark": {name, controller, query, public}) -
// find that one and use its children; if the method doesn't nest (some
// older/simpler resources don't), fall back to its own top-level params.
func fieldParams(params []paramDoc, singularName string) []paramDoc {
	for _, p := range params {
		if p.Name == singularName && len(p.Params) > 0 {
			return p.Params
		}
	}
	for _, p := range params {
		if len(p.Params) > 0 {
			return p.Params
		}
	}
	var out []paramDoc
	for _, p := range params {
		if p.Name != "id" {
			out = append(out, p)
		}
	}
	return out
}

// setEndpointPrefixFrom fills in r.EndpointPrefix/r.ParentParams from m's
// first documented API URL, the first time it's called for r (apidoc
// repeats the same resource's URL shape across its index/show/create/
// update/destroy methods, so every call after the first is a no-op).
func (r *resource) setEndpointPrefixFrom(m methodDoc) {
	if r.EndpointPrefix != "" {
		return
	}
	if len(m.APIs) == 0 {
		return
	}
	url := m.APIs[0].APIURL
	url = strings.TrimPrefix(url, "/api/")
	url = strings.TrimPrefix(url, "/katello/api/")

	segments := strings.Split(url, "/")
	var kept []string
	var parentParams []string
	for i, seg := range segments {
		if !strings.HasPrefix(seg, ":") {
			kept = append(kept, seg)
			continue
		}
		if i == len(segments)-1 {
			// A path parameter in the final segment is this resource's
			// own id (e.g. the ":id" in ".../bookmarks/:id") - the
			// generated show/update/destroy methods already append
			// "/%d" for that themselves, so stop rather than keep it.
			break
		}
		// A path parameter anywhere else identifies a parent resource
		// this one is nested under (e.g. the ":host_id" in
		// "/hosts/:host_id/parameters") - keep its place with a "%d"
		// placeholder and remember its name so the generated CRUD
		// methods can take it as a leading argument. Without this, the
		// segments after it (here, "parameters") would be dropped
		// entirely, leaving EndpointPrefix pointing at the parent
		// resource instead of this one.
		kept = append(kept, "%d")
		parentParams = append(parentParams, lowerFirst(pascalCase(strings.TrimPrefix(seg, ":"))))
	}
	r.EndpointPrefix = strings.Join(kept, "/")
	r.ParentParams = parentParams
}

// lowerFirst lower-cases just the first rune, e.g. "HostId" -> "hostId" -
// used to turn a PascalCase type/field name into a Go parameter name.
func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

// knownBooleanNames covers Foreman API fields that are booleans in
// practice but whose apidoc validator text doesn't reliably say so (empty,
// or worded in a way goTypeFor's validator-text match doesn't catch).
// Confirmed against a real Foreman host JSON response, where e.g.
// "enabled"/"managed"/"build" all come back as true/false, not strings.
var knownBooleanNames = map[string]bool{
	"enabled": true, "managed": true, "build": true, "disk": true,
	"primary": true, "provision": true, "virtual": true, "public": true,
	"locked": true, "unattended": true, "use_image": true,
	"hidden_value": true, "strict": true, "async": true,
	"ignore_types": true, "onscreen": true, "open": true, "default": true,
	"require_ssl": true, "overwrite": true, "subscribed": true,
}

// knownStringArrayNames covers Foreman API fields that are arrays of
// strings in practice but have no apidoc description or validator text at
// all to infer that from (empty of both, unlike the nested-object params
// looksLikeArray handles - this is for a flat scalar param whose apidoc
// entry gives no signal whatsoever). Confirmed against a real Foreman
// host JSON response, where "capabilities" comes back as e.g.
// ["build","image"], not a string.
var knownStringArrayNames = map[string]bool{
	"capabilities": true,
}

// goTypeFor infers a field's Go type from its apidoc param name and
// free-text validator. The validator is apipie's own best attempt at a
// machine-readable type and is checked first; "_id"/"_ids" name suffixes
// and the known-name fallbacks (knownBooleanNames, knownStringArrayNames)
// fill in where real Foreman apidoc validators are empty or worded in a
// way the validator check won't catch.
func goTypeFor(name, validator string) string {
	v := strings.ToLower(validator)
	switch {
	case strings.Contains(v, "integer") || strings.Contains(v, "number"):
		return "int"
	case strings.Contains(v, "true, false") || strings.Contains(v, "boolean"):
		return "bool"
	}

	n := strings.ToLower(name)
	switch {
	case strings.HasSuffix(n, "_ids"):
		return "[]int"
	case strings.HasSuffix(n, "_id"):
		return "int"
	case knownBooleanNames[n]:
		return "bool"
	case knownStringArrayNames[n]:
		return "[]string"
	default:
		return "string"
	}
}

// goTypeForAttributesBag is goTypeFor, except a "*_attributes" param with
// no nested apidoc schema (e.g. a host's "compute_attributes", which
// varies entirely by compute resource provider - VMware wants cluster/
// path/volumes, libvirt wants something else, and apidoc documents
// neither) gets "map[string]interface{}" instead of goTypeFor's "string"
// fallback. "_attributes" is Foreman's own naming convention for a
// nested JSON object/array-of-objects parameter, never a plain scalar -
// typing it string would silently send a string literal where the
// Foreman API expects a JSON object, which encoding/json accepts without
// complaint and the server then rejects. Only reached when len(p.Params)
// == 0 - whenever apidoc does describe the nested shape, the caller
// builds a real struct instead of calling this at all.
func goTypeForAttributesBag(name, validator string) string {
	if strings.HasSuffix(strings.ToLower(name), "_attributes") {
		return "map[string]interface{}"
	}
	return goTypeFor(name, validator)
}

// singularize is a best-effort English de-pluralizer for the common
// patterns apipie resource names actually use (snake_case, mostly regular
// plurals). It won't get every irregular noun right (e.g. "data" has no
// singular form distinct from its plural, so it's left as "data") -
// rename the generated type/methods by hand for anything it mangles,
// same as any other part of the output you review before committing.
func singularize(plural string) string {
	switch {
	case strings.HasSuffix(plural, "ies"):
		return strings.TrimSuffix(plural, "ies") + "y"
	case strings.HasSuffix(plural, "sses"), strings.HasSuffix(plural, "ches"),
		strings.HasSuffix(plural, "shes"), strings.HasSuffix(plural, "xes"):
		return strings.TrimSuffix(plural, "es")
	case strings.HasSuffix(plural, "s") && !strings.HasSuffix(plural, "ss"):
		return strings.TrimSuffix(plural, "s")
	default:
		return plural
	}
}

// pascalCase converts a snake_case (or already-plain) identifier to
// PascalCase, e.g. "content_view_filters" -> "ContentViewFilters". A
// param apipie names starting with a digit (e.g. "2fa_enabled") would
// otherwise produce an invalid Go identifier, so that case gets an "X"
// prefix - the same convention protoc-gen-go and similar tools use.
func pascalCase(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == '_' || r == '-' })
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		r := []rune(p)
		b.WriteRune(unicode.ToUpper(r[0]))
		b.WriteString(string(r[1:]))
	}
	out := b.String()
	if out == "" {
		return out
	}
	if unicode.IsDigit([]rune(out)[0]) {
		out = "X" + out
	}
	return out
}

var htmlTagRe = regexp.MustCompile(`<[^>]*>`)

// sanitizeText makes an apipie description safe to embed as a single-line
// Go "//" comment: apipie descriptions routinely contain literal HTML
// (e.g. "<p>REPLACE locations with given ids</p>") and embedded newlines,
// either of which breaks the generated source outright if written
// verbatim - an unescaped "<" or a bare newline doesn't end a comment the
// way a human-written one would, it just corrupts whatever follows. This
// strips tags (matching the same stripping python-foreman's own
// client.py does for descriptions - see MethodAPIDescription's
// exclude_html_reg), unescapes HTML entities apipie leaves behind (e.g.
// "&#39;" for an apostrophe), collapses all whitespace/newlines to single
// spaces, and truncates to maxLen.
func sanitizeText(s string, maxLen int) string {
	s = htmlTagRe.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > maxLen {
		s = strings.TrimSpace(s[:maxLen]) + "..."
	}
	return s
}
