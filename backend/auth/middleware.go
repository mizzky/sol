package auth

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/pkg/respond"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AdminOnly(queries db.Querier) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
			c.Abort()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
			c.Abort()
			return
		}
		tokenStr := parts[1]

		token, err := Validate(tokenStr)
		if err != nil || token == nil || !token.Valid {
			respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
			c.Abort()
			return
		}

		rawID, ok := claims["user.id"]
		if !ok {
			respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
			c.Abort()
			return
		}

		var userID int64
		switch v := rawID.(type) {
		case float64:
			userID = int64(v)
		case string:
			id, perr := strconv.ParseInt(v, 10, 64)
			if perr != nil {
				respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
				c.Abort()
				return
			}
			userID = id
		}

		user, err := queries.GetUserForUpdate(c.Request.Context(), userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
				c.Abort()
				return
			}
			respond.RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
			c.Abort()
			return
		}

		if user.Role != "admin" {
			respond.RespondError(c, http.StatusForbidden, "管理者権限が必要です")
			c.Abort()
			return
		}

		c.Set("userID", user.ID)
		c.Next()

	}
}

func RequireAuth(queries db.Querier) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
			c.Abort()
			return
		}
		tokenStr := parts[1]

		token, err := Validate(tokenStr)
		if err != nil || token == nil || !token.Valid {
			respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
			c.Abort()
			return
		}

		rawID, ok := claims["user.id"]
		if !ok {
			respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
			c.Abort()
			return
		}

		var userID int64
		switch v := rawID.(type) {
		case int:
			userID = int64(v)
		case int32:
			userID = int64(v)
		case int64:
			userID = v
		case float64:
			userID = int64(v)
		case string:
			id, perr := strconv.ParseInt(v, 10, 64)
			if perr != nil {
				respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
				c.Abort()
				return
			}
			userID = id
		default:
			respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
			c.Abort()
			return
		}

		user, err := queries.GetUserForUpdate(c.Request.Context(), userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
				c.Abort()
				return
			}
			respond.RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
			c.Abort()
			return
		}
		c.Set("userID", user.ID)
		c.Next()
	}
}
