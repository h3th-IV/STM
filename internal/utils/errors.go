package utils

import "fmt"

// AppError represents an application error with HTTP status.
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Predefined errors.
var (
	ErrBadRequest          = &AppError{Code: 400, Message: "Bad request"}
	ErrUnauthorized        = &AppError{Code: 401, Message: "Unauthorized"}
	ErrForbidden           = &AppError{Code: 403, Message: "Forbidden"}
	ErrNotFound            = &AppError{Code: 404, Message: "Resource not found"}
	ErrConflict            = &AppError{Code: 409, Message: "Resource already exists"}
	ErrInternalServerError = &AppError{Code: 500, Message: "Internal server error"}
)

// NewAppError creates an AppError with optional wrapping.
func NewAppError(code int, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}
