package cache

import (
	"sync"
	"time"
)

// RelatedCacheManager implements the RelatedCache interface
type RelatedCacheManager[T ForeignKeyable] struct {
	data      map[uint]T      // Primary key -> Entity
	fkIndex   map[uint][]uint // Foreign key -> Primary keys
	mu        sync.RWMutex
	loader    DataLoader[T]
	ttl       time.Duration
	lastFetch time.Time
}

// NewRelatedCacheManager creates a new related cache manager instance
func NewRelatedCacheManager[T ForeignKeyable](loader DataLoader[T], ttl time.Duration) *RelatedCacheManager[T] {
	return &RelatedCacheManager[T]{
		data:      make(map[uint]T),
		fkIndex:   make(map[uint][]uint),
		loader:    loader,
		ttl:       ttl,
		lastFetch: time.Time{},
	}
}

// Get retrieves an item by ID
func (rcm *RelatedCacheManager[T]) Get(id uint) (T, bool) {
	rcm.mu.RLock()
	defer rcm.mu.RUnlock()

	if rcm.isExpired() {
		var zero T
		return zero, false
	}

	item, exists := rcm.data[id]
	return item, exists
}

// GetByForeignKey retrieves items by foreign key
func (rcm *RelatedCacheManager[T]) GetByForeignKey(fkID uint) []T {
	rcm.mu.RLock()
	defer rcm.mu.RUnlock()

	if rcm.isExpired() {
		return nil
	}

	pks := rcm.fkIndex[fkID]
	result := make([]T, 0, len(pks))
	for _, pk := range pks {
		if item, exists := rcm.data[pk]; exists {
			result = append(result, item)
		}
	}
	return result
}

// GetAll returns all items in the cache
func (rcm *RelatedCacheManager[T]) GetAll() []T {
	rcm.mu.RLock()
	defer rcm.mu.RUnlock()

	if rcm.isExpired() {
		return nil
	}

	items := make([]T, 0, len(rcm.data))
	for _, item := range rcm.data {
		items = append(items, item)
	}
	return items
}

// Refresh reloads the cache data
func (rcm *RelatedCacheManager[T]) Refresh() error {
	rcm.mu.Lock()
	defer rcm.mu.Unlock()

	items, err := rcm.loader.Load()
	if err != nil {
		return err
	}

	newData := make(map[uint]T)
	newFKIndex := make(map[uint][]uint)

	for _, item := range items {
		pk := item.GetID()
		fk := item.GetUserID()
		newData[pk] = item
		newFKIndex[fk] = append(newFKIndex[fk], pk)
	}

	rcm.data = newData
	rcm.fkIndex = newFKIndex
	rcm.lastFetch = time.Now()
	return nil
}

// Clear removes all items from the cache
func (rcm *RelatedCacheManager[T]) Clear() {
	rcm.mu.Lock()
	defer rcm.mu.Unlock()
	rcm.data = make(map[uint]T)
	rcm.fkIndex = make(map[uint][]uint)
}

func (rcm *RelatedCacheManager[T]) isExpired() bool {
	return !rcm.lastFetch.IsZero() && time.Since(rcm.lastFetch) > rcm.ttl
}

// Query returns items that match the given condition
func (rcm *RelatedCacheManager[T]) Query(condition QueryCondition[T]) []T {
	rcm.mu.RLock()
	defer rcm.mu.RUnlock()

	if rcm.isExpired() {
		return nil
	}

	result := make([]T, 0)
	for _, item := range rcm.data {
		if condition.Match(item) {
			result = append(result, item)
		}
	}
	return result
}
