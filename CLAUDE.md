# WeWork CLI Project Guide

## Build Commands
- Build: `go build ./cmd/wework`
- Install locally: `go install ./cmd/wework`
- Run: `go run ./cmd/wework/main.go`
- Format code: `go fmt ./...`
- Lint: `golint ./...`
- Vet: `go vet ./...`

## Code Style Guidelines
- Use camelCase for variable names
- Use PascalCase for exported functions and types
- Error handling: Always check errors and return them with context
- Imports: Group standard library, third-party, and local imports
- Prefer explicit error returns over panics
- Use cobra for CLI commands structure
- Authentication via environment variables or flags

## Project Structure
- `/cmd/wework`: Main CLI application
- `/pkg/wework`: Core library functionality
- Commands follow the pattern in `/cmd/wework/commands/`
- Authentication flow in main.go with token handling

## Common Tasks
- Adding a new command: Create file in `/cmd/wework/commands/` and register in main.go
- API interactions: Add methods to the WeWork struct in pkg/wework
- Testing: Add tests with `_test.go` suffix (currently missing)
