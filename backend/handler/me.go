package handler

import (
	"database/sql"
	"net/http"
	"sol_coffeesys/backend/auth"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/pkg/apperror"
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
			respond.RespondWithError(c, apperror.NewUnauthorizedError("token_not_found", apperror.UnauthorizedMessageAuth))
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			respond.RespondWithError(c, apperror.NewUnauthorizedError("invalid_format_token", apperror.UnauthorizedMessageAuth))
			return
		}
		tokenStr := parts[1]

		token, err := auth.Validate(tokenStr)
		if err != nil || token == nil || !token.Valid {
			respond.RespondWithError(c, apperror.NewUnauthorizedError("invalid_token", apperror.UnauthorizedMessageAuth))
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			respond.RespondWithError(c, apperror.NewUnauthorizedError("failed_to_decode_token", apperror.UnauthorizedMessageAuth))
			return
		}

		rawID, ok := claims["user.id"]
		if !ok {
			respond.RespondWithError(c, apperror.NewUnauthorizedError("userID_claims_not_found", apperror.UnauthorizedMessageAuth))
			return
		}

		var userID int64
		switch v := rawID.(type) {
		case float64:
			userID = int64(v)
		case string:
			id, perr := strconv.ParseInt(v, 10, 64)
			if perr != nil {
				respond.RespondWithError(c, apperror.NewUnauthorizedError("userID_parse_failed", apperror.UnauthorizedMessageAuth))
				return
			}
			userID = id
		default:
			respond.RespondWithError(c, apperror.NewUnauthorizedError("userID_type_is_invalid", apperror.UnauthorizedMessageAuth))
			return
		}
		user, err := q.GetUserForUpdate(c.Request.Context(), userID)
		if err != nil {
			if err == sql.ErrNoRows {
				respond.RespondWithError(c, apperror.NewUnauthorizedError("userID_is_not_authenticated", apperror.UnauthorizedMessageAuth))
				return
			}
			respond.RespondWithError(c, apperror.NewInternalError("GetUserForUpdate", err, apperror.InternalServerMessageCommon))
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"user": gin.H{
				"id":    user.ID,
				"name":  user.Name,
				"email": user.Email,
				"role":  user.Role,
			},
		})
	}
}
