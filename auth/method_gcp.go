package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	cst "thy/constants"
	"thy/paths"

	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	iam "google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
)

func buildGcpParams(token string, gcpAuthType string) (*requestBody, error) {
	if token == "" {
		if gcpAuthType == "" {
			gcpAuthType = GcpGceAuth
		}
		gcp := GcpClient{}

		var err error
		token, err = gcp.GetJwtToken(gcpAuthType)
		if err != nil {
			return nil, err
		}
	}

	data := &requestBody{
		GrantType: authTypeToGrantType[FederatedGcp],
		Jwt:       token,
	}
	return data, nil
}

type GcpClient struct{}

const (
	GcpGceAuth               = "gce"
	GcpIamAuth               = "iam"
	gcpDefaultServiceAccName = "default"
	googleIssuer             = "https://accounts.google.com"
)

func (c *GcpClient) GetJwtToken(authType string) (string, error) {
	if authType != GcpGceAuth && authType != GcpIamAuth {
		return "", fmt.Errorf("invalid GCP auth type: %s", authType)
	}

	serviceAcctName := viper.GetString(cst.GcpServiceAccount)
	projectId := viper.GetString(cst.GcpProject)

	audience := GetAudience()
	var errPrimary, errSecondary error

	if authType == GcpGceAuth {
		metadataIdentityTemplate := "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/%s/identity?audience=%s&format=full"
		if serviceAcctName == "" {
			serviceAcctName = gcpDefaultServiceAccName
		}
		tokenRequestURL := fmt.Sprintf(metadataIdentityTemplate, serviceAcctName, audience)
		client := &http.Client{}
		if req, err := http.NewRequest(http.MethodGet, tokenRequestURL, nil); err != nil {
			errPrimary = err
		} else {
			req.Header.Add("Metadata-Flavor", "Google")
			if resp, err := client.Do(req); err != nil {
				errPrimary = err
			} else {
				return ParseMetadataIdentityResponse(resp)
			}
		}
	}
	if authType == GcpIamAuth || errPrimary != nil {
		// reset service account name
		serviceAcctName = viper.GetString(cst.GcpServiceAccount)
		if errPrimary != nil {
			log.Print("Failed auth with auth.gcp.type='gce'. Trying with auth.gcp.type='iam'")
		}
		ctx := context.Background()
		scopes := []string{iam.CloudPlatformScope}
		creds, err := google.FindDefaultCredentials(ctx, scopes...)
		if err != nil || creds == nil {
			return "", errors.New("unable to find default gcp credentials for iam authentication")
		}
		if projectId == "" {
			projectId = creds.ProjectID
		}
		var accountEmail string
		var accountType string
		credMeta := googleCredMeta{}
		if err := json.Unmarshal(creds.JSON, &credMeta); err == nil {
			accountEmail = credMeta.ClientEmail
			accountType = credMeta.Type
			if serviceAcctName == "" {
				serviceAcctName = accountEmail
			}
		}
		if serviceAcctName == "" {
			if accountEmail != "" {
				serviceAcctName = accountEmail
			} else {
				err = errors.New("Did not find service account identifier (email or uniqueId)")
			}
		}

		if err != nil {
			errSecondary = err
		} else {
			payload := map[string]interface{}{
				"iss":        googleIssuer,
				"aud":        audience,
				"sub":        serviceAcctName,
				"email":      accountEmail,
				"sub_type":   accountType,
				"project_id": projectId,
			}
			payloadMarshalled, _ := json.Marshal(payload)
			jwtRequest := iam.SignJwtRequest{
				Payload: string(payloadMarshalled),
			}

			oauthHttpClient := oauth2.NewClient(ctx, creds.TokenSource)
			iamService, err := iam.NewService(ctx, option.WithHTTPClient(oauthHttpClient))
			name := fmt.Sprintf("projects/%s/serviceAccounts/%s", projectId, serviceAcctName)
			resp, err := iamService.Projects.ServiceAccounts.SignJwt(name, &jwtRequest).Do()
			if err != nil {
				errSecondary = err
			} else {
				return resp.SignedJwt, nil
			}
		}
		log.Printf("gcp iam auth failed: %v", errSecondary)
	}
	if errPrimary != nil {
		return "", errPrimary
	}
	return "", errSecondary
}

func ParseMetadataIdentityResponse(resp *http.Response) (string, error) {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("Metadata identity request failed. Status: %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(respData), nil
}

func GetAudience() string {
	t := viper.GetString(cst.Tenant)
	d := paths.GetDomain()
	return fmt.Sprintf("https://%s.%s", t, d)
}

type googleCredMeta struct {
	Type         string `json:"type"`
	ProjectID    string `json:"project_id"`
	ClientEmail  string `json:"client_email"`
	AuthClientID string `json:"client_id"`
}
