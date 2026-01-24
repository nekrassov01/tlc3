package tlc3

import (
	"context"
	"testing"
	"time"
)

func Benchmark_Single(b *testing.B) {
	for b.Loop() {
		_, err := GetCerts(context.Background(), []string{"localhost:8443"}, 5*time.Second, true, time.Local)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Multiple(b *testing.B) {
	for b.Loop() {
		_, err := GetCerts(context.Background(), []string{"localhost:8443", "127.0.0.1:8443"}, 5*time.Second, true, time.Local)
		if err != nil {
			b.Fatal(err)
		}
	}
}
