package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/DelineaXPM/dsv-cli/internal/pki"
	"github.com/DelineaXPM/dsv-cli/paths"
)

func (a *authenticator) buildCertParams(cert string, privateKey string) (*requestBody, error) {
	challengeID, challenge, err := a.initiateCertAuth(cert, privateKey)
	if err != nil {
		return nil, err
	}
	return &requestBody{
		GrantType:          authTypeToGrantType[Certificate],
		CertChallengeID:    challengeID,
		DecryptedChallenge: challenge,
	}, nil
}

// initiateCertAuth makes initial request and prepares info for final token request.
func (a *authenticator) initiateCertAuth(cert, privKey string) (string, string, error) {
	log.Println("Reading private key.")
	privateKey, err := pki.ParseBase64EncodedPrivKeyPEM(privKey)
	if err != nil {
		return "", "", fmt.Errorf("unable to parse private key: %w", err)
	}

	request := struct {
		Cert string `json:"client_certificate"`
	}{
		Cert: cert,
	}
	response := struct {
		ID        string `json:"cert_challenge_id"`
		Encrypted string `json:"encrypted"`
	}{}

	log.Println("Requesting challenge for certificate authentication.")
	uri := paths.CreateURI("certificate/auth", nil)
	requestErr := a.requestClient.DoRequestOut(http.MethodPost, uri, request, &response)
	if requestErr != nil {
		return "", "", errors.New(requestErr.Error())
	}
	encrypted, err := base64.StdEncoding.DecodeString(response.Encrypted)
	if err != nil {
		return "", "", fmt.Errorf("unable to read challenge: %w", err)
	}

	log.Println("Decrypting challenge using private key.")
	plaintext, err := rsa.DecryptOAEP(sha512.New(), rand.Reader, privateKey, encrypted, nil)
	if err != nil {
		return "", "", fmt.Errorf("unable to decrypt challenge: %w", err)
	}

	decrypted := base64.StdEncoding.EncodeToString(plaintext)
	return response.ID, decrypted, nil
}
