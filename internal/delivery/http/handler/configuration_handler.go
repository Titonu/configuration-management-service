package handler

import (
	"encoding/json"
	stdErrors "errors"
	"github.com/Titonu/configuration-management-service/internal/domain/usecase"
	"github.com/Titonu/configuration-management-service/pkg/errors"

	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// ConfigurationHandler handles HTTP requests for configuration management
type ConfigurationHandler struct {
	configService usecase.ConfigurationUsecase
}

// NewConfigurationHandler creates a new configuration handler
func NewConfigurationHandler(configService usecase.ConfigurationUsecase) *ConfigurationHandler {
	return &ConfigurationHandler{
		configService: configService,
	}
}

// CreateConfiguration handles creating a new configuration
func (h *ConfigurationHandler) CreateConfiguration(c *gin.Context) {
	var req struct {
		Name string          `json:"name" binding:"required"`
		Data json.RawMessage `json:"data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(
			"Invalid request body",
			errors.ErrorCodeInvalidRequest,
			err.Error(),
		))
		return
	}

	config, err := h.configService.CreateConfiguration(req.Name, req.Data)
	if err != nil {
		var appErr *errors.AppError
		if stdErrors.As(err, &appErr) {
			switch appErr.Code {
			case errors.ErrorCodeAlreadyExists:
				c.JSON(http.StatusConflict, appErr.ToErrorResponse())
			case errors.ErrorCodeValidationFailed:
				c.JSON(http.StatusBadRequest, appErr.ToErrorResponse())
			default:
				c.JSON(http.StatusInternalServerError, appErr.ToErrorResponse())
			}
		} else {
			c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(
				"Failed to create configuration",
				errors.ErrorCodeInternalError,
				err.Error(),
			))
		}
		return
	}

	c.JSON(http.StatusCreated, config)
}

// UpdateConfiguration handles updating an existing configuration
func (h *ConfigurationHandler) UpdateConfiguration(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(
			"Configuration name is required",
			errors.ErrorCodeInvalidRequest,
			nil,
		))
		return
	}

	var req struct {
		Data json.RawMessage `json:"data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(
			"Invalid request body",
			errors.ErrorCodeInvalidRequest,
			err.Error(),
		))
		return
	}

	config, err := h.configService.UpdateConfiguration(name, req.Data)
	if err != nil {
		var appErr *errors.AppError
		if stdErrors.As(err, &appErr) {
			switch appErr.Code {
			case errors.ErrorCodeNotFound:
				c.JSON(http.StatusNotFound, appErr.ToErrorResponse())
			case errors.ErrorCodeValidationFailed:
				c.JSON(http.StatusBadRequest, appErr.ToErrorResponse())
			default:
				c.JSON(http.StatusInternalServerError, appErr.ToErrorResponse())
			}
		} else {
			c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(
				"Failed to update configuration",
				errors.ErrorCodeInternalError,
				err.Error(),
			))
		}
		return
	}

	c.JSON(http.StatusOK, config)
}

// GetConfiguration handles retrieving a configuration
func (h *ConfigurationHandler) GetConfiguration(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(
			"Configuration name is required",
			errors.ErrorCodeInvalidRequest,
			nil,
		))
		return
	}

	config, err := h.configService.GetConfiguration(name)
	if err != nil {
		var appErr *errors.AppError
		if stdErrors.As(err, &appErr) {
			if appErr.Code == errors.ErrorCodeNotFound {
				c.JSON(http.StatusNotFound, appErr.ToErrorResponse())
			} else {
				c.JSON(http.StatusInternalServerError, appErr.ToErrorResponse())
			}
		} else {
			c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(
				"Failed to get configuration",
				errors.ErrorCodeInternalError,
				err.Error(),
			))
		}
		return
	}

	c.JSON(http.StatusOK, config)
}

// GetConfigurationVersion handles retrieving a specific version of a configuration
func (h *ConfigurationHandler) GetConfigurationVersion(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(
			"Configuration name is required",
			errors.ErrorCodeInvalidRequest,
			nil,
		))
		return
	}

	versionStr := c.Param("version")
	if versionStr == "" {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(
			"Version is required",
			errors.ErrorCodeInvalidRequest,
			nil,
		))
		return
	}

	version, err := strconv.Atoi(versionStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(
			"Invalid version format",
			errors.ErrorCodeInvalidRequest,
			err.Error(),
		))
		return
	}

	config, err := h.configService.GetConfigurationVersion(name, version)
	if err != nil {
		var appErr *errors.AppError
		if stdErrors.As(err, &appErr) {
			if appErr.Code == errors.ErrorCodeNotFound {
				c.JSON(http.StatusNotFound, appErr.ToErrorResponse())
			} else {
				c.JSON(http.StatusInternalServerError, appErr.ToErrorResponse())
			}
		} else {
			c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(
				"Failed to get configuration version",
				errors.ErrorCodeInternalError,
				err.Error(),
			))
		}
		return
	}

	c.JSON(http.StatusOK, config)
}

// ListConfigurationVersions handles listing all versions of a configuration
func (h *ConfigurationHandler) ListConfigurationVersions(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(
			"Configuration name is required",
			errors.ErrorCodeInvalidRequest,
			nil,
		))
		return
	}

	versions, err := h.configService.ListConfigurationVersions(name)
	if err != nil {
		var appErr *errors.AppError
		if stdErrors.As(err, &appErr) {
			if appErr.Code == errors.ErrorCodeNotFound {
				c.JSON(http.StatusNotFound, appErr.ToErrorResponse())
			} else {
				c.JSON(http.StatusInternalServerError, appErr.ToErrorResponse())
			}
		} else {
			c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(
				"Failed to list configuration versions",
				errors.ErrorCodeInternalError,
				err.Error(),
			))
		}
		return
	}

	c.JSON(http.StatusOK, versions)
}

// RollbackConfiguration handles rolling back a configuration to a previous version
func (h *ConfigurationHandler) RollbackConfiguration(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(
			"Configuration name is required",
			errors.ErrorCodeInvalidRequest,
			nil,
		))
		return
	}

	var req struct {
		TargetVersion int `json:"target_version" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(
			"Invalid request body",
			errors.ErrorCodeInvalidRequest,
			err.Error(),
		))
		return
	}

	config, err := h.configService.RollbackConfiguration(name, req.TargetVersion)
	if err != nil {
		var appErr *errors.AppError
		if stdErrors.As(err, &appErr) {
			switch appErr.Code {
			case errors.ErrorCodeNotFound:
				c.JSON(http.StatusNotFound, appErr.ToErrorResponse())
			default:
				c.JSON(http.StatusInternalServerError, appErr.ToErrorResponse())
			}
		} else {
			c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(
				"Failed to rollback configuration",
				errors.ErrorCodeInternalError,
				err.Error(),
			))
		}
		return
	}

	c.JSON(http.StatusOK, config)
}

// RegisterSchema handles registering a JSON schema for a configuration
func (h *ConfigurationHandler) RegisterSchema(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(
			"Configuration name is required",
			errors.ErrorCodeInvalidRequest,
			nil,
		))
		return
	}

	var schema json.RawMessage
	if err := c.ShouldBindJSON(&schema); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(
			"Invalid schema format",
			errors.ErrorCodeInvalidRequest,
			err.Error(),
		))
		return
	}

	err := h.configService.RegisterSchema(name, schema)
	if err != nil {
		var appErr *errors.AppError
		if stdErrors.As(err, &appErr) {
			switch appErr.Code {
			case errors.ErrorCodeInvalidRequest:
				c.JSON(http.StatusBadRequest, appErr.ToErrorResponse())
			default:
				c.JSON(http.StatusInternalServerError, appErr.ToErrorResponse())
			}
		} else {
			c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(
				"Failed to register schema",
				errors.ErrorCodeInternalError,
				err.Error(),
			))
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"name":   name,
		"status": "schema registered successfully",
	})
}

// GetSchema handles retrieving a JSON schema for a configuration
func (h *ConfigurationHandler) GetSchema(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(
			"Configuration name is required",
			errors.ErrorCodeInvalidRequest,
			nil,
		))
		return
	}

	schema, err := h.configService.GetSchema(name)
	if err != nil {
		var appErr *errors.AppError
		if stdErrors.As(err, &appErr) {
			if appErr.Code == errors.ErrorCodeNotFound {
				c.JSON(http.StatusNotFound, appErr.ToErrorResponse())
			} else {
				c.JSON(http.StatusInternalServerError, appErr.ToErrorResponse())
			}
		} else {
			c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(
				"Failed to get schema",
				errors.ErrorCodeInternalError,
				err.Error(),
			))
		}
		return
	}

	// Parse JSON to return as object
	var schemaObj interface{}
	if err := json.Unmarshal(schema, &schemaObj); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(
			"Failed to parse schema",
			errors.ErrorCodeInternalError,
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, schemaObj)
}
