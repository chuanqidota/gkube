package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Hardcoded JWT secret key for development.
// In production, this should be read from configuration.
var jwtSecret = []byte("gkube-jwt-secret-key-change-in-production")

// Claims represents the JWT claims structure.
type Claims struct {
	UserID       uint   `json:"user_id"`
	Username     string `json:"username"`
	IsSuperAdmin bool   `json:"is_super_admin"`
	jwt.RegisteredClaims
}

// TokenPair contains an access token and a refresh token.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// GenerateToken creates a new access token (2h) and refresh token (7d) for the given user.
func GenerateToken(userID uint, username string, isSuperAdmin bool) (*TokenPair, error) {
	now := time.Now()

	// Access token claims with 2-hour expiry
	accessClaims := Claims{
		UserID:       userID,
		Username:     username,
		IsSuperAdmin: isSuperAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(2 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "gkube",
		},
	}

	accessSigningKey := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err := accessSigningKey.SignedString(jwtSecret)
	if err != nil {
		return nil, err
	}

	// Refresh token claims with 7-day expiry
	refreshClaims := Claims{
		UserID:       userID,
		Username:     username,
		IsSuperAdmin: isSuperAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "gkube",
		},
	}

	refreshSigningKey := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err := refreshSigningKey.SignedString(jwtSecret)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// ParseToken validates and parses a JWT token string, returning the claims.
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
