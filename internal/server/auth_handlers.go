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
	_ "github.com/joho/godotenv/autoload"
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

func (s *Server) InfoHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := GetClaims(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]any{
		"user": claims.UserName,
	}

	if err = json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to encode success response", "error", err)
	}
}

func (s *Server) ExchangeHandler(w http.ResponseWriter, r *http.Request) {
	var oAuth models.OAuthCode
	statusCode, err := DecodeJSON(r, &oAuth)
	if err != nil {
		writeError(w, statusCode, err.Error())
		return
	}

	if strings.TrimSpace(oAuth.Code) == "" {
		writeError(w, http.StatusBadRequest, "Code is required")
		return
	}
	if strings.TrimSpace(oAuth.Verifier) == "" {
		writeError(w, http.StatusBadRequest, "Verifier is required")
		return
	}

	reqData := url.Values{}
	reqData.Add("client_id", "exileprofit")
	reqData.Add("client_secret", clientSecret)
	reqData.Add("grant_type", "authorization_code")
	reqData.Add("code", oAuth.Code)
	reqData.Add("redirect_uri", "https://exile-profit.com/auth/poe")
	reqData.Add("scope", "account:profile account:stashes account:characters")
	reqData.Add("code_verifier", oAuth.Verifier)

	req, err := http.NewRequest("POST", "https://www.pathofexile.com/oauth/token", strings.NewReader(reqData.Encode()))
	if err != nil {
		slog.Error("failed to create OAuth token request", "error", err)
		writeError(w, http.StatusInternalServerError, "Internal server error: failed to create token exchange request")
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "OAuth exileprofit/0.0.1 (contact: vyaryw@gmail.com)")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("OAuth token request failed", "error", err, "url", "https://www.pathofexile.com/oauth/token")
		writeError(w, http.StatusInternalServerError, "Failed to communicate with Path of Exile OAuth service")
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		var body []byte

		body, err = io.ReadAll(res.Body)
		if err != nil {
			slog.Error("failed to read OAuth error response body", "error", err, "status_code", res.StatusCode)
			writeError(w, http.StatusInternalServerError, "OAuth token exchange failed and unable to read error details")
			return
		}

		slog.Error("OAuth token exchange failed", "status_code", res.StatusCode, "response_body", string(body))

		// test different errors and create custom messages
		writeError(w, http.StatusInternalServerError, "OAuth token exchange failed with upstream error")
		return
	}

	var token models.OAuthToken
	err = json.NewDecoder(res.Body).Decode(&token)
	if err != nil {
		slog.Error("failed to decode OAuth token response", "error", err)
		writeError(w, http.StatusInternalServerError, "Invalid response format from OAuth service")
		return
	}

	reqUser, err := http.NewRequest("GET", "https://api.pathofexile.com/profile", nil)
	if err != nil {
		slog.Error("failed to create user profile request", "error", err)
		writeError(w, http.StatusInternalServerError, "Internal server error: failed to create profile request")
		return
	}

	reqUser.Header.Set("User-Agent", "OAuth exileprofit/0.0.1 (contact: vyaryw@gmail.com)")
	reqUser.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resUser, err := http.DefaultClient.Do(reqUser)
	if err != nil {
		slog.Error("user profile request failed", "error", err)
		writeError(w, http.StatusInternalServerError, "Failed to retrieve user profile from Path of Exile API")
		return
	}
	defer resUser.Body.Close()

	if resUser.StatusCode != http.StatusOK {
		var body []byte

		body, err = io.ReadAll(r.Body)
		if err != nil {
			slog.Error("failed to read user profile error response", "error", err, "status_code", resUser.StatusCode)
			writeError(w, http.StatusInternalServerError, "Failed to retrieve user profile and unable to read error details")
			return
		}

		slog.Error("failed to get user profle", "status_code", resUser.StatusCode, "body", string(body))

		// track errors and add custom responses
		writeError(w, http.StatusInternalServerError, "Unable to retrieve user profile")
		return
	}

	var user models.UserProfile
	if err = json.NewDecoder(resUser.Body).Decode(&user); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to decode user")
		return
	}

	tokenPair, err := s.GenTokenPair(user)
	if err != nil {
		slog.Error("failed to gen JWT tokens", "error", err, "user_uuid", user.ID)
		writeError(w, http.StatusInternalServerError, "Failed to create authentication token")
		return
	}

	if err = s.db.StoreOAuthToken(user.ID, token); err != nil {
		slog.Error("failed to save token", "error", err)
		writeError(w, http.StatusInternalServerError, "Failed to save token")
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
}

func (s *Server) TokenRefreshHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := GetClaims(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	if !s.db.IsRefreshTokenValid(claims.TokenID) {
		writeError(w, http.StatusUnauthorized, "Refresh token revoked")
		return
	}

	if err = s.db.RevokeRefreshToken(claims.UserID, claims.TokenID); err != nil {
		slog.Error("failed to revoke token", "user_id", claims.UserID, "token_id", claims.TokenID)
	}

	tokenPair, err := s.GenTokenPair(models.UserProfile{ID: claims.UserID, Name: claims.UserName})
	if err != nil {
		slog.Error("failed to gen JWT tokens", "error", err, "user_uuid", claims.UserID)
		writeError(w, http.StatusInternalServerError, "Failed to create authentication token")
		return
	}

	setJWTCookies(w, *tokenPair)

	w.WriteHeader(http.StatusOK)
}

func (s *Server) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := GetClaims(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	if err := s.db.RevokeRefreshToken(claims.UserID, claims.TokenID); err != nil {
		slog.Error("failed to revoke refresh token", "user_id", claims.UserID, "token_id", claims.TokenID)
	}

	clearJWTCookies(w)

	w.WriteHeader(http.StatusOK)
}

func (s *Server) LogoutAllHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := GetClaims(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
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
}
