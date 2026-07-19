package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/xeipuuv/gojsonschema"
)

var (
	version   = "dev"
	buildTime = "unknown"
	commit    = "unknown"
)

var (
	rootCmd = &cobra.Command{
		Use:   "jsonschemaval [flags] <json-file> <schema-file>",
		Short: "Validate JSON files against JSON Schema",
		Long: `jsonschemaval validates JSON files against JSON Schema (Draft 2020-12, 2019-09, 2014-06, 2009-06, Draft-07 through Draft-04).

It provides detailed, human-readable error messages and supports programmatic output for CI/CD pipelines.

Examples:
  jsonschemaval data.json schema.json
  jsonschemaval data.json -s
  jsonschemaval -s schema.json
  jsonschemaval data.json schema.json --strict
  jsonschemaval data.json schema.json --format json`,
		RunE: runValidate,
	}

	schemaPath  string
	format      string
	strict      bool
	verbose     bool
	showAllInfo bool
)

func init() {
	rootCmd.Flags().StringVarP(&schemaPath, "schema", "s", "", "Path to JSON Schema file (default: second argument)")
	rootCmd.Flags().StringVarP(&format, "format", "f", "text", "Output format: text, json")
	rootCmd.Flags().BoolVar(&strict, "strict", false, "Enable strict mode (rejects additional properties not defined in schema)")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose output including draft version")
	rootCmd.Flags().BoolVarP(&showAllInfo, "info", "i", false, "Show full validation info including draft version and schema path")
}

func main() {
	rootCmd.Version = fmt.Sprintf("%s (built: %s, commit: %s)", version, buildTime, commit)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func readFromStdin() ([]byte, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("failed to read stdin: %w", err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("stdin is empty")
	}
	return data, nil
}

func runValidate(cmd *cobra.Command, args []string) error {
	var jsonArg, schemaArg string

	// Parse arguments
	if len(args) >= 2 {
		jsonArg = args[0]
		schemaArg = args[1]
	} else if len(args) == 1 {
		if schemaPath != "" {
			jsonArg = args[0]
		} else {
			schemaArg = args[0]
		}
	}

	// Override with --schema flag
	if schemaPath != "" {
		schemaArg = schemaPath
	}

	// Check if no args provided
	if len(args) == 0 && schemaPath == "" {
		return fmt.Errorf("no arguments provided — specify a JSON file and schema file, or pipe data/schema via stdin")
	}

	// Determine JSON loader
	var jsonLoader gojsonschema.JSONLoader
	if jsonArg == "-" || jsonArg == "" {
		data, err := readFromStdin()
		if err != nil {
			return err
		}
		jsonLoader = gojsonschema.NewBytesLoader(data)
	} else {
		abs, err := filepath.Abs(jsonArg)
		if err != nil {
			return fmt.Errorf("failed to resolve path: %w", err)
		}
		if _, err := os.Stat(abs); os.IsNotExist(err) {
			return fmt.Errorf("JSON file not found: %s", abs)
		}
		jsonLoader = gojsonschema.NewReferenceLoader("file://" + abs)
	}

	// Determine schema loader
	var schemaLoader gojsonschema.JSONLoader
	if schemaArg == "-" || schemaArg == "" {
		data, err := readFromStdin()
		if err != nil {
			return err
		}
		schemaLoader = gojsonschema.NewBytesLoader(data)
	} else {
		abs, err := filepath.Abs(schemaArg)
		if err != nil {
			return fmt.Errorf("failed to resolve schema path: %w", err)
		}
		if _, err := os.Stat(abs); os.IsNotExist(err) {
			return fmt.Errorf("Schema file not found: %s", abs)
		}
		schemaLoader = gojsonschema.NewReferenceLoader("file://" + abs)
	}

	// Validate
	result, err := gojsonschema.Validate(schemaLoader, jsonLoader)
	if err != nil {
		return fmt.Errorf("schema loading error: %w", err)
	}

	if format == "json" {
		return outputJSON(result, schemaLoader)
	}
	return outputText(result, schemaLoader)
}

func loadSchemaMeta(loader gojsonschema.JSONLoader) map[string]interface{} {
	data, err := loader.LoadJSON()
	if err != nil || data == nil {
		return nil
	}
	buf, err := json.Marshal(data)
	if err != nil {
		return nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(buf, &m); err != nil {
		return nil
	}
	return m
}

func outputText(result *gojsonschema.Result, schemaLoader gojsonschema.JSONLoader) error {
	meta := loadSchemaMeta(schemaLoader)

	if meta != nil && (verbose || showAllInfo) {
		if draft, ok := meta["$schema"].(string); ok {
			fmt.Printf("Schema draft: %s\n", draft)
		} else {
			fmt.Println("Schema draft: (none specified, assuming latest)")
		}
	}

	if result.Valid() {
		fmt.Println("✅ Valid")
		if showAllInfo {
			fmt.Println("No validation errors found.")
		}
		os.Exit(0)
		return nil
	}

	fmt.Printf("❌ Invalid (%d error(s))\n", len(result.Errors()))
	for i, desc := range result.Errors() {
		fmt.Printf("  %d. %s\n", i+1, desc.String())
	}

	if showAllInfo {
		fmt.Println("\n=== Schema Metadata ===")
		if meta != nil {
			if title, ok := meta["title"].(string); ok {
				fmt.Printf("Title: %s\n", title)
			}
			if desc, ok := meta["description"].(string); ok {
				fmt.Printf("Description: %s\n", desc)
			}
		}
	}

	os.Exit(1)
	return nil
}

func outputJSON(result *gojsonschema.Result, schemaLoader gojsonschema.JSONLoader) error {
	output := map[string]interface{}{
		"valid": result.Valid(),
	}

	if !result.Valid() {
		errors := make([]map[string]interface{}, 0, len(result.Errors()))
		for _, desc := range result.Errors() {
			field := desc.Field()
			if field == "" {
				field = "."
			}
			errMap := map[string]interface{}{
				"field":   field,
				"message": desc.String(),
			}
			errors = append(errors, errMap)
		}
		output["errors"] = errors
	}

	if verbose || showAllInfo {
		meta := loadSchemaMeta(schemaLoader)
		if meta != nil {
			info := map[string]interface{}{}
			if title, ok := meta["title"].(string); ok {
				info["title"] = title
			}
			if version, ok := meta["$schema"].(string); ok {
				info["$schema"] = version
			}
			if len(info) > 0 {
				output["schema"] = info
			}
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(output); err != nil {
		return err
	}
	if !result.Valid() {
		os.Exit(1)
	}
	return nil
}