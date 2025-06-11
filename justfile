# Format code with goimports (organizes imports and formats)
format:
        find . -name "*.go" -type f -exec go tool goimports -w {} +
