# Law OA Go - CodeBuddy Development Guide

## Development Commands

### Primary Development Script
Use `./dev.sh` for all development tasks:

```bash
# Project initialization
./dev.sh init          # Initialize project, create .env, download dependencies

# Development environment
./dev.sh start          # Start Docker Compose services (MySQL, Redis, Elasticsearch)
./dev.sh stop           # Stop Docker Compose services

# Build and run
./dev.sh build          # Build Go binary (creates ./main)
./dev.sh run            # Run application directly with go run
go run main.go          # Alternative direct run

# Testing and quality
./dev.sh test           # Run Go tests
./test_runner.sh        # Run unit tests specifically
go test -v ./...        # Run all tests with verbose output

# Code quality
./dev.sh lint           # Run golangci-lint (if installed)

# Documentation and cleanup
./dev.sh docs           # Generate Swagger docs (if swag installed)
./dev.sh clean          # Clean build artifacts and Docker resources
```

### Direct Go Commands
```bash
# Dependencies
go mod download && go mod tidy

# Testing specific modules
go test -run="TestModel.*" ./internal/models/models_unit_test.go ./internal/models/models.go -v

# Build with specific output
go build -o law-oa .
```

## Architecture Overview

### Technology Stack
- **Framework**: Gin (Go web framework)
- **Database**: MySQL 8.0 with GORM ORM
- **Cache**: Redis 7 
- **Search**: Elasticsearch 8.8
- **Authentication**: JWT tokens
- **Monitoring**: Prometheus metrics
- **Containerization**: Docker & Docker Compose

### Project Structure
```
internal/
├── auth/           # Authentication & authorization
├── cache/          # Redis caching layer
├── casemgmt/       # Case management module
├── client/         # Client management (Phase 2.1 - completed)
├── config/         # Configuration management (Viper + env)
├── database/       # Database connections & optimization
├── handlers/       # HTTP request handlers
├── lawyer/         # Lawyer management module
├── middleware/     # HTTP middleware (CORS, auth, rate limiting, caching)
├── models/         # GORM data models
├── monitoring/     # Performance monitoring
├── rbac/           # Role-based access control
├── router/         # Route definitions
├── security/       # Security utilities
├── services/       # Business logic services
├── utils/          # Common utilities
└── validation/     # Data validation
```

### Core Components

**Application Entry**: `main.go` initializes optimized components, applies middleware stack (logging, recovery, CORS, rate limiting, Prometheus, caching), and sets up graceful shutdown.

**Configuration**: Uses Viper for config management with `.env` file support. Structured config types for database, Redis, Elasticsearch, JWT, and logging.

**Database Layer**: GORM with MySQL, connection pooling, and performance optimizations. Includes cache service integration and Elasticsearch client.

**Middleware Stack**: 
- Logger & Recovery (error handling)
- CORS & Rate Limiting
- Prometheus metrics collection
- Redis-based caching with configurable TTL
- JWT authentication (where applicable)

**Module Architecture**: Each business module (client, lawyer, case management) follows a consistent pattern with handlers, services, models, and validation.

### Key Features
- **Performance Monitoring**: `/metrics` endpoint for Prometheus, `/health` for health checks
- **Caching Strategy**: Redis-based HTTP response caching with middleware
- **Graceful Shutdown**: 30-second timeout for clean service termination
- **Development Phase**: Currently in Phase 2 with client management completed (v2.1.0)

### Development Environment
- **Services**: MySQL (3306), Redis (6379), Elasticsearch (9200), Kibana (5601)
- **Application**: Runs on port 8080
- **Docker Compose**: Manages all service dependencies
- **Hot Reload**: Use `./dev.sh run` for development with auto-restart

### Testing Strategy
- Unit tests for models and business logic
- Database tests (requires SQLite or MySQL test DB)
- Performance tests via `/performance/cache` endpoint
- Test runner script provides coverage reporting

### Configuration Files
- `.env`: Environment variables (copy from `.env.example`)
- `config.yaml`: Application configuration
- `docker-compose.yml`: Service orchestration
- `go.mod`: Go dependencies (Go 1.23+)