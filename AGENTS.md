# jsonschemaval - AI Agent Guide

## Project Overview

**jsonschemaval** is a fast, standalone Go CLI tool for validating JSON files against JSON Schema (Draft 2020-12 through Draft-04). It provides detailed, human-readable error messages and supports programmatic output for CI/CD pipelines.

## Building

```bash
go build -o jsonschemaval ./cmd/jsonschemaval/
```

## Testing

```bash
# All tests
go test ./cmd/jsonschemaval/... -v

# With coverage
go test ./cmd/jsonschemaval/... -cover

# Race detector
go test ./cmd/jsonschemaval/... -race
```

## Project Structure

```
jsonschemaval/
├── cmd/jsonschemaval/
│   ├── main.go        — CLI entry point with cobra commands
│   └── main_test.go   — Integration tests
├── go.mod             — Go module definition
├── README.md          — Documentation
├── LICENSE            — MIT License
├── AGENTS.md          — AI agent guide (this file)
└── .gitignore         — Git ignore rules
```

## Key Design Decisions

1. **Single binary** — No runtime dependencies, self-contained executable
2. **Cobra CLI** — Standard Go CLI framework with subcommands and flags
3. **gojsonschema** — Battle-tested JSON Schema validation library (go 1.x)
4. **Exit codes** — 0 for valid, 1 for invalid — designed for CI/CD
5. **Stdin support** — Read JSON or schema from stdin for piping
6. **No network calls** — All validation is local, no external API dependencies

## Adding New Features

### Adding a new flag
1. Add flag variable in `init()` function
2. Register with `rootCmd.Flags().StringVarP(...)` or similar
3. Use in `runValidate()` function

### Adding a new output format
1. Modify `format` flag parsing
2. Add handler in `runValidate()` switch
3. Add `outputFormat()` function
4. Update tests

### Adding a new JSON Schema keyword check
The `gojsonschema` library handles all JSON Schema keywords automatically. No additional code needed — validation is delegated to the library.

## Common Tasks

### Adding a new subcommand
1. Create a new `cobra.Command` struct
2. Register it with `rootCmd.AddCommand()`
3. Write handler function
4. Add tests

### Updating dependencies
```bash
go get -u ./...
go mod tidy
```

### Publishing a release
```bash
go build -ldflags "-X main.version=1.0.0 -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ) -X main.commit=$(git rev-parse HEAD)" -o jsonschemaval ./cmd/jsonschemaval/
```

## Error Handling

- Use `fmt.Errorf` with `%w` for wrapped errors
- Check file existence with `os.Stat` before opening
- Handle stdin EOF gracefully
- Never use bare `panic()` — always return errors

## Test Data

Tests use `t.TempDir()` for temporary files. All JSON and schema data is inline in test cases — no external fixtures needed.