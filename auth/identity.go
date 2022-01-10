package auth

import (
	"github.com/golang-jwt/jwt/v4"
	"thy/errors"
)

func GetCurrentIdentity() (string, error) {
	authenticator := NewAuthenticatorDefault()
	tokenResp, apiErr := authenticator.GetToken()
	if apiErr != nil {
		return "", apiErr
	}
	return ParseSubjectFromToken(tokenResp.Token)
}

func ParseSubjectFromToken(accessToken string) (string, error) {
	standardClaims := &jwt.RegisteredClaims{}
	parser := jwt.Parser{}
	token, _, err := parser.ParseUnverified(accessToken, standardClaims)
	if err != nil {
		return "", err
	}
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return "", errors.NewS("Failed to parse the JWT token")
	}
	return claims.Subject, nil
}
