# envconfig

Package `envconfig` will populate your config struct based on sources such as environment variables, files, etc.

- [Installation](#installation)
- [Features](#features)
    - [Options](#options)
    - [Struct Tags](#tags)
    - [Other](#other)
    - [Examples](#examples)
- [Merging Values](#merging-values)

## Installation

```bash
go get github.com/h-dav/envconfig/v3
```

## Features

### Options

| Option                            | Description                                           |
|-----------------------------------|-------------------------------------------------------|
| `WithFilepath("config/file.env")` | Use file to populate config struct.                   |
| `WithActiveProfile("dev_env")`    | Provide the profile to select a specific config file. |

### Struct Tags

- `env`: Used to determine the key of the value to use when populating config fields.
- `required`: `true` or `false`
- `default`: Default value if environment variable is not set.
- `prefix`: Used for nested structures.
- `envjson`: Used for deserialising JSON into config.

### Other

- Text Replacement: `${EXAMPLE}` can be used to insert other discovered values.

## Merging Values

> [!IMPORTANT]
> When merging values, `envconfig` uses the following precedence:
> 1. Flags
> 2. Environment Variables
> 3. Config File (provided via `WithFilepath()`)

## Examples

### Basic

```go
func main() {
    type Config struct {
        Development bool `env:"DEVELOPMENT" default:"true"`
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
        Service string `env:"SERVICE"`
    }

    var cfg Config

    if err := envconfig.Set(&cfg, WithFilepath("internal/config/config.env")); err != nil {
        panic(err)
    }
}
```

### Profile

```go
func main() {
    type Config struct {
        Service string `env:"SERVICE"`
    }

    var cfg Config

    if err := envconfig.Set(
        &cfg,
        WithActiveProfile(os.Getenv("ACTIVE_PROFILE")),
        WithFilepath("internal/config/"), // Will use file `internal/config/development.env`.
    ); err != nil {
        panic(err)
    }
}
```

### Nested Structs

```go
func main() {
    type Config struct {
        Service struct {
            Name string `env:"NAME"`
            Version string `env:"VERSION"`
        } `prefix:"SERVICE_"`
    }

    var cfg Config

    if err := envconfig.Set(&cfg); err != nil {
        panic(err)
    }
}
```
