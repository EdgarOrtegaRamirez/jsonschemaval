package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var (
	binaryPath string
)

func init() {
	// Get project root (parent of cmd/jsonschemaval)
	cwd, _ := os.Getwd()
	projectRoot := filepath.Join(cwd, "..", "..")
	binaryPath = filepath.Join(projectRoot, "jsonschemaval")
}

func runCLI(args ...string) (*exec.Cmd, string, int) {
	cmd := exec.Command(binaryPath, args...)
	out, err := cmd.CombinedOutput()
	status := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			status = exitErr.ExitCode()
		} else {
			status = 1
		}
	}
	return cmd, string(out), status
}

func writeTempFile(t *testing.T, name, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
}

func TestValidJSON(t *testing.T) {
	jsonPath := writeTempFile(t, "valid.json", `{"name":"Alice","age":30}`)
	schemaPath := writeTempFile(t, "schema.json", `{"type":"object","properties":{"name":{"type":"string"},"age":{"type":"number"}},"required":["name","age"]}`)
	_, out, status := runCLI(jsonPath, schemaPath)
	if status != 0 {
		t.Errorf("expected exit 0, got %d: %s", status, out)
	}
	if !strings.Contains(out, "Valid") {
		t.Errorf("expected 'Valid' in output, got: %s", out)
	}
}

func TestInvalidJSON(t *testing.T) {
	jsonPath := writeTempFile(t, "invalid.json", `{"name":123}`)
	schemaPath := writeTempFile(t, "schema.json", `{"type":"object","properties":{"name":{"type":"string"},"age":{"type":"number"}},"required":["name","age"]}`)
	_, out, status := runCLI(jsonPath, schemaPath)
	if status != 1 {
		t.Errorf("expected exit 1, got %d", status)
	}
	if !strings.Contains(out, "Invalid") {
		t.Errorf("expected 'Invalid' in output, got: %s", out)
	}
}

func TestMissingRequiredField(t *testing.T) {
	jsonPath := writeTempFile(t, "partial.json", `{"name":"Bob"}`)
	schemaPath := writeTempFile(t, "schema.json", `{"type":"object","properties":{"name":{"type":"string"},"age":{"type":"number"}},"required":["name","age"]}`)
	_, out, status := runCLI(jsonPath, schemaPath)
	if status != 1 {
		t.Errorf("expected exit 1, got %d", status)
	}
	if !strings.Contains(out, "required") {
		t.Errorf("expected 'required' in output, got: %s", out)
	}
}

func TestWrongType(t *testing.T) {
	jsonPath := writeTempFile(t, "wrong.json", `{"name":42}`)
	schemaPath := writeTempFile(t, "schema.json", `{"type":"object","properties":{"name":{"type":"string"}}}`)
	_, out, status := runCLI(jsonPath, schemaPath)
	if status != 1 {
		t.Errorf("expected exit 1, got %d", status)
	}
	if !strings.Contains(out, "Invalid type") {
		t.Errorf("expected 'Invalid type' in output, got: %s", out)
	}
}

func TestJSONFormatInvalid(t *testing.T) {
	jsonPath := writeTempFile(t, "invalid.json", `{"name":123}`)
	schemaPath := writeTempFile(t, "schema.json", `{"type":"object","properties":{"name":{"type":"string"}}}`)
	_, out, status := runCLI(jsonPath, schemaPath, "--format", "json")
	if status != 1 {
		t.Errorf("expected exit 1, got %d", status)
	}
	if !strings.Contains(out, `"valid": false`) {
		t.Errorf("expected 'valid: false' in JSON output, got: %s", out)
	}
	if !strings.Contains(out, `"errors"`) {
		t.Errorf("expected 'errors' key in JSON output, got: %s", out)
	}
}

func TestJSONFormatValid(t *testing.T) {
	jsonPath := writeTempFile(t, "valid.json", `{"name":"test"}`)
	schemaPath := writeTempFile(t, "schema.json", `{"type":"object","properties":{"name":{"type":"string"}}}`)
	_, out, status := runCLI(jsonPath, schemaPath, "--format", "json")
	if status != 0 {
		t.Errorf("expected exit 0, got %d", status)
	}
	if !strings.Contains(out, `"valid": true`) {
		t.Errorf("expected 'valid: true' in JSON output, got: %s", out)
	}
}

func TestSchemaFlag(t *testing.T) {
	jsonPath := writeTempFile(t, "data.json", `{"x":1}`)
	schemaPath := writeTempFile(t, "schema.json", `{"type":"object","properties":{"x":{"type":"number"}}}`)
	_, out, status := runCLI("-s", schemaPath, jsonPath)
	if status != 0 {
		t.Errorf("expected exit 0 with -s flag, got %d: %s", status, out)
	}
}

func TestArrayValidation(t *testing.T) {
	jsonPath := writeTempFile(t, "arr.json", `[1, "two", 3]`)
	schemaPath := writeTempFile(t, "schema.json", `{"type":"array","items":{"type":"number"}}`)
	_, _, status := runCLI(jsonPath, schemaPath)
	if status != 1 {
		t.Errorf("expected exit 1 for array with wrong item type, got %d", status)
	}
}

func TestNestedObject(t *testing.T) {
	jsonPath := writeTempFile(t, "nested.json", `{"user":{"name":"Alice","age":30}}`)
	schemaPath := writeTempFile(t, "schema.json", `{"type":"object","properties":{"user":{"type":"object","properties":{"name":{"type":"string"},"age":{"type":"number"}},"required":["name","age"]}}}`)
	_, _, status := runCLI(jsonPath, schemaPath)
	if status != 0 {
		t.Errorf("expected exit 0 for valid nested object, got %d", status)
	}
}

func TestEnumValidation(t *testing.T) {
	jsonPath := writeTempFile(t, "enum.json", `{"color":"red"}`)
	schemaPath := writeTempFile(t, "schema.json", `{"type":"object","properties":{"color":{"type":"string","enum":["red","green","blue"]}}}`)
	_, _, status := runCLI(jsonPath, schemaPath)
	if status != 0 {
		t.Errorf("expected exit 0 for valid enum value, got %d", status)
	}

	jsonPath2 := writeTempFile(t, "enum2.json", `{"color":"purple"}`)
	_, _, status2 := runCLI(jsonPath2, schemaPath)
	if status2 != 1 {
		t.Errorf("expected exit 1 for invalid enum value, got %d", status2)
	}
}

func TestMinMaxLength(t *testing.T) {
	jsonPath := writeTempFile(t, "short.json", `{"name":"A"}`)
	schemaPath := writeTempFile(t, "schema.json", `{"type":"object","properties":{"name":{"type":"string","minLength":2}}}`)
	_, _, status := runCLI(jsonPath, schemaPath)
	if status != 1 {
		t.Errorf("expected exit 1 for string too short, got %d", status)
	}
}

func TestPatternValidation(t *testing.T) {
	jsonPath := writeTempFile(t, "email.json", `{"email":"not-an-email"}`)
	schemaPath := writeTempFile(t, "schema.json", `{"type":"object","properties":{"email":{"type":"string","pattern":"^[\\w.+-]+@[\\w-]+\\.[\\w.]+$"}}}`)
	_, _, status := runCLI(jsonPath, schemaPath)
	if status != 1 {
		t.Errorf("expected exit 1 for invalid email pattern, got %d", status)
	}
}

func TestNumberRange(t *testing.T) {
	jsonPath := writeTempFile(t, "num.json", `{"score":150}`)
	schemaPath := writeTempFile(t, "schema.json", `{"type":"object","properties":{"score":{"type":"number","minimum":0,"maximum":100}}}`)
	_, _, status := runCLI(jsonPath, schemaPath)
	if status != 1 {
		t.Errorf("expected exit 1 for score out of range, got %d", status)
	}
}

func TestVersionFlag(t *testing.T) {
	_, out, status := runCLI("--version")
	if status != 0 {
		t.Errorf("expected exit 0 for --version, got %d", status)
	}
	if status == 0 && !strings.Contains(out, "jsonschemaval") {
		t.Errorf("expected version output to contain program name, got: %s", out)
	}
}

func TestNoArgs(t *testing.T) {
	_, out, status := runCLI()
	if status != 1 {
		t.Errorf("expected exit 1 with no args, got %d", status)
	}
	if !strings.Contains(out, "no arguments") {
		t.Errorf("expected 'no arguments' error, got: %s", out)
	}
}

func TestFileNotFound(t *testing.T) {
	_, _, status := runCLI("/nonexistent/file.json", "/nonexistent/schema.json")
	if status != 1 {
		t.Errorf("expected exit 1 for missing file, got %d", status)
	}
}

func TestAdditionalProperties(t *testing.T) {
	jsonPath := writeTempFile(t, "extra.json", `{"name":"test","extra":"field"}`)
	schemaPath := writeTempFile(t, "schema.json", `{"type":"object","properties":{"name":{"type":"string"}},"additionalProperties":false}`)
	_, _, status := runCLI(jsonPath, schemaPath)
	if status != 1 {
		t.Errorf("expected exit 1 for additional properties, got %d", status)
	}
}

func TestAllOf(t *testing.T) {
	jsonPath := writeTempFile(t, "allof.json", `{"name":"test","age":25}`)
	schemaPath := writeTempFile(t, "schema.json", `{"type":"object","allOf":[{"properties":{"name":{"type":"string"}},"required":["name"]},{"properties":{"age":{"type":"number"}},"required":["age"]}]}`)
	_, _, status := runCLI(jsonPath, schemaPath)
	if status != 0 {
		t.Errorf("expected exit 0 for allOf valid, got %d", status)
	}
}

func TestAnyOf(t *testing.T) {
	jsonPath := writeTempFile(t, "anyof.json", `{"name":"test"}`)
	schemaPath := writeTempFile(t, "schema.json", `{"type":"object","anyOf":[{"properties":{"email":{"type":"string"}},"required":["email"]},{"properties":{"name":{"type":"string"}},"required":["name"]}]}`)
	_, _, status := runCLI(jsonPath, schemaPath)
	if status != 0 {
		t.Errorf("expected exit 0 for anyOf valid, got %d", status)
	}
}

func TestOneOf(t *testing.T) {
	jsonPath := writeTempFile(t, "oneof.json", `{"name":"test"}`)
	schemaPath := writeTempFile(t, "schema.json", `{"type":"object","properties":{"name":{"type":"string"},"email":{"type":"string"},"phone":{"type":"string"}},"oneOf":[{"properties":{"email":{"type":"string"}},"required":["email"]},{"properties":{"phone":{"type":"string"}},"required":["phone"]}],"additionalProperties":false}`)
	_, _, status := runCLI(jsonPath, schemaPath)
	if status != 1 {
		t.Errorf("expected exit 1 for oneOf (matches neither), got %d", status)
	}
}
