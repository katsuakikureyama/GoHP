package compat

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
)

// LoadArchive reads a saved run without tests, PHP, Git or source checkout
// access. It checks byte bindings and replays the recorded event evaluation.
// This proves internal consistency, NOT authenticity or current applicability:
// fixture hashes and environment metadata remain assertions of the producer.
func LoadArchive(dir string) (Report, error) {
	root, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return Report{}, err
	}
	var r Report
	raw, err := archiveFile(root, "run.json", 16<<20)
	if err != nil {
		return r, err
	}
	if err := strictJSON(raw, &r); err != nil {
		return Report{}, fmt.Errorf("decode run.json: %w", err)
	}
	if err := validateBaseline(r.Schema, r.Metadata); err != nil {
		return Report{}, err
	}
	if !validSHA256(r.Metadata.SourceSHA256) {
		return Report{}, fmt.Errorf("missing or invalid recorded source fingerprint")
	}
	catalogBytes, err := archiveFile(root, "catalog.json", 4<<20)
	if err != nil {
		return Report{}, err
	}
	events, err := archiveFile(root, "go-test.jsonl", 128<<20)
	if err != nil {
		return Report{}, err
	}
	if hashBytes(catalogBytes) != r.CatalogSHA256 || hashBytes(events) != r.EventSHA256 {
		return Report{}, fmt.Errorf("catalog/event hash mismatch; archive is incomplete or changed")
	}
	var c Catalog
	if err := strictJSON(catalogBytes, &c); err != nil {
		return Report{}, fmt.Errorf("decode catalog snapshot: %w", err)
	}
	if err := c.validateCatalog(false); err != nil {
		return Report{}, fmt.Errorf("invalid catalog snapshot: %w", err)
	}
	c.SHA256 = r.CatalogSHA256
	byID := make(map[string]CaseEvidence)
	for _, evidence := range r.Cases {
		if _, exists := byID[evidence.ID]; exists {
			return Report{}, fmt.Errorf("duplicate evidence case %q", evidence.ID)
		}
		byID[evidence.ID] = evidence
		if evidence.Fixture != "" && !validSHA256(evidence.FixtureSHA256) {
			// Preserve the only two source-read failures the producer can emit.
			// Such cases cannot acquire a pass through historical reconstruction.
			if evidence.FixtureSHA256 != "" || evidence.Status != "incomplete" || !oneOf(evidence.Detail,
				"fixture no longer resolves beneath repository root", "fixture is no longer readable") {
				return Report{}, fmt.Errorf("case %q has invalid fixture evidence", evidence.ID)
			}
		}
	}
	replayed, err := evaluate(c, bytes.NewReader(events), r.Metadata, func(test Case) (string, string) {
		evidence := byID[test.ID]
		if evidence.FixtureSHA256 == "" {
			return "", evidence.Detail
		}
		return evidence.FixtureSHA256, ""
	}, r.Schema)
	if err != nil {
		return Report{}, fmt.Errorf("replay saved evidence: %w", err)
	}
	if !reflect.DeepEqual(r, replayed) {
		return Report{}, fmt.Errorf("saved verdicts do not match catalog and event evidence")
	}
	return r, nil
}

// Fixed filenames only; never follow an artifact link outside its directory or
// read devices/FIFOs. Size bounds limit accidental giant archive reads. As with
// the producer, use a private directory without concurrent hostile mutation.
func archiveFile(root, name string, limit int64) ([]byte, error) {
	path, err := (Catalog{root: root}).checkedPath(name)
	if err != nil {
		return nil, fmt.Errorf("archive %s: %w", name, err)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit {
		return nil, fmt.Errorf("archive %s exceeds %d-byte limit", name, limit)
	}
	return b, nil
}
