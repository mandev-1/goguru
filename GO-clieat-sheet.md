Great question! Here are the essential Go CLI commands you'll use:

## **Module Management**

```bash
# Initialize a new module (do this first)
go mod init github.com/username/camagru

# Add missing dependencies and remove unused ones
go mod tidy

# Download dependencies to local cache
go mod download

# Verify dependencies
go mod verify

# Show why a package is needed
go mod why github.com/lib/pq

# Update a specific dependency
go get -u github.com/lib/pq

# Update all dependencies
go get -u ./...
```

## **Building & Running**

```bash
# Run the application directly
go run main.go
# Or with the standard structure:
go run cmd/server/main.go

# Build an executable
go build -o camagru
# Or specify the entry point:
go build -o camagru cmd/server/main.go

# Install (builds and puts binary in $GOPATH/bin)
go install

# Run with automatic reload on file changes (use air or fresh tool)
# First: go install github.com/cosmtrek/air@latest
air
```

## **Testing**

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test
go test -run TestFunctionName

# Run benchmarks
go test -bench=.
```

## **Code Quality**

```bash
# Format all Go files (ALWAYS do this)
go fmt ./...
# Or use gofmt directly:
gofmt -w .

# Lint code (install first: go install golang.org/x/lint/golint@latest)
golint ./...

# Static analysis
go vet ./...

# Show documentation
go doc package/name
go doc package.FunctionName
```

## **Dependencies**

```bash
# Add a new dependency
go get github.com/gorilla/mux

# Add a specific version
go get github.com/gorilla/mux@v1.8.0

# Add latest version
go get -u github.com/gorilla/mux

# Remove a dependency (then run go mod tidy)
go get github.com/unused/package@none

# List all dependencies
go list -m all

# Show available updates
go list -u -m all
```

## **Working with Packages**

```bash
# List packages in current module
go list ./...

# Show package info
go list -json ./internal/handlers

# Find where Go tools are installed
go env GOPATH
go env GOROOT

# Show all Go environment variables
go env
```

## **Common Workflow**

```bash
# Starting a new project
go mod init myproject
go get github.com/lib/pq  # Add dependencies
go mod tidy               # Clean up
go fmt ./...              # Format code
go vet ./...              # Check for issues
go test ./...             # Run tests
go build                  # Build binary

# During development
go run main.go            # Quick testing
go fmt ./...              # Before committing
go mod tidy               # Keep go.mod clean

# Before committing
go fmt ./...
go vet ./...
go test ./...
go mod tidy
```

## **Useful Tools to Install**

```bash
# Hot reload during development
go install github.com/cosmtrek/air@latest

# Better linter
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Database migrations
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Code generation (mocks, etc)
go install github.com/golang/mock/mockgen@latest
```

## **For Your Camagru Project**

```bash
# Initialize
cd camagru
go mod init camagru

# Add dependencies as you need them
go get github.com/lib/pq              # PostgreSQL driver
go get golang.org/x/crypto/bcrypt     # Password hashing
go get github.com/joho/godotenv       # .env file support

# Daily workflow
go fmt ./...                          # Format before running
go run cmd/server/main.go            # Run the server
go test ./...                        # Test your code

# Before committing
go fmt ./...
go vet ./...
go test ./...
go mod tidy
```

**Pro tip:** Add this to your `.gitignore`:
```
# Binaries
camagru
*.exe

# Go workspace file
go.work

# Environment
.env
.env.local

# IDE
.vscode/
.idea/
```

The most important commands you'll use daily are: `go run`, `go build`, `go fmt`, `go mod tidy`, and `go test`. Start with those!