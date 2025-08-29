# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Core Architecture

Tabularium is a Go-based schema definition and code generation system that provides universal data models for Chariot systems. The architecture follows a registry-driven approach where Go structs serve as the single source of truth for data representation across multiple languages and formats.

### Key Components

- **`pkg/model/model/`**: Core data models defined as Go structs with JSON and description tags
- **`pkg/registry/`**: Type registry system that enables model discovery and introspection
- **`pkg/schema/`**: OpenAPI schema generation from registered models
- **`cmd/schemagen/`**: CLI tool to generate OpenAPI 3.0 specifications from Go models
- **`cmd/codegen/`**: CLI tool to generate client libraries from OpenAPI schemas

### Model Registration Pattern

Every new model must be registered in its `init()` function:

```go
func init() {
    registry.Registry.MustRegisterModel(MyNewModel{})
}
```

This registration enables automatic schema generation and code generation for the model.

## Development Commands

### Schema and Code Generation
```bash
# Generate OpenAPI schema from registered Go models
go run ./cmd/schemagen -output client/api.yaml

# Generate Python client from OpenAPI schema
go run ./cmd/codegen -input client/api.yaml -gen py:client/python/tabularium
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests for specific package
go test ./pkg/model/model/

# Run Go vet
go vet ./...
```

### Go Module Management
```bash
# Tidy dependencies
go mod tidy
```

## Code Generation Workflow

The automated workflow (`.github/workflows/schema.yml`) runs on every push to main:

1. **Schema Generation**: Runs `schemagen` to create `client/api.yaml` from registered Go models
2. **Client Generation**: Runs `codegen` to generate Python Pydantic v2 models from the schema
3. **Auto-commit**: Automatically commits generated files back to the repository

This ensures that schema and client libraries stay synchronized with Go model changes.

## Adding New Models

1. Create Go struct in `pkg/model/model/` with appropriate tags
2. Add `init()` function to register the model with the registry
3. The automated workflow will generate corresponding OpenAPI schema and Python client
4. Models require `json` and `desc` struct tags for proper schema generation

## Python Client

The generated Python client uses Pydantic v2 and requires Python 3.11+. Install with:
```bash
pip install git+https://github.com/praetorian-inc/tabularium.git#subdirectory=client/python
```

## Dependencies

- Go 1.24.6+
- `datamodel-code-generator` for Python client generation
- Standard Go testing framework (no additional test frameworks used)