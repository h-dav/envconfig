package envconfig

import (
	"os"
	"testing"
)

func BenchmarkParseTag_Simple(b *testing.B) {
	tag := "PORT,required"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseTag(tag)
	}
}

func BenchmarkParseTag_Complex(b *testing.B) {
	tag := "PORT,required,default=8080,prefix=API_,json"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseTag(tag)
	}
}

func BenchmarkParseTag_Whitespace(b *testing.B) {
	tag := "  PORT  ,  required  ,  default = 8080  ,  prefix = API_  ,  json  "
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseTag(tag)
	}
}

func BenchmarkFileSource_Load(b *testing.B) {
	content := "KEY1=VAL1\nKEY2=VAL2\nKEY3=VAL3\n# Comment\nKEY4=VAL4 # inline comment"
	tmpfile, err := os.CreateTemp("", "bench*.env")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		b.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		b.Fatal(err)
	}

	src := FileSource{filepath: tmpfile.Name()}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = src.Load()
	}
}

func BenchmarkEnvironmentVariableSource_Load(b *testing.B) {
	src := EnvironmentVariableSource{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = src.Load()
	}
}

func BenchmarkSet_Small(b *testing.B) {
	type Config struct {
		Key string `config:"SMALL_KEY"`
	}
	_ = os.Setenv("SMALL_KEY", "value")
	defer os.Unsetenv("SMALL_KEY")

	var cfg Config
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Set(&cfg)
	}
}

func BenchmarkSet_Medium(b *testing.B) {
	type Config struct {
		K1 string `config:"K1"`
		K2 int    `config:"K2"`
		K3 bool   `config:"K3"`
		K4 string `config:"K4,default=def"`
		K5 []int  `config:"K5"`
	}
	_ = os.Setenv("K1", "v1")
	_ = os.Setenv("K2", "123")
	_ = os.Setenv("K3", "true")
	_ = os.Setenv("K5", "1,2,3")
	defer func() {
		os.Unsetenv("K1")
		os.Unsetenv("K2")
		os.Unsetenv("K3")
		os.Unsetenv("K5")
	}()

	var cfg Config
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Set(&cfg)
	}
}

func BenchmarkSet_Large(b *testing.B) {
	type Nested struct {
		N1 string `config:"N1"`
		N2 int    `config:"N2"`
	}
	type Config struct {
		K1  string   `config:"K1"`
		K2  int      `config:"K2"`
		K3  bool     `config:"K3"`
		K4  string   `config:"K4,default=def"`
		K5  []int    `config:"K5"`
		K6  float64  `config:"K6"`
		K7  string   `config:"K7,required"`
		K8  Nested   `config:",prefix=PRE_"`
		K9  []string `config:"K9"`
		K10 string   `config:"K10"`
	}
	_ = os.Setenv("K1", "v1")
	_ = os.Setenv("K2", "123")
	_ = os.Setenv("K3", "true")
	_ = os.Setenv("K5", "1,2,3")
	_ = os.Setenv("K6", "1.23")
	_ = os.Setenv("K7", "req")
	_ = os.Setenv("PRE_N1", "nv1")
	_ = os.Setenv("PRE_N2", "456")
	_ = os.Setenv("K9", "a,b,c")
	_ = os.Setenv("K10", "v10")
	defer func() {
		os.Unsetenv("K1")
		os.Unsetenv("K2")
		os.Unsetenv("K3")
		os.Unsetenv("K5")
		os.Unsetenv("K6")
		os.Unsetenv("K7")
		os.Unsetenv("PRE_N1")
		os.Unsetenv("PRE_N2")
		os.Unsetenv("K9")
		os.Unsetenv("K10")
	}()

	var cfg Config
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Set(&cfg)
	}
}
