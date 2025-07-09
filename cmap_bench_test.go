package cmap

import (
	"sync"
	"testing"
)

func Benchmark_Items(b *testing.B) {
	m1 := New[Integer, int]()
	for i := 0; i < 10000; i++ {
		m1.Set(Integer(i), i)
	}
	for i := 0; i < b.N; i++ {
		m1.Items()
	}
}

func Benchmark_SyncMap(b *testing.B) {
	m2 := sync.Map{}
	for i := 0; i < 10000; i++ {
		m2.Store(i, i)
	}
	for i := 0; i < b.N; i++ {
		m2.Range(func(key, value any) bool {
			return true
		})
	}
}
