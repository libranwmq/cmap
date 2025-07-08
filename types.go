package cmap

import (
	"fmt"
	"hash/fnv"
	"strconv"
)

// Stringer 字符串化接口
type Stringer interface {
	fmt.Stringer
	comparable
}

var defaultSlotCount = 32

func defaultSloting[K Stringer](key K) uint32 {
	hasher := fnv.New32a()
	if _, err := hasher.Write([]byte(key.String())); err != nil {
		panic(err)
	}
	return hasher.Sum32()
}

type Integer int
type Integer32 int32

func (i Integer) String() string {
	return strconv.Itoa(int(i))
}

func (i Integer32) String() string {
	return strconv.Itoa(int(i))
}

type ConcurrentMapOption[K Stringer, V any] func(*ConcurrentMap[K, V])

func WithConcurrentMapSlotNum[K Stringer, V any](slotNum int) ConcurrentMapOption[K, V] {
	return func(cm *ConcurrentMap[K, V]) {
		if slotNum > 0 {
			cm.slotNum = slotNum
		}
	}
}

func WithConcurrentMapSloting[K Stringer, V any](sloting func(key K) uint32) ConcurrentMapOption[K, V] {
	return func(cm *ConcurrentMap[K, V]) {
		if sloting != nil {
			cm.sloting = sloting
		}
	}
}
