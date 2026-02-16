package handler

import (
	"database/sql"
	"net/http"
	"sol_coffeesys/backend/auth"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/pkg/respond"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func MeHandler(q db.Querier) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
			return
		}
		tokenStr := parts[1]

		token, err := auth.Validate(tokenStr)
		if err != nil || token == nil || !token.Valid {
			respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
			return
		}

		rawID, ok := claims["user.id"]
		if !ok {
			respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
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
				return
			}
			userID = id
		default:
			respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
			return
		}
		user, err := q.GetUserForUpdate(c.Request.Context(), userID)
		if err != nil {
			if err == sql.ErrNoRows {
				respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
				return
			}
			respond.RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"user": gin.H{
				"id":    user.ID,
				"name":  user.Name,
				"email": user.Email,
			},
		})
	}
}
