package vaultcli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSurveyRequired(t *testing.T) {
	tcase := func(in string, wantError bool) {
		t.Helper()
		got := SurveyRequired(in)
		if wantError && got == nil {
			t.Errorf("Expected error SurveyRequired(%s), but <nil> returned", in)
		} else if !wantError && got != nil {
			t.Errorf("Unexpected error SurveyRequired(%s): %v", in, got)
		}
	}
	tcase(" a", false)
	tcase("  a", false)
	tcase("  a ", false)
	tcase("  a  ", false)
	tcase(" a  ", false)
	tcase("a  ", false)
	tcase("	a", false)
	tcase("	a	", false)
	tcase("a	", false)

	tcase(" ", true)
	tcase("  ", true)
	tcase("		", true)
	tcase("	", true)
}

func TestSurveyRequiredInt(t *testing.T) {
	tcase := func(in string, wantError bool) {
		t.Helper()
		got := SurveyRequiredInt(in)
		if wantError && got == nil {
			t.Errorf("Expected error SurveyRequiredInt(%s), but <nil> returned", in)
		} else if !wantError && got != nil {
			t.Errorf("Unexpected error SurveyRequiredInt(%s): %v", in, got)
		}
	}
	tcase(" 1", false)
	tcase("  1", false)
	tcase("  1 ", false)
	tcase("  1  ", false)
	tcase(" 1  ", false)
	tcase("1  ", false)
	tcase("	1", false)
	tcase("	1	", false)
	tcase("1	", false)

	tcase("a", true)
	tcase(" a ", true)
	tcase("	a	", true)
	tcase("a	", true)
}

func TestSurveyRequiredPortNumber(t *testing.T) {
	tcase := func(in string, wantError bool) {
		t.Helper()
		got := SurveyRequiredPortNumber(in)
		if wantError && got == nil {
			t.Errorf("Expected error SurveyRequiredPortNumber(%s), but <nil> returned", in)
		} else if !wantError && got != nil {
			t.Errorf("Unexpected error SurveyRequiredPortNumber(%s): %v", in, got)
		}
	}
	tcase("0", false)
	tcase(" 1", false)
	tcase("  1", false)
	tcase("  1 ", false)
	tcase("  1  ", false)
	tcase(" 1  ", false)
	tcase("1  ", false)
	tcase("	1", false)
	tcase("	1	", false)
	tcase("1	", false)
	tcase("65534", false)
	tcase("65535", false)

	tcase("-1", true)
	tcase("65536", true)

	tcase("", true)
	tcase(" ", true)
	tcase("a", true)
	tcase(" a ", true)
	tcase("	a	", true)
	tcase("a	", true)
}

func TestSurveyRequiredFile(t *testing.T) {
	tcase := func(in string, wantError bool) {
		t.Helper()
		got := SurveyRequiredFile(in)
		if wantError && got == nil {
			t.Errorf("Expected error SurveyRequiredFile(%s), but <nil> returned", in)
		} else if !wantError && got != nil {
			t.Errorf("Unexpected error SurveyRequiredFile(%s): %v", in, got)
		}
	}

	tcase("", true)
	tcase(" ", true)
	tcase("	", true)
	tcase(filepath.Join(os.TempDir(), "dsv-cli-test-file-does-not-exist"), true)

	f, err := os.CreateTemp("", "dsv-cli-test-file-*")
	if err != nil {
		t.Fatalf("os.CreateTemp(): %v", err)
	}
	tcase(f.Name(), false)
	err = os.Remove(f.Name())
	if err != nil {
		t.Fatalf("os.Remove(): %v", err)
	}
}

func TestSurveyRequiredPath(t *testing.T) {
	tcase := func(in string, wantError bool) {
		t.Helper()
		got := SurveyRequiredPath(in)
		if wantError && got == nil {
			t.Errorf("Expected error SurveyRequiredPath(%s), but <nil> returned", in)
		} else if !wantError && got != nil {
			t.Errorf("Unexpected error SurveyRequiredPath(%s): %v", in, got)
		}
	}

	tcase("", true)
	tcase(" ", true)
	tcase("	", true)
	tcase("::invalid:path", true)
	tcase("valid:path", false)
}

func TestSurveyRequiredName(t *testing.T) {
	tcase := func(in string, wantError bool) {
		t.Helper()
		got := SurveyRequiredName(in)
		if wantError && got == nil {
			t.Errorf("Expected error SurveyRequiredName(%s), but <nil> returned", in)
		} else if !wantError && got != nil {
			t.Errorf("Unexpected error SurveyRequiredName(%s): %v", in, got)
		}
	}

	tcase("", true)
	tcase(" ", true)
	tcase("	", true)
	tcase("&invalid:name", true)
	tcase("val1d_name", false)
}

func TestSurveyRequiredProfileName(t *testing.T) {
	existingProfiles := []string{"aa", "bb"}

	tcase := func(in string, wantError bool) {
		t.Helper()
		got := SurveyRequiredProfileName(existingProfiles)(in)
		if wantError && got == nil {
			t.Errorf("Expected error SurveyRequiredProfileName(%s), but <nil> returned", in)
		} else if !wantError && got != nil {
			t.Errorf("Unexpected error SurveyRequiredProfileName(%s): %v", in, got)
		}
	}

	tcase("", true)
	tcase(" ", true)
	tcase("	", true)
	tcase("invalid name", true)
	tcase("UpperCaseName", true)
	tcase("aa", true) // Already defined in existing profiles list.
	tcase("val1d_name", false)
}

func TestSurveyOptionalCIDR(t *testing.T) {
	tcase := func(in string, wantError bool) {
		t.Helper()
		got := SurveyOptionalCIDR(in)
		if wantError && got == nil {
			t.Errorf("Expected error SurveyOptionalCIDR(%s), but <nil> returned", in)
		} else if !wantError && got != nil {
			t.Errorf("Unexpected error SurveyOptionalCIDR(%s): %v", in, got)
		}
	}
	// Empty answer is allowed.
	tcase("", false)
	tcase(" ", false)
	tcase("	", false)

	// Valid addresses.
	tcase("10.10.10.0/24", false)
	tcase("  10.10.10.0/24   ", false)
	tcase("  10.10.10.0/24	", false)

	// Invalid addresses.
	tcase("1", true)
	tcase("1.2.3.4", true)
	tcase("1.2.3.4/99   ", true)
	tcase("  text      ", true)
}

func TestSurveyOptionalJSON(t *testing.T) {
	tcase := func(in string, wantError bool) {
		t.Helper()
		got := SurveyOptionalJSON(in)
		if wantError && got == nil {
			t.Errorf("Expected error SurveyOptionalJSON(%s), but <nil> returned", in)
		} else if !wantError && got != nil {
			t.Errorf("Unexpected error SurveyOptionalJSON(%s): %v", in, got)
		}
	}
	// Empty answer is allowed.
	tcase(``, false)
	tcase(` `, false)
	tcase(`	`, false)

	// Valid JSONs.
	tcase(`{"k":"v"}`, false)
	tcase(`  {"k":"v"}   `, false)
	tcase(`  {"k":"v"}	`, false)
	tcase(`  {"k": "v"}	`, false)
	tcase(`  {"k": ["v1", "v2"]}	`, false)

	// Invalid JSONs.
	tcase(`1`, true)
	tcase(`{"k"} `, true)
	tcase(`{"k"}   `, true)
	tcase(`  text      `, true)
}

func TestSurveyTrimSpace(t *testing.T) {
	tcase := func(in, out string) {
		t.Helper()
		got := SurveyTrimSpace(in)
		if got.(string) != out {
			t.Errorf("Unexpected result. Want: %s, got: %s", out, got.(string))
		}
	}
	tcase(" a", "a")
	tcase("  a", "a")
	tcase("  a ", "a")
	tcase("  a  ", "a")
	tcase(" a  ", "a")
	tcase("a  ", "a")
	tcase("	a", "a")
	tcase("	a	", "a")
	tcase("a	", "a")
}
