package pki

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
)

func ParseBase64EncodedPrivKeyPEM(privKey string) (*rsa.PrivateKey, error) {
	pemEncoded, err := base64.StdEncoding.DecodeString(privKey)
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}

	block, _ := pem.Decode(pemEncoded)
	if block == nil {
		return nil, errors.New("invalid PEM format")
	}

	if block.Type != privKeyPEMLabel {
		return nil, fmt.Errorf("expected PEM encoded '%s', got PEM encoded '%s'", privKeyPEMLabel, block.Type)
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}
