package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// Claims defines the JWT claims for access tokens.
type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// HashPassword hashes a plaintext password using bcrypt.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hashing password: %w", err)
	}
	return string(hash), nil
}

// CheckPassword compares a plaintext password against a bcrypt hash.
func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// GenerateAccessToken creates a signed JWT access token.
func GenerateAccessToken(userID int64, email, secret string, expiry time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", userID),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateAccessToken parses and validates a JWT access token, returning the claims.
func ValidateAccessToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}

// GenerateRefreshToken creates a cryptographically random refresh token string.
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating refresh token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// HashToken returns the SHA-256 hash of a token string for database storage.
func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// StoreRefreshToken inserts a hashed refresh token into the database.
func StoreRefreshToken(db *sql.DB, userID int64, tokenHash string, expiresAt time.Time) error {
	_, err := db.Exec(
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("storing refresh token: %w", err)
	}
	return nil
}

// ValidateRefreshToken checks if a refresh token hash exists, is not expired, and is not revoked.
// Returns the user_id if valid.
func ValidateRefreshToken(db *sql.DB, tokenHash string) (int64, error) {
	var userID int64
	err := db.QueryRow(
		`SELECT user_id FROM refresh_tokens
		 WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW()`,
		tokenHash,
	).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("validating refresh token: %w", err)
	}
	return userID, nil
}

// RevokeRefreshToken marks a refresh token as revoked.
func RevokeRefreshToken(db *sql.DB, tokenHash string) error {
	_, err := db.Exec(
		`UPDATE refresh_tokens SET revoked_at = NOW() WHERE token_hash = $1`,
		tokenHash,
	)
	if err != nil {
		return fmt.Errorf("revoking refresh token: %w", err)
	}
	return nil
}

// RevokeAllUserRefreshTokens revokes all refresh tokens for a given user.
func RevokeAllUserRefreshTokens(db *sql.DB, userID int64) error {
	_, err := db.Exec(
		`UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("revoking all user refresh tokens: %w", err)
	}
	return nil
}
