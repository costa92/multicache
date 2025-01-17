package examples

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alecthomas/assert"
	"github.com/costa92/multicache/cache"
	"github.com/costa92/multicache/loader"
	"github.com/costa92/multicache/models"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
		{ID: 2, Name: "Jane Premium", Email: "jane@example.com"},
		{ID: 3, Name: "John Smith", Email: "john.smith@example.com"},
	}

	// Add status and subscription_type fields for testing
	err = db.Exec(`ALTER TABLE users ADD COLUMN status TEXT`).Error
	require.NoError(t, err)
	err = db.Exec(`ALTER TABLE users ADD COLUMN subscription_type TEXT`).Error
	require.NoError(t, err)

	for _, user := range users {
		result := db.Create(&user)
		require.NoError(t, result.Error)

		// Set status and subscription type
		if user.ID == 2 {
			db.Model(&user).Updates(map[string]interface{}{
				"status":            "active",
				"subscription_type": "premium",
			})
		} else {
			db.Model(&user).Update("status", "active")
		}
	}

	// Insert orders with timestamps
	now := time.Now()
	orders := []models.Order{
		{ID: 1, UserID: 1, Amount: 100, CreatedAt: now.Add(-2 * 24 * time.Hour)},
		{ID: 2, UserID: 1, Amount: 2000, CreatedAt: now.Add(-1 * 24 * time.Hour)},
		{ID: 3, UserID: 2, Amount: 3000, CreatedAt: now},
		{ID: 4, UserID: 3, Amount: 1500, CreatedAt: now},
	}

	// Insert orders with timestamps
	for _, order := range orders {
		result := db.Create(&order)
		require.NoError(t, result.Error)
	}

	return db
}

func setupTestMongoDB(t *testing.T) (*mongo.Client, *mongo.Collection, context.Context) {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	// Clean up test database
	err = client.Database("testdb").Drop(ctx)
	require.NoError(t, err)

	coll := client.Database("testdb").Collection("orders")

	// Insert test data
	orders := []interface{}{
		models.Order{ID: 1, UserID: 1, Amount: 1500},
		models.Order{ID: 2, UserID: 1, Amount: 2500},
		models.Order{ID: 3, UserID: 2, Amount: 3500},
	}
	_, err = coll.InsertMany(ctx, orders)
	require.NoError(t, err)

	return client, coll, ctx
}

func TestUseGormLoaderExample(t *testing.T) {
	db := setupTestDB(t)

	t.Run("loads active users with orders", func(t *testing.T) {
		UseGormLoaderExample(db)
		// Since this is mainly a demonstration function that prints output,
		// we're just verifying it runs without errors
	})
}

func TestUseCacheExample(t *testing.T) {
	db := setupTestDB(t)

	t.Run("caches and queries users", func(t *testing.T) {
		UseCacheExample(db)
		// Verify the function runs without errors and prints expected output
	})
}

func TestUseCombinedExample(t *testing.T) {
	db := setupTestDB(t)

	t.Run("processes orders with different caches", func(t *testing.T) {
		UseCombinedExample(db)
		// Verify the function runs without errors
	})

	t.Run("processes orders correctly", func(t *testing.T) {
		orders := []models.Order{
			{ID: 1, UserID: 1, Amount: 100},
			{ID: 2, UserID: 2, Amount: 200}, // Premium user
		}

		// Create test caches with test data
		activeUserLoader := loader.NewGormLoader(db, models.User{}).
			WithCondition("status = ?", "active")
		premiumUserLoader := loader.NewGormLoader(db, models.User{}).
			WithCondition("status = ? AND subscription_type = ?", "active", "premium")

		activeCache := cache.NewCacheManager[models.User](activeUserLoader)
		premiumCache := cache.NewCacheManager[models.User](premiumUserLoader)

		require.NoError(t, activeCache.Refresh())
		require.NoError(t, premiumCache.Refresh())

		processOrders(orders, activeCache, premiumCache)
		// Since this function prints output, we're verifying it processes without errors
	})
}

func TestUseMongoExample(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping MongoDB integration test")
	}

	client, coll, ctx := setupTestMongoDB(t)
	defer client.Disconnect(ctx)

	t.Run("queries high-value orders", func(t *testing.T) {
		UseMongoExample(ctx, coll)
		// Verify the function runs without errors

		// Verify the data directly
		var orders []models.Order
		cursor, err := coll.Find(ctx, bson.M{"amount": bson.M{"$gt": 2000}})
		require.NoError(t, err)
		err = cursor.All(ctx, &orders)
		require.NoError(t, err)
		assert.NotEmpty(t, orders, "should have high-value orders")
	})
}

// Helper function to capture stdout for testing
type testLogger struct {
	logs []string
}

func (l *testLogger) Printf(format string, v ...interface{}) {
	l.logs = append(l.logs, fmt.Sprintf(format, v...))
}

func (l *testLogger) Fatalf(format string, v ...interface{}) {
	l.logs = append(l.logs, fmt.Sprintf(format, v...))
	panic("fatal error")
}
