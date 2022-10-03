package pki

import (
	"encoding/base64"
	"encoding/pem"
	"testing"
)

func TestParseBase64EncodedPrivKeyPEM_inputMalformed(t *testing.T) {
	privKey := "malformed RSA private key"
	_, err := ParseBase64EncodedPrivKeyPEM(privKey)
	if err == nil {
		t.Fatal("expected error, but <nil> returned")
	}
}

func TestParseBase64EncodedPrivKeyPEM_inputBase64Malformed(t *testing.T) {
	privKey := []byte("malformed RSA private key bytes")
	base64PrivKey := base64.StdEncoding.EncodeToString(privKey)
	_, err := ParseBase64EncodedPrivKeyPEM(base64PrivKey)
	if err == nil {
		t.Fatal("expected error, but <nil> returned")
	}
}

func TestParseBase64EncodedPrivKeyPEM_inputBase64InvalidLabelPEM(t *testing.T) {
	privKey := genPrivateKey(t)
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY INVALID", Bytes: privKey})
	base64PrivKey := base64.StdEncoding.EncodeToString(pemBytes)

	_, err := ParseBase64EncodedPrivKeyPEM(base64PrivKey)
	if err == nil {
		t.Fatal("expected error, but <nil> returned")
	}
}

func TestParseBase64EncodedPrivKeyPEM_inputBase64InvalidKey(t *testing.T) {
	privKey := genPrivateKey(t)
	privKey = append(privKey, 1)
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privKey})
	base64PrivKey := base64.StdEncoding.EncodeToString(pemBytes)

	_, err := ParseBase64EncodedPrivKeyPEM(base64PrivKey)
	if err == nil {
		t.Fatal("expected error, but <nil> returned")
	}
}

func TestParseBase64EncodedPrivKeyPEM_successCase(t *testing.T) {
	privKey := genPrivateKey(t)
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privKey})
	base64PrivKey := base64.StdEncoding.EncodeToString(pemBytes)

	key, err := ParseBase64EncodedPrivKeyPEM(base64PrivKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key == nil {
		t.Fatal("unexpected <nil> key")
	}
}
