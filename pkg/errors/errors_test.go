package errors

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppError(t *testing.T) {
	t.Run("NewNotFoundError", func(t *testing.T) {
		err := NewNotFoundError("resource", "123")

		assert.Equal(t, "resource not found", err.Error())
		assert.Equal(t, ErrorCodeNotFound, err.Code)

		details, ok := err.Details.(map[string]string)
		assert.True(t, ok)
		assert.Equal(t, "123", details["id"])
	})

	t.Run("NewAlreadyExistsError", func(t *testing.T) {
		err := NewAlreadyExistsError("resource", "123")

		assert.Equal(t, "resource already exists", err.Error())
		assert.Equal(t, ErrorCodeAlreadyExists, err.Code)

		details, ok := err.Details.(map[string]string)
		assert.True(t, ok)
		assert.Equal(t, "123", details["id"])
	})

	t.Run("NewInvalidRequestError", func(t *testing.T) {
		details := map[string]string{"field": "name"}
		err := NewInvalidRequestError("invalid request", details)

		assert.Equal(t, "invalid request", err.Error())
		assert.Equal(t, ErrorCodeInvalidRequest, err.Code)
		assert.Equal(t, details, err.Details)
	})

	t.Run("NewValidationFailedError", func(t *testing.T) {
		details := []ValidationError{
			{Field: "name", Reason: "required"},
		}
		err := NewValidationFailedError("validation failed", details)

		assert.Equal(t, "validation failed", err.Error())
		assert.Equal(t, ErrorCodeValidationFailed, err.Code)
		assert.Equal(t, details, err.Details)
	})

	t.Run("NewInternalError", func(t *testing.T) {
		err := NewInternalError("internal error", nil)

		assert.Equal(t, "internal error", err.Error())
		assert.Equal(t, ErrorCodeInternalError, err.Code)
		assert.Nil(t, err.Details)
	})
}

func TestAppError_Error(t *testing.T) {
	err := NewAppError("test error", ErrorCodeNotFound, nil)
	assert.Equal(t, "test error", err.Error())
}

func TestAppError_ToErrorResponse(t *testing.T) {
	details := map[string]string{"field": "name"}
	err := NewAppError("test error", ErrorCodeValidationFailed, details)

	resp := err.ToErrorResponse()

	assert.Equal(t, "test error", resp.Error)
	assert.Equal(t, ErrorCodeValidationFailed, resp.Code)
	assert.Equal(t, details, resp.Details)
}

func TestNewErrorResponse(t *testing.T) {
	details := map[string]string{"field": "name"}
	resp := NewErrorResponse("test error", ErrorCodeValidationFailed, details)

	assert.Equal(t, "test error", resp.Error)
	assert.Equal(t, ErrorCodeValidationFailed, resp.Code)
	assert.Equal(t, details, resp.Details)
}

func TestNewValidationError(t *testing.T) {
	valErr := NewValidationError("name", "required")

	assert.Equal(t, "name", valErr.Field)
	assert.Equal(t, "required", valErr.Reason)
}

func TestErrorResponse_ToJSON(t *testing.T) {
	details := map[string]string{"field": "name"}
	resp := NewErrorResponse("test error", ErrorCodeValidationFailed, details)

	jsonBytes, err := resp.ToJSON()
	assert.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	assert.NoError(t, err)

	assert.Equal(t, "test error", result["error"])
	assert.Equal(t, string(ErrorCodeValidationFailed), result["code"])

	detailsMap, ok := result["details"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "name", detailsMap["field"])
}
