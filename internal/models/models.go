package models

import "github.com/golang-jwt/jwt/v5"

type OAuthCode struct {
	Code     string `json:"code"`
	Verifier string `json:"verifier"`
}

type OAuthToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	Sub         string `json:"sub"`
	Username    string `json:"username"`
}

type User struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

type JWTClaims struct {
	UserID   string
	UserName string
	jwt.RegisteredClaims
}

type ErrorResponse struct {
	Error string
	Code  int
}
