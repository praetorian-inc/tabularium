# Tabularium

`tabularium` contains the schema of universal data objects for chariot and chariot subsystems, as well as supporting tooling for validation, code generation, and testing.

## Models and Schema

The core data structures are defined as Go structs within the `pkg/model` directory. These models serve as the single source of truth for data representation across Chariot systems.

### Adding a New Model

To add a new data model:

1.  **Define the Struct:** Create a new Go struct in a relevant file within `pkg/models` (e.g., `pkg/models/mynewmodel.go`). Use standard Go types and include necessary struct tags like `json` and `description`.
2.  **Register the Model:** Within the same file (e.g., `pkg/models/mynewmodel.go`), add an `init()` function that registers your model type with the central registry. This makes the model discoverable by the schema generation tools.

    ```go
    func init() {
        // Register the model using its zero value
        registry.Registry.MustRegisterModel(MyNewModel{})
    }

    type MyNewModel struct {
        Field string `json:"field" desc:"A description of the model field."`
    }
    ```

## Schema Generation and Code Generation

The registered models are used to automatically generate consistent artifacts, ensuring that different parts of the system agree on data structures.

1.  **Schema Generation (`cmd/schemagen`):** The `cmd/schemagen` tool inspects all registered models in `pkg/model` via the registry mechanism in `pkg/schema`. It generates a standard OpenAPI 3.0 specification file located at `client/api.yaml`. This YAML file describes all registered models, their fields, types, and descriptions.
2.  **Code Generation (`cmd/codegen`):** The `cmd/codegen` tool takes the generated `client/api.yaml` as input and can produce client libraries or data model implementations for various languages. Currently, it supports generating Python Pydantic v2 models using `datamodel-codegen`. The GitHub Actions workflow automatically runs `schemagen` and `codegen` to keep the schema and Python client (`client/python/tabularium`) up-to-date (see `.github/workflows/schema.yml`).

This process allows developers to define a model once in Go and automatically get corresponding representations in other languages or formats.

## Future Possibilities

Using a registry of go types + an openapi schema allows for a number of future possibilities:

*   **Client Libraries:** Generating type-safe client libraries for languages beyond Python (e.g., TypeScript, Java).
*   **APIs, and API Documentation:** Automatically generating APIs for interacting with data objects, and the accompanying documentation.
*   **Dynamic Functionality:** By maintaining the source of truth in go, we can standardize and encapsulate additional dynamic functionality with the go API, using a common type registry.

## Client Libraries (WIP)

To install the python client library: `pip install git+https://github.com/praetorian-inc/tabularium.git#subdirectory=client/python`
