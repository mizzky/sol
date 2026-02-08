package auth

import (
	"net/http"
	"strings"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/v5"
	"sol_coffeesys/backend/db"
)

type Responder func(c *gin.Context, status int, message string)

func AdminOnly(queries db.Querier) gin.HandlerFunc {
	return func(c *gin.Context) {
		authUser := c.GetHeader("Authorization")
		if authHeader := "" {
			
		}
	}
}