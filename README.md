# envconfig

Package `envconfig` will populate your config struct based on sources such as environment variables, flags, files, etc.

- [Installation](#installation)
- [Features](#features)
    - [Options](#options)
    - [Struct Tags](#tags)
    - [Other](#other)
    - [Examples](#examples)
- [Merging Values](#merging-values)
- [Error Handling](#error-handling)
- [Benchmarking](#benchmarking)

## Installation

```bash
go get github.com/h-dav/envconfig/v3
```

## Features

### Options

| Option                            | Description                                           |
|-----------------------------------|-------------------------------------------------------|
| `WithFilepath("config/file.env")` | Use file to populate config struct.                   |
| `WithActiveProfile("internal/config/", "dev")`    | Provide the path and profile to select a specific config file (e.g., `internal/config/dev.env`). |
| `WithPrefix("MYAPP_")`            | Add a global prefix to all environment variable lookups. |
| `WithLogger(logger)`              | Provide a `*slog.Logger` for internal diagnostics.     |
| `WithDecoders(map)`               | Register custom decoders for specific types.           |

### Struct Tags

- `config`: Used to define configuration mapping and options.
  - Format: `config:"NAME[,option1,option2=value,...]"`
  - Options:
    - `required`: Marks the field as mandatory.
    - `default=<value>`: Specifies a default value if the configuration key is not set in any source.
    - `prefix=<prefix>`: Specifies a prefix for nested structs.
    - `json`: Indicates that the value should be deserialized from JSON.

### Other

- Text Replacement: `${EXAMPLE}` can be used to insert other discovered values.

## Merging Values

> [!IMPORTANT]
> When merging values, `envconfig` uses a strict precedence order. If a key is present in multiple sources, the value from the source with the highest precedence will be used (overwriting any lower precedence values).
>
> The precedence order (from highest to lowest) is:
> 1. **Flags:** Command-line flags.
> 2. **Environment Variables:** Application environment variables.
> 3. **Config Files:** Provided via `WithFilepath()` or `WithActiveProfile()`.
> 4. **Defaults:** Defined in struct tags using `,default=...`.

## Error Handling

`envconfig` uses structured error types to allow for robust error checking using standard library functions like `errors.Is` and `errors.As`.

### Checking for Specific Errors

```go
err := envconfig.Set(&cfg)
if err != nil {
    if errors.Is(err, envconfig.ErrRequired) {
        // Handle missing required field
    }
    
    var jsonErr *envconfig.JSONUnmarshalError
    if errors.As(err, &jsonErr) {
        fmt.Printf("JSON error in field %s: %v\n", jsonErr.FieldName, jsonErr.Err)
    }
}
```

### Base Error Categories

- `ErrInvalidConfig`: Output is not a pointer to a struct.
- `ErrUnsupported`: Field type is not supported.
- `ErrRequired`: A required field was missing.
- `ErrFile`: Issues opening or reading a config file.
- `ErrParse`: Syntax errors in `.env` files.
- `ErrConversion`: Errors converting string values to target types.
- `ErrJSON`: Errors unmarshaling JSON values.
- `ErrTag`: Errors in struct tag definitions.

## Examples

### Basic

```go
func main() {
    type Config struct {
        Development bool `config:"DEVELOPMENT,default=true"`
    }

    var cfg Config

    if err := envconfig.Set(&cfg); err != nil {
        panic(err)
    }
}
```

### Config File

```go
func main() {
    type Config struct {
        Service string `config:"SERVICE,required"`
    }

    var cfg Config

    if err := envconfig.Set(&cfg, envconfig.WithFilepath("internal/config/config.env")); err != nil {
        panic(err)
    }
}
```

### Diagnostics with slog

```go
func main() {
    logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
    
    var cfg Config
    if err := envconfig.Set(&cfg, envconfig.WithLogger(logger)); err != nil {
        panic(err)
    }
}
```
\n## Benchmarking\n\nThe library includes a benchmarking suite to track performance and memory allocations. To run the benchmarks, use the following command:\n\n```bash\nmake bench\n```\n\nThis will execute benchmarks for tag parsing, source loading, and end-to-end struct population, providing metrics such as execution time (ns/op) and memory allocations (B/op and allocs/op).
