package sourceinventory

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

//go:embed tokenize.php
var tokenizerScript string

type tokenization struct {
	Schema     string  `json:"schema"`
	PHPVersion string  `json:"php_version"`
	Bytes      int     `json:"bytes"`
	SHA256     string  `json:"sha256"`
	Tokens     []Token `json:"tokens"`
}

func tokenize(ctx context.Context, source []byte) (tokenization, error) {
	if len(source) > 16<<20 {
		return tokenization{}, fmt.Errorf("source exceeds tokenizer byte limit")
	}
	cmd := exec.CommandContext(ctx, "php", "-n", "-d", "short_open_tag=0", "-d", "memory_limit=256M", "-d", "display_errors=stderr", "-d", "log_errors=0", "-r", strings.TrimPrefix(tokenizerScript, "<?php\n"))
	cmd.Stdin = bytes.NewReader(source)
	stdout := &boundedOutput{limit: 16 << 20}
	stderr := &boundedOutput{limit: 64 << 10}
	cmd.Stdout, cmd.Stderr = stdout, stderr
	cmd.WaitDelay = time.Second
	if err := cmd.Run(); err != nil || stderr.Len() != 0 {
		// Never echo raw PHP diagnostics, source snippets or local config paths.
		return tokenization{}, fmt.Errorf("PHP tokenizer failed, exceeded limits, or emitted diagnostics")
	}
	return decodeTokens(stdout.Bytes(), source)
}

func decodeTokens(raw, source []byte) (tokenization, error) {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	var result tokenization
	if err := d.Decode(&result); err != nil {
		return tokenization{}, fmt.Errorf("invalid PHP tokenizer JSON")
	}
	if _, err := d.Token(); err != io.EOF {
		return tokenization{}, fmt.Errorf("trailing PHP tokenizer output")
	}
	if result.Schema != "gohp.php.tokens/v0.1" || result.Bytes != len(source) || result.SHA256 != fmt.Sprintf("%x", sha256.Sum256(source)) || result.Tokens == nil || len(result.Tokens) > 250000 {
		return tokenization{}, fmt.Errorf("PHP tokenizer response does not bind the complete source")
	}
	return result, nil
}

type boundedOutput struct {
	buffer bytes.Buffer
	limit  int
}

func (b *boundedOutput) Len() int      { return b.buffer.Len() }
func (b *boundedOutput) Bytes() []byte { return b.buffer.Bytes() }

func (b *boundedOutput) Write(p []byte) (int, error) {
	if len(p) > b.limit-b.Len() {
		return 0, fmt.Errorf("tokenizer output limit exceeded")
	}
	return b.buffer.Write(p)
}
