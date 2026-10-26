package compat

import (
	"encoding/json"
	"fmt"
)

const (
	RunSchemaLegacy   = "gohp.compatibility.run/v0.1"
	RunSchemaBaseline = "gohp.compatibility.run/v0.2"
	RunSchemaQuiet    = "gohp.compatibility.run/v0.3"
	// RunSchemaOutputs introduced recorded stdout and process-exit observations.
	// It remains readable after the current schema advances.
	RunSchemaOutputs = "gohp.compatibility.run/v0.4"
	RunSchema        = "gohp.compatibility.run/v0.5"
)

// PHPBaseline is an allowlisted observation from one separate php -r process
// before the suite. It is NOT the effective configuration of every fixture.
// Nil INI/extension versions mean unavailable; empty strings remain distinct.
type PHPBaseline struct {
	Scope          string             `json:"scope"`
	PHPVersion     string             `json:"php_version"`
	SAPI           string             `json:"sapi"`
	IntSize        int                `json:"int_size"`
	Timezone       string             `json:"timezone"`
	INI            map[string]*string `json:"ini"`
	Locales        map[string]*string `json:"locales"`
	Extensions     []PHPExtension     `json:"extensions"`
	ZendExtensions []PHPExtension     `json:"zend_extensions"`
}

type PHPExtension struct {
	Name    string  `json:"name"`
	Version *string `json:"version"`
}

// The keys are the complete v0.2 projection, not a dump of php.ini or env.
func PHPINIKeys() []string {
	return []string{"precision", "serialize_precision", "default_charset", "date.timezone", "error_reporting", "display_errors", "display_startup_errors", "log_errors", "zend.assertions", "assert.exception"}
}

func PHPLocaleKeys() []string {
	return []string{"LC_CTYPE", "LC_NUMERIC", "LC_TIME", "LC_COLLATE", "LC_MONETARY", "LC_MESSAGES"}
}

// PHPBaselineHash identifies the canonical JSON projection, not the entire
// environment or configuration files. It is not a signature or attestation.
func PHPBaselineHash(b *PHPBaseline) string {
	encoded, _ := json.Marshal(b) // This type contains only JSON-encodable fields.
	return hashBytes(encoded)
}

func (b *PHPBaseline) Validate(version string) error {
	if b == nil || b.Scope != "cli-startup-baseline" || b.PHPVersion == "" || b.PHPVersion != version || b.SAPI != "cli" || (b.IntSize != 4 && b.IntSize != 8) || blank(b.Timezone) {
		return fmt.Errorf("missing or invalid PHP startup baseline identity")
	}
	for _, projection := range []struct {
		values map[string]*string
		keys   []string
	}{{b.INI, PHPINIKeys()}, {b.Locales, PHPLocaleKeys()}} {
		if len(projection.values) != len(projection.keys) {
			return fmt.Errorf("PHP baseline projection has missing or unknown keys")
		}
		for _, key := range projection.keys {
			if _, exists := projection.values[key]; !exists {
				return fmt.Errorf("PHP baseline projection requires %s", key)
			}
		}
	}
	for _, locale := range b.Locales {
		if locale != nil && blank(*locale) {
			return fmt.Errorf("PHP baseline has empty locale")
		}
	}
	for _, extensions := range [][]PHPExtension{b.Extensions, b.ZendExtensions} {
		if extensions == nil {
			return fmt.Errorf("PHP baseline requires extension arrays (empty is allowed)")
		}
		for i, extension := range extensions {
			if blank(extension.Name) || (i > 0 && extensions[i-1].Name >= extension.Name) {
				return fmt.Errorf("PHP baseline extensions must have unique, sorted names")
			}
		}
	}
	return nil
}

func validateBaseline(schema string, meta Metadata) error {
	switch schema {
	case RunSchemaLegacy:
		if meta.PHPBaseline != nil || meta.PHPBaselineSHA256 != "" {
			return fmt.Errorf("legacy run cannot carry v0.2 PHP baseline evidence")
		}
	case RunSchemaBaseline, RunSchemaQuiet, RunSchemaOutputs, RunSchema:
		if err := meta.PHPBaseline.Validate(meta.PHPVersion); err != nil {
			return err
		}
		if meta.PHPBaselineSHA256 != PHPBaselineHash(meta.PHPBaseline) {
			return fmt.Errorf("PHP baseline hash mismatch")
		}
	default:
		return fmt.Errorf("unsupported run schema %q", schema)
	}
	return nil
}
