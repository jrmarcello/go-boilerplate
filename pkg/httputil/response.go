package httputil

import (
	"github.com/gin-gonic/gin"
)

// SuccessResponse is the standard envelope for successful API responses.
// Format: {"data": ..., "meta": ..., "links": ...}
type SuccessResponse struct {
	Data  any `json:"data"`
	Meta  any `json:"meta,omitempty"`
	Links any `json:"links,omitempty"`
}

// ErrorDetail holds error details for API error responses.
type ErrorDetail struct {
	Message string         `json:"message"`
	Code    string         `json:"code,omitempty"`
	Details map[string]any `json:"details,omitempty"`
}

// ErrorResponse is the standard envelope for API error responses.
// Format: {"errors": {"message": ..., "code": ..., "details": ...}}
type ErrorResponse struct {
	Errors ErrorDetail `json:"errors"`
}

// SendSuccess sends a standardized success response.
func SendSuccess(c *gin.Context, status int, data any) {
	c.JSON(status, SuccessResponse{Data: data})
}

// SendSuccessWithMeta sends a standardized success response with metadata and links.
func SendSuccessWithMeta(c *gin.Context, status int, data, meta, links any) {
	c.JSON(status, SuccessResponse{
		Data:  data,
		Meta:  meta,
		Links: links,
	})
}

// SendError sends a standardized error response.
func SendError(c *gin.Context, status int, message string) {
	c.JSON(status, ErrorResponse{
		Errors: ErrorDetail{Message: message},
	})
}

// SendErrorWithCode sends a standardized error response with an error code.
func SendErrorWithCode(c *gin.Context, status int, code, message string) {
	c.JSON(status, ErrorResponse{
		Errors: ErrorDetail{
			Message: message,
			Code:    code,
		},
	})
}

// SendErrorWithDetails sends a standardized error response with code and details.
func SendErrorWithDetails(c *gin.Context, status int, code, message string, details map[string]any) {
	c.JSON(status, ErrorResponse{
		Errors: ErrorDetail{
			Message: message,
			Code:    code,
			Details: details,
		},
	})
}
