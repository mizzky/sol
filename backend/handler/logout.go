package handler

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/http"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/pkg/respond"

	"github.com/gin-gonic/gin"
)

func LogoutHandler(queries db.Querier) gin.HandlerFunc {
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
			c.JSON(http.StatusOK, gin.H{"message": "ログアウトしました"})
			return
		}

		sum := sha256.Sum256([]byte(cookie.Value))
		hash := hex.EncodeToString(sum[:])

		if err := queries.RevokeRefreshTokenByHash(c.Request.Context(), hash); err != nil {

			if errors.Is(err, sql.ErrNoRows) {

			} else {
				respond.RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
				c.Abort()
				return
			}
		}

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
		c.JSON(http.StatusOK, gin.H{"message": "ログアウトしました"})
	}
}
