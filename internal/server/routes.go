package server

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/Vyary/api/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

var (
	clientSecret string
	jwtSecret    string
)

func init() {
	clientSecret = os.Getenv("CLIENT_SECRET")
	jwtSecret = os.Getenv("JWT_SECRET")

	if clientSecret == "" {
		slog.Error("CLIENT_SECRET environment variable is required")
		os.Exit(1)
	}
	if jwtSecret == "" {
		slog.Error("JWT_SECRET environment variable is required")
		os.Exit(1)
	}
}

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("POST /auth/poe/exchange", s.ExchangeHandler())
	mux.Handle("POST /auth/poe/refresh", s.TokenRefreshHandler())
	mux.Handle("POST /auth/poe/logout", s.LogoutHandler())
	mux.Handle("POST /auth/poe/logout-all", s.LogoutAllHandler())

	return mux
}

func (s *Server) ExchangeHandler() http.Handler {
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

		tokenPair, err := s.GenTokenPair(user)
		if err != nil {
			slog.Error("failed to gen JWT tokens", "error", err, "user_uuid", user.UUID)
			writeError(w, "Failed to create authentication token", http.StatusInternalServerError)
			return
		}

		if err = s.db.StoreOAuthToken(user.UUID, token); err != nil {
			slog.Error("failed to save token", "error", err)
			writeError(w, "Failed to save token", http.StatusInternalServerError)
			return
		}

		setJWTCookies(w, *tokenPair)

		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"user": user.Name,
		}

		if err = json.NewEncoder(w).Encode(response); err != nil {
			slog.Error("failed to encode success response", "error", err)
		}
	})
}

func (s *Server) TokenRefreshHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, err := r.Cookie("jwt_refresh")
		if err != nil {
			writeError(w, "Refresh token not found", http.StatusUnauthorized)
			return
		}

		token, err := jwt.ParseWithClaims(tokenCookie.Value, &models.JWTClaims{}, func(t *jwt.Token) (any, error) {
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			writeError(w, "Invalid refresh token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*models.JWTClaims)
		if !ok {
			writeError(w, "Invalid refresh token claims", http.StatusUnauthorized)
			return
		}

		if !s.db.IsRefreshTokenValid(claims.TokenID) {
			writeError(w, "Refresh token revoked", http.StatusUnauthorized)
			return
		}

		if err = s.db.RevokeRefreshToken(claims.UserID, claims.TokenID); err != nil {
			slog.Error("failed to revoke token", "user_id", claims.UserID, "token_id", claims.TokenID)
		}

		tokenPair, err := s.GenTokenPair(models.User{UUID: claims.UserID, Name: claims.UserName})
		if err != nil {
			slog.Error("failed to gen JWT tokens", "error", err, "user_uuid", claims.UserID)
			writeError(w, "Failed to create authentication token", http.StatusInternalServerError)
			return
		}

		setJWTCookies(w, *tokenPair)

		w.WriteHeader(http.StatusOK)
	})
}

func (s *Server) LogoutHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, err := r.Cookie("jwt_token")
		if err != nil {
			writeError(w, "JWT cookie not found", http.StatusUnauthorized)
			return
		}

		token, err := jwt.ParseWithClaims(tokenCookie.Value, &models.JWTClaims{}, func(t *jwt.Token) (any, error) {
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			writeError(w, "Invalid jwt token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*models.JWTClaims)
		if !ok {
			writeError(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		if err := s.db.RevokeRefreshToken(claims.UserID, claims.TokenID); err != nil {
			slog.Error("failed to revoke refresh token", "user_id", claims.UserID, "token_id", claims.TokenID)
		}

		clearJWTCookies(w)

		w.WriteHeader(http.StatusOK)
	})
}

func (s *Server) LogoutAllHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, err := r.Cookie("jwt_token")
		if err != nil {
			writeError(w, "JWT cookie not found", http.StatusUnauthorized)
			return
		}

		token, err := jwt.ParseWithClaims(tokenCookie.Value, &models.JWTClaims{}, func(t *jwt.Token) (any, error) {
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			writeError(w, "Invalid jwt token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*models.JWTClaims)
		if !ok {
			writeError(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		if err := s.db.RemoveOAuthToken(claims.UserID); err != nil {
			slog.Error("failed to remove OAuth token", "user_id", claims.UserID)
		}
		if err := s.db.RevokeAllRefreshTokens(claims.UserID); err != nil {
			slog.Error("failed to revoke refresh token", "user_id", claims.UserID, "token_id", claims.TokenID)
		}

		clearJWTCookies(w)

		w.WriteHeader(http.StatusOK)
	})
}
