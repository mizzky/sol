package db_test

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sol_coffeesys/backend/db"
)

// setupMock は テスト用の sqlmock と Queries を初期化
func setupMock(t *testing.T) (*db.Queries, sqlmock.Sqlmock, func()) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	q := db.New(sqlDB)
	cleanup := func() { sqlDB.Close() }
	return q, mock, cleanup
}

// TestGetUserByID テーブル駆動テスト
func TestGetUserByID(t *testing.T) {
	tests := []struct {
		name        string
		id          int64
		mockSetup   func(sqlmock.Sqlmock)
		expectedErr bool
		expectedID  int64
	}{
		{
			name: "正常系: ユーザーが存在する",
			id:   1,
			mockSetup: func(m sqlmock.Sqlmock) {
				cols := []string{"id", "name", "email", "password_hash", "role", "status", "created_at", "updated_at", "reset_token"}
				rows := sqlmock.NewRows(cols).AddRow(
					int64(1), "Alice", "alice@example.com", "hash", "member", "active",
					time.Now(), time.Now(), sql.NullString{Valid: false},
				)
				m.ExpectQuery(regexp.QuoteMeta("SELECT id, name, email, password_hash, role, status, created_at, updated_at, reset_token FROM users\nWHERE id = $1 LIMIT 1")).
					WithArgs(int64(1)).
					WillReturnRows(rows)
			},
			expectedErr: false,
			expectedID:  1,
		},
		{
			name: "異常系: ユーザーが存在しない",
			id:   999,
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta("SELECT id, name, email, password_hash, role, status, created_at, updated_at, reset_token FROM users\nWHERE id = $1 LIMIT 1")).
					WithArgs(int64(999)).
					WillReturnError(sql.ErrNoRows)
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, mock, cleanup := setupMock(t)
			defer cleanup()

			tt.mockSetup(mock)

			user, err := q.GetUserByID(context.Background(), tt.id)
			if tt.expectedErr {
				assert.Error(t, err)
				assert.Equal(t, sql.ErrNoRows, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, user.ID)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestUpdateUserRole テーブル駆動テスト
func TestUpdateUserRole(t *testing.T) {
	tests := []struct {
		name         string
		id           int64
		newRole      string
		mockSetup    func(sqlmock.Sqlmock)
		expectedErr  bool
		expectedRole string
	}{
		{
			name:    "正常系: role を admin に更新",
			id:      2,
			newRole: "admin",
			mockSetup: func(m sqlmock.Sqlmock) {
				cols := []string{"id", "name", "email", "password_hash", "role", "status", "created_at", "updated_at", "reset_token"}
				rows := sqlmock.NewRows(cols).AddRow(
					int64(2), "Bob", "bob@example.com", "hash", "admin", "active",
					time.Now(), time.Now(), sql.NullString{Valid: false},
				)
				m.ExpectQuery(regexp.QuoteMeta("UPDATE users\nSET role = $1,\n    updated_at = NOW()\nWHERE id = $2\nRETURNING id, name, email, password_hash, role, status, created_at, updated_at, reset_token")).
					WithArgs("admin", int64(2)).
					WillReturnRows(rows)
			},
			expectedErr:  false,
			expectedRole: "admin",
		},
		{
			name:    "異常系: ユーザーが存在しない",
			id:      999,
			newRole: "admin",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta("UPDATE users\nSET role = $1,\n    updated_at = NOW()\nWHERE id = $2\nRETURNING id, name, email, password_hash, role, status, created_at, updated_at, reset_token")).
					WithArgs("admin", int64(999)).
					WillReturnError(sql.ErrNoRows)
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, mock, cleanup := setupMock(t)
			defer cleanup()

			tt.mockSetup(mock)

			res, err := q.UpdateUserRole(context.Background(), db.UpdateUserRoleParams{
				Role: tt.newRole,
				ID:   tt.id,
			})
			if tt.expectedErr {
				assert.Error(t, err)
				assert.Equal(t, sql.ErrNoRows, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRole, res.Role)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestSetResetToken テーブル駆動テスト
func TestSetResetToken(t *testing.T) {
	tests := []struct {
		name               string
		id                 int64
		token              sql.NullString
		mockSetup          func(sqlmock.Sqlmock)
		expectedErr        bool
		expectedTokenValid bool
		expectedToken      string
	}{
		{
			name:  "正常系: トークンを保存",
			id:    3,
			token: sql.NullString{String: "tok123abc", Valid: true},
			mockSetup: func(m sqlmock.Sqlmock) {
				cols := []string{"id", "name", "email", "password_hash", "role", "status", "created_at", "updated_at", "reset_token"}
				rows := sqlmock.NewRows(cols).AddRow(
					int64(3), "Carol", "carol@example.com", "hash", "member", "active",
					time.Now(), time.Now(), sql.NullString{String: "tok123abc", Valid: true},
				)
				m.ExpectQuery(regexp.QuoteMeta("UPDATE users\nSET reset_token = $1,\n    updated_at = NOW()\nWHERE id = $2\nRETURNING id, name, email, password_hash, role, status, created_at, updated_at, reset_token")).
					WithArgs(sql.NullString{String: "tok123abc", Valid: true}, int64(3)).
					WillReturnRows(rows)
			},
			expectedErr:        false,
			expectedTokenValid: true,
			expectedToken:      "tok123abc",
		},
		{
			name:  "正常系: トークンを NULL（クリア）",
			id:    3,
			token: sql.NullString{Valid: false},
			mockSetup: func(m sqlmock.Sqlmock) {
				cols := []string{"id", "name", "email", "password_hash", "role", "status", "created_at", "updated_at", "reset_token"}
				rows := sqlmock.NewRows(cols).AddRow(
					int64(3), "Carol", "carol@example.com", "hash", "member", "active",
					time.Now(), time.Now(), sql.NullString{Valid: false},
				)
				m.ExpectQuery(regexp.QuoteMeta("UPDATE users\nSET reset_token = $1,\n    updated_at = NOW()\nWHERE id = $2\nRETURNING id, name, email, password_hash, role, status, created_at, updated_at, reset_token")).
					WithArgs(sql.NullString{Valid: false}, int64(3)).
					WillReturnRows(rows)
			},
			expectedErr:        false,
			expectedTokenValid: false,
		},
		{
			name:  "異常系: ユーザーが存在しない",
			id:    999,
			token: sql.NullString{String: "tok", Valid: true},
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta("UPDATE users\nSET reset_token = $1,\n    updated_at = NOW()\nWHERE id = $2\nRETURNING id, name, email, password_hash, role, status, created_at, updated_at, reset_token")).
					WithArgs(sql.NullString{String: "tok", Valid: true}, int64(999)).
					WillReturnError(sql.ErrNoRows)
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, mock, cleanup := setupMock(t)
			defer cleanup()

			tt.mockSetup(mock)

			res, err := q.SetResetToken(context.Background(), db.SetResetTokenParams{
				ResetToken: tt.token,
				ID:         tt.id,
			})
			if tt.expectedErr {
				assert.Error(t, err)
				assert.Equal(t, sql.ErrNoRows, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTokenValid, res.ResetToken.Valid)
				if tt.expectedTokenValid {
					assert.Equal(t, tt.expectedToken, res.ResetToken.String)
				}
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
