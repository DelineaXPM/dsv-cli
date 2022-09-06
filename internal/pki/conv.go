package pki

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
)

const (
	certPEMLabel    = "CERTIFICATE"
	privKeyPEMLabel = "RSA PRIVATE KEY"
)

var (
	errCertMalformed    = errors.New("certificate is malformed or its format is not supported (supported formats: PEM or DER, optionaly base64 encoded)")
	errPrivKeyMalformed = errors.New("private key is malformed or its format is not supported (supported formats: PEM or DER, optionaly base64 encoded)")
)

// CertToBase64EncodedPEM returns certificate in a base64 encoded PEM format.
func CertToBase64EncodedPEM(cert string) (string, error) {
	// Case 1. Given 'cert' is a valid PEM.
	block, _ := pem.Decode([]byte(cert))
	if block != nil {
		if block.Type != certPEMLabel {
			return "", fmt.Errorf("expected .pem encoded '%s', got .pem encoded '%s'", certPEMLabel, block.Type)
		}
		return base64.StdEncoding.EncodeToString([]byte(cert)), nil
	}

	// Case 2. Given 'cert' is a valid DER.
	derCert, err := x509.ParseCertificate([]byte(cert))
	if err == nil {
		out := pem.EncodeToMemory(&pem.Block{Type: certPEMLabel, Bytes: derCert.Raw})
		return base64.StdEncoding.EncodeToString(out), nil
	}

	// Case 3. Given 'cert' is a valid base64 encoded DER.
	content, err := base64.StdEncoding.DecodeString(cert)
	if err != nil {
		return "", errCertMalformed
	}
	derCert, err = x509.ParseCertificate(content)
	if err == nil {
		out := pem.EncodeToMemory(&pem.Block{Type: certPEMLabel, Bytes: derCert.Raw})
		return base64.StdEncoding.EncodeToString(out), nil
	}

	// Case 4. Given 'cert' is a valid base64 encoded PEM.
	block, _ = pem.Decode(content)
	if block != nil {
		if block.Type != certPEMLabel {
			return "", fmt.Errorf("expected .pem encoded '%s', got .pem encoded '%s'", certPEMLabel, block.Type)
		}
		return cert, nil
	}

	return "", errCertMalformed
}

// PrivateKeyToBase64EncodedPEM returns private key in a base64 encoded PEM format.
func PrivateKeyToBase64EncodedPEM(key string) (string, error) {
	// Case 1. Given 'key' is a valid PEM.
	block, _ := pem.Decode([]byte(key))
	if block != nil {
		if block.Type != privKeyPEMLabel {
			return "", fmt.Errorf("expected .pem encoded '%s', got .pem encoded '%s'", privKeyPEMLabel, block.Type)
		}
		return base64.StdEncoding.EncodeToString([]byte(key)), nil
	}

	// Case 2. Given 'key' is a valid DER.
	derKey, err := x509.ParsePKCS1PrivateKey([]byte(key))
	if err == nil {
		keyBytes := x509.MarshalPKCS1PrivateKey(derKey)
		out := pem.EncodeToMemory(&pem.Block{Type: privKeyPEMLabel, Bytes: keyBytes})
		return base64.StdEncoding.EncodeToString(out), nil
	}

	// Case 3. Given 'key' is a valid base64 encoded DER.
	content, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return "", errPrivKeyMalformed
	}
	derKey, err = x509.ParsePKCS1PrivateKey(content)
	if err == nil {
		keyBytes := x509.MarshalPKCS1PrivateKey(derKey)
		out := pem.EncodeToMemory(&pem.Block{Type: privKeyPEMLabel, Bytes: keyBytes})
		return base64.StdEncoding.EncodeToString(out), nil
	}

	// Case 4. Given 'key' is a valid base64 encoded PEM.
	block, _ = pem.Decode(content)
	if block != nil {
		if block.Type != privKeyPEMLabel {
			return "", fmt.Errorf("expected .pem encoded '%s', got .pem encoded '%s'", privKeyPEMLabel, block.Type)
		}
		return key, nil
	}

	return "", errPrivKeyMalformed
}
