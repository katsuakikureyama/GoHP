// Command gohp-compat runs the repository tests once and publishes bounded
// compatibility evidence. It is a development tool, not a PHP support oracle.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/katsuakikureyama/GoHP/internal/compat"
	"github.com/katsuakikureyama/GoHP/internal/phpversion"
)

func main() {
	out := flag.String("out", "build/compatibility", "fresh evidence directory (must not exist)")
	flag.Parse()
	if flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "gohp-compat accepts no positional arguments")
		os.Exit(2)
	}
	if err := run(*out); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(out string) error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	catalog, err := compat.LoadCatalog(root)
	if err != nil {
		return err
	}
	meta, err := metadata()
	if err != nil {
		return err
	}
	meta.SourceSHA256, err = sourceFingerprint(root)
	if err != nil {
		return err
	}
	// A caller-owned, non-concurrently-mutated workspace is required. Never
	// overwrite or resume an older run, even if it ended with incomplete files.
	if err := os.MkdirAll(filepath.Dir(out), 0700); err != nil {
		return err
	}
	if err := os.Mkdir(out, 0700); err != nil {
		return fmt.Errorf("create fresh evidence directory: %w", err)
	}
	catalogBytes, err := os.ReadFile(filepath.Join(root, "compatibility", "catalog.json"))
	if err != nil {
		return err
	}
	if fmt.Sprintf("%x", sha256.Sum256(catalogBytes)) != catalog.SHA256 {
		return fmt.Errorf("catalog changed after validation; use a non-concurrently-mutated workspace")
	}
	if err := writeExclusive(filepath.Join(out, "catalog.json"), catalogBytes); err != nil {
		return err
	}
	events, err := os.OpenFile(filepath.Join(out, "go-test.jsonl"), os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer events.Close()
	stderr, err := os.OpenFile(filepath.Join(out, "go-test.stderr"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer stderr.Close()
	cmd := exec.Command("go", "test", "-count=1", "-json", "./cmd/...", "./internal/...", "./runtime", "./tests")
	cmd.Env = append(os.Environ(), "GOHP_PHP_TARGET="+meta.PHPTarget)
	cmd.Stdout = events
	cmd.Stderr = io.MultiWriter(stderr, os.Stderr)
	fmt.Fprintln(os.Stderr, "Running", meta.Command, "for PHP", meta.PHPTarget)
	runErr := cmd.Run()
	meta.TestExit = 0
	if runErr != nil {
		meta.TestExit = -1
		var exit *exec.ExitError
		if errors.As(runErr, &exit) {
			meta.TestExit = exit.ExitCode()
		}
	}
	if err := errors.Join(events.Sync(), stderr.Sync()); err != nil {
		return err
	}
	if _, err := events.Seek(0, io.SeekStart); err != nil {
		return err
	}
	report, err := compat.Evaluate(catalog, events, meta)
	if err != nil {
		return fmt.Errorf("evidence rejected (raw logs retained in %s): %w", out, err)
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := writeExclusive(filepath.Join(out, "summary.md"), compat.Markdown(report)); err != nil {
		return err
	}
	// Report is written last. Consumers still require the process/job outcome:
	// a crash or subsequent CI smoke-test failure is not a successful build.
	if err := writeExclusive(filepath.Join(out, "run.json"), append(data, '\n')); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "Compatibility evidence: %s (%s)\n", out, report.Overall)
	if runErr != nil {
		return fmt.Errorf("test execution failed; see retained evidence: %w", runErr)
	}
	if report.Overall != "pass" {
		return fmt.Errorf("compatibility evidence is %s; see %s", report.Overall, out)
	}
	return nil
}

func writeExclusive(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	return errors.Join(err, f.Sync(), f.Close())
}

func commandText(program string, args ...string) string {
	b, err := exec.Command(program, args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func metadata() (compat.Metadata, error) {
	m := compat.Metadata{
		Commit: os.Getenv("GOHP_COMMIT"), Tree: os.Getenv("GOHP_TREE"), Dirty: os.Getenv("GOHP_DIRTY"),
		GoVersion: commandText("go", "version"), GOOS: runtime.GOOS, GOARCH: runtime.GOARCH,
		PHPVersion: commandText("php", "-r", "echo PHP_VERSION;"),
		PHPTarget:  os.Getenv("GOHP_PHP_TARGET"), Image: os.Getenv("GOHP_PHP_IMAGE"),
		GoImage: os.Getenv("GOHP_GO_IMAGE"), RunURL: os.Getenv("GOHP_RUN_URL"),
		RunID: os.Getenv("GOHP_RUN_ID"), Attempt: os.Getenv("GOHP_RUN_ATTEMPT"),
		Command: "go test -count=1 -json ./cmd/... ./internal/... ./runtime ./tests", StartedAt: time.Now().UTC().Format(time.RFC3339Nano),
		BuildFlags: fmt.Sprintf("CGO_ENABLED=%s GOTOOLCHAIN=%s GOFLAGS=%s", os.Getenv("CGO_ENABLED"), os.Getenv("GOTOOLCHAIN"), os.Getenv("GOFLAGS")),
		Provenance: "ci-checkout-env",
	}
	if m.Commit == "" {
		m.Provenance = "local-git"
		m.Commit = commandText("git", "rev-parse", "HEAD")
		m.Tree = commandText("git", "rev-parse", "HEAD^{tree}")
		status, err := exec.Command("git", "status", "--porcelain", "--untracked-files=all").Output()
		m.Dirty = "unknown"
		if err == nil {
			m.Dirty = "clean"
			if len(status) != 0 {
				m.Dirty = "dirty"
			}
		}
		if m.Commit == "" || m.Tree == "" {
			m.Provenance = "unbound-checkout"
		}
	} else {
		for _, value := range []string{m.Commit, m.Tree} {
			b, err := hex.DecodeString(value)
			if err != nil || len(b) != 20 {
				return m, fmt.Errorf("GOHP_COMMIT/GOHP_TREE must both be full Git SHA-1 identities")
			}
		}
		if m.Dirty != "clean" && m.Dirty != "dirty" {
			return m, fmt.Errorf("GOHP_DIRTY must be clean or dirty for injected checkout metadata")
		}
	}
	if m.PHPTarget == "" {
		m.PHPTarget = phpversion.DefaultTarget
	}
	// Live admission is separate from historical archive replay. Do not infer a
	// new supported target from whichever PHP happens to be on PATH.
	if err := phpversion.ValidateOracle(m.PHPTarget, m.PHPVersion); err != nil {
		return m, err
	}
	return m, nil
}

// This fingerprint identifies the actual Go/PHP inputs independently of Git
// availability. WalkDir's lexical ordering and JSON encoding make path/content
// boundaries unambiguous. Build products and VCS internals are not inputs.
func sourceFingerprint(root string) (string, error) {
	type input struct {
		Path   string `json:"path"`
		SHA256 string `json:"sha256"`
	}
	var inputs []input
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if rel == ".git" || rel == "build" || rel == ".cache" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := filepath.Ext(rel)
		inventoryFixture := strings.HasPrefix(rel, "internal/inventory/testdata/") && (filepath.Base(rel) == "composer.json" || filepath.Base(rel) == "composer.lock")
		if ext != ".go" && ext != ".php" && rel != "go.mod" && rel != "go.sum" && rel != "compatibility/catalog.json" && !inventoryFixture {
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("source fingerprint refuses symlink input %s", rel)
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		inputs = append(inputs, input{rel, fmt.Sprintf("%x", sha256.Sum256(b))})
		return nil
	})
	if err != nil {
		return "", err
	}
	b, err := json.Marshal(inputs)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(b)), nil
}
