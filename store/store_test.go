package store

import (
	"fmt"
	"os"
	"runtime"
	"testing"

	cst "thy/constants"
	"thy/errors"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

type tokenData struct {
	Token    []byte `json:"token"`
	Password []byte `json:"password"`
}

func TestStores(t *testing.T) {
	storeTypes := []string{File}

	isWindows := runtime.GOOS == "windows"
	if isWindows {
		storeTypes = append(storeTypes, WinCred)
	}

	// TODO : get pass installation working in ci-cd
	// isLinux := runtime.GOOS == "linux"
	// if isLinux {
	// 	storeTypes = append(storeTypes, store.PassLinux)
	// }

	for i, st := range storeTypes {
		t.Run(fmt.Sprintf("case=%d:(%s)", i, st), func(t *testing.T) {
			testStore(t, st)
		})
	}
}

func testStore(t *testing.T, storeType string) {
	t.Helper()

	if storeType == File {
		// Setup for file store to not use local user default path.
		viper.Set(cst.StorePath, "./testing-store-asb5a23afs3")
		defer os.Remove("./testing-store-asb5a23afs3")
	}

	// arrange
	s, err := GetStore(storeType)
	assert.Nil(t, err)

	obj := tokenData{
		Token:    []byte("GIyZDY5O"),
		Password: []byte("CIsImF0d"),
	}
	err = s.Store("token", obj)
	assert.Nil(t, err)

	// act
	var obj2 tokenData
	err = s.Get("token", &obj2)
	assert.Nil(t, err)
	assert.Equal(t, obj, obj2)
	_ = s.Delete("token")

	// assert
	obj3 := tokenData{}
	err = s.Get("token", &obj2)
	assert.Empty(t, obj3)

	// arrange
	err = s.Store("token", obj)
	assert.Nil(t, err)

	// act
	_ = s.Wipe("")

	//assert
	obj2 = tokenData{}
	err = s.Get("token", &obj2)
	assert.Empty(t, obj2)
}

func TestGetSecureSetting(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		profile       string
		expectedError error
		expectedVal   string
	}{
		{"missing-key", "", "some-profile", errors.NewS("key cannot be empty"), ""},
		{"empty-value", "hello", "some-profile", nil, ""},
		{"missing-profile", "hello", "", errors.NewS("profile cannot be empty"), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := getSecureSettingForProfile(tt.key, tt.profile)

			if tt.expectedError != nil {
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			}
			if err == nil {
				assert.Equal(t, tt.expectedVal, val)
			}
		})

	}
}
