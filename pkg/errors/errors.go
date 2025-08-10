package errors

import (
	"encoding/json"
	"fmt"
)

// ErrorCode represents standardized error codes for the API
type ErrorCode string

// Error codes
const (
	ErrorCodeNotFound         ErrorCode = "NOT_FOUND"
	ErrorCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrorCodeAlreadyExists    ErrorCode = "ALREADY_EXISTS"
	ErrorCodeInvalidRequest   ErrorCode = "INVALID_REQUEST"
	ErrorCodeInternalError    ErrorCode = "INTERNAL_ERROR"
	ErrorCodeUnauthorized     ErrorCode = "UNAUTHORIZED"
)

// ErrorResponse represents a standardized API error response
type ErrorResponse struct {
	Error   string      `json:"error"`
	Details interface{} `json:"details,omitempty"`
	Code    ErrorCode   `json:"code"`
}

// ValidationError represents a validation error detail
type ValidationError struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

// AppError is a custom error type that includes error code and details
type AppError struct {
	Message string
	Code    ErrorCode
	Details interface{}
}

// Error implements the error interface for AppError
func (e *AppError) Error() string {
	return e.Message
}

// ToErrorResponse converts AppError to an ErrorResponse
func (e *AppError) ToErrorResponse() *ErrorResponse {
	return &ErrorResponse{
		Error:   e.Message,
		Code:    e.Code,
		Details: e.Details,
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(message string, code ErrorCode, details interface{}) *ErrorResponse {
	return &ErrorResponse{
		Error:   message,
		Code:    code,
		Details: details,
	}
}

// NewAppError creates a new application error
func NewAppError(message string, code ErrorCode, details interface{}) *AppError {
	return &AppError{
		Message: message,
		Code:    code,
		Details: details,
	}
}

// NewValidationError creates a new validation error
func NewValidationError(field, reason string) *ValidationError {
	return &ValidationError{
		Field:  field,
		Reason: reason,
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resourceType, resourceID string) *AppError {
	return NewAppError(
		fmt.Sprintf("%s not found", resourceType),
		ErrorCodeNotFound,
		map[string]string{"id": resourceID},
	)
}

// NewAlreadyExistsError creates an already exists error
func NewAlreadyExistsError(resourceType, resourceID string) *AppError {
	return NewAppError(
		fmt.Sprintf("%s already exists", resourceType),
		ErrorCodeAlreadyExists,
		map[string]string{"id": resourceID},
	)
}

// NewInvalidRequestError creates an invalid request error
func NewInvalidRequestError(message string, details interface{}) *AppError {
	return NewAppError(
		message,
		ErrorCodeInvalidRequest,
		details,
	)
}

// NewValidationFailedError creates a validation failed error
func NewValidationFailedError(message string, details interface{}) *AppError {
	return NewAppError(
		message,
		ErrorCodeValidationFailed,
		details,
	)
}

// NewInternalError creates an internal error
func NewInternalError(message string, details interface{}) *AppError {
	return NewAppError(
		message,
		ErrorCodeInternalError,
		details,
	)
}

// ToJSON converts the error response to JSON
func (e *ErrorResponse) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}
