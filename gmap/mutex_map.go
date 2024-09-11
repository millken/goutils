package gmap

import (
	"sync"
)

// MutexMap is an interface that defines a thread-safe map with keys of type T associated to
// read-write mutexes (sync.RWMutex), allowing for granular locking on a per-key basis.
// This can be useful for scenarios where fine-grained concurrency control is needed.
//
// Methods:
// - Lock(key T): Acquires an exclusive lock on the mutex associated with the given key.
// - Unlock(key T): Releases the exclusive lock on the mutex associated with the given key.
// - RLock(key T): Acquires a read lock on the mutex associated with the given key.
// - RUnlock(key T): Releases the read lock on the mutex associated with the given key.
// - Delete(key T): Removes the mutex associated with the given key from the map.
// - Clear(): Removes all mutexes from the map.
// - ItemCount() int: Returns the number of items (mutexes) in the map.
// - DeleteUnlock(key T): Removes the mutex associated with the given key from the map and releases the lock.
// - DeleteRUnlock(key T): Removes the mutex associated with the given key from the map and releases the read lock.
type MutexMap[T comparable] interface {
	Lock(key T)
	Unlock(key T)
	RLock(key T)
	RUnlock(key T)
	Delete(key T)
	Clear()
	ItemCount() int
	DeleteUnlock(key T)
	DeleteRUnlock(key T)
}

type mutexMap[T comparable] struct {
	lock  sync.RWMutex
	items map[T]*sync.RWMutex
}

func NewMutexMap[T comparable]() MutexMap[T] {
	return &mutexMap[T]{
		items: make(map[T]*sync.RWMutex),
	}
}

func (a *mutexMap[T]) Lock(key T) {
	a.lock.RLock()
	mutex, ok := a.items[key]
	a.lock.RUnlock()
	if !ok {
		a.lock.Lock()
		mutex, ok = a.items[key]
		if !ok {
			mutex = &sync.RWMutex{}
			a.items[key] = mutex
		}
		a.lock.Unlock()
	}
	mutex.Lock()
}

func (a *mutexMap[T]) Unlock(key T) {
	a.lock.RLock()
	mutex, ok := a.items[key]
	a.lock.RUnlock()
	if ok {
		mutex.Unlock()
	}
}

func (a *mutexMap[T]) RLock(key T) {
	a.lock.RLock()
	mutex, ok := a.items[key]
	a.lock.RUnlock()
	if !ok {
		a.lock.Lock()
		mutex, ok = a.items[key]
		if !ok {
			mutex = &sync.RWMutex{}
			a.items[key] = mutex
		}
		a.lock.Unlock()
	}
	mutex.RLock()
}

func (a *mutexMap[T]) RUnlock(key T) {
	a.lock.RLock()
	mutex, ok := a.items[key]
	a.lock.RUnlock()
	if ok {
		mutex.RUnlock()
	}
}

func (a *mutexMap[T]) Delete(key T) {
	a.lock.Lock()
	delete(a.items, key)
	a.lock.Unlock()
}

func (a *mutexMap[T]) DeleteUnlock(key T) {
	a.lock.Lock()
	mutex, ok := a.items[key]
	delete(a.items, key)
	a.lock.Unlock()
	if ok {
		mutex.Unlock()
	}
}

func (a *mutexMap[T]) DeleteRUnlock(key T) {
	a.lock.Lock()
	mutex, ok := a.items[key]
	delete(a.items, key)
	a.lock.Unlock()
	if ok {
		mutex.RUnlock()
	}
}

func (a *mutexMap[T]) Clear() {
	a.lock.Lock()
	clear(a.items)
	a.lock.Unlock()
}

func (a *mutexMap[T]) ItemCount() int {
	a.lock.Lock()
	defer a.lock.Unlock()
	return len(a.items)
}
