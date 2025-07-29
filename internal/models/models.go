package models

import (
	"github.com/golang-jwt/jwt/v5"
)

type OAuthCode struct {
	Code     string `json:"code"`
	Verifier string `json:"verifier"`
}

type OAuthToken struct {
	Username    string `json:"username"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	Sub         string `json:"sub"`
}

type User struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

type JWTClaims struct {
	UserID   string
	UserName string
	TokenID  string
	jwt.RegisteredClaims
}

type ErrorResponse struct {
	Error string
	Code  int
}

type TokenPair struct {
	JWT        string
	JWTRefresh string
}
