# Profitify App - Claude Development Guide

## Project Overview

**Profitify** is a full-stack application for stock data analysis and management. It features a modern architecture with a Go backend, React frontend, and DynamoDB for data storage.

### Tech Stack

**Backend:**
- **Language:** Go 1.24.6
- **Framework:** Gin (HTTP web framework)
- **Database:** DynamoDB (AWS)
- **Logging:** Zap (structured logging)
- **Testing:** Testify (mocking and assertions)

**Frontend:**
- **Language:** TypeScript
- **Framework:** React 19.1.1
- **Build Tool:** Vite 7.1.0
- **UI Library:** shadcn/ui with Tailwind CSS 4.1.11
- **Styling:** Tailwind CSS with theme provider support
- **Icons:** Lucide React

**Infrastructure:**
- **Containerization:** Docker & Docker Compose
- **Local Development:** LocalStack (for DynamoDB simulation)
- **CI/CD:** Docker-based deployment

## Project Structure

```
profitify-app/
├── backend/                     # Go backend application
│   ├── internal/               # Private application code
│   │   ├── handlers/          # HTTP request handlers
│   │   ├── middleware/        # HTTP middleware
│   │   ├── models/           # Data models
│   │   └── repository/       # Data access layer
│   ├── pkg/                   # Public/shared packages
│   │   ├── config/           # Application configuration
│   │   ├── logger/           # Structured logging
│   │   ├── router/           # HTTP routing
│   │   └── server/           # HTTP server
│   ├── scripts/              # Utility scripts
│   └── main.go              # Application entry point
├── frontend/                   # React frontend application
│   ├── src/
│   │   ├── components/       # React components
│   │   ├── lib/             # Utility functions
│   │   └── assets/          # Static assets
│   └── public/              # Public assets
└── docker/                    # Docker volumes and state
```

## Development Commands

### Backend (Go)

```bash
# Development
cd backend
go run main.go                 # Run development server
go test ./...                  # Run all tests
go test -v ./internal/...      # Run tests with verbose output
go test -bench=.               # Run benchmarks
go mod tidy                    # Clean up dependencies
go mod download               # Download dependencies

# Building
go build -o bin/profitify-backend main.go

# Testing with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Frontend (React/TypeScript)

```bash
# Development
cd frontend
yarn dev                      # Start development server (port 3000)
yarn build                   # Build for production
yarn lint                    # Run ESLint
yarn preview                 # Preview production build

# Dependencies
yarn install                 # Install dependencies
yarn add <package>           # Add new dependency
```

### Docker & Infrastructure

```bash
# Full stack development
docker-compose up -d          # Start all services
docker-compose --profile tools run --rm dynamodb-init  # Initialize DynamoDB
docker-compose down          # Stop all services
docker-compose down -v       # Stop and remove volumes

# Individual services
docker-compose up backend    # Start only backend
docker-compose up frontend   # Start only frontend
docker-compose up localstack # Start only LocalStack
```

## Architecture Patterns

### Backend Architecture

**Clean Architecture Implementation:**
- **Handlers Layer:** HTTP request handling and response formatting
- **Repository Layer:** Data access abstraction with interface-based design
- **Models Layer:** Domain entities and data structures
- **Middleware Layer:** Cross-cutting concerns (logging, CORS, etc.)

**Key Patterns:**
- **Dependency Injection:** Context-based handler initialization
- **Interface Segregation:** Repository interfaces for testability
- **Error Handling:** Custom error types with structured responses
- **Graceful Shutdown:** Context-based server lifecycle management
- **Structured Logging:** Zap logger with configurable levels

**API Design:**
- RESTful endpoints under `/api` prefix
- Health check endpoints (`/health`, `/health/live`, `/health/ready`)
- JSON request/response format
- Proper HTTP status codes

### Frontend Architecture

**Component Structure:**
- **Theme Provider:** Dark/light theme support with persistence
- **shadcn/ui Components:** Consistent UI component library
- **Tailwind CSS:** Utility-first styling approach
- **TypeScript:** Type safety and better developer experience

**Build & Development:**
- **Vite:** Fast build tool with HMR (Hot Module Replacement)
- **ESLint:** Code quality and style enforcement
- **Path Aliases:** Clean import statements with `@/` prefix

### Database Design

**DynamoDB Schema:**
- **Tickers Table:** Stock ticker information
  - Primary Key: `ticker` (string)
  - Attributes: name, market, locale, active status, etc.
  - GSI considerations for query patterns

## Testing Strategy

### Backend Testing

**Unit Tests:**
- **Handlers:** Mock DynamoDB client testing
- **Repository:** Mock repository implementation with call tracking
- **Models:** Data validation and marshaling tests

**Test Structure:**
- Table-driven tests for comprehensive coverage
- Mock interfaces using testify/mock
- Benchmark tests for performance validation
- Context-aware testing for timeout scenarios

**Mocking Strategy:**
- DynamoDB client mocking for handler tests
- Repository interface mocking for service layer tests
- Logger mocking for clean test output

### Frontend Testing

**Current Setup:**
- ESLint for code quality
- TypeScript for compile-time error detection
- Development server with HMR for rapid feedback

**Testing Framework:** (To be implemented)
- Jest/Vitest for unit testing
- React Testing Library for component testing
- Cypress/Playwright for E2E testing

## Configuration & Environment

### Environment Variables

**Backend:**
```bash
PORT=8080                     # Server port
ENVIRONMENT=development       # Environment mode
LOG_LEVEL=debug              # Logging level
SHUTDOWN_TIMEOUT=30s         # Graceful shutdown timeout
READ_TIMEOUT=15s             # HTTP read timeout
WRITE_TIMEOUT=15s            # HTTP write timeout
IDLE_TIMEOUT=60s             # HTTP idle timeout

# AWS/DynamoDB (LocalStack)
AWS_ENDPOINT_URL=http://localstack:4566
AWS_ACCESS_KEY_ID=test
AWS_SECRET_ACCESS_KEY=test
AWS_DEFAULT_REGION=us-east-1
```

**Frontend:**
- Vite environment variables support
- Theme persistence in localStorage
- TypeScript strict mode enabled

### Local Development Setup

**Prerequisites:**
- Docker and Docker Compose
- Ports 3000, 8080, and 4566 available

**Setup Process:**
1. Start infrastructure: `docker-compose up -d`
2. Initialize database: `docker-compose --profile tools run --rm dynamodb-init`
3. Access services:
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080
   - LocalStack: http://localhost:4566

## API Documentation

### Endpoints

**Health Checks:**
- `GET /health` - General health status
- `GET /health/live` - Liveness probe
- `GET /health/ready` - Readiness probe

**Tickers API:**
- `GET /api/tickers` - Retrieve all tickers from DynamoDB

### Response Format

```json
{
  "tickers": [...],
  "count": 123
}
```

**Error Response:**
```json
{
  "error": "Error message description"
}
```

## Development Guidelines

### Code Style

**Go:**
- Follow Go conventions and gofmt formatting
- Use structured logging with appropriate levels
- Implement proper error handling with custom error types
- Write table-driven tests with comprehensive coverage
- Use interfaces for dependency injection and testability

**TypeScript/React:**
- Follow ESLint configuration
- Use TypeScript strict mode
- Implement proper component composition
- Use path aliases for clean imports
- Follow React best practices and hooks guidelines

### Git Workflow

- Feature branches for new development
- Descriptive commit messages
- Pull requests for code review
- CI/CD integration with Docker builds

### Performance Considerations

**Backend:**
- Connection pooling for DynamoDB
- Proper context handling for timeouts
- Graceful shutdown implementation
- Pagination for large datasets

**Frontend:**
- Vite for fast builds and HMR
- Tree shaking for optimized bundles
- Lazy loading for components
- Theme persistence optimization

## Special Notes

### Current Development Status

- **Backend:** Core structure implemented with ticker management
- **Frontend:** Basic React setup with theming, needs feature development
- **Database:** LocalStack setup for development, production DynamoDB ready
- **Testing:** Comprehensive backend testing, frontend testing to be implemented

### TODO/Future Enhancements

- Implement frontend ticker management UI
- Add authentication and authorization
- Implement real-time stock data updates
- Add comprehensive frontend testing suite
- Implement CI/CD pipeline
- Add API documentation (OpenAPI/Swagger)
- Implement caching strategies
- Add monitoring and observability