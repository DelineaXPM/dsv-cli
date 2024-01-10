package vault

import (
	"encoding/json"
	"log"
	"net/http"
)

// secretsResource is the HTTP URL path component for the secrets resource
const secretsResource = "secrets"

// secretResource is composed with resourceMetadata to for SecretContents
type secretResource struct {
	Attributes map[string]interface{}
	Data       map[string]interface{}
	Path       string
}

// Secret holds the contents of a secret from DSV
type Secret struct {
	resourceMetadata
	secretResource
}

// Secret [gets the secret] at path from the DSV of the given tenant.
//
// [gets the secret]: https://dsv.secretsvaultcloud.com/api#operation/getSecret
func (v Vault) Secret(path string) (*Secret, error) {
	data, err := v.accessResource(http.MethodGet, secretsResource, path, nil)
	if err != nil {
		return nil, err
	}

	secret := &Secret{}
	if err := json.Unmarshal(data, secret); err != nil {
		log.Printf("[DEBUG] error parsing response from /%s/%s: %q", secretsResource, path, data)
		return nil, err
	}
	return secret, nil
}

// DeleteSecret [deletes the secret] at path from the DSV of the given tenant.
//
// [deletes the secret]: https://dsv.secretsvaultcloud.com/api#operation/deleteSecret
func (v Vault) DeleteSecret(path string) error {
	_, err := v.accessResource(http.MethodDelete, secretsResource, path, nil)
	return err
}

// SecretCreateRequest represents the request body of the CreateSecret operation.
type SecretCreateRequest struct {
	Attributes  map[string]interface{} `json:"attributes"`
	Data        map[string]interface{} `json:"data"`
	Description string                 `json:"description"`
}

// CreateSecret [creates the secret] at path in the DSV of the given tenant.
//
// [creates the secret]: https://dsv.secretsvaultcloud.com/api#operation/createSecret
func (v Vault) CreateSecret(path string, req *SecretCreateRequest) (*Secret, error) {
	d, err := v.accessResource(http.MethodPost, secretsResource, path, req)
	if err != nil {
		return nil, err
	}

	secret := &Secret{}
	if err := json.Unmarshal(d, secret); err != nil {
		log.Printf("[DEBUG] error parsing response from /%s/%s: %q", secretsResource, path, d)
		return nil, err
	}
	return secret, nil
}
