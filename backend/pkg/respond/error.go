package respond

import (
	"sol_coffeesys/backend/pkg/apperror"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type ValidationError struct {
	Error  string                 `json:"error"`
	Fields map[string]interface{} `json:"fields,omitempty"`
}

func RespondError(c *gin.Context, status int, message string) {
	c.JSON(status, ErrorResponse{Error: message})
}

func RespondWithError(c *gin.Context, err error) {
	status, msg := apperror.ToHTTP(err)
	c.JSON(status, ErrorResponse{Error: msg})
}
