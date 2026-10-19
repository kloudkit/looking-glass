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

func TestParsePrefixes(t *testing.T) {
	got, err := parsePrefixes([]string{"10.0.0.0/8", "192.0.2.7", "::1"})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"10.0.0.0/8", "192.0.2.7/32", "::1/128"}
	for i, prefix := range got {
		if prefix.String() != want[i] {
			t.Errorf("prefix %d = %s, want %s", i, prefix, want[i])
		}
	}

	if _, err := parsePrefixes([]string{"traefik"}); err == nil {
		t.Error("expected an error for a name")
	}
}

func TestEnvList(t *testing.T) {
	t.Setenv("LG_LIST", " 10.0.0.1, ,10.0.0.0/8 ")
	if got := envList("LG_LIST", nil); len(got) != 2 || got[0] != "10.0.0.1" || got[1] != "10.0.0.0/8" {
		t.Errorf("list = %q", got)
	}

	t.Setenv("LG_LIST", "")
	if got := envList("LG_LIST", []string{"x"}); len(got) != 0 {
		t.Errorf("an empty variable must trust nothing, got %q", got)
	}

	if got := envList("LG_UNSET", []string{"x"}); len(got) != 1 {
		t.Errorf("an unset variable must keep the default, got %q", got)
	}
}
