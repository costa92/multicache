package cache

import (
	"fmt"
	"sync"
	"time"
)

// CacheManager implements the Cache interface using a thread-safe map
type CacheManager[T Identifiable] struct {
	data      map[uint]T
	mu        sync.RWMutex
	loader    DataLoader[T]
	ttl       time.Duration
	lastFetch time.Time
}

// NewCacheManager creates a new cache manager instance with a default TTL of permanent if not set
func NewCacheManager[T Identifiable](loader DataLoader[T]) *CacheManager[T] {
	return &CacheManager[T]{
		data:      make(map[uint]T),
		loader:    loader,
		ttl:       0, // Default to permanent
		lastFetch: time.Time{},
	}
}

// WithTTL sets the TTL for the cache manager
func (cm *CacheManager[T]) WithTTL(ttl time.Duration) *CacheManager[T] {
	cm.ttl = ttl
	return cm
}

// Template method pattern for cache operations
func (cm *CacheManager[T]) executeWithLock(read bool, operation func() interface{}) interface{} {
	if read {
		cm.mu.RLock()
		defer cm.mu.RUnlock()
	} else {
		cm.mu.Lock()
		defer cm.mu.Unlock()
	}
	return operation()
}

// Get retrieves an item by ID
func (cm *CacheManager[T]) Get(id uint) (T, error) {
	result := cm.executeWithLock(true, func() interface{} {
		if cm.isExpired() {
			var zero T
			return struct {
				item T
				err  error
			}{zero, fmt.Errorf("cache expired")}
		}
		item, exists := cm.data[id]
		if !exists {
			var zero T
			return struct {
				item T
				err  error
			}{zero, fmt.Errorf("item with ID %d not found", id)}
		}
		return struct {
			item T
			err  error
		}{item, nil}
	})
	res := result.(struct {
		item T
		err  error
	})
	return res.item, res.err
}

// GetAll returns all items in the cache
func (cm *CacheManager[T]) GetAll() []T {
	result := cm.executeWithLock(true, func() interface{} {
		if cm.isExpired() {
			return nil
		}
		items := make([]T, 0, len(cm.data))
		for _, item := range cm.data {
			items = append(items, item)
		}
		return items
	})
	return result.([]T)
}

// Refresh reloads the cache data
func (cm *CacheManager[T]) Refresh() error {
	err := cm.executeWithLock(false, func() interface{} {
		items, err := cm.loader.Load()
		if err != nil {
			return err
		}
		newData := make(map[uint]T)
		for _, item := range items {
			newData[item.GetID()] = item
		}
		cm.data = newData
		cm.lastFetch = time.Now()
		return nil
	})
	if err != nil {
		return err.(error)
	}
	return nil
}

// Clear removes all items from the cache
func (cm *CacheManager[T]) Clear() {
	cm.executeWithLock(false, func() interface{} {
		cm.data = make(map[uint]T)
		return nil
	})
}

func (cm *CacheManager[T]) isExpired() bool {
	return cm.ttl > 0 && !cm.lastFetch.IsZero() && time.Since(cm.lastFetch) > cm.ttl
}

// Query returns items that match the given condition
func (cm *CacheManager[T]) Query(condition QueryCondition[T]) []T {
	result := cm.executeWithLock(true, func() interface{} {
		if cm.isExpired() {
			return nil
		}
		result := make([]T, 0)
		for _, item := range cm.data {
			if condition.Match(item) {
				result = append(result, item)
			}
		}
		return result
	})
	return result.([]T)
}
