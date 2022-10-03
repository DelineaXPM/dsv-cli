package auth

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Azure/go-autorest/autorest"
	azure "github.com/Azure/go-autorest/autorest/azure/auth"
)

func buildAzureParams() (*requestBody, error) {
	resource := "https://management.azure.com/"
	authorizer, err := azure.NewAuthorizerFromEnvironmentWithResource(resource)
	if err != nil {
		return nil, fmt.Errorf("create authorizer: %w", err)
	}

	p := authorizer.WithAuthorization()

	r := &http.Request{}
	r, err = autorest.CreatePreparer(p).Prepare(r)
	if err != nil {
		return nil, fmt.Errorf("generate Azure auth token: %w", err)
	}

	qualifiedBearer := r.Header.Get("Authorization")
	lenPrefix := len("Bearer ")
	if len(qualifiedBearer) < lenPrefix {
		return nil, errors.New("received invalid bearer token")
	}
	bearer := qualifiedBearer[lenPrefix:]

	data := &requestBody{
		GrantType: authTypeToGrantType[FederatedAzure],
		Jwt:       bearer,
	}

	return data, nil
}
