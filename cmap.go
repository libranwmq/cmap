package cmap

import (
	"sync"

	"github.com/libranwmq/cmap/common/nocopy"
)

type SlotingFunc[K Stringer] func(key K) uint32

type ConcurrentMap[K Stringer, V any] struct {
	_       nocopy.NoCopy
	slotNum int
	sloting SlotingFunc[K]
	slots   []*ConcurrentMapSlotted[K, V]
}

type ConcurrentMapSlotted[K Stringer, V any] struct {
	items map[K]V
	sync.RWMutex
}

// New create a new ConcurrentMap
func New[K Stringer, V any](opts ...ConcurrentMapOption[K, V]) *ConcurrentMap[K, V] {
	cm := &ConcurrentMap[K, V]{
		sloting: defaultSloting[K],
		slotNum: defaultSlotCount,
	}

	for _, opt := range opts {
		opt(cm)
	}

	cm.slots = make([]*ConcurrentMapSlotted[K, V], cm.slotNum)
	for i := range cm.slots {
		cm.slots[i] = &ConcurrentMapSlotted[K, V]{
			items: make(map[K]V),
		}
	}

	return cm
}
