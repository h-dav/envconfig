# envconfig

Package `envconfig` will populate your config struct based on sources such as environment variables, files, etc.

- [Installation](#installation)
- [Features](#features)
    - [Options](#options)
    - [Struct Tags](#tags)
    - [Other](#other)
- [Merging Values](#merging-values)

## Installation

```bash
go get github.com/h-dav/envconfig/v3
```

## Features

### Options

- `WithFilepath("config/file.env")`
> [!NOTE]
> Currently only .env files are supported.
- `WithActiveProfile("dev_env")`
- `WithPrefix("MY_APP")`

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
