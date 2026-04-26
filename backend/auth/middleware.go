package auth

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/pkg/apperror"
	"sol_coffeesys/backend/pkg/respond"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AdminOnly(queries db.Querier) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, err := tokenFromRequest(c)
		if err != nil {
			respond.RespondWithError(c, apperror.NewUnauthorizedError("", apperror.UnauthorizedMessageAuth))
			c.Abort()
			return
		}
		token, err := Validate(tokenStr)
		if err != nil || token == nil || !token.Valid {
			respond.RespondWithError(c, apperror.NewUnauthorizedError("", apperror.UnauthorizedMessageAuth))
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			respond.RespondWithError(c, apperror.NewUnauthorizedError("", apperror.UnauthorizedMessageAuth))
			c.Abort()
			return
		}

		rawID, ok := claims["user.id"]
		if !ok {
			respond.RespondWithError(c, apperror.NewUnauthorizedError("", apperror.UnauthorizedMessageAuth))
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
				respond.RespondWithError(c, apperror.NewUnauthorizedError("", apperror.UnauthorizedMessageAuth))
				c.Abort()
				return
			}
			userID = id
		}

		user, err := queries.GetUserForUpdate(c.Request.Context(), userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				respond.RespondWithError(c, apperror.NewUnauthorizedError("", apperror.UnauthorizedMessageAuth))
				c.Abort()
				return
			}
			respond.RespondWithError(c, apperror.NewInternalError("GetUserForUpdate", err, apperror.InternalServerMessageCommon))
			c.Abort()
			return
		}

		if user.Role != "admin" {
			respond.RespondWithError(c, apperror.NewForbiddenError("admin", "user", apperror.ForbiddenMessageAdmin))
			c.Abort()
			return
		}

		c.Set("userID", user.ID)
		c.Next()

	}
}

func RequireAuth(queries db.Querier) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, err := tokenFromRequest(c)
		if err != nil {
			respond.RespondWithError(c, apperror.NewUnauthorizedError("", apperror.UnauthorizedMessageAuth))
			c.Abort()
			return
		}

		token, err := Validate(tokenStr)
		if err != nil || token == nil || !token.Valid {
			respond.RespondWithError(c, apperror.NewUnauthorizedError("", apperror.UnauthorizedMessageAuth))
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			respond.RespondWithError(c, apperror.NewUnauthorizedError("", apperror.UnauthorizedMessageAuth))
			c.Abort()
			return
		}

		rawID, ok := claims["user.id"]
		if !ok {
			respond.RespondWithError(c, apperror.NewUnauthorizedError("", apperror.UnauthorizedMessageAuth))
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
				respond.RespondWithError(c, apperror.NewUnauthorizedError("", apperror.UnauthorizedMessageAuth))
				c.Abort()
				return
			}
			userID = id
		default:
			respond.RespondWithError(c, apperror.NewUnauthorizedError("", apperror.UnauthorizedMessageAuth))
			c.Abort()
			return
		}

		user, err := queries.GetUserForUpdate(c.Request.Context(), userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				respond.RespondWithError(c, apperror.NewUnauthorizedError("", apperror.UnauthorizedMessageAuth))
				c.Abort()
				return
			}
			respond.RespondWithError(c, apperror.NewInternalError("GetUserForUpdate", err, apperror.InternalServerMessageCommon))
			c.Abort()
			return
		}
		c.Set("userID", user.ID)
		c.Next()
	}
}

func tokenFromRequest(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return parts[1], nil
		}
	}
	if cookie, err := c.Request.Cookie("access_token"); err == nil {
		return cookie.Value, nil
	}
	return "", errors.New("no token")
}
