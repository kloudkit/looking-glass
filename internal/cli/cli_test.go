package cli

import "testing"

func TestResolveVersionPrefersStamped(t *testing.T) {
	defer func(v string) { version = v }(version)

	for stamped, want := range map[string]string{
		"v0.3.0": "0.3.0",
		"0.3.0":  "0.3.0",
		"dev":    "dev", // a test binary carries no release module version
	} {
		version = stamped
		if got := resolveVersion(); got != want {
			t.Errorf("version %q: got %q, want %q", stamped, got, want)
		}
	}
}

func TestReleaseTagRejectsPseudoVersions(t *testing.T) {
	for v, want := range map[string]bool{
		"v0.2.0":                               true,
		"v0.2.1-0.20261013134525-e15fbd3c5a19": false,
		"v0.2.1-0.20261013134525-e15fbd3c5a19+dirty": false,
		"(devel)": false,
	} {
		if got := releaseTag.MatchString(v); got != want {
			t.Errorf("%q: got %v, want %v", v, got, want)
		}
	}
}
