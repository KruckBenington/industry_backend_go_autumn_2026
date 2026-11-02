package main

import "container/list"

type entry[K comparable, V any] struct {
	key   K
	value V
}
type LRUCache[K comparable, V any] struct {
	capacity int
	ll       list.List
	items    map[K]*list.Element
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		capacity: capacity,
		ll:       list.List{},
		items:    make(map[K]*list.Element, capacity),
	}
}
func (c *LRUCache[K, V]) Get(key K) (value V, ok bool) {
	if value, ok := c.items[key]; ok {
		currentEntry, typeOk := value.Value.(entry[K, V])
		if !typeOk {
			panic("Wrong")
		}
		c.ll.MoveToBack(value)
		return currentEntry.value, ok
	}

	return *new(V), false
}
func (c *LRUCache[K, V]) Set(key K, value V) {
	if c.capacity <= 0 {
		return
	}

	if _, ok := c.items[key]; ok {
		c.items[key].Value = entry[K, V]{key: key, value: value}
		return
	}

	if len(c.items) >= c.capacity {
		frontEl := c.ll.Front()
		currEntry, typeOk := frontEl.Value.(entry[K, V])
		if !typeOk {
			panic("Wrong")
		}
		delete(c.items, currEntry.key)
		c.ll.Remove(frontEl)
	}

	el := c.ll.PushBack(entry[K, V]{key: key, value: value})
	c.items[key] = el
	return
}

type LRU[K comparable, V any] interface {
	Get(K) (V, bool)
	Set(K, V)
}
