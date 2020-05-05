package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncrypt(t *testing.T) {
	key, _ := GetEncryptionKey("")
	s1, s2, err := Encrypt(string(key), "data")
	assert.Nil(t, err)
	assert.NotEmpty(t, s1, s2)

	st, err := Decrypt(s1, s2)
	assert.Nil(t, err)

	assert.Equal(t, st, "data")
}

func TestEncipherPassword(t *testing.T) {
	_, err := EncipherPassword("password")
	assert.NotNil(t, err)
}
