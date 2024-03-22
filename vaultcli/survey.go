package vaultcli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2/core"
)

// List of validation errors.
var (
	errValueRequired    = errors.New("Value is required.")
	errInvalidInteger   = errors.New("Please enter a valid integer.")
	errInvalidPort      = errors.New("Please enter a valid port number.")
	errFileNotFound     = errors.New("Cannot find file at given path.")
	errUppercaseProfile = errors.New("Profile name can only use lowercase letters.")
	errProfileExists    = errors.New("Profile with this name already exists in the config.")
	errAtLeastOne       = errors.New("Please select at least one item.")
)

// -----------------------------------------------------------------------//
// Common functions for Question objects from AlecAivazis/survey library. //
// -----------------------------------------------------------------------//

// SurveyRequired verifies that there is some answer.
// Built-in function "survey.Required()" does not trim spaces.
func SurveyRequired(ans any) error {
	answer := strings.TrimSpace(ans.(string))
	if len(answer) == 0 {
		return errValueRequired
	}
	return nil
}

// SurveyRequiredInt verifies that the answer is a valid integer number.
func SurveyRequiredInt(ans any) error {
	answer := strings.TrimSpace(ans.(string))
	_, err := strconv.Atoi(answer)
	if err != nil {
		return errInvalidInteger
	}
	return nil
}

// SurveyRequiredPortNumber verifies that the answer is a valid port number.
func SurveyRequiredPortNumber(ans any) error {
	answer := strings.TrimSpace(ans.(string))
	num, err := strconv.Atoi(answer)
	if err != nil || num > 65535 || num < 0 {
		return errInvalidPort
	}
	return nil
}

// SurveyRequiredFile verifies that the answer is a valid path to a file.
func SurveyRequiredFile(ans any) error {
	answer := strings.TrimSpace(ans.(string))
	if len(answer) == 0 {
		return errValueRequired
	}
	_, err := os.Stat(answer)
	if err != nil {
		return errFileNotFound
	}
	return nil
}

// SurveyRequiredPath checks path.
func SurveyRequiredPath(ans any) error {
	answer := strings.TrimSpace(ans.(string))
	if len(answer) == 0 {
		return errValueRequired
	}
	return ValidatePath(answer)
}

// SurveyRequiredName checks name.
func SurveyRequiredName(ans any) error {
	answer := strings.TrimSpace(ans.(string))
	if len(answer) == 0 {
		return errValueRequired
	}
	return ValidateName(answer)
}

// SurveyRequiredProfileName checks profile name.
func SurveyRequiredProfileName(existingProfiles []string) func(ans any) error {
	return func(ans any) error {
		answer := strings.TrimSpace(ans.(string))
		if len(answer) == 0 {
			return errValueRequired
		}
		lowered := strings.ToLower(answer)
		if lowered != answer {
			return errUppercaseProfile
		}
		err := ValidateProfile(answer)
		if err != nil {
			return err
		}
		for _, p := range existingProfiles {
			if answer == p {
				return errProfileExists
			}
		}
		return nil
	}
}

// SurveySelectAtLeastOne requires the answer is a list with at least one item.
func SurveySelectAtLeastOne(ans any) error {
	list, ok := ans.([]core.OptionAnswer)
	if !ok {
		return fmt.Errorf("unexpected type %T", ans)
	}
	if len(list) == 0 {
		return errAtLeastOne
	}
	return nil
}

// SurveyOptionalCIDR verifies that the answer is either empty or a valid CIDR.
func SurveyOptionalCIDR(ans any) error {
	answer := strings.TrimSpace(ans.(string))
	if len(answer) == 0 {
		// Answer is optional.
		return nil
	}
	if _, _, err := net.ParseCIDR(answer); err != nil {
		return err
	}
	return nil
}

// SurveyOptionalJSON verifies that the answer is either empty or a valid JSON.
func SurveyOptionalJSON(ans any) error {
	answer := strings.TrimSpace(ans.(string))
	if len(answer) == 0 {
		// Answer is optional.
		return nil
	}
	m := map[string]any{}
	err := json.Unmarshal([]byte(answer), &m)
	if err != nil {
		return fmt.Errorf("Invalid JSON: %v", err)
	}
	return nil
}

// SurveyTrimSpace trims spaces.
func SurveyTrimSpace(ans any) (newAns any) {
	return strings.TrimSpace(ans.(string))
}
