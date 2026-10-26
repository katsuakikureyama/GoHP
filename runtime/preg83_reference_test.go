//go:build gohp_pcre2 && cgo

package runtime_test

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/runtime"
)

// TestPregMatch83RetainedReference exercises the value adapter against the
// retained PHP 8.3 observation. The GPL WordPress pattern and all expected
// values stay in the separately licensed Gossip checkout; this repository
// keeps only the evidence identity and the adapter contract.
func TestPregMatch83RetainedReference(t *testing.T) {
	dir := os.Getenv("GOHP_PCRE2_REFERENCE_DIR")
	if dir == "" {
		t.Skip("external Gossip evidence not supplied; no PHP 8.3 24-case comparison claim")
	}

	const evidenceName = "parser-token-reference-2026-10-27.json"
	const evidenceHash = "5cbf560247afcd2c5f1462cab2f95f16bb822ac2d4d1756be16221cffdc6dc60"
	bytes, err := os.ReadFile(filepath.Join(dir, evidenceName))
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(bytes)); got != evidenceHash {
		t.Fatalf("%s: evidence hash drift: got %s", evidenceName, got)
	}

	var evidence php83Evidence
	decoder := json.NewDecoder(strings.NewReader(string(bytes)))
	decoder.UseNumber()
	if err := decoder.Decode(&evidence); err != nil {
		t.Fatal(err)
	}
	if evidence.NativeExitCode != 0 {
		t.Fatalf("retained native observation did not succeed: exit=%d", evidence.NativeExitCode)
	}
	if len(evidence.Native.Cases) != 24 {
		t.Fatalf("retained case count drift: got %d, want 24", len(evidence.Native.Cases))
	}
	pattern, err := base64.StdEncoding.Strict().DecodeString(evidence.Native.Pattern)
	if err != nil || len(pattern) < 3 || pattern[0] != '/' || !strings.HasSuffix(string(pattern), "/s") {
		t.Fatalf("invalid retained slash-delimited /s pattern: %v", err)
	}

	for _, c := range evidence.Native.Cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			subject, err := base64.StdEncoding.Strict().DecodeString(c.Input.Document)
			if err != nil {
				t.Fatal(err)
			}
			offset, err := int64JSON(c.Input.Offset)
			if err != nil || offset < -int64(len(subject)) || offset > int64(len(subject)) {
				t.Fatalf("invalid retained offset %q: %v", c.Input.Offset, err)
			}
			matchLimit, err := uint32JSON(c.Profile.BacktrackLimit)
			if err != nil || (matchLimit != 100000 && matchLimit != 1) {
				t.Fatalf("invalid retained match limit %q: %v", c.Profile.BacktrackLimit, err)
			}
			depthLimit, err := uint32JSON(c.Profile.RecursionLimit)
			if err != nil || depthLimit != 10000 || c.Profile.JIT != "0" {
				t.Fatalf("invalid retained PCRE profile: depth=%q jit=%q: %v", c.Profile.RecursionLimit, c.Profile.JIT, err)
			}
			wantReturn, err := recordedValue(c.PregMatch.Return)
			if err != nil {
				t.Fatal(err)
			}
			wantMatches, err := recordedValue(c.PregMatch.MatchesAfter)
			if err != nil {
				t.Fatal(err)
			}
			if wantMatches.kind != "array" {
				t.Fatalf("retained matches_after is %s, want array", wantMatches.kind)
			}

			execution := runtime.Execution{}
			got, err := execution.PregMatch83(string(pattern), string(subject), offset, runtime.Preg83Limits{
				Match:   matchLimit,
				Depth:   depthLimit,
				HeapKiB: 131072,
			})
			if err != nil {
				t.Fatal(err)
			}
			gotReturn := preg83TypedValue(t, got.Return)
			gotMatches := preg83TypedValue(t, got.Matches)
			if !reflect.DeepEqual(gotReturn, wantReturn) {
				t.Fatalf("return differs: got %#v, want %#v", gotReturn, wantReturn)
			}
			if !reflect.DeepEqual(gotMatches, wantMatches) {
				t.Fatalf("matches_after differs: got %#v, want %#v", gotMatches, wantMatches)
			}
			if gotError := execution.PregLastError83(); gotError != c.PregMatch.ErrorCode {
				t.Fatalf("last error differs: got %d, want %d", gotError, c.PregMatch.ErrorCode)
			}
			assertRawEngineOutcome(t, wantReturn, c.PregMatch.ErrorCode, got.EngineCode)
		})
	}
}

type php83Evidence struct {
	NativeExitCode int `json:"native_exit_code"`
	Native         struct {
		Pattern string      `json:"pattern_base64"`
		Cases   []php83Case `json:"cases"`
	} `json:"native_observation"`
}

type php83Case struct {
	Name  string `json:"name"`
	Input struct {
		Document string          `json:"document_base64"`
		Offset   json.RawMessage `json:"offset"`
	} `json:"input"`
	PregMatch struct {
		ErrorCode    int64           `json:"error_code"`
		MatchesAfter json.RawMessage `json:"matches_after"`
		Return       json.RawMessage `json:"return"`
	} `json:"preg_match"`
	Profile struct {
		BacktrackLimit json.RawMessage `json:"pcre.backtrack_limit"`
		JIT            string          `json:"pcre.jit"`
		RecursionLimit json.RawMessage `json:"pcre.recursion_limit"`
	} `json:"profile"`
}

type typedValue struct {
	kind    string
	string_ string
	int_    int64
	bool_   bool
	entries []typedEntry
}

type typedEntry struct {
	key   typedValue
	value typedValue
}

func recordedValue(raw json.RawMessage) (typedValue, error) {
	var wire struct {
		Type    string          `json:"type"`
		Base64  string          `json:"base64"`
		Value   json.RawMessage `json:"value"`
		Entries json.RawMessage `json:"entries"`
	}
	if err := decodeJSON(raw, &wire); err != nil {
		return typedValue{}, err
	}
	switch wire.Type {
	case "string":
		b, err := base64.StdEncoding.Strict().DecodeString(wire.Base64)
		if err != nil {
			return typedValue{}, fmt.Errorf("invalid base64 string: %w", err)
		}
		return typedValue{kind: "string", string_: string(b)}, nil
	case "int":
		v, err := int64JSON(wire.Value)
		if err != nil {
			return typedValue{}, err
		}
		return typedValue{kind: "int", int_: v}, nil
	case "bool":
		var v bool
		if err := decodeJSON(wire.Value, &v); err != nil {
			return typedValue{}, err
		}
		return typedValue{kind: "bool", bool_: v}, nil
	case "null":
		if string(wire.Value) != "null" {
			return typedValue{}, fmt.Errorf("invalid null value %s", wire.Value)
		}
		return typedValue{kind: "null"}, nil
	case "array":
		if len(wire.Entries) == 0 || string(wire.Entries) == "null" {
			return typedValue{}, fmt.Errorf("array entries missing or null")
		}
		var entries []struct {
			Key   json.RawMessage `json:"key"`
			Value json.RawMessage `json:"value"`
		}
		if err := decodeJSON(wire.Entries, &entries); err != nil {
			return typedValue{}, err
		}
		out := typedValue{kind: "array", entries: make([]typedEntry, len(entries))}
		for i, entry := range entries {
			key, err := recordedValue(entry.Key)
			if err != nil {
				return typedValue{}, fmt.Errorf("entry %d key: %w", i, err)
			}
			if key.kind != "int" && key.kind != "string" {
				return typedValue{}, fmt.Errorf("entry %d has unsupported key kind %s", i, key.kind)
			}
			value, err := recordedValue(entry.Value)
			if err != nil {
				return typedValue{}, fmt.Errorf("entry %d value: %w", i, err)
			}
			out.entries[i] = typedEntry{key: key, value: value}
		}
		return out, nil
	default:
		return typedValue{}, fmt.Errorf("unknown retained value type %q", wire.Type)
	}
}

func runtimeValue(value any) (typedValue, error) {
	switch value := value.(type) {
	case nil:
		return typedValue{kind: "null"}, nil
	case string:
		return typedValue{kind: "string", string_: value}, nil
	case int64:
		return typedValue{kind: "int", int_: value}, nil
	case bool:
		return typedValue{kind: "bool", bool_: value}, nil
	case runtime.Array:
		out := typedValue{kind: "array", entries: make([]typedEntry, len(value.Entries))}
		for i, entry := range value.Entries {
			key, err := runtimeValue(entry.Key)
			if err != nil {
				return typedValue{}, fmt.Errorf("entry %d key: %w", i, err)
			}
			if key.kind != "int" && key.kind != "string" {
				return typedValue{}, fmt.Errorf("entry %d has unsupported key kind %s", i, key.kind)
			}
			item, err := runtimeValue(entry.Value)
			if err != nil {
				return typedValue{}, fmt.Errorf("entry %d value: %w", i, err)
			}
			out.entries[i] = typedEntry{key: key, value: item}
		}
		return out, nil
	default:
		return typedValue{}, fmt.Errorf("unexpected runtime value type %T", value)
	}
}

// preg83TypedValue gives all external runtime tests one exact representation
// for PHP-shaped values without converting binary strings through JSON. It
// deliberately distinguishes nil (null) from an empty runtime.Array.
func preg83TypedValue(t *testing.T, value any) any {
	t.Helper()
	encoded, err := runtimeValue(value)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func assertRawEngineOutcome(t *testing.T, returned typedValue, errorCode int64, engineCode int) {
	t.Helper()
	switch {
	case returned.kind == "int" && returned.int_ == 1 && errorCode == 0:
		if engineCode <= 0 {
			t.Fatalf("matched PHP result requires positive raw engine code, got %d", engineCode)
		}
	case returned.kind == "int" && returned.int_ == 0 && errorCode == 0:
		if engineCode != -1 {
			t.Fatalf("no-match PHP result requires PCRE2_ERROR_NOMATCH (-1), got %d", engineCode)
		}
	case returned.kind == "bool" && !returned.bool_ && errorCode == 2:
		if engineCode != -47 {
			t.Fatalf("backtrack-limit PHP result requires PCRE2_ERROR_MATCHLIMIT (-47), got %d", engineCode)
		}
	default:
		t.Fatalf("unsupported retained PHP/raw-engine outcome: return=%#v error=%d", returned, errorCode)
	}
}

func decodeJSON(raw []byte, value any) error {
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.UseNumber()
	return decoder.Decode(value)
}

func int64JSON(raw json.RawMessage) (int64, error) {
	var number json.Number
	if err := decodeJSON(raw, &number); err == nil {
		return number.Int64()
	}
	var quoted string
	if err := decodeJSON(raw, &quoted); err != nil {
		return 0, fmt.Errorf("invalid JSON integer %s", raw)
	}
	return strconv.ParseInt(quoted, 10, 64)
}

func uint32JSON(raw json.RawMessage) (uint32, error) {
	i, err := int64JSON(raw)
	if err != nil || i < 0 || i > int64(^uint32(0)) {
		return 0, fmt.Errorf("invalid uint32 %s", raw)
	}
	return uint32(i), nil
}
