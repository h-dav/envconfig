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

- `env`: Used to define environment variable mapping and options.
  - Format: `env:"NAME[,option1,option2=value,...]"`
  - Options:
    - `required`: Marks the field as mandatory.
    - `default=<value>`: Specifies a default value if the environment variable is not set.
    - `prefix=<prefix>`: Specifies a prefix for nested structs.
    - `envjson`: Indicates that the value should be deserialized from JSON.

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
        Development bool `env:"DEVELOPMENT,default=true"`
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
        Service string `env:"SERVICE,required"`
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

    activeProfile := os.Getenv("ACTIVE_PROFILE")
    if activeProfile == "" {
        activeProfile = "default"
    }

    if err := envconfig.Set(
        &cfg,
        WithActiveProfile("internal/config/", activeProfile),
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
        } `env:",prefix=SERVICE_"`
    }

    var cfg Config

    if err := envconfig.Set(&cfg); err != nil {
        panic(err)
    }
}
```
