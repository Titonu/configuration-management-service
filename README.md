# Configuration Management Service

## Overview
The Configuration Management Service is a robust API for managing configuration data with schema-based validation, versioning, and rollback support. It enables teams to safely define, update, retrieve, and roll back configuration data while ensuring data integrity through JSON Schema validation.

## Table of Contents
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Environment Setup](#environment-setup)
- [Running the Service](#running-the-service)
- [API Documentation](#api-documentation)
- [Authentication](#authentication)
- [Configuration Schemas](#configuration-schemas)
- [Testing](#testing)
- [Project Structure](#project-structure)
- [Design Decisions and Trade-offs](#design-decisions-and-trade-offs)
- [Future Improvements](#future-improvements)

## Features
- **Schema Validation**: Validate configuration data against JSON Schema definitions
- **Versioning**: Automatically track all changes to configurations
- **Rollback**: Easily revert to previous versions when needed
- **Authentication**: Secure API access with API key authentication
- **RESTful API**: Clean and intuitive API design

## Prerequisites
- Go 1.21+ (required for building and running the service)
- SQLite (embedded, no separate installation required)
- Git (for cloning the repository)

## Environment Setup

### Clone the Repository
```bash
git clone https://github.com/Titonu/configuration-management-service.git
cd configuration-management-service
```

### Environment Variables
Create a `.env` file in the project root (or copy from `.env.example`):

```bash
cp .env.example .env
```

The following environment variables are supported:

| Variable | Description | Default |
|----------|-------------|--------|
| `PORT` | Port to run the server on | `8080` |
| `GIN_MODE` | Gin framework mode (`debug` or `release`) | `debug` |
| `SQLITE_DB_PATH` | Path to SQLite database file | `data/config.db` |
| `API_KEYS` | Comma-separated list of API keys in format `key:client` | `dev-api-key:development` |

## Running the Service

### Using the Makefile
The project includes a comprehensive Makefile to simplify common development tasks:

```bash
# Build the application
make build

# Run the application
make run

# Clean build artifacts
make clean

# Run all tests
make test

# Run only integration tests
make test-integration

# Run only unit tests
make test-unit

# Generate test coverage report
make test-coverage

# Format code
make fmt

# Run linters
make lint

# Show all available commands
make help
```

### Manual Build and Run
If you prefer not to use the Makefile, you can build and run the service manually:

```bash
go build -o config-service ./cmd/server
./config-service
```

Or run directly with:
```bash
go run cmd/server/main.go
```

The service will start on port 8080 by default (configurable via the `PORT` environment variable).

### Docker Support
The project includes Docker support with a multi-stage build Dockerfile for production and a development Dockerfile with hot reload capabilities.

#### Using Docker Directly
You can use the Makefile Docker commands:

```bash
# Build Docker image
make docker-build

# Run Docker container
make docker-run

# Clean Docker resources
make docker-clean
```

#### Using Docker Compose
The project includes a `docker-compose.yml` file that defines two services:
- `config-service`: Production service
- `config-service-dev`: Development service with hot reload

To use Docker Compose:

```bash
# Start production service
make docker-compose-up

# Start development service with hot reload
make docker-compose-dev

# Stop all services
make docker-compose-down
```

#### Development Environment
The development service uses [Air](https://github.com/cosmtrek/air) for hot reload, which automatically rebuilds and restarts the application when code changes are detected.

The development service is exposed on port 8081 to avoid conflicts with the production service.

## API Documentation
The API is documented using OpenAPI 3.0 specification in the `openapi.yaml` file at the project root. You can view this documentation using tools like Swagger UI or Redoc.

### API Endpoints

#### Configuration Management
- `POST /api/v1/configurations` - Create a new configuration
- `GET /api/v1/configurations/{name}` - Get the latest version of a configuration
- `PUT /api/v1/configurations/{name}` - Update an existing configuration
- `GET /api/v1/configurations/{name}/versions` - List all versions of a configuration
- `GET /api/v1/configurations/{name}/versions/{version}` - Get a specific version of a configuration
- `POST /api/v1/configurations/{name}/rollback` - Rollback a configuration to a previous version

#### Schema Management
- `POST /api/v1/schemas/{name}` - Register a schema for a configuration type
- `GET /api/v1/schemas/{name}` - Get the schema for a configuration type

#### Health Check
- `GET /health` - Service health check (no authentication required)

## Authentication
All endpoints (except the health check) require authentication using an API key. Include the API key in the Authorization header using the Bearer token format:

```
Authorization: Bearer your-api-key
```

API keys are configured via the `API_KEYS` environment variable using the format:
```
API_KEYS=key1:client1,key2:client2
```

For example: `API_KEYS=dev-api-key:development,test-key:testing`

For testing, you can use: `dev-api-key`

### Production Readiness and Multi-User Support
The API key authentication mechanism is designed for production readiness in multi-user environments:

1. **Client Isolation**: Each API key is associated with a client identifier, allowing for multi-tenant usage where different teams or services can have their own keys.

2. **Access Control**: The authentication middleware validates all requests, ensuring only authorized clients can access or modify configurations.

3. **Audit Trail**: Client identifiers are logged with each request, providing accountability and traceability in multi-user environments.

4. **Security**: API keys can be rotated or revoked independently without affecting other users of the system.

5. **Scalability**: The authentication scheme supports multiple concurrent users and services without performance degradation.

6. **Environment Separation**: Different API keys can be used for different environments (development, staging, production) while using the same service instance.

This approach ensures the Configuration Management Service can be safely deployed in production environments where multiple teams or services need to access and manage configurations securely.

### Using the OpenAPI UI
To authenticate in the OpenAPI UI:

1. Click the "Authorize" button at the top of the page
2. Enter your API key in the value field (without "Bearer" prefix)
3. Click "Authorize" and close the dialog
4. Your requests will now include the Authorization header

## Configuration Schemas
Configurations are validated against JSON Schema definitions. Here's an example schema for a payment configuration:

```json
{
  "type": "object",
  "properties": {
    "max_limit": { "type": "integer" },
    "enabled": { "type": "boolean" }
  },
  "required": ["max_limit", "enabled"]
}
```

With this schema, a valid configuration would be:

```json
{
  "max_limit": 1000,
  "enabled": true
}
```

## Testing

### Running Tests with Make
The Makefile provides convenient commands for testing:

```bash
# Run all tests
make test

# Run only integration tests
make test-integration

# Run only unit tests
make test-unit

# Generate test coverage report
make test-coverage
```

### Manual Test Commands
If you prefer not to use the Makefile, you can run tests manually:

```bash
# Run all tests
go test ./...

# Run integration tests specifically
go test ./tests/integration

# Generate test coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

The integration tests cover all API endpoints and include tests for:
- Schema registration and retrieval
- Configuration creation, retrieval, and update
- Version listing and retrieval
- Configuration rollback
- Error cases (invalid data, not found, etc.)

## Project Structure

```
.
├── cmd/                      # Application entrypoints
│   └── server/               # Server application
│       └── main.go           # Main application file
├── internal/                 # Private application code
│   ├── delivery/             # HTTP delivery layer
│   │   └── http/             # HTTP handlers and middleware
│   │       ├── handler/      # HTTP handlers
│   │       └── middleware/   # HTTP middleware
│   ├── domain/              # Domain models and interfaces
│   │   ├── model/           # Domain models
│   │   ├── repository/      # Repository interfaces
│   │   └── usecase/         # Usecase interfaces
│   ├── repository/          # Repository implementations
│   │   └── sqlite/          # SQLite repository implementation
│   └── usecase/             # Usecase implementations
├── pkg/                     # Public packages
│   └── errors/              # Error handling utilities
├── tests/                   # Test files
│   └── integration/         # Integration tests
├── data/                    # Data storage
├── .env.example             # Example environment variables
├── .env                     # Environment variables (not in git)
├── .air.toml                # Air configuration for hot reload
├── docker-compose.yml       # Docker Compose configuration
├── Dockerfile               # Production Docker image definition
├── Dockerfile.dev           # Development Docker image with hot reload
├── go.mod                   # Go modules file
├── go.sum                   # Go modules checksum
├── Makefile                 # Build and development tasks
├── openapi.yaml             # OpenAPI specification
└── README.md                # This file
```

## Design Decisions and Trade-offs

### Clean Architecture
The project follows clean architecture principles, separating concerns into layers:
- **Domain Layer**: Core business logic and interfaces
- **Usecase Layer**: Application-specific business rules
- **Repository Layer**: Data access implementation
- **Delivery Layer**: HTTP handlers and middleware

This approach makes the codebase modular, testable, and maintainable.

### SQLite Storage
SQLite was chosen for simplicity and ease of setup. For production use with higher loads, a more robust database like PostgreSQL would be recommended.

### Error Handling
A custom error handling package provides structured error responses with error codes, messages, and details. This ensures consistent error reporting across the API.

### Authentication
A client-based API key authentication mechanism was implemented to support multi-tenant usage in production environments. Each API key is associated with a specific client identifier, enabling request tracking, access control, and client isolation. For even more robust security in larger deployments, this could be extended to OAuth2 or JWT authentication with role-based access control.

### JSON Schema Validation
JSON Schema validation ensures that configuration data adheres to predefined structures, preventing invalid configurations from being stored.

### Versioning
Each configuration change creates a new version, allowing for complete history tracking and the ability to roll back to previous states.

### Limitations
- Limited database options (currently SQLite only)
- No CI/CD pipeline configuration

## Future Improvements

1. **Enhanced Authentication**: Implement more robust authentication mechanisms like OAuth2 or JWT
2. **Database Options**: Support multiple database backends (PostgreSQL, MySQL, etc.)
3. **Caching Layer**: Add caching for frequently accessed configurations
4. **Metrics and Monitoring**: Integrate with Prometheus/Grafana for monitoring
5. **CI/CD Pipeline**: Add GitHub Actions or similar for automated testing and deployment
6. **CI/CD Pipeline**: Add GitHub Actions or similar for automated testing and deployment
7. **Database Options**: Add support for other databases like PostgreSQL or MySQL
8. **Configuration Import/Export**: Add bulk import/export functionality
9. **User Management**: Add user management for more granular access control
10. **Audit Logging**: Implement comprehensive audit logging for all configuration changes