package inventory

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const maxInputBytes = 16 << 20

// Read inventories two fixed metadata files in a private, non-concurrently-
// mutated directory. It does not follow repository, autoload or script paths,
// inspect vendor, execute programs or resolve version constraints.
func Read(root string) (Report, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return Report{}, err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return Report{}, err
	}
	r := Report{Schema: "gohp.composer.inventory/v0.1", Status: "metadata-only", Compatibility: "unverified", LockFreshness: "not-available", Packages: []Package{}, Links: []Link{}, PlatformOverrides: []Override{}, Notices: []string{
		"Metadata only: no PHP/Composer execution, downloads, dependency solving or compatibility tests.",
		"Locked versions and references are recorded declarations, not installed packages or verified source revisions. Lock freshness is not checked.",
		"All dependency compatibility is unverified; no catalog entry or passing builtin case certifies a package or extension.",
		"require-dev of a locked package is a development declaration of that package, not an active root dependency. suggest/provide/replace/conflict are not require.",
		"Platform overrides describe configured values, not the actual PHP runtime. Script events and autoload modes are names only; commands, mappings and URLs are omitted.",
		"PHP callsites, dynamic calls, vendor/generated files, installation paths, repositories, aliases, plugin policy and other metadata fields are not analyzed. This is not full Composer schema validation or a license audit.",
	}}
	manifest, input, err := readInput(root, "composer.json", false)
	if err != nil {
		return Report{}, err
	}
	r.Inputs = append(r.Inputs, input)
	r.Root, err = r.readPackage(manifest, "root", false)
	if err != nil {
		return Report{}, fmt.Errorf("composer.json: %w", err)
	}
	if config, ok := manifest["config"]; ok {
		obj, ok := config.(map[string]any)
		if !ok {
			return Report{}, fmt.Errorf("composer.json config must be an object")
		}
		if err := r.overrides(obj, "platform", "composer.json/config.platform"); err != nil {
			return Report{}, err
		}
	}
	lock, input, err := readInput(root, "composer.lock", true)
	if err != nil {
		return Report{}, err
	}
	r.Inputs = append(r.Inputs, input)
	if input.Status == "absent" {
		r.Notices = append(r.Notices, "composer.lock is absent; locked dependency versions and transitive requirements are unavailable.")
		return r, nil
	}
	r.LockContentHash, err = stringField(lock, "content-hash", true)
	if err != nil {
		return Report{}, fmt.Errorf("composer.lock: %w", err)
	}
	r.LockFreshness = "unverified"
	seen := map[string]bool{}
	for _, group := range []struct{ key, scope string }{{"packages", "production"}, {"packages-dev", "development"}} {
		list, ok := lock[group.key].([]any)
		if !ok {
			return Report{}, fmt.Errorf("composer.lock %s must be an array", group.key)
		}
		for i, value := range list {
			obj, ok := value.(map[string]any)
			if !ok {
				return Report{}, fmt.Errorf("composer.lock %s[%d] must be an object", group.key, i)
			}
			pkg, err := r.readPackage(obj, group.scope, true)
			if err != nil {
				return Report{}, fmt.Errorf("composer.lock %s[%d]: %w", group.key, i, err)
			}
			if seen[pkg.Name] {
				return Report{}, fmt.Errorf("composer.lock contains duplicate package names")
			}
			seen[pkg.Name] = true
			r.Packages = append(r.Packages, pkg)
		}
	}
	for _, key := range []string{"platform", "platform-dev"} {
		if err := r.links(lock, key, "composer.lock", "lock-root"); err != nil {
			return Report{}, err
		}
	}
	if err := r.overrides(lock, "platform-overrides", "composer.lock/platform-overrides"); err != nil {
		return Report{}, err
	}
	sort.Slice(r.Packages, func(i, j int) bool {
		if r.Packages[i].Scope != r.Packages[j].Scope {
			return r.Packages[i].Scope < r.Packages[j].Scope
		}
		return r.Packages[i].Name < r.Packages[j].Name
	})
	sort.SliceStable(r.Links, func(i, j int) bool {
		a, b := r.Links[i], r.Links[j]
		return a.Scope+"\x00"+a.Owner+"\x00"+a.Relation+"\x00"+a.Name < b.Scope+"\x00"+b.Owner+"\x00"+b.Relation+"\x00"+b.Name
	})
	return r, nil
}

func (r *Report) readPackage(obj map[string]any, scope string, locked bool) (Package, error) {
	p := Package{Scope: scope, ScriptEvents: []string{}, AutoloadModes: []string{}, AutoloadDevModes: []string{}}
	var err error
	if p.Name, err = stringField(obj, "name", locked); err != nil {
		return p, err
	}
	if p.Version, err = stringField(obj, "version", locked); err != nil {
		return p, err
	}
	if p.Type, err = stringField(obj, "type", false); err != nil {
		return p, err
	}
	if value, ok := obj["license"]; ok {
		if p.Licenses, err = stringList(value); err != nil {
			return p, fmt.Errorf("license: %w", err)
		}
	}
	if p.Source, err = reference(obj, "source"); err != nil {
		return p, err
	}
	if p.Dist, err = reference(obj, "dist"); err != nil {
		return p, err
	}
	for _, field := range []struct {
		key string
		dst *[]string
	}{{"scripts", &p.ScriptEvents}, {"autoload", &p.AutoloadModes}, {"autoload-dev", &p.AutoloadDevModes}} {
		value, exists := obj[field.key]
		if !exists {
			continue
		}
		m, ok := value.(map[string]any)
		if !ok {
			return p, fmt.Errorf("%s must be an object", field.key)
		}
		*field.dst = keys(m)
	}
	owner := p.Name
	if scope == "root" {
		owner = "composer.json"
	}
	for _, relation := range []string{"require", "require-dev", "conflict", "provide", "replace", "suggest"} {
		if err := r.links(obj, relation, owner, scope); err != nil {
			return p, err
		}
	}
	return p, nil
}

func (r *Report) links(obj map[string]any, relation, owner, scope string) error {
	value, exists := obj[relation]
	if !exists {
		return nil
	}
	m, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("%s must be an object of strings", relation)
	}
	for _, name := range keys(m) {
		constraint, ok := m[name].(string)
		if !ok || strings.TrimSpace(name) == "" {
			return fmt.Errorf("%s must have nonempty keys and string values", relation)
		}
		r.Links = append(r.Links, Link{Owner: owner, Scope: scope, Relation: relation, Name: name, Constraint: constraint, Kind: dependencyKind(name), Compatibility: "unverified"})
	}
	return nil
}

func (r *Report) overrides(obj map[string]any, key, source string) error {
	value, exists := obj[key]
	if !exists {
		return nil
	}
	m, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("%s must be an object", key)
	}
	for _, name := range keys(m) {
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("%s contains an empty key", key)
		}
		switch v := m[name].(type) {
		case string:
		case bool:
			if v {
				return fmt.Errorf("%s values must be strings or false", key)
			}
		default:
			return fmt.Errorf("%s values must be strings or false", key)
		}
		b, err := json.Marshal(m[name])
		if err != nil {
			return err
		}
		r.PlatformOverrides = append(r.PlatformOverrides, Override{Source: source, Name: name, Value: b})
	}
	return nil
}

func dependencyKind(name string) string {
	switch {
	case name == "php" || strings.HasPrefix(name, "php-"):
		return "php"
	case strings.HasPrefix(name, "ext-"):
		return "extension"
	case strings.HasPrefix(name, "lib-"):
		return "library"
	case name == "composer" || name == "composer-plugin-api" || name == "composer-runtime-api":
		return "composer"
	case strings.Contains(name, "/"):
		return "package"
	default:
		return "unknown"
	}
}

func stringField(obj map[string]any, key string, required bool) (string, error) {
	v, exists := obj[key]
	if !exists && !required {
		return "", nil
	}
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("%s must be a string", key)
	}
	if required && strings.TrimSpace(s) == "" {
		return "", fmt.Errorf("%s must be nonempty", key)
	}
	return s, nil
}

func stringList(value any) ([]string, error) {
	if s, ok := value.(string); ok {
		return []string{s}, nil
	}
	list, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("must be a string or array of strings")
	}
	result := []string{}
	for _, v := range list {
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("must contain only strings")
		}
		result = append(result, s)
	}
	return result, nil
}

func reference(obj map[string]any, key string) (*Reference, error) {
	v, exists := obj[key]
	if !exists {
		return nil, nil
	}
	m, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s must be an object", key)
	}
	r := &Reference{}
	for _, field := range []struct {
		key string
		dst *string
	}{{"type", &r.Type}, {"reference", &r.Reference}, {"shasum", &r.Shasum}} {
		// Composer dist metadata can record a null reference/shasum.
		if field.key != "type" && m[field.key] == nil {
			continue
		}
		var err error
		*field.dst, err = stringField(m, field.key, false)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", key, err)
		}
	}
	url, err := stringField(m, "url", false)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", key, err)
	}
	r.URLPresent = url != ""
	return r, nil
}

func keys(m map[string]any) []string {
	result := make([]string, 0, len(m))
	for key := range m {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}

func readInput(root, name string, optional bool) (map[string]any, Input, error) {
	input := Input{Name: name, Status: "present"}
	path := filepath.Join(root, name)
	if _, err := os.Lstat(path); err != nil {
		if os.IsNotExist(err) && optional {
			input.Status = "absent"
			return nil, input, nil
		}
		return nil, input, err
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return nil, input, err
	}
	rel, err := filepath.Rel(root, resolved)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return nil, input, fmt.Errorf("%s escapes the project directory", name)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return nil, input, err
	}
	if !info.Mode().IsRegular() {
		return nil, input, fmt.Errorf("%s must be a regular file", name)
	}
	f, err := os.Open(resolved)
	if err != nil {
		return nil, input, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, maxInputBytes+1))
	if err != nil {
		return nil, input, err
	}
	if len(b) > maxInputBytes {
		return nil, input, fmt.Errorf("%s exceeds the 16 MiB input limit", name)
	}
	sum := sha256.Sum256(b)
	input.SHA256 = hex.EncodeToString(sum[:])
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	value, err := jsonValue(d, 0)
	if err != nil {
		return nil, input, fmt.Errorf("%s: %w", name, err)
	}
	if _, err := d.Token(); err != io.EOF {
		return nil, input, fmt.Errorf("%s contains trailing JSON data", name)
	}
	m, ok := value.(map[string]any)
	if !ok {
		return nil, input, fmt.Errorf("%s must contain a JSON object", name)
	}
	return m, input, nil
}

// Reject duplicate keys at every depth; last-key-wins can hide requirements.
func jsonValue(d *json.Decoder, depth int) (any, error) {
	if depth > 64 {
		return nil, fmt.Errorf("JSON exceeds 64 nesting levels")
	}
	t, err := d.Token()
	if err != nil {
		return nil, err
	}
	switch t {
	case json.Delim('{'):
		m := map[string]any{}
		for d.More() {
			key, err := d.Token()
			if err != nil {
				return nil, err
			}
			s, ok := key.(string)
			if !ok {
				return nil, fmt.Errorf("invalid JSON object key")
			}
			if _, exists := m[s]; exists {
				return nil, fmt.Errorf("duplicate JSON object key")
			}
			m[s], err = jsonValue(d, depth+1)
			if err != nil {
				return nil, err
			}
		}
		_, err := d.Token()
		return m, err
	case json.Delim('['):
		list := []any{}
		for d.More() {
			v, err := jsonValue(d, depth+1)
			if err != nil {
				return nil, err
			}
			list = append(list, v)
		}
		_, err := d.Token()
		return list, err
	default:
		return t, nil
	}
}
