package frontend

import (
	"crypto/sha256"
	"fmt"

	"github.com/katsuakikureyama/GoHP/internal/ir"
)

// ParseSource is the filesystem-free single-source adapter for future checked
// closure lowering. It does not admit require_once or return a loaded closure.
func ParseSource(sourceID, src, version string, options Options) (ir.Program, error) {
	if !ir.ValidSourceID(sourceID) {
		return ir.Program{}, fmt.Errorf("PHP003: source ID must be a canonical relative slash path")
	}
	p, err := ParseWithOptions(src, version, options)
	if err != nil {
		return ir.Program{}, ir.WithSource(err, sourceID)
	}
	p.Sources = []ir.SourceFile{{ID: sourceID, Bytes: int64(len(src)), SHA256: fmt.Sprintf("%x", sha256.Sum256([]byte(src)))}}
	for i := range p.Functions {
		p.Functions[i].SourceID = sourceID
	}
	for i := range p.Classes {
		p.Classes[i].SourceID = sourceID
	}
	for i := range p.Body {
		p.Body[i].SourceID = sourceID
	}
	return p, nil
}
