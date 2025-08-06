// Package models defines data types used across the API service
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
	ID   string `json:"id"`
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

type Item struct {
	ID        int     `json:"id"`
	Icon      string  `json:"icon"`
	Name      string  `json:"name"`
	Base      string  `json:"base"`
	Category  string  `json:"category"`
	Value     float32 `json:"value"`
	Currency  string  `json:"currency"`
	Listed    int     `json:"listed"`
	UserID    string  `json:"user_id"`
	CreatedAt int64   `json:"created_at"`
	UpdatedAt int64   `json:"updated_at"`
}

type Strategy struct {
	ID          int    `json:"id"`
	UserID      string `json:"user_id"`
	CreatedBy   string `json:"created_by"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Atlas       string `json:"atlas"`
	Public      bool   `json:"public"`
	Featured    bool   `json:"featured"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

type StrategyItem struct {
	ID         int     `json:"id"`
	StratID    int     `json:"strat_id"`
	ItemID     int     `json:"item_id"`
	Amount     int     `json:"amount"`
	Role       string  `json:"role"`
	DropChance float32 `json:"drop_chance"`
	Pair       int     `json:"pair"`
}
