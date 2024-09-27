/*
Package github.com/h-dav/envconfig is for parsing .env configuration files for you Go application.

Your define your configuration structure like so:

	type ExampleConfig struct {
		Example      string  `env:"EXAMPLE,required"`
		AnotherValue string  `env:"ANOTHER_VALUE"`
		IntExample   int     `env:"INT_EXAMPLE"`
		Int32Example int32   `env:"INT32_EXAMPLE"`
		FloatExample float64 `env:"FLOAT_EXAMPLE"`
		Service      struct {
			Port int64  `env:"PORT"`
			Name string `env:"NAME,required"`
		} `env:"prefix=HTTP_"`
		ExampleEndpoint string `env:"EXAMPLE_ENDPOINT"`
		DefaultValue    string `env:"DEFAULT_VALUE,default=thevalue"`
	}

Then you can either read your configuration file and populate the structure like so (you can also just set environment variables by using envconfig.SetVars(filename)):

	cfg := ExampleConfig{}
	if err := envconfig.SetPopulate(filepath.Join("config", "example.env"), &cfg); err != nil {
		log.Fatal(err)
	}

Or just populate your structure with environment variables that are already set like so:

	cfg := ExampleConfig{}
	if err := envconfig.Populate(&cfg); err != nil {
		log.Fatal(err)
	}
*/
package envconfig
