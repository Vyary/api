package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Vyary/api/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("POST /auth/poe/exchange", s.LoginHandler())
	mux.Handle("GET /info", s.Info())
	mux.Handle("GET /set", s.SetCookie())

	return mux
}

func (s *Server) Info() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookies := r.Cookies()

		all := make(map[string]string)
		for _, c := range cookies {
			all[c.Name] = c.Value
		}

		if err := json.NewEncoder(w).Encode(all); err != nil {
			slog.Error("failed to encode cookies", "error", err)
		}
	})
}

func (s *Server) SetCookie() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie := http.Cookie{
			Name:     "session_token",                         // The name of the cookie.
			Value:    "a_very_secret_and_secure_token_string", // The value. This would typically be a session ID or JWT.
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
			Secure:   true, // Set to false if not using HTTPS
			Path:     "/",
			SameSite: http.SameSiteNoneMode,
		}

		http.SetCookie(w, &cookie)

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Cookie has been set successfully!")
	})
}

func writeError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := models.ErrorResponse{
		Error: message,
		Code:  code,
	}

	json.NewEncoder(w).Encode(response)
}

func (s *Server) LoginHandler() http.Handler {
	clientSecret := os.Getenv("CLIENT_SECRET")
	jwtSecret := os.Getenv("JWT_SECRET")

	if clientSecret == "" {
		slog.Error("CLIENT_SECRET environment variable is required")
		os.Exit(1)
	}
	if jwtSecret == "" {
		slog.Error("JWT_SECRET environment variable is required")
		os.Exit(1)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			slog.Warn("invalid content type", "content_type", r.Header.Get("Content-Type"), "expected", "application/json")
			writeError(w, "Content-Type must be application/json", http.StatusBadRequest)
			return
		}

		var code models.OAuthCode
		if err := json.NewDecoder(r.Body).Decode(&code); err != nil {
			slog.Error("failed to decode request body", "error", err)
			writeError(w, "Invalid JSON format in request body", http.StatusBadRequest)
			return
		}

		if strings.TrimSpace(code.Code) == "" {
			writeError(w, "Code is required", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(code.Verifier) == "" {
			writeError(w, "Verifier is required", http.StatusBadRequest)
			return
		}

		reqData := url.Values{}
		reqData.Add("client_id", "exileprofit")
		reqData.Add("client_secret", clientSecret)
		reqData.Add("grant_type", "authorization_code")
		reqData.Add("code", code.Code)
		reqData.Add("redirect_uri", "https://exile-profit.com/auth/poe")
		reqData.Add("scope", "account:profile account:stashes account:characters")
		reqData.Add("code_verifier", code.Verifier)

		req, err := http.NewRequest("POST", "https://www.pathofexile.com/oauth/token", strings.NewReader(reqData.Encode()))
		if err != nil {
			slog.Error("failed to create OAuth token request", "error", err)
			writeError(w, "Internal server error: failed to create token exchange request", http.StatusInternalServerError)
			return
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "OAuth exileprofit/0.0.1 (contact: vyaryw@gmail.com)")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			slog.Error("OAuth token request failed", "error", err, "url", "https://www.pathofexile.com/oauth/token")
			writeError(w, "Failed to communicate with Path of Exile OAuth service", http.StatusInternalServerError)
			return
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			var body []byte

			body, err = io.ReadAll(res.Body)
			if err != nil {
				slog.Error("failed to read OAuth error response body", "error", err, "status_code", res.StatusCode)
				writeError(w, "OAuth token exchange failed and unable to read error details", http.StatusInternalServerError)
				return
			}

			slog.Error("OAuth token exchange failed", "status_code", res.StatusCode, "response_body", string(body))

			// test different errors and create custom messages
			writeError(w, "OAuth token exchange failed with upstream error", http.StatusInternalServerError)
			return
		}

		var token models.OAuthToken
		err = json.NewDecoder(res.Body).Decode(&token)
		if err != nil {
			slog.Error("failed to decode OAuth token response", "error", err)
			writeError(w, "Invalid response format from OAuth service", http.StatusInternalServerError)
			return
		}

		reqUser, err := http.NewRequest("GET", "https://api.pathofexile.com/profile", nil)
		if err != nil {
			slog.Error("failed to create user profile request", "error", err)
			writeError(w, "Internal server error: failed to create profile request", http.StatusInternalServerError)
			return
		}

		reqUser.Header.Set("User-Agent", "OAuth exileprofit/0.0.1 (contact: vyaryw@gmail.com)")
		reqUser.Header.Set("Authorization", "Bearer "+token.AccessToken)

		resUser, err := http.DefaultClient.Do(reqUser)
		if err != nil {
			slog.Error("user profile request failed", "error", err)
			writeError(w, "Failed to retrieve user profile from Path of Exile API", http.StatusInternalServerError)
			return
		}
		defer resUser.Body.Close()

		if resUser.StatusCode != http.StatusOK {
			var body []byte

			body, err = io.ReadAll(r.Body)
			if err != nil {
				slog.Error("failed to read user profile error response", "error", err, "status_code", resUser.StatusCode)
				writeError(w, "Failed to retrieve user profile and unable to read error details", http.StatusInternalServerError)
				return
			}

			slog.Error("failed to get user profle", "status_code", resUser.StatusCode, "body", string(body))

			// track errors and add custom responses
			writeError(w, "Unable to retrieve user profile", http.StatusInternalServerError)
			return
		}

		var user models.User
		if err = json.NewDecoder(resUser.Body).Decode(&user); err != nil {
			writeError(w, "Failed to decode user", http.StatusInternalServerError)
			return
		}

		now := time.Now()
		claims := models.JWTClaims{
			UserID:   user.UUID,
			UserName: user.Name,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(now),
				NotBefore: jwt.NewNumericDate(now),
				Issuer:    "exile-profit",
				Subject:   user.UUID,
			},
		}

		jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		signedToken, err := jwtToken.SignedString([]byte(jwtSecret))
		if err != nil {
			slog.Error("failed to sign JWT token", "error", err, "user_uuid", user.UUID)
			writeError(w, "Internal server error: failed to create authentication token", http.StatusInternalServerError)
			return
		}

		if err = s.db.SaveToken(user.UUID, token); err != nil {
			slog.Error("failed to save token", "error", err)
			writeError(w, "Failed to save token", http.StatusInternalServerError)
			return
		}

		cookie := http.Cookie{
			Name:     "jwt_token",
			Value:    signedToken,
			Path:     "/",
			MaxAge:   3600,
			HttpOnly: true,
			Secure:   true,
			Domain:   ".exile-profit.com",
			SameSite: http.SameSiteLaxMode,
		}

		http.SetCookie(w, &cookie)

		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"user": user.Name,
		}

		if err = json.NewEncoder(w).Encode(response); err != nil {
			slog.Error("failed to encode success response", "error", err)
		}
	})
}
