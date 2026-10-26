//go:build gohp_pcre2 && cgo

package pcre2_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/pcre2"
)

// This is an opt-in consumer evidence replay, not a PHP differential test.
// The GPL WordPress pattern stays in the separately licensed Gossip checkout.
func TestRetainedConsumerReference(t *testing.T) {
	dir := os.Getenv("GOHP_PCRE2_REFERENCE_DIR")
	if dir == "" {
		t.Skip("external consumer evidence not supplied; no 24-case comparison claim")
	}
	const nativeHash = "5cbf560247afcd2c5f1462cab2f95f16bb822ac2d4d1756be16221cffdc6dc60"
	read := func(name, hash string, out any) {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if got := fmt.Sprintf("%x", sha256.Sum256(b)); got != hash {
			t.Fatalf("%s: evidence hash drift: %s", name, got)
		}
		if err := json.Unmarshal(b, out); err != nil {
			t.Fatal(err)
		}
	}
	var native struct {
		Native struct {
			Pattern string `json:"pattern_base64"`
		} `json:"native_observation"`
	}
	read("parser-token-reference-2026-10-27.json", nativeHash, &native)
	var reference struct {
		ReferenceSHA string `json:"reference_sha256"`
		PatternSHA   string `json:"pattern_body_sha256"`
		Cases        []struct {
			Name   string `json:"name"`
			Input  string `json:"input_base64"`
			Offset int    `json:"normalized_offset"`
			Raw    struct {
				Version      string       `json:"version"`
				MatchLimit   uint32       `json:"match_limit"`
				DepthLimit   uint32       `json:"depth_limit"`
				HeapLimit    uint32       `json:"heap_limit_kib"`
				ReturnCode   int          `json:"return_code"`
				CaptureCount int          `json:"capture_count"`
				Names        []pcre2.Name `json:"names"`
				Captures     []struct {
					Number int    `json:"number"`
					Start  int    `json:"start"`
					End    int    `json:"end"`
					Hex    string `json:"hex"`
				} `json:"captures"`
			} `json:"raw"`
		} `json:"cases"`
	}
	read("parser-pcre2-comparison-2026-10-27.json", "5390d6c560ca7b0be9ab6363ddc65f12e0908ba4b68cffaa2d0998c1130d021d", &reference)
	pattern, err := base64.StdEncoding.Strict().DecodeString(native.Native.Pattern)
	if err != nil || len(pattern) < 3 || pattern[0] != '/' || string(pattern[len(pattern)-2:]) != "/s" {
		t.Fatal("invalid retained slash-delimited /s pattern")
	}
	// No syntax rewriting: remove only the PHP delimiters and modifier.
	pattern = pattern[1 : len(pattern)-2]
	if reference.ReferenceSHA != nativeHash || len(reference.Cases) != 24 || fmt.Sprintf("%x", sha256.Sum256(pattern)) != reference.PatternSHA {
		t.Fatal("reference identity mismatch")
	}
	for _, c := range reference.Cases {
		t.Run(c.Name, func(t *testing.T) {
			subject, err := base64.StdEncoding.Strict().DecodeString(c.Input)
			if err != nil {
				t.Fatal(err)
			}
			r, err := pcre2.Match(pcre2.Request{Pattern: pattern, Subject: subject, Offset: c.Offset, Profile: pcre2.ByteDotAll,
				Limits: pcre2.Limits{Match: c.Raw.MatchLimit, Depth: c.Raw.DepthLimit, HeapKiB: c.Raw.HeapLimit}})
			if err != nil {
				t.Fatal(err)
			}
			if r.Version != c.Raw.Version || r.ReturnCode != c.Raw.ReturnCode || r.CaptureCount != c.Raw.CaptureCount || !reflect.DeepEqual(r.Names, c.Raw.Names) || len(r.Captures) != len(c.Raw.Captures) {
				t.Fatalf("raw metadata differs: got %+v, expected %+v", r, c.Raw)
			}
			for i, want := range c.Raw.Captures {
				b, err := hex.DecodeString(want.Hex)
				if err != nil {
					t.Fatal(err)
				}
				got := r.Captures[i]
				if got.Number != want.Number || got.Present != (want.Start >= 0) || got.Start != want.Start || got.End != want.End || !bytes.Equal(got.Bytes, b) {
					t.Fatalf("capture %d differs: got %+v, expected %+v", i, got, want)
				}
			}
		})
	}
}
