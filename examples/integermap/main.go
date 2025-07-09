package main

import (
	"fmt"

	"github.com/libranwmq/cmap"
)

func fnv32(key cmap.Integer) uint32 {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	keyLength := len(key.String())
	for i := 0; i < keyLength; i++ {
		hash *= prime32
		hash ^= uint32(key.String()[i])
	}
	return hash
}

func main() {
	opts := []cmap.ConcurrentMapOption[cmap.Integer, int]{
		cmap.WithConcurrentMapSlotNum[cmap.Integer, int](16),
		cmap.WithConcurrentMapSloting[cmap.Integer, int](fnv32),
	}
	scm := cmap.New[cmap.Integer, int](opts...)
	for i := 0; i < 100000; i++ {
		scm.Set(cmap.Integer(i), i)
	}
	fmt.Printf("concurrent map count: %d\n", scm.Count())
	scm.Clear()
	fmt.Printf("after clear concurrent map count: %d\n", scm.Count())
}
