package pki

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"testing"
	"time"
)

func TestCertToBase64EncodedPEM_inputMalformed(t *testing.T) {
	cert := []byte("malformed certificate bytes")
	_, err := CertToBase64EncodedPEM(string(cert))
	if err == nil {
		t.Fatal("expected error, but <nil> returned")
	}
}

func TestCertToBase64EncodedPEM_inputBase64Malformed(t *testing.T) {
	cert := []byte("malformed certificate bytes")
	base64Cert := base64.StdEncoding.EncodeToString(cert)
	_, err := CertToBase64EncodedPEM(string(base64Cert))
	if err == nil {
		t.Fatal("expected error, but <nil> returned")
	}
}

func TestCertToBase64EncodedPEM_inputDER(t *testing.T) {
	cert := genCertificate(t)
	_, err := CertToBase64EncodedPEM(string(cert))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCertToBase64EncodedPEM_inputPEM(t *testing.T) {
	cert := genCertificate(t)
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert})

	_, err := CertToBase64EncodedPEM(string(pemBytes))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCertToBase64EncodedPEM_inputInvalidLabelPEM(t *testing.T) {
	cert := genCertificate(t)
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE INVALID", Bytes: cert})

	_, err := CertToBase64EncodedPEM(string(pemBytes))
	if err == nil {
		t.Fatal("expected error, but <nil> returned")
	}
}

func TestCertToBase64EncodedPEM_inputBase64DER(t *testing.T) {
	cert := genCertificate(t)
	base64Cert := base64.StdEncoding.EncodeToString(cert)

	_, err := CertToBase64EncodedPEM(string(base64Cert))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCertToBase64EncodedPEM_inputBase64PEM(t *testing.T) {
	cert := genCertificate(t)
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert})
	base64Cert := base64.StdEncoding.EncodeToString(pemBytes)

	_, err := CertToBase64EncodedPEM(string(base64Cert))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCertToBase64EncodedPEM_inputBase64InvalidLabelPEM(t *testing.T) {
	cert := genCertificate(t)
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE INVALID", Bytes: cert})
	base64Cert := base64.StdEncoding.EncodeToString(pemBytes)

	_, err := CertToBase64EncodedPEM(string(base64Cert))
	if err == nil {
		t.Fatal("expected error, but <nil> returned")
	}
}

func TestPrivateKeyToBase64EncodedPEM_inputMalformed(t *testing.T) {
	privKey := []byte("malformed RSA private key bytes")
	_, err := PrivateKeyToBase64EncodedPEM(string(privKey))
	if err == nil {
		t.Fatal("expected error, but <nil> returned")
	}
}

func TestPrivateKeyToBase64EncodedPEM_inputBase64Malformed(t *testing.T) {
	privKey := []byte("malformed RSA private key bytes")
	base64PrivKey := base64.StdEncoding.EncodeToString(privKey)
	_, err := PrivateKeyToBase64EncodedPEM(string(base64PrivKey))
	if err == nil {
		t.Fatal("expected error, but <nil> returned")
	}
}

func TestPrivateKeyToBase64EncodedPEM_inputDER(t *testing.T) {
	privKey := genPrivateKey(t)
	_, err := PrivateKeyToBase64EncodedPEM(string(privKey))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPrivateKeyToBase64EncodedPEM_inputPEM(t *testing.T) {
	privKey := genPrivateKey(t)
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privKey})

	_, err := PrivateKeyToBase64EncodedPEM(string(pemBytes))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPrivateKeyToBase64EncodedPEM_inputInvalidLabelPEM(t *testing.T) {
	privKey := genPrivateKey(t)
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY INVALID", Bytes: privKey})

	_, err := PrivateKeyToBase64EncodedPEM(string(pemBytes))
	if err == nil {
		t.Fatal("expected error, but <nil> returned")
	}
}

func TestPrivateKeyToBase64EncodedPEM_inputBase64DER(t *testing.T) {
	privKey := genPrivateKey(t)
	base64PrivKey := base64.StdEncoding.EncodeToString(privKey)

	_, err := PrivateKeyToBase64EncodedPEM(string(base64PrivKey))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPrivateKeyToBase64EncodedPEM_inputBase64PEM(t *testing.T) {
	privKey := genPrivateKey(t)
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privKey})
	base64PrivKey := base64.StdEncoding.EncodeToString(pemBytes)

	_, err := PrivateKeyToBase64EncodedPEM(string(base64PrivKey))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPrivateKeyToBase64EncodedPEM_inputBase64InvalidLabelPEM(t *testing.T) {
	privKey := genPrivateKey(t)
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY INVALID", Bytes: privKey})
	base64PrivKey := base64.StdEncoding.EncodeToString(pemBytes)

	_, err := PrivateKeyToBase64EncodedPEM(string(base64PrivKey))
	if err == nil {
		t.Fatal("expected error, but <nil> returned")
	}
}

func genPrivateKey(t *testing.T) []byte {
	t.Helper()
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("could not generate RSA key: %v", err)
	}
	return x509.MarshalPKCS1PrivateKey(privKey)
}

func genCertificate(t *testing.T) []byte {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("could not generate RSA key: %v", err)
	}
	pub := priv.Public()

	tpl := &x509.Certificate{
		SerialNumber: big.NewInt(123),
		Subject: pkix.Name{
			Organization: []string{"dsv testing certificate"},
		},
		NotBefore: time.Now(), NotAfter: time.Now().AddDate(0, 0, 1),
	}
	cert, err := x509.CreateCertificate(rand.Reader, tpl, tpl, pub, priv)
	if err != nil {
		t.Fatalf("could not create x509 certificate: %v", err)
	}
	return cert
}
