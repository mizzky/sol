package handler

import "github.com/gin-gonic/gin"

type ErrorResponse struct {
	Error string `json:"error"`
}

type ValidationError struct {
	Error  string                 `json:"error"`
	Fields map[string]interface{} `json:"fields,omitempty"`
}

func RespondError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, ErrorResponse{Error: message})
}
