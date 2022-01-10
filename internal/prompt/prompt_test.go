package prompt

import (
	"testing"

	"github.com/thycotic-rd/cli"
)

type cliUiMock struct {
	answer string
}

func (m *cliUiMock) Ask(string) (string, error)       { return m.answer, nil }
func (m *cliUiMock) AskSecret(string) (string, error) { return m.answer, nil }
func (m *cliUiMock) Output(string)                    {}
func (m *cliUiMock) Info(string)                      {}
func (m *cliUiMock) Error(string)                     {}
func (m *cliUiMock) Warn(string)                      {}

func TestPromptYesNo(t *testing.T) {
	testCase := func(answer string, result, defaultYes bool) {
		t.Helper()
		var mock cli.Ui = &cliUiMock{answer}
		testResult, err := YesNo(mock, "question", defaultYes)
		if err != nil {
			t.Error(err)
		}
		if testResult != result {
			t.Errorf("unexpected %t got %t; for answer '%s'", result, testResult, answer)
		}
	}

	testCase("yes", true, false)
	testCase("YES", true, false)
	testCase("t", true, false)
	testCase("TRUe", true, false)
	testCase("T", true, true)
	testCase("NO", false, false)
	testCase("n", false, false)
	testCase("false", false, false)
	testCase("", false, false)
	testCase("", true, true)
	testCase("F", false, false)
}

func TestAsk(t *testing.T) {
	testCase := func(answer string) {
		t.Helper()
		var mock cli.Ui = &cliUiMock{answer}
		testResult, err := Ask(mock, answer)
		if err != nil {
			t.Error(err)
		}
		if testResult != answer {
			t.Errorf("unexpected %s got %s; for answer '%s'", answer, testResult, answer)
		}
	}

	testCase("yes")
	testCase("TeSt")
}

func TestAskDefault(t *testing.T) {
	testCase := func(answer string) {
		t.Helper()
		var mock cli.Ui = &cliUiMock{answer}
		testResult, err := AskDefault(mock, answer, answer)
		if err != nil {
			t.Error(err)
		}
		if testResult != answer {
			t.Errorf("unexpected %s got %s; for answer '%s'", answer, testResult, answer)
		}
	}

	testCase("yes")
	testCase("TeSt")
	testCase("")
}

func TestAskSecureConfirm(t *testing.T) {
	testCase := func(answer string) {
		t.Helper()
		var mock cli.Ui = &cliUiMock{answer}
		testResult, err := AskSecureConfirm(mock, answer)
		if err != nil {
			t.Error(err)
		}
		if testResult != answer {
			t.Errorf("unexpected %s got %s; for answer '%s'", answer, testResult, answer)
		}
	}

	testCase("yes")
	testCase("TeSt")
}

func TestChoose(t *testing.T) {
	testCase := func(opt, result string, o1 Option, options ...Option) {
		t.Helper()
		var mock cli.Ui = &cliUiMock{opt}
		testResult, err := Choose(mock, "question", o1, options...)
		if err != nil {
			t.Error(err)
		}
		if testResult != result {
			t.Errorf("unexpected value '%s' for option '%s'", testResult, opt)
		}
	}

	options := []Option{
		{"1", "test1"},
		{"2", "test2"},
		{"3", "TEST"},
		{"4", ""},
		{"5", ""},
	}

	testCase("", "1", options[0], options[1], options[2], options[3])
	testCase("1", "1", options[0], options[1])
	testCase("2", "2", options[0], options[1])
	testCase("3", "3", options[0], options[1], options[2])
	testCase("4", "4", options[0], options[1:]...)
}
