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
