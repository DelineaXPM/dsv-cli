package utils

import (
	"encoding/json"
	"errors"
	"strconv"
)

// getVersion tries to extract the version property in a JSON response.
func getVersion(resp []byte) (string, error) {
	var m map[string]interface{}
	err := json.Unmarshal(resp, &m)
	if err != nil {
		return "", err
	}
	version, ok := m["version"]
	if !ok {
		return "", errors.New("version not found")
	}
	v, ok := version.(string)
	if !ok {
		return "", errors.New("version is not a string")
	}
	return v, nil
}

// GetPreviousVersion tries to extract the version property in a JSON response and
// return the previous version that must be a non-negative integer.
func GetPreviousVersion(resp []byte) (string, error) {
	version, err := getVersion(resp)
	if err != nil {
		return "", err
	}
	v, err := strconv.Atoi(version)
	if err != nil {
		return "", err
	}

	v = v - 1
	if v < 0 {
		return "", errors.New("no previous version")
	}
	return strconv.Itoa(v), nil
}
