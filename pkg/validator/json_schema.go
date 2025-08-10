package validator

import (
	"encoding/json"
	"fmt"

	"github.com/Titonu/configuration-management-service/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

// JSONSchemaValidator provides JSON schema validation functionality
type JSONSchemaValidator struct{}

// NewJSONSchemaValidator creates a new JSON schema validator
func NewJSONSchemaValidator() *JSONSchemaValidator {
	return &JSONSchemaValidator{}
}

// ValidateJSON validates JSON data against a schema
func (v *JSONSchemaValidator) ValidateJSON(schema json.RawMessage, data json.RawMessage) error {
	// Parse schema
	schemaLoader := gojsonschema.NewStringLoader(string(schema))

	// Parse data
	dataLoader := gojsonschema.NewStringLoader(string(data))

	// Validate
	result, err := gojsonschema.Validate(schemaLoader, dataLoader)
	if err != nil {
		return errors.NewInternalError("Failed to validate JSON", err.Error())
	}

	// Check validation result
	if !result.Valid() {
		// Collect validation errors
		validationErrors := make([]errors.ValidationError, 0)
		for _, desc := range result.Errors() {
			validationErrors = append(validationErrors, errors.ValidationError{
				Field:  desc.Field(),
				Reason: desc.Description(),
			})
		}

		return errors.NewValidationFailedError(
			"JSON validation failed",
			validationErrors,
		)
	}

	return nil
}

// ValidateSchemaDefinition validates that a schema definition is valid JSON Schema
func (v *JSONSchemaValidator) ValidateSchemaDefinition(schema json.RawMessage) error {
	// Parse schema
	schemaLoader := gojsonschema.NewStringLoader(string(schema))

	// Compile schema to check if it's valid
	_, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return errors.NewInvalidRequestError(
			"Invalid JSON Schema",
			fmt.Sprintf("Schema validation error: %s", err.Error()),
		)
	}

	return nil
}
