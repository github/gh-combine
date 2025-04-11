# gh-combine GitHub Copilot Guidelines

This is a Go based repository that provide a gh cli extension to combine multiple pull requests into one for a given repository(or multiple repos at once). It is designed to be used with the GitHub CLI and is intended to be a simple and easy to use tool for developers who want to combine multiple pull requests into one.

Here is the the very basic usage of the `gh-combine` extension:

```bash
gh combine owner/repo
```

## Code Standards

### Required Before Each Commit

- Run `make fmt` before committing any changes to ensure proper code formatting
- This will run gofmt on all Go files to maintain consistent style

### Development Flow

- Test: `go test -v -cover ./...`
- Lint: `go fmt ./... && go mod tidy && golangci-lint run --fix`
- Build: `go build -v ./cmd/gh-combine`

## Repository Structure

- `cmd/gh-combine`: Main cli entry points and executables
- `internal/`: Logic related to the core functionality of the extension
- `script/`: Scripts for building, testing, and releasing the extension
- `.github/`: GitHub Actions workflows for CI/CD
- `vendor/`: Vendor directory for Go modules (committed to the repository for reproducibility)

## Key Guidelines

1. Follow Go best practices and idiomatic patterns
2. Maintain existing code structure and organization
3. Use dependency injection patterns where appropriate
4. Write unit tests for new functionality. Use table-driven unit tests when possible.
5. Document public APIs and complex logic. Suggest changes to the `docs/` folder when appropriate
6. When responding to code refactoring suggestions, function suggestions, or other code changes, please keep your responses as concise as possible. We are capable engineers and can understand the code changes without excessive explanation. If you feel that a more detailed explanation is necessary, you can provide it, but keep it concise.
7. When suggesting code changes, always opt for the most maintainable approach. Try your best to keep the code clean and follow DRY principles. Avoid unnecessary complexity and always consider the long-term maintainability of the code.
8. When writing unit tests, always strive for 100% code coverage where it makes sense. Try to consider edge cases as well.
