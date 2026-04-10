package handler_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/handler"
	testutil "sol_coffeesys/backend/handler/testutil"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGenerateRefreshToken_Success(t *testing.T) {
	mockDB := new(testutil.MockDB)

	var captured db.CreateRefreshTokenParams
	mockDB.On("CreateRefreshToken", mock.Anything, mock.Anything).Run(
		func(args mock.Arguments) {
			captured = args.Get(1).(db.CreateRefreshTokenParams)
		}).Return(
		db.RefreshToken{ID: 1, UserID: 1}, nil)
	raw, hash, expiresAt, err := handler.GenerateRefreshToken(context.Background(), mockDB, 1)
	assert.NoError(t, err)
	assert.Equal(t, 64, len(raw))

	sum := sha256.Sum256([]byte(raw))
	assert.Equal(t, hash, hex.EncodeToString(sum[:]))
	assert.Equal(t, hash, captured.TokenHash)
	assert.WithinDuration(t, time.Now().Add(14*24*time.Hour), expiresAt, 5*time.Second)

	mockDB.AssertExpectations(t)
}

func TestGenerateRefreshToken_DBError(t *testing.T) {
	mockDB := new(testutil.MockDB)
	mockDB.On("CreateRefreshToken", mock.Anything, mock.Anything).Return(db.RefreshToken{}, errors.New("db error"))
	_, _, _, err := handler.GenerateRefreshToken(context.Background(), mockDB, 1)
	assert.Error(t, err)
	mockDB.AssertExpectations(t)
}

func TestRevokeRefreshByRaw_Success(t *testing.T) {
	mockDB := new(testutil.MockDB)
	oldRaw := strings.Repeat("a", 64)
	sum := sha256.Sum256([]byte(oldRaw))
	oldHash := hex.EncodeToString(sum[:])

	mockDB.On("RevokeRefreshTokenByHash", mock.Anything, oldHash).Return(nil)

	err := handler.RevokeRefreshByRaw(context.Background(), mockDB, oldRaw)
	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
}

func TestRevokeRefreshByRaw_DBError(t *testing.T) {
	mockDB := new(testutil.MockDB)
	oldRaw := strings.Repeat("a", 64)

	sum := sha256.Sum256([]byte(oldRaw))
	oldHash := hex.EncodeToString(sum[:])

	mockDB.On("RevokeRefreshTokenByHash", mock.Anything, oldHash).Return(errors.New("db error"))

	err := handler.RevokeRefreshByRaw(context.Background(), mockDB, oldRaw)
	assert.Error(t, err)
	mockDB.AssertExpectations(t)
}
