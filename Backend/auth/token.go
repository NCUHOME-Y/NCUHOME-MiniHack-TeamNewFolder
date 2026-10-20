package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"
)

var (
	ErrTokenInvalid = errors.New("token invalid")
	ErrTokenExpired = errors.New("token expired")
)

func TokenHash(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(sum[:])
	return tokenHash
}

func GenerateToken(db *sql.DB, userID uint64, ttl time.Duration) (rawToken string, err error) {

	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", err
	}
	rawToken = base64.RawURLEncoding.EncodeToString(b)
	tokenHash := TokenHash(rawToken)

	expiresAt := time.Now().UTC().Add(ttl)
	_, err = db.Exec(
		`INSERT INTO user_tokens (user_id, token_hash, expires_at) VALUES (?, ?, ?)`,
		userID, tokenHash, expiresAt,
	)

	if err != nil {
		return "", err
	}
	return rawToken, nil

}

func VerifyToken(db *sql.DB, rawToken string) (userID uint64, err error) {
	if rawToken == "" {
		return 0, ErrTokenInvalid
	}

	tokenHash := TokenHash(rawToken)

	var (
		expiresAt time.Time
		revoked   int
		status    int
	)
	err = db.QueryRow(`
		SELECT t.user_id, t.expires_at, t.revoked, u.status
		FROM user_tokens t
		JOIN users u ON u.id = t.user_id
		WHERE t.token_hash = ?`, tokenHash).
		Scan(&userID, &expiresAt, &revoked, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrTokenInvalid
	}
	if err != nil {
		return 0, err
	}

	if revoked == 1 {
		return 0, ErrTokenInvalid
	}
	if time.Now().UTC().After(expiresAt) {
		return 0, ErrTokenExpired
	}
	if status != 1 {
		return 0, ErrTokenInvalid // 账号被禁用
	}
	return userID, nil

}
