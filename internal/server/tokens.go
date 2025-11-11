package server

import (
	"time"

	"github.com/Vyary/api/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	jwtExpiration        = 24 * time.Hour * 7
	jwtRefreshExpiration = 24 * time.Hour * 30
)

func (s *Server) GenTokenPair(user models.UserProfile) (*models.TokenPair, error) {
	now := time.Now()
	tokenID := uuid.New().String()

	jwtClaims := models.JWTClaims{
		UserID:   user.ID,
		UserName: user.Name,
		TokenID:  tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(jwtExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "exile-profit",
			Subject:   user.ID,
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)

	signedJwtToken, err := jwtToken.SignedString([]byte(jwtSecret))
	if err != nil {
		return nil, err
	}

	jwtRefreshClaim := models.JWTClaims{
		UserID:   user.ID,
		UserName: user.Name,
		TokenID:  tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(jwtRefreshExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "exile-profit",
			Subject:   user.ID,
		},
	}

	jwtRefreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtRefreshClaim)

	signedJwtRefreshToken, err := jwtRefreshToken.SignedString([]byte(jwtSecret))
	if err != nil {
		return nil, err
	}

	if err := s.db.StoreRefreshToken(user.ID, tokenID, jwtRefreshExpiration); err != nil {
		return nil, err
	}

	return &models.TokenPair{
		JWT:        signedJwtToken,
		JWTRefresh: signedJwtRefreshToken,
	}, nil
}
