package phpversion

import (
	"fmt"
	"strconv"
	"strings"
)

// VersionConstant resolves only target-series integers. The oracle's minimum
// patch and the host PHP installation must never become generated semantics.
func VersionConstant(target, name string) (int64, error) {
	policy, err := Lookup(target)
	if err != nil {
		return 0, err
	}
	major, minor, _ := strings.Cut(policy.Target, ".")
	var component string
	switch name {
	case "PHP_MAJOR_VERSION":
		component = major
	case "PHP_MINOR_VERSION":
		component = minor
	default:
		return 0, fmt.Errorf("unsupported target version constant %q", name)
	}
	return strconv.ParseInt(component, 10, 64)
}
