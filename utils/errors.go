package utils

import (
	"strings"
	"thy/errors"
)

func NewMissingArgError(arg ...string) *errors.ApiError {
	flagified := []string{}
	for _, s := range arg {
		flagified = append(flagified, "--"+s)
	}
	joinedArgs := strings.Join(flagified, " or ")
	return errors.NewF("%s must be set", strings.ToLower(joinedArgs))
}

func GetExecStatus(err error) int {
	if err == nil || err.Error() == "" {
		return 0
	}
	return 1
}
