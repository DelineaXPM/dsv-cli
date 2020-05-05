package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	mrand "math/rand"
	"time"
)

const leafCommonName = "example.com"

func generateRootWithPrivateKey() ([]byte, []byte, error) {
	privateKey, publicKey := generateRSAKeyPair(2048)
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(int64(mrand.Int31n(big.MaxExp))),
		Subject: pkix.Name{
			CommonName:         "thycotic.com",
			Country:            []string{"US"},
			Province:           []string{"DC"},
			Locality:           []string{"Washington"},
			Organization:       []string{"Thycotic"},
			OrganizationalUnit: []string{"Software"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Duration(10000) * time.Hour),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, publicKey, privateKey)
	if err != nil {
		return nil, nil, err
	}

	caPEM := &bytes.Buffer{}
	err = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	if err != nil {
		return nil, nil, err
	}
	return caPEM.Bytes(), marshalRsaPrivateKeyToPem(privateKey), nil
}

func generateCSR() ([]byte, error) {
	keyBytes, _ := rsa.GenerateKey(rand.Reader, 1024)

	subj := pkix.Name{
		CommonName:         leafCommonName,
		Country:            []string{"US"},
		Province:           []string{"NY"},
		Locality:           []string{"New York"},
		Organization:       []string{"Company Ltd"},
		OrganizationalUnit: []string{"IT"},
	}

	template := x509.CertificateRequest{
		Subject:            subj,
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, keyBytes)
	if err != nil {
		return nil, err
	}
	csr := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})
	return csr, nil
}

func generateRSAKeyPair(keySize int) (*rsa.PrivateKey, *rsa.PublicKey) {
	privKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		panic(err)
	}
	return privKey, &privKey.PublicKey
}

// marshalRsaPrivateKeyToPem converts a private key to a PEM formatted byte slice.
func marshalRsaPrivateKeyToPem(privKey *rsa.PrivateKey) []byte {
	privKeyBytes := x509.MarshalPKCS1PrivateKey(privKey)
	privKeyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privKeyBytes,
		},
	)
	return privKeyPem
}
