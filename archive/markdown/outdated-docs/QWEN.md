# Law OA Go - Project Context for Qwen Code

## Project Overview

**Law OA Go** is a modern law office automation system built with Go 1.23+. It's a monolithic application designed to provide a complete digital solution for small and medium-sized law firms.

### Core Technologies
- **Language**: Go 1.23+
- **Framework**: Gin (high-performance HTTP framework)
- **ORM**: GORM (modern ORM framework)
- **Database**: MySQL 8.0+ (primary), Redis 7+ (cache), Elasticsearch 8+ (search)
- **Authentication**: JWT with refresh tokens
- **Security**: RBAC (Role-Based Access Control), bcrypt password encryption
- **Monitoring**: Prometheus metrics, structured logging (JSON format)
- **Documentation**: Swagger/OpenAPI 3.0
- **Deployment**: Docker + Docker Compose

## Project Structure

```
law-oa-go/
├── main.go                   # Application entry point
├── cmd/                      # Command-line tools
│   └── migrate/              # Database migration tools
├── internal/                 # Internal packages (not exposed publicly)
│   ├── handlers/             # HTTP handlers
│   ├── services/             # Business logic layer
│   ├── models/               # Data models
│   ├── middleware/           # Middleware functions
│   ├── auth/                 # Authentication module
│   ├── config/               # Configuration management
│   ├── database/             # Database connections
│   ├── cache/                # Cache operations
│   ├── common/               # Common components
│   ├── utils/                # Utility functions
│   ├── errors/               # Error handling system
│   ├── concurrency/          # Concurrency management
│   ├── health/               # Health monitoring
│   ├── monitoring/           # Performance monitoring
│   ├── infrastructure/       # Infrastructure components
│   ├── logger/               # Logging system
│   ├── logging/              # Structured logging
│   ├── rbac/                 # Role-based access control
│   ├── repositories/         # Data access layer
│   ├── retry/                # Retry mechanisms
│   ├── router/               # HTTP routing
│   ├── security/             # Security components
│   ├── server/               # Server management
│   ├── validation/           # Validation system
│   └── validators/           # Specific validators
├── docs/                     # Documentation
├── scripts/                  # Build and deployment scripts
├── configs/                  # Configuration files
├── docker-compose.yml        # Docker Compose configuration
├── Dockerfile               # Docker image build
├── Makefile                 # Build commands
└── .golangci.yml            # Code quality configuration
```

## Key Features

### Authentication & Authorization
- JWT token authentication with automatic refresh
- Role-based access control (RBAC) with admin, lawyer, and user roles
- Secure password handling with bcrypt encryption
- Session management and device tracking

### Core Modules
1. **User Management**: Complete CRUD operations, profile management, role assignment
2. **Client Management**: Client profiles, statistics, search and filtering
3. **Case Management**: Case tracking, lawyer assignment, status management
4. **Reporting**: Data statistics and business reports
5. **Search**: Basic search functionality with Elasticsearch integration

### System Capabilities
- High-performance API with sub-100ms response times
- Structured JSON logging for observability
- Health checks and monitoring endpoints
- Graceful shutdown with proper resource cleanup
- Comprehensive error handling with detailed error responses
- Circuit breaker and retry mechanisms for resilience
- Rate limiting and security headers
- Prometheus metrics for performance monitoring

## Building and Running

### Prerequisites
- Go 1.23 or higher
- Docker (recommended for deployment)
- MySQL 8.0+, Redis 7+, Elasticsearch 8+ (when not using Docker)

### Development Setup
```bash
# Install dependencies
make deps

# Run with Docker (recommended)
docker-compose up -d

# Or run locally (requires separate database services)
make run
```

### Testing
```bash
# Run all tests
make test

# Run specific test suites
make test-unit
make test-integration
make test-e2e

# Generate coverage report
make test-coverage
```

### Code Quality
```bash
# Format code
make fmt

# Run linters
make lint

# Static analysis
make vet
```

### Building
```bash
# Build application
make build

# Build for Linux
make build-linux

# Build Docker image
make docker-build
```

## API Design

### Authentication
All authenticated API requests require a JWT token in the Authorization header:
```
Authorization: Bearer YOUR_JWT_TOKEN
```

### Response Format
The API uses a unified response format:
```json
// Success response
{
  "success": true,
  "data": { ... },
  "error": null,
  "meta": {
    "timestamp": "2025-09-15T10:30:00Z",
    "request_id": "req_123456789"
  }
}

// Error response
{
  "success": false,
  "data": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "details": "email field format is incorrect"
  },
  "meta": {
    "timestamp": "2025-09-15T10:30:00Z",
    "request_id": "req_123456789"
  }
}
```

## Configuration

### Environment Variables
Key configuration options:
- `APP_ENV`: Environment (development, production)
- `PORT`: Application port (default: 8080)
- `DB_HOST`, `DB_PORT`, `DB_USERNAME`, `DB_PASSWORD`, `DB_DATABASE`: Database configuration
- `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`: Redis configuration
- `JWT_SECRET`: JWT secret key (must be at least 32 characters)
- `ES_HOST`, `ES_PORT`, `ES_USERNAME`, `ES_PASSWORD`: Elasticsearch configuration

### Configuration Files
- `.env`: Environment variables
- `.env.example`: Template for environment configuration
- `config/config.yaml`: Optional YAML configuration file

## Monitoring and Observability

### Health Checks
- `/health`: Basic health check
- `/health/live`: Liveness probe
- `/health/ready`: Readiness probe
- `/api/v1/health`: Detailed health check
- `/api/v1/health/metrics`: Health metrics

### Metrics
- `/metrics`: Prometheus metrics endpoint
- Built-in performance monitoring with response time tracking
- Database query metrics with slow query detection
- Connection pool monitoring

## Development Guidelines

### Code Quality Standards
- Follow Go official coding standards
- Use golangci-lint for code quality checks
- Maintain test coverage above 70%
- Write comprehensive unit and integration tests
- Use godoc format for documentation comments

### Error Handling
- Use the custom error system in `internal/errors`
- Provide meaningful error messages and codes
- Include context information in errors when relevant
- Handle errors at appropriate levels in the call stack

### Concurrency
- Use the concurrency service in `internal/concurrency` for parallel operations
- Implement proper synchronization with mutexes when needed
- Use context for timeout and cancellation handling
- Follow worker pool patterns for resource-intensive tasks

### Caching
- Use Redis for caching with proper TTL settings
- Implement cache-aside pattern for data retrieval
- Use cache keys defined in `internal/cache` package
- Invalidate cache appropriately when data changes

## Deployment

### Docker Deployment (Recommended)
```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### Manual Deployment
1. Build the binary: `make build`
2. Set up required environment variables
3. Ensure database services are running
4. Run the application: `./bin/law-oa-go`

## Testing Strategy

### Test Types
1. **Unit Tests**: Test individual functions and components
2. **Integration Tests**: Test interactions between components
3. **End-to-End Tests**: Test complete user workflows
4. **Performance Tests**: Benchmark critical operations
5. **Load Tests**: Test system under high load conditions

### Test Coverage Goals
- Overall coverage: 80%+
- Critical business logic: 95%+
- API endpoints: 90%+
- Database operations: 85%+

## Common Commands

### Makefile Targets
```bash
# Development workflow
make clean        # Clean build artifacts
make deps         # Install dependencies
make fmt          # Format code
make lint         # Run linters
make test         # Run tests
make build        # Build application
make run          # Run application

# Docker operations
make docker-build # Build Docker image
make pgo-build    # PGO optimized build

# Testing
make test-unit    # Run unit tests
make test-integration # Run integration tests
make bench        # Run benchmarks
make fuzz-all     # Run fuzzing tests

# Quality assurance
make quality      # Run all quality checks
make ci           # Run CI pipeline
```

This documentation provides the essential context for working with the Law OA Go project. For specific implementation details, refer to the source code and README.md file.