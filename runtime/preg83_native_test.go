//go:build gohp_pcre2 && cgo

package runtime_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/katsuakikureyama/GoHP/runtime"
)

// No PHP program is transpiled here. This measures the B2 Go value API
// against new native PHP state traces; compiler propagation remains B3 work.
func TestPreg83NativeExecutionState(t *testing.T) {
	image := os.Getenv("GOHP_PREG83_PHP_IMAGE")
	if image == "" {
		t.Skip("pinned native oracle not supplied; no new state-comparison claim")
	}
	const pinned = "php:8.3.36-cli@sha256:98d555ff97b3865140ad31032508c3ca089fa97a8944b14ca9dc03ea05d788cd"
	if image != pinned {
		t.Fatal("requires exact pinned PHP8.3.36 oracle")
	}
	fixture, err := filepath.Abs("../tests/oracle/preg83_state.php")
	if err != nil {
		t.Fatal(err)
	}
	fixtureBytes, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("native fixture sha256=%x; image=%s", sha256.Sum256(fixtureBytes), image)
	type step struct {
		Label   string          `json:"label"`
		Before  int64           `json:"before"`
		After   int64           `json:"after"`
		Return  json.RawMessage `json:"return"`
		Matches json.RawMessage `json:"matches"`
	}
	type observation struct {
		Schema   string `json:"schema"`
		PHP      string `json:"php_version"`
		PCRE     string `json:"pcre_version"`
		JIT      string `json:"jit"`
		Depth    string `json:"depth"`
		Locale   string `json:"locale"`
		Scenario string `json:"scenario"`
		Initial  int64  `json:"initial"`
		Final    int64  `json:"final"`
		Steps    []step `json:"steps"`
	}
	oracles := make(map[string]observation)
	for _, scenario := range []string{"a", "b"} {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		cmd := exec.CommandContext(ctx, "docker", "run", "--rm", "--pull=never", "--network=none", "--read-only", "--cap-drop=ALL", "--security-opt=no-new-privileges", "--memory=128m", "--cpus=1", "--pids-limit=64", "--user=65534:65534", "--mount", "type=bind,src="+fixture+",dst=/input.php,readonly", image, "php", "-n", "-d", "memory_limit=128M", "/input.php", scenario)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		out, err := cmd.Output()
		cancel()
		if err != nil || stderr.Len() != 0 {
			t.Fatalf("native %s: error=%v stderr=%s", scenario, err, stderr.String())
		}
		var o observation
		if err := json.Unmarshal(out, &o); err != nil {
			t.Fatal(err)
		}
		if o.Schema != "gohp.preg83-state-reference/v1" || o.PHP != "8.3.36" || !strings.HasPrefix(o.PCRE, "10.42 ") || o.JIT != "0" || o.Depth != "10000" || o.Locale != "C" || o.Scenario != scenario || o.Initial != 0 {
			t.Fatalf("native profile drift: %+v", o)
		}
		oracles[scenario] = o
		t.Logf("native-%s sha256=%x json=%s", scenario, sha256.Sum256(out), bytes.TrimSpace(out))
	}
	a, b := &runtime.Execution{}, &runtime.Execution{}
	if a.PregLastError83() != oracles["a"].Initial || b.PregLastError83() != oracles["b"].Initial {
		t.Fatal("initial state differs")
	}
	if len(oracles["a"].Steps) != 5 || len(oracles["b"].Steps) != 2 {
		t.Fatal("native step count differs")
	}
	run := func(which string, i int, context *runtime.Execution, label, pattern, subject string, limit uint32, nested bool) {
		t.Helper()
		want := oracles[which].Steps[i]
		if want.Label != label || context.PregLastError83() != want.Before {
			t.Fatalf("%s before-state differs", label)
		}
		limits := runtime.Preg83Limits{Match: limit, Depth: 10000, HeapKiB: 131072}
		var got runtime.Preg83Result
		var err error
		if nested {
			got, err = preg83NestedNativeCall(context, pattern, subject, limits)
		} else {
			got, err = context.PregMatch83(pattern, subject, 0, limits)
		}
		if err != nil {
			t.Fatal(err)
		}
		returned, err := recordedValue(want.Return)
		if err != nil {
			t.Fatal(err)
		}
		matches, err := recordedValue(want.Matches)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(preg83TypedValue(t, got.Return), returned) || !reflect.DeepEqual(preg83TypedValue(t, got.Matches), matches) || context.PregLastError83() != want.After {
			t.Fatalf("%s native/value/state differs: got=%+v error=%d want=%+v", label, got, context.PregLastError83(), want)
		}
		assertRawEngineOutcome(t, returned, want.After, got.EngineCode)
	}
	long := strings.Repeat("a", 100) + "!"
	run("a", 0, a, "failure", "/^(a+)+$/s", long, 1, false)
	run("b", 0, b, "empty_capture", "/()/s", "", 100000, false)
	if a.PregLastError83() != 2 {
		t.Fatal("execution B changed A failure state")
	}
	run("a", 1, a, "nested_success", "/(?<letter>a)(b)?/s", "a", 100000, true)
	run("a", 2, a, "failure_again", "/^(a+)+$/s", long, 1, false)
	run("b", 1, b, "no_match", "/z/s", "a", 100000, false)
	if a.PregLastError83() != 2 {
		t.Fatal("execution B reset A failure state")
	}
	run("a", 3, a, "no_match_reset", "/z/s", "a", 100000, false)
	run("a", 4, a, "nested_failure", "/^(a+)+$/s", long, 1, true)
	if a.PregLastError83() != oracles["a"].Final || b.PregLastError83() != oracles["b"].Final {
		t.Fatal("final state differs")
	}
}

func preg83NestedNativeCall(context *runtime.Execution, pattern, subject string, limits runtime.Preg83Limits) (runtime.Preg83Result, error) {
	return context.PregMatch83(pattern, subject, 0, limits)
}
