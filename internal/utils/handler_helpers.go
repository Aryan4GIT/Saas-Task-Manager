package utils

import (
	"net/http"

	"saas-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ParseUUID parses a UUID from a URL parameter and returns an error response if invalid
func ParseUUID(c *gin.Context, paramName string, errorLabel string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(paramName))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid " + errorLabel,
			Message: err.Error(),
		})
		return uuid.UUID{}, false
	}
	return id, true
}

// BindJSON binds JSON request body and returns an error response if invalid
func BindJSON(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid request",
			Message: err.Error(),
		})
		return false
	}
	return true
}

// HandlePermissionError handles permission-related errors uniformly
func HandlePermissionError(c *gin.Context, err error, defaultMsg string) {
	status := http.StatusBadRequest
	errMsg := defaultMsg
	if err.Error() == "insufficient permissions" {
		status = http.StatusForbidden
		errMsg = "insufficient permissions"
	}
	c.JSON(status, models.ErrorResponse{Error: errMsg, Message: err.Error()})
}

// RespondWithError sends a JSON error response
func RespondWithError(c *gin.Context, status int, error string, message string) {
	c.JSON(status, models.ErrorResponse{
		Error:   error,
		Message: message,
	})
}

// RespondWithSuccess sends a JSON success response
func RespondWithSuccess(c *gin.Context, status int, data interface{}) {
	c.JSON(status, data)
}

// RespondWithMessage sends a JSON message response
func RespondWithMessage(c *gin.Context, status int, message string) {
	c.JSON(status, models.SuccessResponse{
		Message: message,
	})
}
