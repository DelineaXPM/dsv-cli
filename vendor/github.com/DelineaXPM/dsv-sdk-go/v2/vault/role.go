package vault

import (
	"encoding/json"
	"log"
	"net/http"
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
}

// Role gets the role named name from the DSV of the given tenant
func (v Vault) Role(name string) (*Role, error) {
	data, err := v.accessResource(http.MethodGet, rolesResource, name, nil)
	if err != nil {
		return nil, err
	}

	role := &Role{}
	if err := json.Unmarshal(data, role); err != nil {
		log.Printf("[DEBUG] error parsing response from /%s/%s: %q", rolesResource, name, data)
		return nil, err
	}
	return role, nil
}
