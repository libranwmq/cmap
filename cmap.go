package cmap

import (
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/libranwmq/cmap/common/jsonx"
	"github.com/libranwmq/cmap/common/nocopy"
)

// ConcurrentMapInterface is the interface of ConcurrentMap
type ConcurrentMapInterface[K Stringer, V any] interface {
	Count() int
	GetSlotNum() int
	IsEmpty() bool
	GetSlot(key K) *ConcurrentMapSlotted[K, V]
	Set(key K, value V)
	SetIfAbsent(key K, value V) bool
	MSet(data map[K]V)
	Upsert(key K, value V, cb UpsertCb[K, V]) V
	Get(key K) (V, bool)
	Pop(key K) (v V, ok bool)
	Has(key K) bool
	Remove(key K)
	RemoveCb(key K, cb RemoveCb[K, V]) bool
	Iter() <-chan Tuple[K, V]
	IterCb(key K, v V, cb IterCb[K, V])
	Items() map[K]V
	Keys() []K
	Tuples() []Tuple[K, V]
	Clear()
	UnmarshalJSON(b []byte) error
	MarshalJSON() ([]byte, error)
}

type SlotingFunc[K Stringer] func(key K) uint32

type concurrentMap[K Stringer, V any] struct {
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
func New[K Stringer, V any](opts ...ConcurrentMapOption[K, V]) ConcurrentMapInterface[K, V] {
	cm := &concurrentMap[K, V]{
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

// GetSlot get slot by key
func (c *concurrentMap[K, V]) GetSlot(key K) *ConcurrentMapSlotted[K, V] {
	return c.slots[c.sloting(key)%uint32(c.slotNum)]
}

// GetSlotNum get slot number
func (c *concurrentMap[K, V]) GetSlotNum() int {
	return c.slotNum
}

// GetSlotkeyNum get slot key number
func (c *concurrentMap[K, V]) GetSlotkeyNum(idx int) int {
	if idx < c.slotNum {
		slot := c.slots[idx]
		slot.RLock()
		defer slot.RUnlock()
		return len(slot.items)
	}
	return 0
}

// Count return the number of elements in the ConcurrentMap
func (c *concurrentMap[K, V]) Count() int {
	var count int
	for _, slot := range c.slots {
		slot.RLock()
		count += len(slot.items)
		slot.RUnlock()
	}
	return count
}

// IsEmpty return true if the ConcurrentMap is empty
func (c *concurrentMap[K, V]) IsEmpty() bool {
	return c.Count() == 0
}

// Set set a key-value pair
func (c *concurrentMap[K, V]) Set(key K, value V) {
	slot := c.GetSlot(key)
	slot.Lock()
	slot.items[key] = value
	slot.Unlock()
}

// SetIfAbsent set a key-value pair if the key is absent
func (c *concurrentMap[K, V]) SetIfAbsent(key K, value V) bool {
	slot := c.GetSlot(key)
	slot.Lock()
	defer slot.Unlock()
	_, ok := slot.items[key]
	if !ok {
		slot.items[key] = value
	}
	return !ok
}

// MSet set a key-value pair
func (c *concurrentMap[K, V]) MSet(data map[K]V) {
	for key, value := range data {
		c.Set(key, value)
	}
}

// Upsert set a key-value pair if the key is absent
func (c *concurrentMap[K, V]) Upsert(key K, value V, cb UpsertCb[K, V]) V {
	slot := c.GetSlot(key)
	slot.Lock()
	defer slot.Unlock()
	val, ok := slot.items[key]
	slot.items[key] = cb(ok, val, value)
	return slot.items[key]
}

// Get get a key-value pair
func (c *concurrentMap[K, V]) Get(key K) (V, bool) {
	slot := c.GetSlot(key)
	slot.RLock()
	val, ok := slot.items[key]
	slot.RUnlock()
	return val, ok
}

// Pop get and remove a key-value pair
func (c *concurrentMap[K, V]) Pop(key K) (v V, ok bool) {
	slot := c.GetSlot(key)
	slot.Lock()
	val, ok := slot.items[key]
	delete(slot.items, key)
	slot.Unlock()
	return val, ok
}

// Has check if a key exists
func (c *concurrentMap[K, V]) Has(key K) bool {
	slot := c.GetSlot(key)
	slot.RLock()
	_, ok := slot.items[key]
	slot.RUnlock()
	return ok
}

// Remove remove a key-value pair
func (c *concurrentMap[K, V]) Remove(key K) {
	slot := c.GetSlot(key)
	slot.Lock()
	delete(slot.items, key)
	slot.Unlock()
}

// RemoveCb remove a key-value pair if the callback returns true
func (c *concurrentMap[K, V]) RemoveCb(key K, cb RemoveCb[K, V]) bool {
	slot := c.GetSlot(key)
	slot.Lock()
	v, ok := slot.items[key]
	remove := cb(key, v, ok)
	if remove && ok {
		delete(slot.items, key)
	}
	slot.Unlock()
	return remove
}

// Iter return a channel of key-value pairs
func (c *concurrentMap[K, V]) Iter() <-chan Tuple[K, V] {
	chans := c.snapshot()
	total := 0
	for _, c := range chans {
		total += cap(c)
	}
	ch := make(chan Tuple[K, V], total)
	go fanIn(chans, ch)
	return ch
}

// IterCb iterate over the ConcurrentMap with a callback
func (c *concurrentMap[K, V]) IterCb(key K, v V, cb IterCb[K, V]) {
	for idx := range c.slots {
		slot := (c.slots)[idx]
		slot.RLock()
		for key, value := range slot.items {
			cb(key, value)
		}
		slot.RUnlock()
	}
}

// Items return a map of key-value pairs
func (c *concurrentMap[K, V]) Items() map[K]V {
	items := make(map[K]V)
	for t := range c.Iter() {
		items[t.Key] = t.Value
	}
	return items
}

// Keys return a slice of keys
func (c *concurrentMap[K, V]) Keys() []K {
	count := c.Count()
	ch := make(chan K, count)
	go func() {
		var eg errgroup.Group
		for _, slot := range c.slots {
			slot := slot
			eg.Go(func() error {
				slot.RLock()
				for key := range slot.items {
					ch <- key
				}
				slot.RUnlock()
				return nil
			})
		}
		_ = eg.Wait()
		close(ch)
	}()

	keys := make([]K, 0, count)
	for k := range ch {
		keys = append(keys, k)
	}
	return keys
}

// Tuples return a slice of key-value pairs
func (c *concurrentMap[K, V]) Tuples() []Tuple[K, V] {
	count := c.Count()
	ch := make(chan Tuple[K, V], count)
	go func() {
		var eg errgroup.Group
		for _, slot := range c.slots {
			slot := slot
			eg.Go(func() error {
				slot.RLock()
				for key, val := range slot.items {
					ch <- Tuple[K, V]{
						Key:   key,
						Value: val,
					}
				}
				slot.RUnlock()
				return nil
			})
		}
		_ = eg.Wait()
		close(ch)
	}()

	tuples := make([]Tuple[K, V], 0, count)
	for t := range ch {
		tuples = append(tuples, t)
	}
	return tuples
}

// Clear remove all key-value pairs
func (c *concurrentMap[K, V]) Clear() {
	for item := range c.Iter() {
		c.Remove(item.Key)
	}
}

// UnmarshalJSON unmarshal JSON to ConcurrentMap
func (c *concurrentMap[K, V]) UnmarshalJSON(b []byte) error {
	tmp := make(map[K]V)
	if err := jsonx.UnmarshalJSON(b, &tmp); err != nil {
		return err
	}
	c.MSet(tmp)
	return nil
}

// MarshalJSON marshal ConcurrentMap to JSON
func (c *concurrentMap[K, V]) MarshalJSON() ([]byte, error) {
	return jsonx.MarshalJSON(c.Items())
}

func (c *concurrentMap[K, V]) snapshot() []chan Tuple[K, V] {
	chans := make([]chan Tuple[K, V], c.slotNum)
	var eg errgroup.Group
	for index, slot := range c.slots {
		index := index
		slot := slot
		eg.Go(func() error {
			slot.RLock()
			chans[index] = make(chan Tuple[K, V], len(slot.items))
			for key, val := range slot.items {
				chans[index] <- Tuple[K, V]{key, val}
			}
			slot.RUnlock()
			close(chans[index])
			return nil
		})
	}
	_ = eg.Wait()
	return chans
}

func fanIn[K Stringer, V any](chans []chan Tuple[K, V], out chan Tuple[K, V]) {
	var eg errgroup.Group
	for _, ch := range chans {
		ch := ch
		eg.Go(func() error {
			for t := range ch {
				out <- t
			}
			return nil
		})
	}
	_ = eg.Wait()
	close(out)
}
