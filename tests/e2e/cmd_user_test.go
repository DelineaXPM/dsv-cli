//go:build endtoend
// +build endtoend

package e2e

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestUser(t *testing.T) {
	username := makeUserName()
	password := newPassword()

	output := runWithProfile(t, "user")
	requireContains(t, output, "Execute an action on a user")

	output = runWithProfile(t, fmt.Sprintf("user create --username %s --password %s", username, password))
	requireContains(t, output, `"displayName": ""`)
	requireContains(t, output, `"externalId": ""`)
	requireContains(t, output, fmt.Sprintf(`"userName": "%s"`, username))
	requireContains(t, output, `"provider": ""`)
	requireContains(t, output, `"version": "0"`)

	output = runWithProfile(t, fmt.Sprintf("user create --username %s --password %s", username, password))
	requireContains(t, output, `"code": 400`)
	requireContains(t, output, `"message": "a security principal with this name already exists"`)

	output = runWithProfile(t, fmt.Sprintf("user read --username %s", username))
	requireContains(t, output, `"displayName": ""`)
	requireContains(t, output, `"externalId": ""`)
	requireContains(t, output, fmt.Sprintf(`"userName": "%s"`, username))
	requireContains(t, output, `"provider": ""`)
	requireContains(t, output, `"version": "0"`)

	output = runWithProfile(t, fmt.Sprintf("user read %s", username))
	requireContains(t, output, `"displayName": ""`)
	requireContains(t, output, `"externalId": ""`)
	requireContains(t, output, fmt.Sprintf(`"userName": "%s"`, username))
	requireContains(t, output, `"provider": ""`)
	requireContains(t, output, `"version": "0"`)

	output = runWithProfile(t, fmt.Sprintf("user %s", username))
	requireContains(t, output, `"displayName": ""`)
	requireContains(t, output, `"externalId": ""`)
	requireContains(t, output, fmt.Sprintf(`"userName": "%s"`, username))
	requireContains(t, output, `"provider": ""`)
	requireContains(t, output, `"version": "0"`)

	output = runWithProfile(t, fmt.Sprintf("user search %s --limit 10", username[:len(username)-2]))
	requireContains(t, output, `"cursor"`)
	requireContains(t, output, `"data": [`)
	requireContains(t, output, `"displayName": ""`)
	requireContains(t, output, `"externalId": ""`)
	requireContains(t, output, fmt.Sprintf(`"userName": "%s"`, username))
	requireContains(t, output, `"provider": ""`)
	requireContains(t, output, `"version": "0"`)
	requireContains(t, output, `"length"`)
	requireContains(t, output, `"total"`)
	requireContains(t, output, `"limit": 10`)

	output = runWithProfile(t, fmt.Sprintf("user delete %s", username))
	requireContains(t, output, `marked for deletion and will be removed in about 72 hours`)

	output = runWithProfile(t, fmt.Sprintf("user delete %s", username))
	requireContains(t, output, `marked for deletion and will be removed in about 72 hours`)

	output = runWithProfile(t, fmt.Sprintf("user read %s", username))
	requireContains(t, output, `marked for deletion and will be removed in about 72 hours`)

	output = runWithProfile(t, fmt.Sprintf("user restore %s", username))
	requireEmpty(t, output)

	output = runWithProfile(t, fmt.Sprintf("user read %s", username))
	requireContains(t, output, `"displayName": ""`)
	requireContains(t, output, `"externalId": ""`)
	requireContains(t, output, fmt.Sprintf(`"userName": "%s"`, username))
	requireContains(t, output, `"provider": ""`)
	requireContains(t, output, `"version": "0"`)

	output = runWithProfile(t, fmt.Sprintf("user delete %s --force", username))
	requireEmpty(t, output)

	output = runWithProfile(t, "user create --username bob --external-id 1234")
	requireContains(t, output, "provider is required")

	output = runWithProfile(t, "user create --username bob --provider 4321")
	requireContains(t, output, "unable to find authentication provider")
}

// newPassword generates a 20 characters long string.
// Returned string will have at least one letter, one digit
// and one special character.
func newPassword() string {
	const (
		length   = 20
		digits   = "0123456789"
		specials = "~!@#$%^&*()"
		letters  = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
		all      = letters + digits + specials
	)

	rand.Seed(time.Now().UnixNano())

	buf := make([]byte, length)
	buf[0] = letters[rand.Intn(len(letters))]
	buf[1] = letters[rand.Intn(len(letters))]
	buf[2] = digits[rand.Intn(len(digits))]
	buf[3] = specials[rand.Intn(len(specials))]
	for i := 4; i < length; i++ {
		buf[i] = all[rand.Intn(len(all))]
	}
	rand.Shuffle(len(buf), func(i, j int) {
		buf[i], buf[j] = buf[j], buf[i]
	})

	return string(buf)
}
