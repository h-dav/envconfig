package envconfig_test

import "time"

type SuccessWithOneField struct {
	Example string `config:"KEY"`
}
type SuccessWithOneIntField struct {
	Example int `config:"KEY"`
}

type SuccessWithDefaultValueAndEmptyEnvFile struct {
	Example string `config:"DEFAULT_VALUE,default=value2"`
}

type SuccessWithRequiredField struct {
	Example string `config:"REQUIRED_VALUE,required"`
}

type SuccessWithTextReplacement struct {
	ReplaceField string `config:"REPLACE_FIELD"`
}

type SuccessWithSettingTimeDuration struct {
	Duration time.Duration `config:"DURATION"`
}

type SuccessWithPrefixOption struct {
	Duration time.Duration `config:"DURATION"`
}
