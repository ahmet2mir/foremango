// Parsing for Foreman's apipie-generated API documentation JSON (the same
// format served at <foreman>/apidoc/v1.json or /apidoc/v2.json, and the
// format python-foreman's client.py reads to build its API bindings at
// runtime - see https://github.com/david-caro/python-foreman). This file
// only models the subset of that schema apigen actually uses.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// apidoc is the top-level shape of the JSON document.
type apidoc struct {
	Docs struct {
		Resources map[string]resourceDoc `json:"resources"`
	} `json:"docs"`
}

type resourceDoc struct {
	FullDescription string      `json:"full_description"`
	Methods         []methodDoc `json:"methods"`
}

type methodDoc struct {
	Name   string     `json:"name"`
	APIs   []apiDoc   `json:"apis"`
	Params []paramDoc `json:"params"`
}

type apiDoc struct {
	APIURL           string `json:"api_url"`
	HTTPMethod       string `json:"http_method"`
	ShortDescription string `json:"short_description"`
}

type paramDoc struct {
	Name        string     `json:"name"`
	Required    bool       `json:"required"`
	Description string     `json:"description"`
	Validator   string     `json:"validator"`
	Params      []paramDoc `json:"params"`
}

// loadApidoc reads an apidoc document from a local file path or, if src
// looks like a URL, fetches <src>/apidoc/v2.json over HTTP (Foreman serves
// this unauthenticated).
func loadApidoc(src string) (*apidoc, error) {
	var r io.Reader
	if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
		url := strings.TrimSuffix(src, "/") + "/apidoc/v2.json"
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			return nil, fmt.Errorf("fetching %s: %w", url, err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return nil, fmt.Errorf("fetching %s: unexpected status %s", url, resp.Status)
		}
		r = resp.Body
	} else {
		f, err := os.Open(src) // #nosec G304 -- src is the -input flag; this CLI's own user chooses the path
		if err != nil {
			return nil, fmt.Errorf("opening %s: %w", src, err)
		}
		defer func() { _ = f.Close() }()
		r = f
	}

	var doc apidoc
	if err := json.NewDecoder(r).Decode(&doc); err != nil {
		return nil, fmt.Errorf("decoding apidoc JSON: %w", err)
	}
	return &doc, nil
}
