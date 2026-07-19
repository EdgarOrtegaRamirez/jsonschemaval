# jsonschemaval

A fast, standalone CLI tool for validating JSON files against JSON Schema (Draft 2020-12, 2019-09, 2014-06, 2009-06, Draft-07 through Draft-04).

## Features

- **JSON Schema Validation** — Full support for JSON Schema Draft 2020-12, 2019-09, 2014-06, 2009-06, and Draft-07 through Draft-04
- **$ref Resolution** — Automatically resolves internal and external references
- **Detailed Errors** — Clear, human-readable error messages with field paths
- **CI/CD Integration** — Exit codes (0 = valid, 1 = invalid) for automation
- **Multiple Output Formats** — Text (terminal-friendly) and JSON (for pipelines)
- **Stdin Support** — Read data or schema from stdin for piping
- **Schema Inspection** — Shows draft version, title, and description metadata

## Installation

### From source (Go 1.24+)

```bash
go install github.com/EdgarOrtegaRamirez/jsonschemaval/cmd/jsonschemaval@latest
```

### Build from source

```bash
git clone https://github.com/EdgarOrtegaRamirez/jsonschemaval.git
cd jsonschemaval
go build -o jsonschemaval ./cmd/jsonschemaval/
```

### Download binary

```bash
# Replace with latest release URL
curl -sSfL https://github.com/EdgarOrtegaRamirez/jsonschemaval/releases/latest/download/jsonschemaval-linux-amd64 -o jsonschemaval
chmod +x jsonschemaval
sudo mv jsonschemaval /usr/local/bin/
```

## Usage

```bash
# Basic validation
jsonschemaval data.json schema.json

# JSON Schema from stdin
echo '{"type":"object","properties":{"name":{"type":"string"}}}' | jsonschemaval data.json -s

# Schema file, data from stdin
jsonschemaval -s schema.json

# Strict mode (rejects additional properties not in schema)
jsonschemaval data.json schema.json --strict

# Output as JSON for CI/CD pipelines
jsonschemaval data.json schema.json --format json

# Show schema metadata
jsonschemaval data.json schema.json --info

# Verbose with draft version
jsonschemaval data.json schema.json --verbose
```

## Examples

### Validating against a schema

```bash
# schema.json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "name": { "type": "string", "minLength": 1 },
    "age": { "type": "number", "minimum": 0 }
  },
  "required": ["name", "age"],
  "additionalProperties": false
}

# data.json
{"name": "Alice", "age": 30}

# Run
jsonschemaval data.json schema.json
# ✅ Valid
```

### Invalid data with detailed errors

```bash
# data.json
{"name": 123, "age": "old"}

# Run
jsonschemaval data.json schema.json
# ❌ Invalid (2 error(s))
#   1. name: Invalid type. Expected: string, given: integer
#   2. age: Invalid type. Expected: number, given: string
```

### CI/CD integration

```yaml
# .github/workflows/validate.yaml
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install jsonschemaval
        run: go install github.com/EdgarOrtegaRamirez/jsonschemaval/cmd/jsonschemaval@latest
      - name: Validate JSON
        run: jsonschemaval config.json schema.json --format json
```

### Piping from stdin

```bash
# Validate JSON response from curl
curl -s https://api.example.com/config | jsonschemaval -s schema.json

# Validate environment config
cat config.yaml | jsonschemaval - -s <(cat schema.json)
```

## Commands

| Command | Description |
|---------|-------------|
| `jsonschemaval <json> <schema>` | Validate JSON against schema |
| `jsonschemaval --version` | Show version information |
| `jsonschemaval --help` | Show help |

## Flags

| Flag | Description |
|------|-------------|
| `-s, --schema` | Path to JSON Schema file (alternative to second argument) |
| `-f, --format` | Output format: `text` (default) or `json` |
| `--strict` | Enable strict mode (rejects additional properties) |
| `-v, --verbose` | Show verbose output including draft version |
| `-i, --info` | Show full validation info including draft and schema metadata |
| `--version` | Show version information |

## Testing

```bash
# Run all tests
go test ./cmd/jsonschemaval/...

# Run with coverage
go test ./cmd/jsonschemaval/... -cover

# Race detector
go test ./cmd/jsonschemaval/... -race
```

## Differences from other tools

| Feature | jsonschemaval | gojq | ajv | jsonschema |
|---------|---------------|------|-----|------------|
| CLI focused | ✅ | ❌ | ✅ | ❌ |
| $ref resolution | ✅ | ❌ | ✅ | ✅ |
| Draft 2020-12 | ✅ | ❌ | ✅ | ✅ |
| Single binary | ✅ | ✅ | ❌ | ❌ |
| CI exit codes | ✅ | ❌ | ❌ | ❌ |

## License

MIT