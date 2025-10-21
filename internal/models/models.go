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

type LatestPrice struct {
	ID       int64   `json:"id"`
	ItemID   string  `json:"item_id"`
	Realm    string  `json:"realm"`
	League   string  `json:"league"`
	Currency string  `json:"currency"`
	Value    float64 `json:"value"`
	Stock    int64   `json:"stock"`
}

type Price struct {
	ID           int64    `json:"id"`
	ItemID       string   `json:"item_id"`
	Realm        *string  `json:"realm"`
	League       string   `json:"league"`
	Currency     string   `json:"currency"`
	Value        *float64 `json:"value"`
	LowestRatio  *float64 `json:"lowest_ratio"`
	HighestRatio *float64 `json:"highest_ratio"`
	VolumeTraded *int64   `json:"volume_traded"`
	LowestStock  *int64   `json:"lowest_stock"`
	HighestStock *int64   `json:"highest_stock"`
	Timestamp    *int64   `json:"timestamp"`
}

type User struct {
	ID          string  `json:"id"`
	Name        string  `json:"username"`
	AccessToken string  `json:"access_token"`
	ExpiresIn   int64   `json:"expires_in"`
	TokenType   string  `json:"token_type"`
	Scope       string  `json:"scope"`
	Sub         string  `json:"sub"`
	Role        *string `json:"role"`
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

type StrategyDTO struct {
	ID          int    `json:"id"`
	CreatedBy   string `json:"created_by"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Atlas       string `json:"atlas"`
	Public      bool   `json:"public"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

type StrategyTable struct {
	ID         int    `json:"id"`
	StrategyID int    `json:"strategy_id"`
	Type       string `json:"type"`
	Title      string `json:"title"`
}

type StrategyItem struct {
	SID        int
	StrategyID int     `json:"strategy_id"`
	TableID    int     `json:"table_id"`
	ItemID     int     `json:"item_id"`
	Amount     int     `json:"amount"`
	Role       string  `json:"role"`
	DropChance float32 `json:"drop_chance"`
	Pair       int     `json:"pair"`
}

type SItem struct {
	Item
	StrategyItem
}
