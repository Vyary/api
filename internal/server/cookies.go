package server

import (
	"net/http"
	"time"

	"github.com/Vyary/api/internal/models"
)

func setJWTCookies(w http.ResponseWriter, tokenPair models.TokenPair) {
	jwtCookie := http.Cookie{
		Name:     "jwt_token",
		Value:    tokenPair.JWT,
		Path:     "/",
		Expires:  time.Now().Add(jwtExpiration),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	jwtRefreshCookie := http.Cookie{
		Name:     "jwt_refresh",
		Value:    tokenPair.JWTRefresh,
		Path:     "/auth/poe/refresh",
		Expires:  time.Now().Add(jwtRefreshExpiration),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, &jwtCookie)
	http.SetCookie(w, &jwtRefreshCookie)
}

func clearJWTCookies(w http.ResponseWriter) {
	jwtCookie := http.Cookie{
		Name:     "jwt_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	jwtRefreshCookie := http.Cookie{
		Name:     "jwt_refresh",
		Value:    "",
		Path:     "/auth/poe/refresh",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, &jwtCookie)
	http.SetCookie(w, &jwtRefreshCookie)
}
