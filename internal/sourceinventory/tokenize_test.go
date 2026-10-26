package sourceinventory

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestTokenizerResponseBinding(t *testing.T) {
	source := []byte("some text")
	valid := tokenization{Schema: "gohp.php.tokens/v0.1", PHPVersion: "8.3.36", Bytes: len(source), SHA256: fmt.Sprintf("%x", sha256.Sum256(source)), Tokens: []Token{{Kind: "T_INLINE_HTML", Bytes: len(source)}}}
	raw, err := json.Marshal(valid)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := decodeTokens(raw, source); err != nil || !reflect.DeepEqual(got, valid) {
		t.Fatalf("binding failed: %+v %v", got, err)
	}
	for _, test := range []struct {
		name   string
		change func(*tokenization)
	}{
		{"wrong schema", func(r *tokenization) { r.Schema = "other" }},
		{"wrong bytes", func(r *tokenization) { r.Bytes++ }},
		{"wrong hash", func(r *tokenization) { r.SHA256 = "wrong" }},
		{"missing tokens", func(r *tokenization) { r.Tokens = nil }},
	} {
		t.Run(test.name, func(t *testing.T) {
			bad := valid
			test.change(&bad)
			raw, _ := json.Marshal(bad)
			if _, err := decodeTokens(raw, source); err == nil {
				t.Fatal("accepted unbound tokens")
			}
		})
	}
	for _, bad := range [][]byte{append(append([]byte{}, raw...), []byte("{}")...), []byte("{\"unknown\":true}")} {
		if _, err := decodeTokens(bad, source); err == nil {
			t.Fatal("accepted malformed output")
		}
	}
}

func TestTokenizerFailures(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := tokenize(ctx, []byte("<?php")); err == nil {
		t.Fatal("cancelled tokenizer succeeded")
	}
	b := &boundedOutput{limit: 3}
	if _, err := b.Write([]byte("four")); err == nil || b.Len() != 0 {
		t.Fatal("output bound failed")
	}
	if _, err := io.Copy(b, struct{ io.Reader }{strings.NewReader("four")}); err == nil || b.Len() != 0 {
		t.Fatal("pipe copy bypassed output bound")
	}
	// A trusted executable's startup diagnostics and failing exit must not be
	// confused with a successful lexical observation, even when stdout is JSON.
	root := t.TempDir()
	t.Setenv("PATH", root)
	for _, script := range []string{"#!/bin/sh\nprintf 'diagnostic' >&2\n", "#!/bin/sh\nexit 7\n", "#!/bin/sh\nprintf '{}'\n"} {
		if err := os.WriteFile(filepath.Join(root, "php"), []byte(script), 0700); err != nil {
			t.Fatal(err)
		}
		if _, err := tokenize(context.Background(), []byte("<?php")); err == nil {
			t.Fatal("accepted failed tokenizer")
		}
	}
}
