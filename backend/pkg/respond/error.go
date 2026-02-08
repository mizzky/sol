package respond

import "github.com/gin-gonic/gin"

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
