package sourceinventory

import (
	"encoding/base64"
	"reflect"
	"testing"
)

func TestCandidateNameEvidence(t *testing.T) {
	encoded := func(value []byte) string { return base64.StdEncoding.EncodeToString(value) }
	files := []FileInventory{
		{
			CandidateNames: []NameFrequency{
				{Name: "MiXeD", NameBase64: encoded([]byte("MiXeD")), Count: 3},
				{Name: "native", NameBase64: encoded([]byte("native")), Count: 2},
				{Name: "method", NameBase64: encoded([]byte("method")), Count: 1},
				{Name: "Missing", NameBase64: encoded([]byte("Missing")), Count: 1},
				{Name: "A\\B", NameBase64: encoded([]byte("A\\B")), Count: 1},
				{Name: "\\native", NameBase64: encoded([]byte("\\native")), Count: 1},
				{Name: "namespace\\native", NameBase64: encoded([]byte("namespace\\native")), Count: 1},
				{NameBase64: encoded([]byte{0xff}), Count: 1},
				{NameBase64: encoded([]byte{0xfe}), Count: 1},
				// The display spelling cannot stand in for the base64 identity.
				{Name: "native", NameBase64: encoded([]byte("missing-base64-identity")), Count: 1},
				{Name: "Ä", NameBase64: encoded([]byte("Ä")), Count: 1},
			},
			SemanticUnits: nil,
			ResolvedCalls: nil,
		},
		{
			Declarations: []Marker{
				{Kind: "named-function-marker", Name: "mixed", NameBase64: encoded([]byte("mixed"))},
				{Kind: "named-function-marker", Name: "MIXED", NameBase64: encoded([]byte("MIXED"))},
				// Marker kind deliberately remains lexical; it could be a method.
				{Kind: "named-function-marker", Name: "method", NameBase64: encoded([]byte("method"))},
				// Distinct non-UTF-8 byte spellings have the same empty display name.
				{Kind: "named-function-marker", NameBase64: encoded([]byte{0xff})},
				{Kind: "anonymous-function-marker"},
			},
		},
	}

	originalCandidates := append([]NameFrequency(nil), files[0].CandidateNames...)
	if err := attachNameEvidence(files, []string{"NATIVE", "MIXED", "ä"}); err != nil {
		t.Fatalf("attachNameEvidence() error = %v", err)
	}
	if !reflect.DeepEqual(files[0].CandidateNames, originalCandidates) {
		t.Fatalf("candidate names changed = %#v, want %#v", files[0].CandidateNames, originalCandidates)
	}
	if files[0].NameEvidenceStatus != "name-only-not-resolved" || files[1].NameEvidenceStatus != "name-only-not-resolved" {
		t.Fatalf("evidence statuses = %q / %q", files[0].NameEvidenceStatus, files[1].NameEvidenceStatus)
	}
	if files[0].SemanticUnits != nil || files[0].ResolvedCalls != nil {
		t.Fatalf("unresolved fields changed = %v / %v", files[0].SemanticUnits, files[0].ResolvedCalls)
	}

	got := files[0].NameEvidence
	if len(got) != len(originalCandidates) {
		t.Fatalf("evidence length = %d, want %d", len(got), len(originalCandidates))
	}
	want := []struct {
		category     string
		native       bool
		declarations int
	}{
		{"both-name-sources", true, 2},
		{"native-internal-name", true, 0},
		{"selected-declaration-name", false, 1},
		{"unmatched-name", false, 0},
		{"qualified-name-unresolved", false, 0},
		{"qualified-name-unresolved", false, 0},
		{"qualified-name-unresolved", false, 0},
		{"selected-declaration-name", false, 1},
		{"unmatched-name", false, 0},
		{"unmatched-name", false, 0},
		{"unmatched-name", false, 0},
	}
	for index, expected := range want {
		if got[index].NameFrequency != originalCandidates[index] || got[index].Category != expected.category || got[index].NativeInternalMatch != expected.native || got[index].DeclarationMarkers != expected.declarations {
			t.Fatalf("evidence[%d] = %#v, want candidate %#v and %q/%v/%d", index, got[index], originalCandidates[index], expected.category, expected.native, expected.declarations)
		}
	}
	if files[1].NameEvidence == nil || len(files[1].NameEvidence) != 0 {
		t.Fatalf("empty evidence = %#v, want non-nil empty", files[1].NameEvidence)
	}
}

func TestCandidateNameEvidenceInvalid(t *testing.T) {
	valid := base64.StdEncoding.EncodeToString([]byte("name"))
	cases := []struct {
		name   string
		files  []FileInventory
		native []string
	}{
		{"candidate invalid base64", []FileInventory{{CandidateNames: []NameFrequency{{NameBase64: "!", Count: 1}}}}, nil},
		{"candidate empty base64", []FileInventory{{CandidateNames: []NameFrequency{{NameBase64: "", Count: 1}}}}, nil},
		{"candidate non-positive count", []FileInventory{{CandidateNames: []NameFrequency{{NameBase64: valid, Count: 0}}}}, nil},
		{"declaration invalid base64", []FileInventory{{Declarations: []Marker{{Kind: "named-function-marker", NameBase64: "!"}}}}, nil},
		{"declaration empty name", []FileInventory{{Declarations: []Marker{{Kind: "named-function-marker", NameBase64: ""}}}}, nil},
		{"native empty name", nil, []string{""}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if err := attachNameEvidence(test.files, test.native); err == nil {
				t.Fatal("attachNameEvidence() error = nil")
			}
		})
	}
}
