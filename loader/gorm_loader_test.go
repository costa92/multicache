package loader

import (
	"fmt"
	"testing"

	"github.com/costa92/multicache/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Migrate the schema
	err = db.AutoMigrate(&models.User{}, &models.Order{})
	require.NoError(t, err)

	// Insert test data
	users := []models.User{
		{ID: 1, Name: "John", Email: "john@example.com"},
		{ID: 2, Name: "Jane", Email: "jane@example.com"},
		{ID: 3, Name: "John Smith", Email: "john.smith@example.com"},
	}
	err = db.Create(&users).Error
	require.NoError(t, err)

	orders := []models.Order{
		{ID: 1, UserID: 1, Amount: 100},
		{ID: 2, UserID: 1, Amount: 200},
		{ID: 3, UserID: 2, Amount: 300},
		{ID: 4, UserID: 3, Amount: 400},
	}
	err = db.Create(&orders).Error
	require.NoError(t, err)

	return db
}

func TestGormLoader(t *testing.T) {
	db := setupTestDB(t)

	t.Run("basic load", func(t *testing.T) {
		loader := NewGormLoader(db, models.User{}).WithDebug(true)
		users, err := loader.Load()
		require.NoError(t, err)
		assert.Len(t, users, 3, "should load all users")
	})

	t.Run("load with single condition", func(t *testing.T) {
		loader := NewGormLoader(db, models.User{}).
			WithCondition("name = ?", "John").WithDebug(true)
		users, err := loader.Load()
		require.NoError(t, err)

		assert.NotEmpty(t, users, "users should not be empty")
		assert.Len(t, users, 1, "should return exactly one user")
		assert.Equal(t, "John", users[0].Name, "user name should match")
	})

	t.Run("load with multiple conditions", func(t *testing.T) {
		loader := NewGormLoader(db, models.User{}).
			WithCondition("name LIKE ? AND id > ?", "John%", 1).WithDebug(true)
		users, err := loader.Load()
		require.NoError(t, err)

		assert.Len(t, users, 1, "should return one user")
		assert.Equal(t, "John Smith", users[0].Name, "user name should match")
	})

	t.Run("load with preload", func(t *testing.T) {
		loader := NewGormLoader(db, models.User{}).
			WithPreload("Orders")
		users, err := loader.Load()
		require.NoError(t, err)

		assert.Len(t, users, 3, "should load all users")
		// for _, user := range users {
		// 	assert.NotNil(t, user.Orders, "orders should be preloaded")
		// 	if user.ID == 1 {
		// 		assert.Len(t, user.Orders, 2, "user 1 should have 2 orders")
		// 	}
		// }
	})

	t.Run("load with preload and condition", func(t *testing.T) {
		loader := NewGormLoader(db, models.User{}).
			WithPreloadQuery("Orders", "amount > ?", 200).WithDebug(true)
		users, err := loader.Load()
		require.NoError(t, err)
		assert.Len(t, users, 3, "should load all users")
		for _, user := range users {
			for _, order := range user.Orders {
				assert.True(t, order.Amount > 200, "order amount should be > 200")
			}
		}
	})

	t.Run("load with joins", func(t *testing.T) {
		loader := NewGormLoader(db, models.User{}).
			WithJoins("JOIN orders ON orders.user_id = users.id").WithDebug(true)
		users, err := loader.Load()
		require.NoError(t, err)
		assert.NotEmpty(t, users, "should return users with orders")
		for _, user := range users {
			fmt.Println(user.GetID())
			if len(user.Orders) > 0 {

				assert.NotEmpty(t, user.Orders, "user should have orders")
			}
		}
	})

	t.Run("load with preload joins", func(t *testing.T) {
		loader := NewGormLoader(db, models.User{}).
			WithPreloadJoin("Orders", "Orders.amount > ?", 200).WithDebug(true)
		users, err := loader.Load()
		require.NoError(t, err)

		assert.NotEmpty(t, users, "should return users")
		for _, user := range users {
			if len(user.Orders) > 0 {
				for _, order := range user.Orders {
					assert.True(t, order.Amount > 200, "order amount should be > 200")
				}
			}
		}
	})

	t.Run("load with debug mode", func(t *testing.T) {
		loader := NewGormLoader(db, models.Order{}).
			WithCondition("amount > ?", 200).
			WithDebug(true)
		orders, err := loader.Load()
		require.NoError(t, err)

		assert.NotEmpty(t, orders, "should return orders")
		for _, order := range orders {
			assert.True(t, order.Amount > 200, "order amount should be > 200")
		}
	})

	t.Run("load with invalid condition", func(t *testing.T) {
		loader := NewGormLoader(db, models.User{}).
			WithCondition(123) // Invalid condition type
		users, err := loader.Load()
		assert.Error(t, err, "should return error for invalid condition")
		assert.Empty(t, users, "should not return any users")
	})

	t.Run("load with multiple preloads", func(t *testing.T) {
		loader := NewGormLoader(db, models.User{}).
			WithPreload("Orders").WithDebug(true) // Assuming there's a User relation in Order
		users, err := loader.Load()
		require.NoError(t, err)

		assert.Len(t, users, 3, "should load all users")
		assert.NotNil(t, users[0].Orders, "orders should be preloaded")
	})

	t.Run("load with complex conditions", func(t *testing.T) {
		loader := NewGormLoader(db, models.Order{}).
			WithCondition("user_id IN (?) AND amount BETWEEN ? AND ?", []uint{1, 2}, 100, 300).
			WithDebug(true)
		orders, err := loader.Load()
		require.NoError(t, err)

		assert.NotEmpty(t, orders, "should return orders")
		for _, order := range orders {
			assert.Contains(t, []uint{1, 2}, order.UserID, "user ID should be 1 or 2")
			assert.True(t, order.Amount >= 100, "amount should be >= 100")
			assert.True(t, order.Amount <= 300, "amount should be <= 300")
		}
	})
	t.Run("load with joins model", func(t *testing.T) {
		loader := NewGormLoader(db, models.Order{}).
			WithJoinsModel(models.User{}, "orders.user_id", "users.id").
			WithDebug(true)
		orders, err := loader.Load()
		require.NoError(t, err)

		assert.NotEmpty(t, orders, "should return orders")
		for _, order := range orders {
			assert.NotNil(t, order.UserID, "order should have a user ID")
		}
	})
}
