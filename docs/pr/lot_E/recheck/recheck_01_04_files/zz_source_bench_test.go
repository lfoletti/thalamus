package authz

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func benchKeyFile(b *testing.B, n int) string {
	var parts []string
	for i := 0; i < n; i++ {
		parts = append(parts, fmt.Sprintf(`{"id":"k%d","secret":"secret-%032d","clearance":"read-only","scope":{"collection":"c%d"}}`, i, i, i))
	}
	path := filepath.Join(b.TempDir(), "keys.json")
	if err := os.WriteFile(path, []byte(`{"keys":[`+strings.Join(parts, ",")+`]}`), 0o600); err != nil {
		b.Fatal(err)
	}
	return path
}

func BenchmarkStatOnly(b *testing.B) {
	path := benchKeyFile(b, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := os.Stat(path); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLookupFrozen(b *testing.B) {
	for _, n := range []int{1, 100, 1000} {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			ks, err := Load(benchKeyFile(b, n))
			if err != nil {
				b.Fatal(err)
			}
			secret := fmt.Sprintf("secret-%032d", n-1)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, ok := ks.Lookup(secret); !ok {
					b.Fatal("miss")
				}
			}
		})
	}
}

func BenchmarkLookupThroughSource(b *testing.B) {
	for _, n := range []int{1, 100, 1000} {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			s, err := protoOpen(benchKeyFile(b, n))
			if err != nil {
				b.Fatal(err)
			}
			secret := fmt.Sprintf("secret-%032d", n-1)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, ok := s.Current().Lookup(secret); !ok {
					b.Fatal("miss")
				}
			}
		})
	}
}

func BenchmarkLookupThroughSourceParallel(b *testing.B) {
	s, err := protoOpen(benchKeyFile(b, 100))
	if err != nil {
		b.Fatal(err)
	}
	secret := fmt.Sprintf("secret-%032d", 99)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, ok := s.Current().Lookup(secret); !ok {
				b.Fatal("miss")
			}
		}
	})
}

func BenchmarkSourceStatOutsideLockParallel(b *testing.B) {
	s, err := protoOpen(benchKeyFile(b, 100))
	if err != nil {
		b.Fatal(err)
	}
	secret := fmt.Sprintf("secret-%032d", 99)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, ok := s.currentStatOutsideLock().Lookup(secret); !ok {
				b.Fatal("miss")
			}
		}
	})
}

func BenchmarkLookupFrozenParallel(b *testing.B) {
	ks, err := Load(benchKeyFile(b, 100))
	if err != nil {
		b.Fatal(err)
	}
	secret := fmt.Sprintf("secret-%032d", 99)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, ok := ks.Lookup(secret); !ok {
				b.Fatal("miss")
			}
		}
	})
}

func BenchmarkStatOnlyParallel(b *testing.B) {
	path := benchKeyFile(b, 1)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := os.Stat(path); err != nil {
				b.Fatal(err)
			}
		}
	})
}
