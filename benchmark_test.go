package tlc3

import (
	"context"
	"testing"
	"time"
)

func Benchmark(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetCertList(context.Background(), []string{"localhost:8443"}, 5*time.Second, true, time.Local)
		if err != nil {
			b.Fatal(err)
		}
	}
}
