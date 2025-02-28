/*
Package envconfig is for parsing .env configuration files for your Go application.

Create your .env file to set environment variables:

	LOG_LEVEL: info
	SERVER_PORT: 8080

Populate your config structure:

	type Config struct {
	    LogLevel string `env:"LOG_LEVEL"`
	    Server struct {
	        Port string `env:"PORT"`
	    } `prefix:"SERVER_"`
	}

	var cfg Config

	if err := envconfig.Set("default.env", &cfg); err != nil {
	    // handle error.
	}

# Options

[required], [default], [prefix] are options that can be set via struct tags:

	type Config struct {
	    RequiredField string `env:"REQUIRED_FIELD" required:"true"`
	    DefaultField string `env:"DEFAULT_FIELD" default:"defaulted value"`
	    NestedConfig struct {
	        ExampleField string `env:"EXAMPLE_FIELD"`
	    } `prefix:"NESTED_CONFIG_"`
	}
*/
package envconfig
