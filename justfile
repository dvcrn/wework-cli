# Format code with goimports (organizes imports and formats)
format:
        find . -name "*.go" -type f -exec go tool goimports -w {} +

# Install wework CLI tool locally
install:
        go install ./cmd/wework
