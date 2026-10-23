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

// 转hash
func TokenHash(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(sum[:])
	return tokenHash
}

// 制作加密token
func GenerateToken(db *sql.DB, userID uint64, ttl time.Duration) (rawToken string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", err
	}

	rawToken = base64.RawURLEncoding.EncodeToString(b)
	tokenHash := TokenHash(rawToken)

	now := time.Now()
	expiresAt := time.Now().Add(ttl)

	_, err = db.Exec(
		`INSERT INTO user_tokens (user_id, token_hash, expires_at, created_at) VALUES (?, ?, ?, ?)`,
		userID, tokenHash, expiresAt, now,
	)

	if err != nil {
		return "", err
	}
	return rawToken, nil

}

// 检查前端token
func VerifyToken(db *sql.DB, rawToken string) (userID uint64, err error) {
	if rawToken == "" {
		return 0, ErrTokenInvalid
	}
	tokenHash := TokenHash(rawToken)

	var (
		expiresAt time.Time
		revoked   int
	)
	//数据库查询
	// 注意：users 表目前没有 status 列（见 README），查询它会报 Unknown column，
	// 导致所有受保护接口恒 401。JOIN 保留用于确认用户仍然存在。
	err = db.QueryRow(`
		SELECT t.user_id, t.expires_at, t.revoked
		FROM user_tokens t
		JOIN users u ON u.id = t.user_id
		WHERE t.token_hash = ?`, tokenHash).
		Scan(&userID, &expiresAt, &revoked)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrTokenInvalid
	}
	if err != nil {
		return 0, err
	}

	if revoked == 1 {
		return 0, ErrTokenInvalid //账号已退出
	}
	if time.Now().After(expiresAt) {
		return 0, ErrTokenExpired //登录已过期
	}
	return userID, nil

}
