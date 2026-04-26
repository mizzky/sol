package respond

import (
	"sol_coffeesys/backend/pkg/apperror"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func RespondWithError(c *gin.Context, err error) {
	status, msg := apperror.ToHTTP(err)
	c.JSON(status, ErrorResponse{Error: msg})
}
