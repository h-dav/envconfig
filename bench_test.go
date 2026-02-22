package envconfig

import (
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
