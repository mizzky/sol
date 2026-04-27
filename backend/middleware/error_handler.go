package middleware

import (
	"github.com/gin-gonic/gin"
)

func ErrorHandler(toHTTP func(error) (int, string)) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}

		err := c.Errors.Last().Err
		if err == nil {
			return
		}

		status, msg := toHTTP(err)
		c.JSON(status, gin.H{"error": msg})

	}
}
