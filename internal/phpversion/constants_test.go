package phpversion

import "testing"

func TestVersionConstantsUseSelectedSeries(t *testing.T) {
	for _, tt := range []struct {
		target string
		minor  int64
	}{{"8.3", 3}, {"8.4", 4}, {"8.3", 3}} {
		for name, want := range map[string]int64{"PHP_MAJOR_VERSION": 8, "PHP_MINOR_VERSION": tt.minor} {
			if got, err := VersionConstant(tt.target, name); err != nil || got != want {
				t.Fatalf("%s/%s: %d, %v; want %d", tt.target, name, got, err, want)
			}
		}
	}
	for _, target := range []string{"", "8.2", "8.5", "8.3.36", "8.4.27"} {
		if _, err := VersionConstant(target, "PHP_MAJOR_VERSION"); err == nil {
			t.Fatalf("unadmitted target accepted: %q", target)
		}
	}
	for _, name := range []string{"", "php_minor_version", "PHP_VERSION", "PHP_VERSION_ID", "PHP_RELEASE_VERSION", "PHP_EXTRA_VERSION", "PHP_INT_SIZE"} {
		if _, err := VersionConstant("8.3", name); err == nil {
			t.Fatalf("unknown or patch/platform-dependent constant accepted: %q", name)
		}
	}
}
