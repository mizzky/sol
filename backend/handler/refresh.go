package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/http"
	"sol_coffeesys/backend/auth"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/pkg/respond"
	"time"

	"github.com/gin-gonic/gin"
)

func RefreshTokenHandler(q db.Querier, tokenGenerator auth.TokenGenerator) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Request.Cookie("refresh_token")
		if err != nil {
			respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
			return
		}
		sum := sha256.Sum256([]byte(cookie.Value))
		hash := hex.EncodeToString(sum[:])

		rt, err := q.GetRefreshTokenByHash(c.Request.Context(), hash)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
			} else {
				respond.RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
			}
			c.Abort()
			return
		}

		if rt.RevokedAt.Valid || rt.ExpiresAt.Before(time.Now()) {
			respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
			c.Abort()
			return
		}

		user, err := q.GetUserByID(c.Request.Context(), rt.UserID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
			} else {
				respond.RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
			}
			c.Abort()
			return
		}

		newRaw := make([]byte, 32)
		if _, err := rand.Read(newRaw); err != nil {
			respond.RespondError(c, http.StatusInternalServerError, "トークンの生成に失敗しました")
			c.Abort()
			return
		}
		newRefresh := hex.EncodeToString(newRaw)
		newSum := sha256.Sum256([]byte(newRefresh))
		newHash := hex.EncodeToString(newSum[:])

		expiresAt := time.Now().Add(14 * 24 * time.Hour)

		if _, err := q.CreateRefreshToken(c.Request.Context(), db.CreateRefreshTokenParams{
			UserID:    user.ID,
			TokenHash: newHash,
			ExpiresAt: expiresAt,
		}); err != nil {
			respond.RespondError(c, http.StatusInternalServerError, "リフレッシュトークンの保存に失敗しました")
			c.Abort()
			return
		}

		if err := q.RevokeRefreshTokenByHash(c.Request.Context(), hash); err != nil {
			respond.RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
			c.Abort()
			return
		}

		accessToken, err := tokenGenerator.GenerateToken(user.ID)
		if err != nil {
			respond.RespondError(c, http.StatusInternalServerError, "トークンの生成に失敗しました")
			c.Abort()
			return
		}

		accessCookie := &http.Cookie{
			Name:     "access_token",
			Value:    accessToken,
			HttpOnly: true,
			Path:     "/",
			Expires:  time.Now().Add(15 * time.Minute),
			MaxAge:   15 * 60,
			SameSite: http.SameSiteLaxMode,
		}
		refreshCookie := &http.Cookie{
			Name:     "refresh_token",
			Value:    newRefresh,
			HttpOnly: true,
			Path:     "/api/refresh",
			Expires:  expiresAt,
			MaxAge:   14 * 24 * 60 * 60,
			SameSite: http.SameSiteStrictMode,
		}
		if gin.Mode() == gin.ReleaseMode {
			accessCookie.Secure = true
			refreshCookie.Secure = true
		}

		http.SetCookie(c.Writer, accessCookie)
		http.SetCookie(c.Writer, refreshCookie)

		c.JSON(http.StatusOK, gin.H{
			"message": "トークンを更新しました",
			"token":   accessToken,
			"user": gin.H{
				"id":    user.ID,
				"name":  user.Name,
				"email": user.Email,
				"role":  user.Role,
			},
		})

	}
}
