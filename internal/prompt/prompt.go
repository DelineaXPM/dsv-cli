package prompt

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/thycotic-rd/cli"
)

type Option struct {
	Value   string
	Display string
}

func YesNo(ui cli.Ui, question string, defaultToYes bool) (bool, error) {
	suffix := " [y/N]"
	if defaultToYes {
		suffix = " [Y/n]"
	}
	response, err := ui.Ask(question + suffix)
	if err != nil {
		return false, err
	}
	switch strings.ToLower(response) {
	case "y", "yes", "t", "true":
		return true, nil
	case "n", "no", "f", "false":
		return false, nil
	case "":
		return defaultToYes, nil
	}
	ui.Error("Invalid response, must choose (y)es or (n)o")
	return YesNo(ui, question, defaultToYes)
}

func Ask(ui cli.Ui, question string) (string, error) {
	resp, err := ui.Ask(question)
	if err != nil {
		return "", err
	}
	if resp == "" {
		ui.Error("Blank input is invalid")
		return Ask(ui, question)
	}
	return resp, nil
}

func AskDefault(ui cli.Ui, question, defaultAnswer string) (string, error) {
	question = strings.TrimSuffix(question, ":")
	if defaultAnswer != "" {
		question += fmt.Sprintf(" (default:%s):", defaultAnswer)
	} else {
		question += " (optional):"
	}
	answer, err := ui.Ask(question)
	if err != nil {
		return "", err
	}
	if answer == "" {
		return defaultAnswer, nil
	}
	return answer, nil
}

func AskSecure(ui cli.Ui, question string) (string, error) {
	resp, err := ui.AskSecret(question)
	if err != nil {
		return "", err
	}
	if resp == "" {
		ui.Error("Blank input is invalid")
		return AskSecure(ui, question)
	}
	return resp, nil
}

func AskSecureConfirm(ui cli.Ui, question string) (string, error) {
	resp, err := AskSecure(ui, question)
	if err != nil {
		return "", err
	}
	if resp2, err := ui.AskSecret(strings.Trim(question, ":") + " (confirm):"); err != nil {
		return "", err
	} else if resp2 != resp {
		ui.Error("Inputs do not match. Please retry.")
		return AskSecureConfirm(ui, question)
	}
	return resp, nil
}

func Choose(ui cli.Ui, question string, opt1 Option, options ...Option) (string, error) {
	allOptions := append([]Option{opt1}, options...)
	tmpQuestion := getQuestionWithOptions(question, allOptions)
	resp, err := ui.Ask(tmpQuestion)
	if err != nil {
		return "", err
	}
	if resp == "" {
		return allOptions[0].Value, nil
	}

	respInt, err := strconv.ParseUint(resp, 10, 32)
	if optLen := uint64(len(allOptions)); err != nil || respInt > optLen || respInt < 1 {
		if err != nil {
			log.Println(err)
		}
		ui.Error(fmt.Sprintf("Invalid option. Please select between '1' and '%d' or leave it empty to use default one", optLen))
		return Choose(ui, question, opt1, options...)
	}
	return allOptions[respInt-1].Value, nil
}

func getQuestionWithOptions(question string, options []Option) string {
	result := question
	for i, o := range options {
		if i == 0 {
			result += fmt.Sprintf("\n\t(%d) %s (default)", i+1, o.Display)
		} else {
			result += fmt.Sprintf("\n\t(%d) %s", i+1, o.Display)
		}
	}
	result += "\nSelection: "
	return result
}
