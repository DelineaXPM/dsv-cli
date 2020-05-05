package test_helpers

import (
	"encoding/json"
	"log"
	"os"
)

type GcpCreds struct {
	Type         string `json:"type"`
	ClientEmail  string `json:"client_email"`
	ClientId     string `json:"client_id"`
	PrivateKeyId string `json:"private_key_id"`
	PrivateKey   string `json:"private_key"`
	ProjectId    string `json:"project_id"`
	TokenURI     string `json:"token_uri"`
}

func GetGcpCreds() *GcpCreds {
	credsJson := os.Getenv("T_GOOGLE_CREDS")
	if credsJson == "" {
		return nil
	}
	gcpCreds := &GcpCreds{}
	if err := json.Unmarshal([]byte(credsJson), gcpCreds); err != nil {
		log.Fatal("unable to marshal google test creds")
	}
	return gcpCreds
}
