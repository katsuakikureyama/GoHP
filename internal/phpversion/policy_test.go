package phpversion

import "testing"

func TestOracleSecurityFloor(t *testing.T) {
	for _, version := range []string{"8.3.36", "8.3.37", "8.3.100"} {
		if err := ValidateOracle("8.3", version); err != nil {
			t.Errorf("admission of %s: %v", version, err)
		}
	}
	for _, version := range []string{"", "7.4.33", "8.2.99", "8.3.9", "8.3.30", "8.3.35", "8.4.0", "8.5.0", "9.0.0", "8.3", "8.3.036", "8.3.36RC1", "8.3.36-dev", "8.3.36+build", "8.3.36\n", "8.3.4294967296"} {
		if err := ValidateOracle("8.3", version); err == nil {
			t.Errorf("inadmissible oracle accepted: %q", version)
		}
	}
}

func TestTargetsRequireExplicitEnablement(t *testing.T) {
	if err := ValidateTarget(DefaultTarget); err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{"", "7.4", "8.2", "8.3.35", "8.3.36", "8.4", "8.5", "9.0"} {
		if err := ValidateTarget(target); err == nil {
			t.Errorf("unexpected enabled target %q", target)
		}
		if err := ValidateOracle(target, "8.3.36"); err == nil {
			t.Errorf("oracle bypassed disabled target %q", target)
		}
	}
}
