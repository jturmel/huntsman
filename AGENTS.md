# Repository Guidelines

## Project Structure & Module Organization
- `main.go`: Application entry point and TUI logic (Bubble Tea)
- `go.mod`, `go.sum`: Go module definition and dependencies
- `Makefile`: Build and development automation
- `.githooks/`: Git hooks for development workflow

## Build, Test, and Development Commands
- Build command: `make build`
- Run command: `make run`
- Setup hooks: `make setup-hooks`

## Coding Style & Naming Conventions
- Go 1.25+
- Follow standard Go formatting (`gofmt` or `goimports`)
- Use `camelCase` for internal variables and `PascalCase` for exported symbols
- Maintain clear separation between TUI model and business logic
- Use `github.com/charmbracelet/bubbletea` for TUI components