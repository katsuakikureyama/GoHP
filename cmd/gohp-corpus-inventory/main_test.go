package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/katsuakikureyama/GoHP/internal/sourceinventory"
)

func TestInventoryCLIContract(t *testing.T) {
	for _, tc := range []struct {
		name   string
		args   []string
		target string
	}{
		{name: "default target", args: []string{"-manifest", "manifest.json", "source-root"}, target: "8.3"},
		{name: "explicit target", args: []string{"-manifest", "manifest.json", "-php-version", "8.4", "source-root"}, target: "8.4"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			called := false
			pipeline := func(ctx context.Context, manifest, root, target string) (sourceinventory.Report, error) {
				called = true
				if ctx == nil || ctx.Err() != nil {
					t.Fatalf("inactive context: %v", ctx)
				}
				if manifest != "manifest.json" || root != "source-root" || target != tc.target {
					t.Fatalf("arguments = manifest=%q root=%q target=%q", manifest, root, target)
				}
				return sourceinventory.Report{
					Schema:        sourceinventory.Schema,
					Status:        "lexical-only",
					PHPTarget:     target,
					Compatibility: "unverified",
				}, nil
			}
			if code := runWithInventory(tc.args, &stdout, &stderr, pipeline); code != 0 {
				t.Fatalf("exit code = %d, stderr=%q", code, stderr.String())
			}
			if !called || stderr.Len() != 0 {
				t.Fatalf("called=%t stderr=%q", called, stderr.String())
			}
			got := stdout.String()
			for _, want := range []string{
				`"schema": "` + sourceinventory.Schema + `"`,
				`"status": "lexical-only"`,
				`"php_target": "` + tc.target + `"`,
				`"compatibility": "unverified"`,
			} {
				if !strings.Contains(got, want) {
					t.Fatalf("stdout missing %q: %q", want, got)
				}
			}
		})
	}
}

func TestInventoryCLIFailures(t *testing.T) {
	t.Run("usage failures do not invoke inventory", func(t *testing.T) {
		for _, args := range [][]string{
			nil,
			{"source-root"},
			{"-manifest", "manifest.json"},
			{"-manifest", "manifest.json", "source-root", "another-root"},
			{"-manifest", "manifest.json", "-php-version", "8.5", "source-root"},
			{"source-root", "-manifest", "manifest.json"},
		} {
			var stdout, stderr bytes.Buffer
			called := false
			pipeline := func(context.Context, string, string, string) (sourceinventory.Report, error) {
				called = true
				return sourceinventory.Report{}, nil
			}
			if code := runWithInventory(args, &stdout, &stderr, pipeline); code != 2 {
				t.Fatalf("args=%v: exit code = %d, stderr=%q", args, code, stderr.String())
			}
			if called || stdout.Len() != 0 || stderr.Len() == 0 {
				t.Fatalf("args=%v: called=%t stdout=%q stderr=%q", args, called, stdout.String(), stderr.String())
			}
		}
	})

	t.Run("help succeeds without inventory", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		if code := runWithInventory([]string{"-h"}, &stdout, &stderr, func(context.Context, string, string, string) (sourceinventory.Report, error) {
			t.Fatal("inventory must not run for help")
			return sourceinventory.Report{}, nil
		}); code != 0 || stdout.Len() != 0 || stderr.Len() == 0 {
			t.Fatalf("exit code = %d, stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})

	t.Run("inventory failure is generic and leaves stdout empty", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		if code := runWithInventory([]string{"-manifest", "manifest.json", "source-root"}, &stdout, &stderr, func(context.Context, string, string, string) (sourceinventory.Report, error) {
			return sourceinventory.Report{}, errors.New("/private/source.php token failure")
		}); code != 1 {
			t.Fatalf("exit code = %d, stderr=%q", code, stderr.String())
		}
		if stdout.Len() != 0 || stderr.String() != "inventory corpus: failed\n" {
			t.Fatalf("stdout=%q stderr=%q", stdout.String(), stderr.String())
		}
	})

	t.Run("final output failures return one", func(t *testing.T) {
		for _, writer := range []brokenWriter{{}, {short: true}} {
			var stderr bytes.Buffer
			if code := runWithInventory([]string{"-manifest", "manifest.json", "source-root"}, writer, &stderr, func(context.Context, string, string, string) (sourceinventory.Report, error) {
				return sourceinventory.Report{Schema: sourceinventory.Schema, Status: "lexical-only", Compatibility: "unverified"}, nil
			}); code != 1 || stderr.String() != "write corpus inventory: failed\n" {
				t.Fatalf("exit code = %d, stderr=%q", code, stderr.String())
			}
		}
	})
}

type brokenWriter struct{ short bool }

func (w brokenWriter) Write(p []byte) (int, error) {
	if w.short {
		return len(p) - 1, nil
	}
	return 0, errors.New("destination unavailable")
}
