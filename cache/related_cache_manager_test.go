package cache

import (
	"testing"
	"time"

	"github.com/costa92/multicache/models"
	"github.com/stretchr/testify/assert"
)

type mockOrderLoader struct {
	orders []models.Order
	err    error
}

func (m *mockOrderLoader) Load() ([]models.Order, error) {
	return m.orders, m.err
}

func TestRelatedCacheManager(t *testing.T) {
	// Setup test data
	testOrders := []models.Order{
		{ID: 1, UserID: 1, Amount: 100},
		{ID: 2, UserID: 1, Amount: 200},
		{ID: 3, UserID: 2, Amount: 300},
	}

	loader := &mockOrderLoader{orders: testOrders}
	cache := NewRelatedCacheManager[models.Order](loader, 5*time.Minute)

	t.Run("Refresh loads data correctly", func(t *testing.T) {
		err := cache.Refresh()
		assert.NoError(t, err)

		items := cache.GetAll()
		assert.Len(t, items, 3)
	})

	t.Run("GetByForeignKey returns correct items", func(t *testing.T) {
		orders := cache.GetByForeignKey(1)
		assert.Len(t, orders, 2)
		assert.Equal(t, uint(1), orders[0].UserID)
		assert.Equal(t, uint(1), orders[1].UserID)
	})

	t.Run("Get returns correct item", func(t *testing.T) {
		order, exists := cache.Get(1)
		assert.True(t, exists)
		assert.Equal(t, float64(100), order.Amount)
	})

	t.Run("Clear removes all items", func(t *testing.T) {
		cache.Clear()
		items := cache.GetAll()
		assert.Len(t, items, 0)
	})
}
