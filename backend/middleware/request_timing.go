package middleware

import (
	"sol_coffeesys/backend/pkg/logging"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestStartedAtMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(logging.CtxKeyRequestStartedAt, time.Now())
		c.Next()
	}
}
