package vaultcli

import (
	"errors"
	"regexp"
)

func ValidatePath(resource string) error {
	if !regexp.MustCompile(`^[a-zA-Z0-9:\/@\+._-]+$`).MatchString(resource) {
		return errors.New("path may contain only letters, numbers, underscores, dashes, @, pluses and periods separated by colon or slash")
	}
	return nil
}

func ValidateName(name string) error {
	if name == "" {
		return errors.New("must specify name")
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(name) {
		return errors.New("name may contain only letters, numbers, underscores and dashes")
	}
	return nil
}

func ValidateUsername(username string) error {
	if username == "" {
		return errors.New("must specify username")
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9@\+._-]+$`).MatchString(username) {
		return errors.New("name may contain only letters, numbers, underscores, dashes, @, pluses and periods")
	}
	return nil
}
