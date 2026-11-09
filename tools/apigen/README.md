# apigen

Scaffolds new `pkg/foreman` resource files from a Foreman apidoc JSON - the
same machine-readable API description Foreman serves at
`<server>/apidoc/v2.json`, and the one
[python-foreman](https://github.com/david-caro/python-foreman)'s
`client.py` parses at runtime to build its bindings dynamically via
`exec()`. Go can't do that trick in a type-safe way, so apigen does the
equivalent job once, ahead of time: parse the JSON, emit real `.go` source
per resource, reviewed and committed like any other code.

## What it does

For each resource in the apidoc JSON:

1. Classifies its `index`/`show`/`create`/`update`/`destroy` methods into
   Query/Read/Create/Update/Delete.
2. Finds the writable fields - apipie nests them under a param named after
   the resource (e.g. `"bookmark": {name, controller, query, public}`) -
   and best-effort infers each one's Go type from its free-text validator
   string (`"Must be Integer"` → `int`, `"Must be one of: true, false"` →
   `bool`, everything else → `string`).
3. Renders a `<resource>_generated.go` file: a `Foreman<X>` struct
   embedding `ForemanObject`, and `Create<X>`/`Read<X>`/`Update<X>`/
   `Delete<X>`/`Query<X>` methods on `*Client`, each a single
   `Client.Do` call. `Query<X>` takes a Foreman search expression and
   returns a `*List[Foreman<X>]`.

**It only generates resources that look genuinely new.** Before writing
anything, it scans every non-generated `.go` file in the output directory
for `XxxEndpointPrefix = "..."` constants and skips any apidoc resource
whose path overlaps one. A resource you write by hand (with such a
constant) is therefore never clobbered or duplicated by a regeneration.

## What it doesn't do

The output is a **starting point to review, not a finished resource**.
apidoc can't tell apigen about things specific to how foremango talks to
Foreman:

- custom JSON framing some resources need (fields that round-trip under a
  different key on the way in than on the way out - see the README's
  "Package layout" notes on this),
- nested sub-resources (filters, rules, parameters as their own
  mini-CRUD),
- the handful of resources where creating one is secretly a multi-step
  dance (e.g. publishing a Katello content view).

Treat a generated file as a draft: read it, fix field names/types it
guessed wrong (`-input`'s validator strings are free text, not a real type
system), and fold in anything apidoc doesn't capture - the same way you'd
review any other generated code before committing it.

## Usage

```sh
# from a local copy of the JSON
go run ./tools/apigen -input apidoc/v2.json -out pkg/foreman

# or fetch it directly from a running Foreman server
# (this hits <url>/apidoc/v2.json, which Foreman serves unauthenticated)
go run ./tools/apigen -input https://foreman.example.com -out pkg/foreman
```

Also wired as `go generate ./pkg/foreman` (see `pkg/foreman/generate.go`),
which expects the JSON at `apidoc/v2.json` in the repo root - fetch your
own instance's copy there first:

```sh
curl -s https://foreman.example.com/apidoc/v2.json -o apidoc/v2.json
go generate ./pkg/foreman
```

Flags:

| Flag | Default | Meaning |
|---|---|---|
| `-input` | *(required)* | Path to a local apidoc JSON file, or a Foreman base URL to fetch `/apidoc/v2.json` from |
| `-out` | `pkg/foreman` | Where to write `<resource>_generated.go` files, and what to scan for already-implemented resources |
| `-only` | *(all)* | Comma-separated apidoc resource names to consider, e.g. `-only bookmarks,ptables` |
| `-force` | *(none)* | Comma-separated resource names to generate even if they look already implemented |

## Testing it

`go test ./tools/apigen/...` runs the whole pipeline against
`testdata/apidoc_sample.json` - a small hand-written fixture with one
resource that looks already implemented (`architectures`) and one that
doesn't (`bookmarks`) - into a temp directory, and checks the output is
valid Go containing the expected types and methods. It never touches
`pkg/foreman` itself.
