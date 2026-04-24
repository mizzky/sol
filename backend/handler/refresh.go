package handler

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/http"
	"sol_coffeesys/backend/auth"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/pkg/apperror"
	"sol_coffeesys/backend/pkg/respond"
	"time"

	"github.com/gin-gonic/gin"
)

func RefreshTokenHandler(q db.Querier, tokenGenerator auth.TokenGenerator) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Request.Cookie("refresh_token")
		if err != nil {
			respond.RespondWithError(c, apperror.NewUnauthorizedError("invalid_refresh_token", apperror.UnauthorizedMessageAuth))
			return
		}
		sum := sha256.Sum256([]byte(cookie.Value))
		hash := hex.EncodeToString(sum[:])

		rt, err := q.GetRefreshTokenByHash(c.Request.Context(), hash)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				respond.RespondWithError(c, apperror.NewUnauthorizedError("token_not_found", apperror.UnauthorizedMessageAuth))
			} else {
				respond.RespondWithError(c, apperror.NewInternalError("GetRefreshTokenByHash", err, apperror.InternalServerMessageCommon))
			}
			c.Abort()
			return
		}

		if rt.RevokedAt.Valid || rt.ExpiresAt.Before(time.Now()) {
			respond.RespondWithError(c, apperror.NewUnauthorizedError("refresh_token_revoked_alredy", apperror.UnauthorizedMessageAuth))
			c.Abort()
			return
		}

		user, err := q.GetUserByID(c.Request.Context(), rt.UserID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				respond.RespondWithError(c, apperror.NewUnauthorizedError("token_not_found", apperror.UnauthorizedMessageAuth))
			} else {
				respond.RespondWithError(c, apperror.NewInternalError("GetUserByID", err, apperror.InternalServerMessageCommon))
			}
			c.Abort()
			return
		}

		newRefresh, _, expiresAt, err := GenerateRefreshToken(c.Request.Context(), q, user.ID)
		if err != nil {
			respond.RespondWithError(c, apperror.NewInternalError("GenerateRefresToken", err, apperror.InternalServerMessageRefresh))
			c.Abort()
			return
		}

		if err := RevokeRefreshByRaw(c.Request.Context(), q, cookie.Value); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// 存在しないトークンは無視
			} else {
				respond.RespondWithError(c, apperror.NewInternalError("RevokeRefreshByRaw", err, apperror.InternalServerMessageCommon))
				c.Abort()
				return
			}
		}

		accessToken, err := tokenGenerator.GenerateToken(user.ID)
		if err != nil {
			respond.RespondWithError(c, apperror.NewInternalError("GenerateToken", err, apperror.InternalServerMessageGenToken))
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

func RevokeRefreshHandler(q db.Querier) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Request.Cookie("refresh_token")
		if err != nil {
			clearAccess := &http.Cookie{
				Name:     "access_token",
				Value:    "",
				HttpOnly: true,
				Path:     "/",
				MaxAge:   -1,
				SameSite: http.SameSiteLaxMode,
			}
			clearRefresh := &http.Cookie{
				Name:     "refresh_token",
				Value:    "",
				HttpOnly: true,
				Path:     "/api/refresh",
				MaxAge:   -1,
				SameSite: http.SameSiteStrictMode,
			}
			if gin.Mode() == gin.ReleaseMode {
				clearAccess.Secure = true
				clearRefresh.Secure = true
			}
			http.SetCookie(c.Writer, clearAccess)
			http.SetCookie(c.Writer, clearRefresh)
			c.JSON(http.StatusOK, gin.H{"message": "リフレッシュトークンを破棄しました"})
			return
		}

		// DB撤回
		if err := RevokeRefreshByRaw(c.Request.Context(), q, cookie.Value); err != nil {
			if errors.Is(err, sql.ErrNoRows) {

			} else {
				respond.RespondWithError(c, apperror.NewInternalError("RevokeRefreshByRaw", err, apperror.InternalServerMessageCommon))
				c.Abort()
				return
			}
		}

		// Cookie削除
		clearAccess := &http.Cookie{
			Name:     "access_token",
			Value:    "",
			HttpOnly: true,
			Path:     "/",
			MaxAge:   -1,
			SameSite: http.SameSiteLaxMode,
		}
		clearRefresh := &http.Cookie{
			Name:     "refresh_token",
			Value:    "",
			HttpOnly: true,
			Path:     "/api/refresh",
			MaxAge:   -1,
			SameSite: http.SameSiteStrictMode,
		}
		if gin.Mode() == gin.ReleaseMode {
			clearAccess.Secure = true
			clearRefresh.Secure = true
		}
		http.SetCookie(c.Writer, clearAccess)
		http.SetCookie(c.Writer, clearRefresh)
		c.JSON(http.StatusOK, gin.H{"message": "リフレッシュトークンを破棄しました"})
	}
}
