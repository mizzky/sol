package handler

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"sol_coffeesys/backend/db"
	"time"
)

func GenerateRefreshToken(ctx context.Context, q db.Querier, userID int64) (rawToken string, tokenHash string, expiresAt time.Time, err error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", time.Time{}, err
	}
	rawToken = hex.EncodeToString(raw)
	sum := sha256.Sum256([]byte(rawToken))
	tokenHash = hex.EncodeToString(sum[:])
	expiresAt = time.Now().Add(14 * 24 * time.Hour)

	if _, err := q.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}); err != nil {
		return "", "", time.Time{}, err
	}
	return rawToken, tokenHash, expiresAt, nil
}

func RevokeRefreshByRaw(ctx context.Context, q db.Querier, raw string) error {
	sum := sha256.Sum256([]byte(raw))
	hash := hex.EncodeToString(sum[:])
	return q.RevokeRefreshTokenByHash(ctx, hash)
}
