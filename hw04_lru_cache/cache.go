package hw04lrucache

import "sync"

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	mu       sync.Mutex
	capacity int
	queue    List
	items    map[Key]*ListItem
}

type cacheElement struct {
	Key Key
	Val interface{}
}

func (lru *lruCache) Set(key Key, value interface{}) bool {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	newCacheElement := cacheElement{Key: key, Val: value}
	if item, found := lru.items[key]; found {
		item.Value = newCacheElement
		lru.queue.MoveToFront(item)
		return true
	}

	newItem := lru.queue.PushFront(newCacheElement)
	lru.items[key] = newItem

	if lru.queue.Len() > lru.capacity {
		lastItem := lru.queue.Back()
		if lastItem != nil {
			delete(lru.items, lastItem.Value.(cacheElement).Key)
			lru.queue.Remove(lastItem)
		}
	}
	return false
}

func (lru *lruCache) Get(key Key) (interface{}, bool) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if item, found := lru.items[key]; found {
		lru.queue.MoveToFront(item)
		return item.Value.(cacheElement).Val, true
	}
	return nil, false
}

func (lru *lruCache) Clear() {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	lru.queue = NewList()
	lru.items = make(map[Key]*ListItem, lru.capacity)
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}
