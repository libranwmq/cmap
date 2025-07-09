package cmap

import (
	"fmt"
	"hash/fnv"
	"strconv"
)

// Stringer is a stringer interface.
type Stringer interface {
	fmt.Stringer
	comparable
}

// defaultSlotCount is the default slot count.
var defaultSlotCount = 32

// defaultSloting is the default sloting function.
func defaultSloting[K Stringer](key K) uint32 {
	hasher := fnv.New32a()
	if _, err := hasher.Write([]byte(key.String())); err != nil {
		panic(err)
	}
	return hasher.Sum32()
}

type String string
type Integer int
type Integer32 int32

func (s String) String() string {
	return string(s)
}

func (i Integer) String() string {
	return strconv.Itoa(int(i))
}

func (i Integer32) String() string {
	return strconv.Itoa(int(i))
}

type ConcurrentMapOption[K Stringer, V any] func(*concurrentMap[K, V])

// WithConcurrentMapSlotNum is a concurrent map option that sets the slot number.
func WithConcurrentMapSlotNum[K Stringer, V any](slotNum int) ConcurrentMapOption[K, V] {
	return func(cm *concurrentMap[K, V]) {
		if slotNum > 0 {
			cm.slotNum = slotNum
		}
	}
}

// WithConcurrentMapSloting is a concurrent map option that sets the sloting function.
func WithConcurrentMapSloting[K Stringer, V any](sloting func(key K) uint32) ConcurrentMapOption[K, V] {
	return func(cm *concurrentMap[K, V]) {
		if sloting != nil {
			cm.sloting = sloting
		}
	}
}

// Tuple is a key-value pair.
type Tuple[K Stringer, V any] struct {
	Key   K
	Value V
}

type UpsertCb[K Stringer, V any] func(exist bool, valueInMap V, newValue V) V
type RemoveCb[K Stringer, V any] func(key K, v V, exists bool) bool
type IterCb[K Stringer, V any] func(key K, v V)

// DefaultUpsertCb is the default upsert callback function.
func DefaultUpsertCb[K Stringer, V any](exist bool, valueInMap V, newValue V) V {
	if exist {
		return valueInMap
	}
	return newValue
}

// DefaultRemoveCb is the default remove callback function.
func DefaultRemoveCb[K Stringer, V any](key K, v V, exists bool) bool {
	return exists
}

// DefaultIterCb is the default iterator callback function.
func DefaultIterCb[K Stringer, V any](key K, v V) {

}
