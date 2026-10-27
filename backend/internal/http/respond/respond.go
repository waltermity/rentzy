package respond

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorBody is the shape of every error response in the API.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// OK writes a 200 response with the given data under "data"
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// Error writes the standard error envelope with the given status, code and message.
func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": ErrorBody{Code: code, Message: message}})
}
