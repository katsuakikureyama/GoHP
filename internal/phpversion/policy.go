// Package phpversion defines the explicitly enabled PHP language targets and
// their minimum native reference versions. A newer series is never auto-enabled.
package phpversion

import (
	"fmt"
	"strconv"
	"strings"
)

const DefaultTarget = "8.3"

type Policy struct {
	Target  string
	Minimum string
}

// Add another series only together with its frontend behavior, pinned CI oracle
// and differential evidence. The minimum is an admission floor, not a promise
// that every later patch or an untested minor is compatible or vulnerability-free.
var policies = [...]Policy{
	{Target: "8.3", Minimum: "8.3.36"},
	{Target: "8.4", Minimum: "8.4.27"},
}

func TargetHelp() string {
	var enabled []string
	for _, policy := range policies {
		enabled = append(enabled, fmt.Sprintf("%s (native oracle >=%s in that series)", policy.Target, policy.Minimum))
	}
	return strings.Join(enabled, ", ")
}

func Lookup(target string) (Policy, error) {
	for _, policy := range policies {
		if policy.Target == target {
			return policy, nil
		}
	}
	return Policy{}, fmt.Errorf("PHP target %q is not enabled; enabled targets: %s; newer series require separate validation", target, TargetHelp())
}

func ValidateTarget(target string) error {
	_, err := Lookup(target)
	return err
}

// ValidateOracle checks the actual PHP_VERSION, never an image tag or a lexical
// string comparison. Only canonical stable X.Y.Z releases in the selected series
// are admitted; prereleases, suffixes, missing patches and unknown minors fail.
func ValidateOracle(target, version string) error {
	policy, err := Lookup(target)
	if err != nil {
		return err
	}
	actual, ok := release(version)
	minimum, _ := release(policy.Minimum)
	if !ok || actual[0] != minimum[0] || actual[1] != minimum[1] || actual[2] < minimum[2] {
		return fmt.Errorf("PHP oracle %q is not admitted: target %s requires a stable release >=%s in that series", version, target, policy.Minimum)
	}
	return nil
}

func release(version string) ([3]uint64, bool) {
	var result [3]uint64
	parts := strings.Split(version, ".")
	if len(parts) != len(result) {
		return result, false
	}
	for i, part := range parts {
		n, err := strconv.ParseUint(part, 10, 32)
		if err != nil || strconv.FormatUint(n, 10) != part {
			return result, false
		}
		result[i] = n
	}
	return result, true
}
