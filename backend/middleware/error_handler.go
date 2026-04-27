package middleware

import (
	"sol_coffeesys/backend/pkg/respond"

	"github.com/gin-gonic/gin"
)

type ErrorResponder interface {
	ToHTTP(err error) (int, string)
}

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

		respond.RespondWithError(c, err)

		// status, msg := toHTTP(err)
		// c.JSON(status, gin.H{"error": msg})

	}
}
