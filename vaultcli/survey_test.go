package vaultcli

import "testing"

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
