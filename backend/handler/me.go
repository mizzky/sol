package handler

import (
	"database/sql"
	"log/slog"
	"net/http"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/pkg/apperror"
	"sol_coffeesys/backend/pkg/logging"
	"strconv"

	"github.com/gin-gonic/gin"
)

func MeHandler(q db.Querier) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, exist := c.Get("userID")
		if !exist {
			_ = c.Error(apperror.NewUnauthorizedError("token_not_found", apperror.UnauthorizedMessageAuth))
			return
		}

		var userID int64

		switch v := raw.(type) {
		case int:
			userID = int64(v)
		case int32:
			userID = int64(v)
		case int64:
			userID = v
		case float64:
			userID = int64(v)
		case string:
			id, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				_ = c.Error(apperror.NewUnauthorizedError("userID_parse_failed", apperror.UnauthorizedMessageAuth))
				return
			}
			userID = id
		default:
			_ = c.Error(apperror.NewUnauthorizedError("userID_type_is_invalid", apperror.UnauthorizedMessageAuth))
			return
		}
		user, err := q.GetUserForUpdate(c.Request.Context(), userID)
		if err != nil {
			if err == sql.ErrNoRows {
				_ = c.Error(apperror.NewUnauthorizedError("userID_is_not_authenticated", apperror.UnauthorizedMessageAuth))
				return
			}
			_ = c.Error(apperror.NewInternalError("GetUserForUpdate", err, apperror.InternalServerMessageCommon))
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

		c.Set("userID", userID)

		logging.LogEvent(c, logging.EventInput{
			Event:  "auth_me_fetched",
			Status: http.StatusOK,
			Level:  slog.LevelInfo,
		})
	}
}
