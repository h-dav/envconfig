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
