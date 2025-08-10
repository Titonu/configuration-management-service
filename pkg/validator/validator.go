package validator

import (
	"encoding/json"
	"fmt"
	"github.com/Titonu/configuration-management-service/pkg/errors"

	"github.com/xeipuuv/gojsonschema"
)

// Validator is an interface for JSON schema validation
type Validator interface {
	ValidateJSON(schema json.RawMessage, data json.RawMessage) error
	ValidateSchemaDefinition(schema json.RawMessage) error
}

// SchemaValidator handles JSON schema validation
type SchemaValidator struct {
	schemas    map[string]*gojsonschema.Schema
	schemaJSON map[string]json.RawMessage
}

// NewSchemaValidator creates a new schema validator
func NewSchemaValidator() *SchemaValidator {
	return &SchemaValidator{
		schemas:    make(map[string]*gojsonschema.Schema),
		schemaJSON: make(map[string]json.RawMessage),
	}
}

// RegisterSchema adds a new schema for a configuration type
func (v *SchemaValidator) RegisterSchema(configName string, schemaJSON []byte) error {
	// Parse the schema
	schemaLoader := gojsonschema.NewBytesLoader(schemaJSON)
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return fmt.Errorf("invalid schema for %s: %w", configName, err)
	}

	// Store the schema and original JSON
	v.schemas[configName] = schema
	v.schemaJSON[configName] = json.RawMessage(schemaJSON)
	return nil
}

// Validate validates configuration data against its schema
func (v *SchemaValidator) Validate(configName string, data json.RawMessage) ([]*errors.ValidationError, error) {
	// Check if schema exists
	schema, exists := v.schemas[configName]
	if !exists {
		return nil, fmt.Errorf("no schema registered for configuration: %s", configName)
	}

	// Create a document loader for the data
	documentLoader := gojsonschema.NewBytesLoader(data)

	// Validate
	result, err := schema.Validate(documentLoader)
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// If valid, return nil
	if result.Valid() {
		return nil, nil
	}

	// Convert validation errors to our model
	validationErrors := make([]*errors.ValidationError, 0, len(result.Errors()))
	for _, err := range result.Errors() {
		validationErrors = append(validationErrors, &errors.ValidationError{
			Field:  err.Field(),
			Reason: err.Description(),
		})
	}

	return validationErrors, nil
}

// HasSchema checks if a schema exists for a configuration
func (v *SchemaValidator) HasSchema(configName string) bool {
	_, exists := v.schemas[configName]
	return exists
}

// GetSchema returns the schema for a configuration
func (v *SchemaValidator) GetSchema(configName string) (json.RawMessage, error) {
	_, exists := v.schemas[configName]
	if !exists {
		return nil, fmt.Errorf("no schema registered for configuration: %s", configName)
	}

	// Return the stored original schema JSON
	return v.schemaJSON[configName], nil
}
