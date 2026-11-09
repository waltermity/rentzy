package respond

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"rentzy/backend/internal/domain"
)

// ErrorBody is the shape of every error response in the API.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorEnvelope is {"error": {...}}; used in the API docs.
type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

// OK writes a 200 response with the given data under "data"
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// Created writes a 201 response with the given data under "data".
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, gin.H{"data": data})
}

// Error writes the standard error envelope with the given status, code and message.
func Error(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, gin.H{"error": ErrorBody{Code: code, Message: message}})
}

// Err maps a service/repository error to a status code and error code.
// Unknown errors are logged and returned as 500 without details.
func Err(c *gin.Context, err error) {
	type mapping struct {
		target error
		status int
		code   string
	}
	for _, m := range []mapping{
		{domain.ErrNotFound, http.StatusNotFound, "not_found"},
		{domain.ErrInvalidInput, http.StatusBadRequest, "invalid_input"},
		{domain.ErrInvalidCredentials, http.StatusUnauthorized, "invalid_credentials"},
		{domain.ErrUnauthorized, http.StatusUnauthorized, "unauthorized"},
		{domain.ErrForbidden, http.StatusForbidden, "forbidden"},
		{domain.ErrConflict, http.StatusConflict, "conflict"},
		{domain.ErrUnavailable, http.StatusConflict, "unavailable"},
		{domain.ErrInvalidState, http.StatusConflict, "invalid_state"},
	} {
		if errors.Is(err, m.target) {
			Error(c, m.status, m.code, err.Error())
			return
		}
	}
	log.Printf("internal error on %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
	Error(c, http.StatusInternalServerError, "internal", "internal server error")
}
