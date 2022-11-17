package vault

import (
	"encoding/json"
	"log"
)

// rolesResource is the HTTP URL path component for the roles resource
const rolesResource = "roles"

// roleResource is composed with resourceMetadata to for RoleContents
type roleResource struct {
	Groups                     []string
	Name, Provider, ExternalID string
}

// Role holds the contents of a role from DSV
type Role struct {
	resourceMetadata
	roleResource
	vault Vault
}

// Role gets the role named name from the DSV of the given tenant
func (v Vault) Role(name string) (*Role, error) {
	role := &Role{vault: v}
	data, err := v.accessResource("GET", rolesResource, name, nil)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, role); err != nil {
		log.Printf("[DEBUG] error parsing response from /%s/%s: %q", rolesResource, name, data)
		return nil, err
	}
	return role, nil
}
