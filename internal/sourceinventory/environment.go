package sourceinventory

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"
)

// NativeEnvironment is a name-set observation of the trusted scanner process,
// not PHP's universal function list or a resolved application dependency set.
type NativeEnvironment struct {
	InternalFunctions []string `json:"internal_functions"`
	Extensions        []string `json:"extensions"`
	ZendExtensions    []string `json:"zend_extensions"`
	SHA256            string   `json:"sha256"`
}

func nativeEnvironment(result tokenization) (NativeEnvironment, error) {
	for _, list := range []struct {
		names []string
		limit int
	}{
		{result.InternalFunctions, 10000},
		{result.Extensions, 512},
		{result.ZendExtensions, 512},
	} {
		if list.names == nil || len(list.names) > list.limit {
			return NativeEnvironment{}, fmt.Errorf("missing or oversized native name set")
		}
		for i, name := range list.names {
			if name == "" || len(name) > 1024 || !utf8.ValidString(name) || strings.IndexFunc(name, func(r rune) bool { return r < 32 || r == 127 }) >= 0 || (i > 0 && list.names[i-1] >= name) {
				return NativeEnvironment{}, fmt.Errorf("native names must be bounded, unique and byte-sorted")
			}
		}
	}
	// Exact canonical JSON array of the three ordered lists, no newline. Keep
	// the observed lists visible so a hash never stands in for the evidence.
	raw, err := json.Marshal([][]string{result.InternalFunctions, result.Extensions, result.ZendExtensions})
	if err != nil {
		return NativeEnvironment{}, err
	}
	return NativeEnvironment{
		InternalFunctions: result.InternalFunctions, Extensions: result.Extensions,
		ZendExtensions: result.ZendExtensions, SHA256: fmt.Sprintf("%x", sha256.Sum256(raw)),
	}, nil
}
