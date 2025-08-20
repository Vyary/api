package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Vyary/api/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

func writeError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := models.ErrorResponse{
		Error: message,
		Code:  code,
	}

	json.NewEncoder(w).Encode(response)
}

func GetUser(r *http.Request) (*models.User, error) {
	// return &models.User{ID: "asd", Name: "Vyary"}, nil

	claims, err := GetClaims(r)
	if err != nil {
		return nil, err
	}

	return &models.User{ID: claims.UserID, Name: claims.UserName}, nil
}

func GetClaims(r *http.Request) (*models.JWTClaims, error) {
	tokenCookie, err := r.Cookie("jwt_token")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return nil, errors.New("JWT cookie not found")
		}
		return nil, fmt.Errorf("failed to read JWT cookie: %w", err)
	}

	if tokenCookie.Value == "" {
		return nil, errors.New("JWT cookie is empty")
	}

	token, err := jwt.ParseWithClaims(tokenCookie.Value, &models.JWTClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("JWT token is not valid")
	}

	claims, ok := token.Claims.(*models.JWTClaims)
	if !ok {
		return nil, fmt.Errorf("failed to extract claims from token")
	}

	return claims, nil
}

func DecodeJSON(r *http.Request, data any) (statusCode int, err error) {
	if r.Header.Get("Content-Type") != "application/json" {
		return http.StatusUnsupportedMediaType, errors.New("invalid Content-Type")
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return http.StatusInternalServerError, errors.New("unable to process request")
	}

	return http.StatusOK, nil
}
