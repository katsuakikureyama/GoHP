package sourceinventory

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/katsuakikureyama/GoHP/internal/corpus"
	"github.com/katsuakikureyama/GoHP/internal/phpversion"
)

// Inventory verifies the complete selection before starting a trusted native
// lexical scanner. Selected PHP source is only supplied as stdin data.
func Inventory(ctx context.Context, manifestPath, sourceRoot, phpTarget string) (Report, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	return inventory(ctx, manifestPath, sourceRoot, phpTarget, tokenize)
}

type tokenizer func(context.Context, []byte) (tokenization, error)

func inventory(ctx context.Context, manifestPath, sourceRoot, target string, scan tokenizer) (Report, error) {
	if err := phpversion.ValidateTarget(target); err != nil {
		return Report{}, err
	}
	verified, sources, err := corpus.LoadSources(manifestPath, sourceRoot)
	if err != nil {
		return Report{}, err
	}
	selected := false
	for _, t := range verified.Manifest.PHPTargets {
		selected = selected || t == target
	}
	if !selected {
		return Report{}, fmt.Errorf("target is outside the manifest measurement scope")
	}
	report := Report{
		Schema: Schema, Status: "lexical-only", PHPTarget: target,
		TokenizerSHA256: fmt.Sprintf("%x", sha256.Sum256([]byte(tokenizerScript))),
		Verification:    verified, Files: make([]FileInventory, 0, len(sources)), Compatibility: "unverified",
		Notice: "Native token_get_all(flags=0), php -n, short_open_tag=0. Native internal-name and selected declaration-marker matches are name-only evidence, never call resolution or GoHP support. Qualified names remain unresolved; namespace/import/conditional/method context is not resolved. No grammar validation, source execution, GoHP compilation or compatibility percentage. Candidates can include attributes and first-class callable syntax. Every other opening parenthesis remains visible. Semantic units, resolved calls and dependency closure remain unmeasured.",
	}
	totalTokens := 0
	for _, source := range sources {
		if err := ctx.Err(); err != nil {
			return Report{}, err
		}
		tokens, err := scan(ctx, source.Bytes)
		if err != nil {
			return Report{}, fmt.Errorf("tokenize %s: %w", source.Path, err)
		}
		if err := phpversion.ValidateOracle(target, tokens.PHPVersion); err != nil {
			return Report{}, err
		}
		if report.PHPVersion != "" && report.PHPVersion != tokens.PHPVersion {
			return Report{}, fmt.Errorf("tokenizer PHP version changed within the selection")
		}
		report.PHPVersion = tokens.PHPVersion
		environment, err := nativeEnvironment(tokens)
		if err != nil {
			return Report{}, err
		}
		if report.NativeEnvironment.SHA256 != "" && report.NativeEnvironment.SHA256 != environment.SHA256 {
			return Report{}, fmt.Errorf("native name environment changed within the selection")
		}
		report.NativeEnvironment = environment
		if len(tokens.Tokens) > 500000-totalTokens {
			return Report{}, fmt.Errorf("selection exceeds 500000 total token limit")
		}
		totalTokens += len(tokens.Tokens)
		file, err := Analyze(source.Path, source.Bytes, tokens.Tokens)
		if err != nil {
			return Report{}, fmt.Errorf("analyze %s: %w", source.Path, err)
		}
		report.Files = append(report.Files, file)
	}
	if err := attachNameEvidence(report.Files, report.NativeEnvironment.InternalFunctions); err != nil {
		return Report{}, err
	}
	return report, nil
}
