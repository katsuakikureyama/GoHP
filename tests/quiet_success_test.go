package tests

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/compat"
)

func TestQuietSuccessHelperProcess(t *testing.T) {
	if os.Getenv("GOHP_WANT_QUIET_HELPER") != "1" {
		return
	}
	scenario := os.Args[len(os.Args)-1]
	switch scenario {
	case "success":
		_, _ = os.Stdout.WriteString("stdout")
	case "same":
		_, _ = os.Stdout.WriteString("same")
	case "noisy-success":
		_, _ = os.Stdout.WriteString("stdout")
		_, _ = os.Stderr.WriteString("noise")
	case "nonzero":
		_, _ = os.Stdout.WriteString("stdout")
		_, _ = os.Stderr.WriteString("failure")
		os.Exit(7)
	default:
		os.Exit(99)
	}
	os.Exit(0)
}

func quietHelperCommand(scenario string) *exec.Cmd {
	cmd := exec.Command(os.Args[0], "-test.run=TestQuietSuccessHelperProcess", "--", scenario)
	cmd.Env = append(os.Environ(), "GOHP_WANT_QUIET_HELPER=1")
	return cmd
}

func TestCaptureCommandStderrObservations(t *testing.T) {
	for _, tt := range []struct {
		name       string
		command    *exec.Cmd
		wantStdout string
		wantStderr string
		wantExit   bool
		observed   bool
	}{
		{name: "success", command: quietHelperCommand("success"), wantStdout: "stdout", observed: true},
		{name: "noisy success", command: quietHelperCommand("noisy-success"), wantStdout: "stdout", wantStderr: "noise", observed: true},
		{name: "nonzero", command: quietHelperCommand("nonzero"), wantStdout: "stdout", wantStderr: "failure", wantExit: true, observed: true},
		{name: "launch failure", command: exec.Command("/definitely/not/a/gohp-command"), wantExit: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := captureCommand(tt.command)
			if got := string(result.stdout); got != tt.wantStdout {
				t.Fatalf("stdout = %q, want %q", got, tt.wantStdout)
			}
			if got := string(result.stderr); got != tt.wantStderr {
				t.Fatalf("stderr = %q, want %q", got, tt.wantStderr)
			}
			if tt.wantExit {
				if result.err == nil {
					t.Fatal("command unexpectedly succeeded")
				}
			} else if result.err != nil {
				t.Fatalf("command failed: %v", result.err)
			}
			var exitError *exec.ExitError
			if tt.name == "nonzero" && !errors.As(result.err, &exitError) {
				t.Fatalf("error = %T, want *exec.ExitError", result.err)
			}
			if tt.observed != (result.observation != nil) {
				t.Fatalf("observation present = %t, want %t", result.observation != nil, tt.observed)
			}
			if !tt.observed {
				return
			}
			if result.observation.Bytes != len(tt.wantStderr) || result.observation.SHA256 != sha256Hex([]byte(tt.wantStderr)) {
				t.Fatalf("stderr observation = %#v", result.observation)
			}
		})
	}
}

func TestDiagnosticStatusRequiresBothQuietObservations(t *testing.T) {
	empty := &compat.StderrObservation{Bytes: 0, SHA256: sha256Hex(nil)}
	noisy := &compat.StderrObservation{Bytes: 1, SHA256: sha256Hex([]byte("x"))}
	for _, tt := range []struct {
		name  string
		php   *compat.StderrObservation
		goRun *compat.StderrObservation
		want  string
	}{
		{name: "both quiet", php: empty, goRun: empty, want: phasePass},
		{name: "native noisy", php: noisy, goRun: empty, want: phaseFail},
		{name: "generated noisy", php: empty, goRun: noisy, want: phaseFail},
		{name: "identical noise", php: noisy, goRun: noisy, want: phaseFail},
		{name: "partial noise", php: noisy, goRun: nil, want: phaseFail},
		{name: "launch failure", php: nil, goRun: empty, want: comparisonNotChecked},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := diagnosticStatus(tt.php, tt.goRun); got != tt.want {
				t.Fatalf("diagnostic status = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestQuietPHPDeprecationIsNoisyDespiteSuccessfulMatchingStdout(t *testing.T) {
	php, err := exec.LookPath("php")
	if err != nil {
		t.Skip("native PHP is not installed")
	}
	args := append(compat.QuietPHPArgs(), "-r", "trigger_error('quiet success check', E_USER_DEPRECATED); echo 'same';")
	result := captureCommand(exec.Command(php, args...))
	if result.err != nil {
		t.Fatalf("native PHP failed: %v", result.err)
	}
	goResult := captureCommand(quietHelperCommand("same"))
	if goResult.err != nil {
		t.Fatalf("quiet generated-side helper failed: %v", goResult.err)
	}
	if !bytes.Equal(result.stdout, goResult.stdout) {
		t.Fatalf("stdout did not match quiet generated-side output: PHP=%q Go=%q", result.stdout, goResult.stdout)
	}
	if result.observation == nil || result.observation.Bytes == 0 || !strings.Contains(string(result.stderr), "quiet success check") {
		t.Fatalf("expected recorded PHP deprecation, observation=%#v", result.observation)
	}
	if goResult.observation == nil || goResult.observation.Bytes != 0 {
		t.Fatal("Go helper did not prove empty stderr")
	}
	if status := diagnosticStatus(result.observation, goResult.observation); status != phaseFail {
		t.Fatalf("diagnostic status = %q, want %q", status, phaseFail)
	}
}
