package cache

import (
	"testing"
	"time"

	"github.com/costa92/multicache/models"
	"github.com/stretchr/testify/assert"
)

type mockUserLoader struct {
	users []models.User
	err   error
}

func (m *mockUserLoader) Load() ([]models.User, error) {
	return m.users, m.err
}

func TestCacheManager(t *testing.T) {
	// Setup test data
	testUsers := []models.User{
		{ID: 1, Name: "John", Email: "john@example.com"},
		{ID: 2, Name: "Jane", Email: "jane@example.com"},
	}

	loader := &mockUserLoader{users: testUsers}
	cache := NewCacheManager[models.User](loader).WithTTL(5 * time.Minute)

	t.Run("Refresh loads data correctly", func(t *testing.T) {
		err := cache.Refresh()
		assert.NoError(t, err)

		items := cache.GetAll()
		assert.Len(t, items, 2)
	})

	t.Run("Get returns correct item", func(t *testing.T) {
		user, err := cache.Get(1)
		assert.NoError(t, err)
		assert.Equal(t, "John", user.Name)
	})

	t.Run("Get returns false for non-existent item", func(t *testing.T) {
		_, err := cache.Get(999)
		assert.Error(t, err)
	})

	t.Run("Clear removes all items", func(t *testing.T) {
		cache.Clear()
		items := cache.GetAll()
		assert.Len(t, items, 0)
	})
}

func TestCacheManagerQuery(t *testing.T) {
	testUsers := []models.User{
		{ID: 1, Name: "John", Email: "john@example.com"},
		{ID: 2, Name: "Jane", Email: "jane@example.com"},
		{ID: 3, Name: "John Smith", Email: "john.smith@example.com"},
	}

	loader := &mockUserLoader{users: testUsers}
	cache := NewCacheManager[models.User](loader).WithTTL(1 * time.Minute)
	err := cache.Refresh()
	assert.NoError(t, err)

	t.Run("Query by name equality", func(t *testing.T) {
		condition := StringFieldCondition[models.User]{
			FieldExtractor: func(u models.User) string { return u.Name },
			Value:          "John",
			Operation:      "eq",
		}

		results := cache.Query(condition)
		assert.Len(t, results, 1)
		assert.Equal(t, "John", results[0].Name)
	})

	t.Run("Query by name contains", func(t *testing.T) {
		condition := StringFieldCondition[models.User]{
			FieldExtractor: func(u models.User) string { return u.Name },
			Value:          "John",
			Operation:      "contains",
		}

		results := cache.Query(condition)
		assert.Len(t, results, 2)
	})

	t.Run("Composite query", func(t *testing.T) {
		nameCondition := StringFieldCondition[models.User]{
			FieldExtractor: func(u models.User) string { return u.Name },
			Value:          "John",
			Operation:      "contains",
		}

		emailCondition := StringFieldCondition[models.User]{
			FieldExtractor: func(u models.User) string { return u.Email },
			Value:          "smith",
			Operation:      "contains",
		}

		composite := CompositeCondition[models.User]{
			Conditions: []QueryCondition[models.User]{nameCondition, emailCondition},
			Operation:  "and",
		}

		results := cache.Query(composite)
		assert.Len(t, results, 1)
		assert.Equal(t, "John Smith", results[0].Name)
	})
}
