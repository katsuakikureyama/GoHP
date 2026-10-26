package phpversion

import "testing"

func TestOracleSecurityFloor(t *testing.T) {
	for _, tt := range []struct {
		target   string
		accepted []string
		rejected []string
	}{
		{
			target:   "8.3",
			accepted: []string{"8.3.36", "8.3.37", "8.3.100"},
			rejected: []string{"8.3.35", "8.3.036", "8.3.36RC1", "8.3.36-dev", "8.3.36+build", "8.3.36\n"},
		},
		{
			target:   "8.4",
			accepted: []string{"8.4.27", "8.4.28", "8.4.100"},
			rejected: []string{"8.4.26", "8.4.027", "8.4.27RC1", "8.4.27-dev", "8.4.27+build", "8.4.27\n"},
		},
	} {
		t.Run(tt.target, func(t *testing.T) {
			for _, version := range tt.accepted {
				if err := ValidateOracle(tt.target, version); err != nil {
					t.Errorf("admission of %s: %v", version, err)
				}
			}
			for _, version := range tt.rejected {
				if err := ValidateOracle(tt.target, version); err == nil {
					t.Errorf("inadmissible oracle accepted: %q", version)
				}
			}
		})
	}

	for _, version := range []string{"", "7.4.33", "8.2.99", "8.3.9", "8.4.0", "8.5.0", "9.0.0", "8.3", "8.4", "8.3.4294967296", "8.4.4294967296"} {
		if err := ValidateOracle("8.3", version); err == nil {
			t.Errorf("inadmissible 8.3 oracle accepted: %q", version)
		}
		if err := ValidateOracle("8.4", version); err == nil {
			t.Errorf("inadmissible 8.4 oracle accepted: %q", version)
		}
	}
}

func TestOracleRejectsOtherEnabledSeries(t *testing.T) {
	for _, tt := range []struct {
		target  string
		version string
	}{
		{target: "8.3", version: "8.4.27"},
		{target: "8.4", version: "8.3.36"},
	} {
		if err := ValidateOracle(tt.target, tt.version); err == nil {
			t.Errorf("cross-series oracle accepted: target=%s oracle=%s", tt.target, tt.version)
		}
	}
}

func TestTargetsRequireExplicitEnablement(t *testing.T) {
	if err := ValidateTarget(DefaultTarget); err != nil {
		t.Fatal(err)
	}
	if err := ValidateTarget("8.4"); err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{"", "7.4", "8.2", "8.3.35", "8.3.36", "8.4.27", "8.5", "9.0"} {
		if err := ValidateTarget(target); err == nil {
			t.Errorf("unexpected enabled target %q", target)
		}
		if err := ValidateOracle(target, "8.3.36"); err == nil {
			t.Errorf("oracle bypassed disabled target %q", target)
		}
	}
}
