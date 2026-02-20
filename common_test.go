package envconfig_test

import "time"

type SuccessWithOneField struct {
	Example string `env:"KEY"`
}
type SuccessWithOneIntField struct {
	Example int `env:"KEY"`
}

type SuccessWithDefaultValueAndEmptyEnvFile struct {
	Example string `env:"DEFAULT_VALUE,default=value2"`
}

type SuccessWithRequiredField struct {
	Example string `env:"REQUIRED_VALUE,required"`
}

type SuccessWithTextReplacement struct {
	ReplaceField string `env:"REPLACE_FIELD"`
}

type SuccessWithSettingTimeDuration struct {
	Duration time.Duration `env:"DURATION"`
}

type SuccessWithPrefixOption struct {
	Duration time.Duration `env:"DURATION"`
}
