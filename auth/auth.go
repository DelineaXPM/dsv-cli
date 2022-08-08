package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"reflect"
	"strings"
	"text/template"
	"time"

	cst "thy/constants"
	"thy/errors"
	"thy/paths"
	"thy/requests"
	"thy/store"
	"thy/utils"

	"github.com/Azure/go-autorest/autorest"
	azure "github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/mitchellh/cli"
	"github.com/pkg/browser"
	"github.com/spf13/viper"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o ../tests/fake/fake_authenticator.go . Authenticator

const (
	leewaySecondsTokenExp   = 10
	refreshTokenLifeSeconds = 60 * 60 * 720
)

// Note that this global error variable is of type *ApiError, not regular error.
var KeyfileNotFoundError = errors.NewS("failed to find the encryption key")

// AuthType is the type of authentication
type AuthType string

// Types of supported authentication
const (
	Implicit         = AuthType("")
	Password         = AuthType("password")
	Refresh          = AuthType("refresh")
	ClientCredential = AuthType("clientcred")
	FederatedThyOne  = AuthType(cst.DefaultThyOneName)
	Certificate      = AuthType("cert")
	FederatedAws     = AuthType("aws")
	FederatedAzure   = AuthType("azure")
	FederatedGcp     = AuthType("gcp")
	Oidc             = AuthType("oidc")
)

// GetTokenKey gets the key for the auth type
func (a AuthType) GetTokenKey(tenant string, keySuffix string) string {
	if a == Implicit {
		panic("implicit is invalid auth type")
	}

	key := cst.TokenRoot + "-"
	if a == Refresh {
		// Refresh and Password are stored the same
		key = key + string(Password)
	} else {
		key = key + string(a)
	}

	key += "-" + tenant

	if keySuffix != "" {
		keySuffix = strings.ReplaceAll(keySuffix, "-", "%2D")
		key = key + "-" + keySuffix
	}
	return key
}

// Authenticator is the interface used for authentication funcs
type Authenticator interface {
	GetToken() (*TokenResponse, *errors.ApiError)
	GetTokenCacheOverride(authType string, useCache bool) (*TokenResponse, *errors.ApiError)
}

type authenticator struct {
	store         store.Store
	requestClient requests.Client
}

// NewAuthenticatorDefault gets a new default authenticator.
func NewAuthenticatorDefault() Authenticator {
	st := viper.GetString(cst.StoreType)
	if s, err := store.GetStore(st); err != nil {
		panic(err)
	} else {
		return &authenticator{s, requests.NewHttpClient()}
	}
}

// NewAuthenticator returns a new authenticator
func NewAuthenticator(store store.Store, client requests.Client) Authenticator {
	return &authenticator{store, client}
}

func (a *authenticator) GetToken() (*TokenResponse, *errors.ApiError) {
	authType := viper.GetString(cst.AuthType)

	if AuthType(authType) == ClientCredential && (strings.TrimSpace(viper.GetString(cst.AuthClientID)) != "" || strings.TrimSpace(viper.GetString(cst.AuthClientSecret)) != "") {
		return a.GetTokenCacheOverride(authType, false)
	}
	if AuthType(authType) == Password && strings.TrimSpace(viper.GetString(cst.Password)) != "" {
		return a.GetTokenCacheOverride(authType, false)
	}
	return a.GetTokenCacheOverride(authType, true)
}

func (a *authenticator) GetTokenCacheOverride(authType string, useCache bool) (*TokenResponse, *errors.ApiError) {
	if authType == "" {
		return a.getTokenForAuthType(Password, useCache)
	}
	return a.getTokenForAuthType(AuthType(authType), useCache)
}

func (a *authenticator) getTokenForAuthType(at AuthType, useCache bool) (*TokenResponse, *errors.ApiError) {
	var data requestBody
	var tr *TokenResponse

	pSpecs := paramSpecDict[at]
	var keyName string
	for _, p := range pSpecs {
		if p.IsKey {
			keyName = p.ArgName
			break
		}
	}
	if keyName == "" {
		return nil, errors.NewF("failure to auth. Token cache key not found for authentication type %s\n", string(at))
	}

	// Else, viper prepends env keys with cli prefix
	var keySuffix string
	if at == FederatedAzure {
		keySuffix = os.Getenv(keyName)
	} else if at == FederatedGcp {
		keySuffix = viper.GetString(cst.GcpServiceAccount)
		if keySuffix == "" {
			keySuffix = cst.DefaultProfile
		}
	} else if at == Oidc || at == FederatedThyOne {
		profile := viper.GetString(cst.Profile)
		keySuffix = viper.GetString(keyName)
		if profile != "" && profile != cst.DefaultProfile {
			keySuffix = fmt.Sprintf("%s-%s", keySuffix, profile)
		}
	} else if at == Certificate {
		keySuffix = viper.GetString(cst.Profile)
	} else {
		keySuffix = viper.GetString(keyName)
	}

	tenant := viper.GetString(cst.Tenant)
	keyToken := at.GetTokenKey(tenant, keySuffix)

	if useCache {
		if err := a.store.Get(keyToken, &tr); err != nil {
			return nil, err
		} else if tr != nil && !tr.IsNil() {
			// If init (cli-config) or config, invalidate existing token and later re-authenticate.
			if cmd := viper.GetString(cst.MainCommand); cmd == cst.NounCliConfig {
				err := a.store.Wipe(keyToken)
				if err != nil {
					return nil, err
				}
			} else {
				if tr.SecondsRemainingToken() > 0 {
					return tr, nil
				} else if tr.SecondsRemainingRefreshToken() > 0 {
					log.Printf("Token expired but valid refresh token. Attempting to refresh. Token cache key: '%s'\n", keySuffix)
					at = Refresh
					data = requestBody{
						GrantType:    authTypeToGrantType[at],
						RefreshToken: tr.RefreshToken,
					}

				} else {
					log.Printf("Refresh token expired. Attempting to reauthenticate. Token cache key: '%s'\n", keySuffix)
				}
			}
		}
	}
	if at == FederatedAws {
		if headers, body, err := buildAwsParams(); err != nil {
			return nil, err
		} else {
			data = requestBody{
				GrantType:  authTypeToGrantType[at],
				AwsBody:    body,
				AwsHeaders: headers,
			}
		}
	}
	if at == FederatedAzure {
		// NOTE : NH - this is a little awkward but better than splitting out token
		// provider factory (logic currently done by authorizer)
		resource := "https://management.azure.com/"
		authorizer, err := azure.NewAuthorizerFromEnvironmentWithResource(resource)
		if err != nil {
			return nil, errors.New(err).Grow("Failed to create azure authorizer")
		}
		r := http.Request{}
		p := authorizer.WithAuthorization()
		if r, err := autorest.CreatePreparer(p).Prepare(&r); err != nil {
			return nil, errors.New(err).Grow("Failed to generate azure auth token")
		} else {
			qualifiedBearer := r.Header.Get("Authorization")
			lenPrefix := len("Bearer ")
			if len(qualifiedBearer) < lenPrefix {
				return nil, errors.NewS("Received invalid bearer token")
			}
			bearer := qualifiedBearer[lenPrefix:]
			data = requestBody{
				GrantType: authTypeToGrantType[at],
				Jwt:       bearer,
			}
		}
	}

	if at == FederatedGcp {
		var token string
		var err error
		token = viper.GetString(cst.GcpToken)
		if token == "" {
			gcpAuthType := viper.GetString(cst.GcpAuthType)
			gcp := GcpClient{}
			token, err = gcp.GetJwtToken(gcpAuthType)
			if err != nil {
				return nil, errors.New(err).Grow("Failed to fetch token for gcp")
			}
			log.Printf("Gcp Token:\n%s\n", token)
		}
		data = requestBody{
			GrantType: authTypeToGrantType[at],
			Jwt:       token,
		}
	}

	if data.GrantType == "" {
		data = requestBody{
			GrantType: authTypeToGrantType[at],
		}
		if at == Password {
			err := setupDataForPasswordAuth(&data)
			if err != nil {
				return nil, errors.New(err)
			}
		} else if at == ClientCredential {
			data.AuthClientID = viper.GetString(cst.AuthClientID)
			if secret, err := store.GetSecureSetting(cst.AuthClientSecret); err != nil || secret == "" {
				if err == nil {
					err = errors.NewS("auth-client-secret setting is empty")
				}
				return nil, err.Grow("Failed to retrieve secure setting: " + strings.Replace(cst.AuthClientSecret, ".", "-", -1))
			} else {
				data.AuthClientSecret = secret
			}
		} else if at == Refresh {
			refreshToken := viper.GetString(cst.RefreshToken)
			if data.RefreshToken == "" {
				if refreshToken == "" {
					return nil, errors.NewS("Refresh authentication failed: refreshtoken flag must be set")
				}
				data.RefreshToken = refreshToken
			}
		} else if at == Oidc || at == FederatedThyOne {
			data.Provider = viper.GetString(cst.AuthProvider)

			callback := viper.GetString(cst.Callback)
			if callback == "" {
				callback = cst.DefaultCallback
			}

			data.CallbackHost = callback
			data.CallbackUrl = fmt.Sprintf("http://%s/callback", callback)
		} else if at == Certificate {
			challengeID, challenge, err := a.initiateCertAuth(
				viper.GetString(cst.AuthCert),
				viper.GetString(cst.AuthPrivateKey),
			)
			if err != nil {
				return nil, err
			}
			data.CertChallengeID = challengeID
			data.DecryptedChallenge = challenge
		}
	}

	if tr, err := a.fetchTokenVault(at, data); err != nil {
		if at == Refresh {
			log.Printf("Refresh authentication failed: %s\n", err.Error())
			if pass, err := store.GetSecureSetting(cst.Password); err == nil && pass != "" && viper.GetString(cst.Username) != "" {
				log.Println("Username and password set. Attempt password authentication")
				return a.getTokenForAuthType(Password, false)
			} else {
				if err == nil {
					if viper.GetString(cst.Username) == "" {
						err = utils.NewMissingArgError(cst.Username)
					} else if pass == "" {
						return a.getTokenForAuthType(Password, false)
					}
				}
				return nil, errors.New(err).Grow("Refresh authentication failed. Please re-authenticate with password or other supported authentication type")
			}
		}
		return tr, err.Grow(fmt.Sprintf("Failed to authenticate with auth type '%s'. Please check parameters and try again", at))
	} else {
		log.Printf("%s authentication succeeded.\n", strings.Title(string(at)))
		if err := a.store.Store(keyToken, tr); err != nil {
			return nil, err.Grow("Failed caching token")
		} else {
			return tr, nil
		}
	}
}

type RedirectResponse struct {
	RedirectUrl string `json:"redirect_url"`
}

type AuthResponse struct {
	state    string
	authCode string
	message  string
	err      *errors.ApiError
}

func (a *authenticator) fetchTokenVault(at AuthType, data requestBody) (*TokenResponse, *errors.ApiError) {
	var response TokenResponse
	if err := data.ValidateForAuthType(at); err != nil {
		return nil, errors.New(err)
	}
	if at == Oidc || at == FederatedThyOne {
		ui := cli.BasicUi{
			Writer:      os.Stdout,
			Reader:      os.Stdin,
			ErrorWriter: os.Stderr,
		}

		var redirectResponse RedirectResponse
		uri := paths.CreateURI("oidc/auth", nil)

		if err := a.requestClient.DoRequestOut(http.MethodPost, uri, data, &redirectResponse); err != nil {
			return nil, err
		}
		callbackListener, err := net.Listen("tcp", data.CallbackHost)
		if err != nil {
			return nil, errors.NewF("unable to open callback listener: %v", err)
		}
		defer callbackListener.Close()

		authChannel := make(chan AuthResponse)
		http.HandleFunc("/callback", a.handleOidcAuth(authChannel))

		go func() {
			err := http.Serve(callbackListener, nil)
			if err != nil && err != http.ErrServerClosed {
				authChannel <- AuthResponse{
					message: "login failed",
					err:     errors.New(err),
				}
			}
		}()

		if err = browser.OpenURL(redirectResponse.RedirectUrl); err != nil {
			ui.Info(fmt.Sprintf("Unable to open browser, complete login process here:\n %s", redirectResponse.RedirectUrl))
		}

		select {
		case ar := <-authChannel:
			if ar.err != nil {
				return nil, ar.err
			}
			data.State = ar.state
			data.AuthorizationCode = ar.authCode
			ui.Info(fmt.Sprintf("Received response from %s provider, submitting authorization code to %s", at, cst.ProductName))
		case <-time.After(5 * time.Minute):
			ui.Info(fmt.Sprintf("Timeout occurred waiting for callback from %s provider", at))
			return nil, errors.NewS("no callback occurred after redirect")
		}
	}

	uri := paths.CreateURI(cst.NounToken, nil)
	if err := a.requestClient.DoRequestOut(http.MethodPost, uri, data, &response); err != nil {
		return nil, err
	} else if response.IsNil() {
		return nil, errors.NewS("Empty token")
	}
	response.Granted = time.Now().UTC()
	return &response, nil
}

// handleOidcAuth handles OIDC and Thycotic One auths
func (a *authenticator) handleOidcAuth(doneCh chan<- AuthResponse) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			w.Write([]byte(err.Error()))
			doneCh <- AuthResponse{
				err:     errors.New(err),
				message: "error in callback",
			}
		}

		code := req.URL.Query().Get("code")
		state := req.URL.Query().Get("state")

		if code == "" || state == "" {
			doneCh <- AuthResponse{
				err: errors.NewS("missing values in callback, authorization code or state are empty"),
			}
			w.Write(b)
			return
		}

		tmpl, err := template.New("youDidIt").Parse(youDidIt)
		if err != nil {
			w.Write([]byte(err.Error()))
			doneCh <- AuthResponse{
				err:     errors.New(err),
				message: "error in html parse template",
			}
		}

		vars := map[string]interface{}{
			"providerName": viper.GetString(cst.AuthType),
		}

		err = tmpl.Execute(w, vars)
		if err != nil {
			w.Write([]byte(err.Error()))
			doneCh <- AuthResponse{
				err:     errors.New(err),
				message: "error in html template execute",
			}
		}

		doneCh <- AuthResponse{
			err:      nil,
			message:  "success",
			authCode: code,
			state:    state,
		}

	}
}

// initiateCertAuth makes initial request and prepares info for final token request.
func (a *authenticator) initiateCertAuth(cert, privKey string) (string, string, *errors.ApiError) {
	log.Println("Reading private key.")
	der, err := base64.StdEncoding.DecodeString(privKey)
	if err != nil {
		return "", "", errors.NewF("unable to read private key: %v", err)
	}
	block, _ := pem.Decode(der)
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", "", errors.NewF("unable to parse private key: %v", err)
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
		return "", "", requestErr
	}
	encrypted, err := base64.StdEncoding.DecodeString(response.Encrypted)
	if err != nil {
		return "", "", errors.NewF("unable to read challenge: %v", err)
	}

	log.Println("Decrypting challenge using private key.")
	plaintext, err := rsa.DecryptOAEP(sha512.New(), rand.Reader, privateKey, encrypted, nil)
	if err != nil {
		return "", "", errors.NewF("unable to decrypt challenge: %v", err)
	}

	decrypted := base64.StdEncoding.EncodeToString(plaintext)
	return response.ID, decrypted, nil
}

// setupDataForPasswordAuth prepares the requestBody object with data for password auth.
// It should be used in the authentication workflow, not on its own.
func setupDataForPasswordAuth(data *requestBody) error {
	userName := viper.GetString(cst.Username)
	data.Username = userName
	data.Provider = viper.GetString(cst.AuthProvider)

	// If plaintext Password exists, that means Viper retrieves it from memory. Use this Password to authenticate.
	// If it is an empty string, look for SecurePassword, which Viper gets only from config. Get the corresponding
	// key file and use it to decrypt SecurePassword.
	if data.Password = viper.GetString(cst.Password); data.Password == "" {
		passSetting := cst.SecurePassword
		storeType := viper.GetString(cst.StoreType)
		if storeType == store.WinCred || storeType == store.PassLinux {
			passSetting = cst.Password
		}
		if pass, err := store.GetSecureSetting(passSetting); err == nil && pass != "" {
			if passSetting == cst.SecurePassword {
				keyPath := GetEncryptionKeyFilename(viper.GetString(cst.Tenant), userName)
				key, err := store.ReadFileInDefaultPath(keyPath)
				if err != nil || key == "" {
					return KeyfileNotFoundError
				}
				decrypted, decryptionErr := Decrypt(pass, key)
				if decryptionErr != nil {
					return errors.NewS("Failed to decrypt the password with key.")
				} else {
					data.Password = decrypted
				}
			} else {
				data.Password = pass
			}
		}
	}
	return nil
}

func buildAwsParams() (headers string, body string, err *errors.ApiError) {
	opts := session.Options{
		SharedConfigState:       session.SharedConfigEnable,
		AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
	}
	awsProfile := viper.GetString(cst.AwsProfile)
	if awsProfile != "" {
		opts.Profile = awsProfile
	}
	if sess, err := session.NewSessionWithOptions(opts); err != nil {
		return "", "", errors.New(err).Grow("Failed to create aws session")
	} else {
		stsClient := sts.New(sess)
		r, _ := stsClient.GetCallerIdentityRequest(nil)
		r.Sign()
		headers, err1 := json.Marshal(r.HTTPRequest.Header)
		body, err2 := io.ReadAll(r.HTTPRequest.Body)
		hString := base64.StdEncoding.EncodeToString(headers)
		bString := base64.StdEncoding.EncodeToString(body)
		return hString, bString, errors.New(err1).Or(errors.New(err2))
	}
}

type requestBody struct {
	GrantType          string `json:"grant_type"`
	Username           string `json:"username"`
	Provider           string `json:"provider"`
	Password           string `json:"password"`
	AuthClientID       string `json:"client_id"`
	AuthClientSecret   string `json:"client_secret"`
	RefreshToken       string `json:"refresh_token"`
	AwsBody            string `json:"aws_body"`
	AwsHeaders         string `json:"aws_headers"`
	Jwt                string `json:"jwt"`
	AzureAuthClientID  string
	AuthorizationCode  string `json:"authorization_code"`
	CallbackUrl        string `json:"callback_url"`
	State              string `json:"state"`
	CallbackHost       string `json:"_"`
	CertChallengeID    string `json:"cert_challenge_id"`
	DecryptedChallenge string `json:"decrypted_challenge"`
}

type TokenResponse struct {
	Token        string    `json:"accessToken"`
	TokenType    string    `json:"tokenType"`
	ExpiresIn    int64     `json:"expiresIn"`
	RefreshToken string    `json:"refreshToken"`
	Granted      time.Time `json:"granted"`
}

func (r *TokenResponse) SecondsRemainingToken() int64 {
	remaining := int64(r.Granted.Sub(time.Now().UTC()).Seconds()) + r.ExpiresIn - leewaySecondsTokenExp
	return remaining
}

func (r *TokenResponse) SecondsRemainingRefreshToken() int64 {
	if r.RefreshToken != "" {
		remaining := int64(r.Granted.Sub(time.Now().UTC()).Seconds()) + refreshTokenLifeSeconds - leewaySecondsTokenExp
		return remaining
	}
	return 0
}

func (r *TokenResponse) IsNil() bool {
	return r.Token == "" && r.RefreshToken == ""
}

func (r *requestBody) ValidateForAuthType(at AuthType) error {
	ref := reflect.ValueOf(r)
	for _, k := range paramSpecDict[at] {
		if !k.RequestVar {
			continue
		}
		f := reflect.Indirect(ref).FieldByName(k.PropName)
		if f.String() == "" {
			return utils.NewMissingArgError(k.ArgName)
		}
	}
	return nil
}

var authTypeToGrantType = map[AuthType]string{
	Password:         "password",
	FederatedThyOne:  "oidc",
	ClientCredential: "client_credentials",
	Certificate:      "certificate",
	Refresh:          "refresh_token",
	FederatedAws:     "aws_iam",
	FederatedAzure:   "azure",
	FederatedGcp:     "gcp",
	Oidc:             "oidc",
}

type paramSpec struct {
	PropName   string
	ArgName    string
	IsKey      bool
	RequestVar bool
}

var paramSpecDict = map[AuthType][]paramSpec{
	Password: {
		{PropName: "Password",
			ArgName:    cst.Password,
			IsKey:      false,
			RequestVar: true,
		},
		{PropName: "Username",
			ArgName:    cst.Username,
			IsKey:      true,
			RequestVar: true,
		},
		{PropName: "Provider",
			ArgName:    cst.AuthProvider,
			IsKey:      false,
			RequestVar: false,
		},
	},
	FederatedThyOne: {
		{PropName: "AuthType",
			ArgName:    cst.AuthType,
			IsKey:      true,
			RequestVar: false,
		},
	},
	ClientCredential: {
		{PropName: "AuthClientID",
			ArgName:    cst.AuthClientID,
			IsKey:      true,
			RequestVar: true,
		},
		{PropName: "AuthClientSecret",
			ArgName:    cst.AuthClientSecret,
			IsKey:      false,
			RequestVar: true,
		},
	},
	Certificate: {
		{PropName: "CertificatePath",
			ArgName:    cst.CertPath,
			IsKey:      true,
			RequestVar: false,
		},
	},
	FederatedAws: {
		{PropName: "Profile",
			ArgName:    cst.AwsProfile,
			IsKey:      true,
			RequestVar: false,
		},
		{PropName: "AwsBody",
			ArgName:    "",
			IsKey:      false,
			RequestVar: true,
		},
		{PropName: "AwsHeaders",
			ArgName:    "",
			IsKey:      false,
			RequestVar: true,
		},
	},
	Refresh: {
		{PropName: "RefreshToken",
			ArgName:    cst.RefreshToken,
			IsKey:      true,
			RequestVar: true,
		},
	},
	FederatedAzure: {
		{PropName: "JwtToken",
			ArgName:    "jwt",
			IsKey:      false,
			RequestVar: true,
		},
		{PropName: "AzureAuthClientID",
			ArgName:    cst.AzureAuthClientID,
			IsKey:      true,
			RequestVar: false,
		},
	},
	FederatedGcp: {
		{PropName: "JwtToken",
			ArgName:    "jwt",
			IsKey:      false,
			RequestVar: true,
		},
		{PropName: "Service",
			ArgName:    "Service",
			IsKey:      true,
			RequestVar: false,
		},
	},
	Oidc: {
		{PropName: "AuthType",
			ArgName:    cst.AuthType,
			IsKey:      true,
			RequestVar: false,
		},
	},
}
