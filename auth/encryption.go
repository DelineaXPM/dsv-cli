package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/spf13/viper"

	cst "github.com/DelineaXPM/dsv-cli/constants"
	"github.com/DelineaXPM/dsv-cli/errors"
	"github.com/DelineaXPM/dsv-cli/store"
)

// StorePassword takes a fileName in which it tries to find an encryption key. It also receives data to encrypt (password).
// It returns the encrypted data, key for later decryption, and any error that might have occurred.
func StorePassword(fileName, data string) (string, string, error) {
	key, err := GetEncryptionKey(fileName)
	if err != nil {
		return "", "", err
	}
	cipherText, k, err := Encrypt(string(key), data)
	return string(cipherText), string(k), err
}

// Encrypt returns a cipher text encrypted with AES-256, a key to decrypt, and any error that might have occurred.
func Encrypt(key, data string) (string, string, error) {
	block, _ := aes.NewCipher([]byte(key))
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", "", err
	}
	// Encode to base64 to make it prettier than literal binary data.
	cipherText := base64.StdEncoding.EncodeToString(gcm.Seal(nonce, nonce, []byte(data), nil))
	return string(cipherText), string(key), nil
}

// GetEncryptionKey attempts to fetch and return the encryption key stored in fileName.
// If it does not find the key, it generates and returns a slice of random bytes as a new encryption key.
func GetEncryptionKey(fileName string) ([]byte, error) {
	if s, err := store.ReadFileInDefaultPath(fileName); err == nil {
		return []byte(s), nil
	}
	return generateRandomBytes(32)
}

// Decrypt takes encrypted data and the key and attempts to decrypt the data back into plain text.
func Decrypt(data, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	data = string(decoded)
	nonce, cipherText := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, []byte(nonce), []byte(cipherText), nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// EncipherPassword takes in a plaintext password and returns the encrypted version of it. This is a higher-level function
// that looks up an encryption key found in the default path for tokens and key files. It then tries to encrypt the password
// using the encryption key, which must exist.
func EncipherPassword(plaintext string) (string, error) {
	userName := viper.GetString(cst.Username)
	keyPath := GetEncryptionKeyFilename(viper.GetString(cst.Tenant), userName)
	key, err := store.ReadFileInDefaultPath(keyPath)
	if err != nil || key == "" {
		return "", KeyfileNotFoundError
	}
	cipherText, _, err := Encrypt(key, plaintext)
	if err != nil {
		return "", errors.NewS("Failed to encrypt the password with key.")
	}
	return cipherText, nil
}

// GetEncryptionKeyFilename creates and returns a filename for an encryption key given the tenant name and user name.
func GetEncryptionKeyFilename(tenant string, user string) string {
	return fmt.Sprintf("%s-%s-%s", cst.EncryptionKey, tenant, user)
}
