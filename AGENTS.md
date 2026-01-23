# AGENTS.md

## IMPORTANT

- Try to keep things in one function unless composable or reusable
- DO NOT use `else` statements unless necessary
- AVOID `else` statements
- AVOID using `any` type
- PREFER single word variable names where possible

## Build / Lint / Test Commands

- **Build**: `go build ./...`
- **Run all tests**: `go test ./...`
- **Run a single test**: `go test -run ^TestName$ ./...`
- **Run tests with race detector**: `go test -race ./...`
- **Lint**: `go vet ./... && golint ./...`
- **Format**: `go fmt ./...`

## Code Style Guidelines

- **Imports**: Group standard library, third‑party, then local packages; use blank line between groups.
- **Formatting**: Run `go fmt` before committing; line length ≤ 120 characters.
- **Naming**:
  - Exported identifiers use MixedCaps (e.g., `LoadConfig`).
  - Unexported identifiers use mixedCaps.
  - Interface names end with `er` when appropriate.
- **Types**: Prefer concrete types over `interface{}`; use the `optional` package for optional values.
- **Error handling**: Return errors as the last return value; wrap with `fmt.Errorf("msg: %w", err)`.
- **Logging**: Use `log` or `slog` with structured fields; never log secrets directly.
- **Testing**: Table‑driven tests; name test functions `TestXxx`; use `gotest.tools/v3/assert` for assertions.
- **Documentation**: Exported functions/types must have Go doc comments.
- **Security**: Validate file permissions for secret files; never commit private keys.

## Repository Rules

- No `.cursor` or Copilot instruction files present.
- Follow the project’s existing conventions as seen in `file/` and `examples/` packages.

## Tool Calling

- ALWAYS USE PARALLEL TOOLS WHEN APPLICABLE. Here is an example illustrating how to execute 3 parallel file reads in this chat environment:

json
{
"recipient_name": "multi_tool_use.parallel",
"parameters": {
"tool_uses": [
{
"recipient_name": "functions.read",
"parameters": {
"filePath": "path/to/file.tsx"
}
},
{
"recipient_name": "functions.read",
"parameters": {
"filePath": "path/to/file.ts"
}
},
{
"recipient_name": "functions.read",
"parameters": {
"filePath": "path/to/file.md"
}
}
]
}
}
