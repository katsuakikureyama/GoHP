package sourceinventory

import (
	"encoding/base64"
	"fmt"
)

// NameEvidence records only name-level evidence for a lexical candidate. It
// neither resolves the candidate to a declaration nor establishes that the
// candidate is a call.
type NameEvidence struct {
	NameFrequency
	Category            string `json:"category"`
	NativeInternalMatch bool   `json:"native_internal_match"`
	DeclarationMarkers  int    `json:"selected_declaration_markers"`
}

// attachNameEvidence adds bounded, name-only observations to every selected
// file. Declaration markers are intentionally not semantic declarations: they
// can include methods, and a matching candidate is not thereby resolved.
func attachNameEvidence(files []FileInventory, internalFunctions []string) error {
	declarations := make(map[string]int)
	for fileIndex := range files {
		for markerIndex, declaration := range files[fileIndex].Declarations {
			if declaration.Kind != "named-function-marker" {
				continue
			}
			name, err := decodedName(declaration.NameBase64)
			if err != nil {
				return fmt.Errorf("file %d declaration marker %d: %w", fileIndex, markerIndex, err)
			}
			declarations[asciiFold(name)]++
		}
	}

	nativeFunctions := make(map[string]struct{}, len(internalFunctions))
	for index, name := range internalFunctions {
		if name == "" {
			return fmt.Errorf("native internal name %d is empty", index)
		}
		nativeFunctions[asciiFold([]byte(name))] = struct{}{}
	}

	evidence := make([][]NameEvidence, len(files))
	for fileIndex := range files {
		evidence[fileIndex] = make([]NameEvidence, 0, len(files[fileIndex].CandidateNames))
		for candidateIndex, candidate := range files[fileIndex].CandidateNames {
			if candidate.Count <= 0 {
				return fmt.Errorf("file %d candidate name %d has non-positive count", fileIndex, candidateIndex)
			}
			name, err := decodedName(candidate.NameBase64)
			if err != nil {
				return fmt.Errorf("file %d candidate name %d: %w", fileIndex, candidateIndex, err)
			}

			item := NameEvidence{NameFrequency: candidate}
			if hasBackslash(name) {
				item.Category = "qualified-name-unresolved"
			} else {
				folded := asciiFold(name)
				_, item.NativeInternalMatch = nativeFunctions[folded]
				item.DeclarationMarkers = declarations[folded]
				item.Category = evidenceCategory(item.NativeInternalMatch, item.DeclarationMarkers)
			}
			evidence[fileIndex] = append(evidence[fileIndex], item)
		}
	}

	// Apply only after validating every source of evidence, so a caller never
	// receives a partially attached view on an error.
	for fileIndex := range files {
		files[fileIndex].NameEvidence = evidence[fileIndex]
		files[fileIndex].NameEvidenceStatus = "name-only-not-resolved"
	}
	return nil
}

func decodedName(encoded string) ([]byte, error) {
	if encoded == "" {
		return nil, fmt.Errorf("name is empty")
	}
	name, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("name base64 is invalid: %w", err)
	}
	if len(name) == 0 {
		return nil, fmt.Errorf("name is empty")
	}
	return name, nil
}

func asciiFold(name []byte) string {
	folded := make([]byte, len(name))
	for index, value := range name {
		if value >= 'A' && value <= 'Z' {
			value += 'a' - 'A'
		}
		folded[index] = value
	}
	return string(folded)
}

func hasBackslash(name []byte) bool {
	for _, value := range name {
		if value == '\\' {
			return true
		}
	}
	return false
}

func evidenceCategory(native bool, declarations int) string {
	switch {
	case native && declarations > 0:
		return "both-name-sources"
	case native:
		return "native-internal-name"
	case declarations > 0:
		return "selected-declaration-name"
	default:
		return "unmatched-name"
	}
}
