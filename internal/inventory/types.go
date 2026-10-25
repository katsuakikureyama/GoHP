// Package inventory reads Composer metadata as data, never as executable code
// or proof that a dependency can be compiled or run by GoHP.
package inventory

import "encoding/json"

type Report struct {
	Schema            string     `json:"schema"`
	Status            string     `json:"status"`
	Compatibility     string     `json:"compatibility"`
	Inputs            []Input    `json:"inputs"`
	Root              Package    `json:"root"`
	Packages          []Package  `json:"packages"`
	Links             []Link     `json:"links"`
	PlatformOverrides []Override `json:"platform_overrides"`
	LockContentHash   string     `json:"lock_content_hash,omitempty"`
	LockFreshness     string     `json:"lock_freshness"`
	Notices           []string   `json:"notices"`
}

type Input struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	SHA256 string `json:"sha256,omitempty"`
}

type Package struct {
	Name             string     `json:"name"`
	Version          string     `json:"version,omitempty"`
	Type             string     `json:"type,omitempty"`
	Scope            string     `json:"scope"`
	Licenses         []string   `json:"licenses"`
	Source           *Reference `json:"source,omitempty"`
	Dist             *Reference `json:"dist,omitempty"`
	ScriptEvents     []string   `json:"script_events"`
	AutoloadModes    []string   `json:"autoload_modes"`
	AutoloadDevModes []string   `json:"autoload_dev_modes"`
}

type Reference struct {
	Type       string `json:"type,omitempty"`
	Reference  string `json:"reference,omitempty"`
	Shasum     string `json:"shasum,omitempty"`
	URLPresent bool   `json:"url_present"`
}

type Link struct {
	Owner         string `json:"owner"`
	Scope         string `json:"scope"`
	Relation      string `json:"relation"`
	Name          string `json:"name"`
	Constraint    string `json:"constraint"`
	Kind          string `json:"kind"`
	Compatibility string `json:"compatibility"`
}

type Override struct {
	Source string          `json:"source"`
	Name   string          `json:"name"`
	Value  json.RawMessage `json:"value"`
}
