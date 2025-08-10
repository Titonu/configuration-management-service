package validator

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONSchemaValidator_ValidateSchemaDefinition(t *testing.T) {
	validator := NewJSONSchemaValidator()

	t.Run("ValidSchema", func(t *testing.T) {
		// Valid JSON Schema
		schema := json.RawMessage(`{
			"type": "object",
			"properties": {
				"name": { "type": "string" },
				"age": { "type": "integer", "minimum": 0 }
			},
			"required": ["name"]
		}`)

		err := validator.ValidateSchemaDefinition(schema)
		assert.NoError(t, err)
	})

	t.Run("InvalidSchema", func(t *testing.T) {
		// Invalid JSON Schema (invalid type)
		schema := json.RawMessage(`{
			"type": "invalid-type",
			"properties": {
				"name": { "type": "string" }
			}
		}`)

		err := validator.ValidateSchemaDefinition(schema)
		assert.Error(t, err)
	})

	t.Run("MalformedJSON", func(t *testing.T) {
		// Malformed JSON
		schema := json.RawMessage(`{
			"type": "object",
			"properties": {
				"name": { "type": "string" },
			}
		}`)

		err := validator.ValidateSchemaDefinition(schema)
		assert.Error(t, err)
	})

	t.Run("EmptySchema", func(t *testing.T) {
		// Empty schema
		schema := json.RawMessage(`{}`)

		err := validator.ValidateSchemaDefinition(schema)
		assert.NoError(t, err) // Empty schema is valid
	})
}

func TestJSONSchemaValidator_ValidateJSON(t *testing.T) {
	validator := NewJSONSchemaValidator()

	t.Run("ValidData", func(t *testing.T) {
		// Schema
		schema := json.RawMessage(`{
			"type": "object",
			"properties": {
				"name": { "type": "string" },
				"age": { "type": "integer", "minimum": 0 }
			},
			"required": ["name"]
		}`)

		// Valid data
		data := json.RawMessage(`{
			"name": "John Doe",
			"age": 30
		}`)

		err := validator.ValidateJSON(schema, data)
		assert.NoError(t, err)
	})

	t.Run("MissingRequiredField", func(t *testing.T) {
		// Schema with required field
		schema := json.RawMessage(`{
			"type": "object",
			"properties": {
				"name": { "type": "string" },
				"age": { "type": "integer", "minimum": 0 }
			},
			"required": ["name"]
		}`)

		// Data missing required field
		data := json.RawMessage(`{
			"age": 30
		}`)

		err := validator.ValidateJSON(schema, data)
		assert.Error(t, err)
	})

	t.Run("TypeMismatch", func(t *testing.T) {
		// Schema
		schema := json.RawMessage(`{
			"type": "object",
			"properties": {
				"name": { "type": "string" },
				"age": { "type": "integer", "minimum": 0 }
			}
		}`)

		// Data with type mismatch
		data := json.RawMessage(`{
			"name": "John Doe",
			"age": "thirty"
		}`)

		err := validator.ValidateJSON(schema, data)
		assert.Error(t, err)
	})

	t.Run("MinimumConstraintViolation", func(t *testing.T) {
		// Schema with minimum constraint
		schema := json.RawMessage(`{
			"type": "object",
			"properties": {
				"name": { "type": "string" },
				"age": { "type": "integer", "minimum": 0 }
			}
		}`)

		// Data violating minimum constraint
		data := json.RawMessage(`{
			"name": "John Doe",
			"age": -5
		}`)

		err := validator.ValidateJSON(schema, data)
		assert.Error(t, err)
	})

	t.Run("InvalidSchema", func(t *testing.T) {
		// Invalid schema
		schema := json.RawMessage(`{
			"type": "invalid-type"
		}`)

		// Valid data
		data := json.RawMessage(`{
			"name": "John Doe"
		}`)

		err := validator.ValidateJSON(schema, data)
		assert.Error(t, err)
	})

	t.Run("InvalidData", func(t *testing.T) {
		// Valid schema
		schema := json.RawMessage(`{
			"type": "object",
			"properties": {
				"name": { "type": "string" }
			}
		}`)

		// Invalid JSON data
		data := json.RawMessage(`{
			"name": "John Doe",
		}`)

		err := validator.ValidateJSON(schema, data)
		assert.Error(t, err)
	})

	t.Run("NestedObjectValidation", func(t *testing.T) {
		// Schema with nested object
		schema := json.RawMessage(`{
			"type": "object",
			"properties": {
				"name": { "type": "string" },
				"address": {
					"type": "object",
					"properties": {
						"street": { "type": "string" },
						"city": { "type": "string" },
						"zipCode": { "type": "string", "pattern": "^\\d{5}$" }
					},
					"required": ["street", "city", "zipCode"]
				}
			}
		}`)

		// Valid nested data
		validData := json.RawMessage(`{
			"name": "John Doe",
			"address": {
				"street": "123 Main St",
				"city": "Anytown",
				"zipCode": "12345"
			}
		}`)

		err := validator.ValidateJSON(schema, validData)
		assert.NoError(t, err)

		// Invalid nested data (invalid zip code)
		invalidData := json.RawMessage(`{
			"name": "John Doe",
			"address": {
				"street": "123 Main St",
				"city": "Anytown",
				"zipCode": "1234"
			}
		}`)

		err = validator.ValidateJSON(schema, invalidData)
		assert.Error(t, err)
	})

	t.Run("ArrayValidation", func(t *testing.T) {
		// Schema with array
		schema := json.RawMessage(`{
			"type": "object",
			"properties": {
				"name": { "type": "string" },
				"tags": {
					"type": "array",
					"items": { "type": "string" },
					"minItems": 1
				}
			}
		}`)

		// Valid array data
		validData := json.RawMessage(`{
			"name": "John Doe",
			"tags": ["developer", "golang"]
		}`)

		err := validator.ValidateJSON(schema, validData)
		assert.NoError(t, err)

		// Invalid array data (empty array)
		invalidData := json.RawMessage(`{
			"name": "John Doe",
			"tags": []
		}`)

		err = validator.ValidateJSON(schema, invalidData)
		assert.Error(t, err)

		// Invalid array data (wrong type in array)
		invalidTypeData := json.RawMessage(`{
			"name": "John Doe",
			"tags": ["developer", 123]
		}`)

		err = validator.ValidateJSON(schema, invalidTypeData)
		assert.Error(t, err)
	})
}
