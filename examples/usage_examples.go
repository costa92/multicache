package examples

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/costa92/multicache/cache"
	"github.com/costa92/multicache/loader"
	"github.com/costa92/multicache/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

// Example 1: Using GormLoader directly for real-time queries
func UseGormLoaderExample(db *gorm.DB) {
	// Create a loader for active users with orders
	userLoader := loader.NewGormLoader(db, models.User{}).
		WithPreload("Orders").
		WithCondition("status = ?", "active").
		WithPreloadQuery("Orders", "created_at > ?", time.Now().AddDate(0, -1, 0))

	// Load users with their recent orders
	users, err := userLoader.Load()
	if err != nil {
		log.Fatalf("Failed to load users: %v", err)
	}

	// Process the results
	for _, user := range users {
		fmt.Printf("User %s has %d orders in the last month\n", user.Name, len(user.Orders))
	}

	// Create a loader for high-value orders with user info
	orderLoader := loader.NewGormLoader(db, models.Order{}).
		WithJoinsModel(models.User{}, "orders.user_id", "users.id").
		WithCondition("orders.amount > ?", 4000).
		WithDebug(true)

	// Load high-value orders
	orders, err := orderLoader.Load()
	if err != nil {
		log.Fatalf("Failed to load orders: %v", err)
	}

	for _, order := range orders {
		fmt.Printf("High-value order: %+v\n", order)
	}
}

// Example 2: Using Cache for frequently accessed data
func UseCacheExample(db *gorm.DB) {
	// Create a loader for active users
	userLoader := loader.NewGormLoader(db, models.User{}).
		WithCondition("status = ?", "active")

	// Create a cache with 5-minute TTL
	userCache := cache.NewCacheManager[models.User](userLoader).WithTTL(1 * time.Minute)

	// Initialize cache
	if err := userCache.Refresh(); err != nil {
		log.Fatalf("Failed to initialize cache: %v", err)
	}

	// Get user by ID (from cache)
	if user, err := userCache.Get(1); err != nil {
		fmt.Printf("Found user in cache: %s\n", user.Name)
	}

	// Query users by condition (from cache)
	nameCondition := cache.StringFieldCondition[models.User]{
		FieldExtractor: func(u models.User) string { return u.Name },
		Value:          "John",
		Operation:      "contains",
	}

	users := userCache.Query(nameCondition)
	fmt.Printf("Found %d users with 'John' in their name\n", len(users))
}

// Example 3: Combined usage for complex scenarios
func UseCombinedExample(db *gorm.DB) {
	// Create loaders with different conditions
	activeUserLoader := loader.NewGormLoader(db, models.User{}).
		WithCondition("status = ?", "active")

	premiumUserLoader := loader.NewGormLoader(db, models.User{}).
		WithCondition("status = ? AND subscription_type = ?", "active", "premium").
		WithPreload("Orders")

	// Create caches with different TTLs
	activeUserCache := cache.NewCacheManager[models.User](activeUserLoader).WithTTL(5 * time.Minute)
	premiumUserCache := cache.NewCacheManager[models.User](premiumUserLoader).WithTTL(1 * time.Minute)

	// Initialize caches
	if err := activeUserCache.Refresh(); err != nil {
		log.Printf("Failed to refresh active user cache: %v", err)
	}
	if err := premiumUserCache.Refresh(); err != nil {
		log.Printf("Failed to refresh premium user cache: %v", err)
	}

	// Use real-time loader for specific queries
	recentOrderLoader := loader.NewGormLoader(db, models.Order{}).
		WithCondition("created_at > ?", time.Now().AddDate(0, 0, -1)).
		WithJoinsModel(models.User{}, "orders.user_id", "users.id")

	recentOrders, err := recentOrderLoader.Load()
	if err != nil {
		log.Printf("Failed to load recent orders: %v", err)
	}

	// Complex business logic combining cached and real-time data
	processOrders(recentOrders, activeUserCache, premiumUserCache)
}

func processOrders(orders []models.Order, activeCache, premiumCache *cache.CacheManager[models.User]) {
	for _, order := range orders {
		// Try premium cache first
		if user, err := premiumCache.Get(order.UserID); err != nil {
			fmt.Printf("Premium user %s placed order: %+v\n", user.Name, order)
			continue
		}

		// Try active cache
		if user, err := activeCache.Get(order.UserID); err != nil {
			fmt.Printf("Active user %s placed order: %+v\n", user.Name, order)
			continue
		}

		fmt.Printf("Order from unknown user: %+v\n", order)
	}
}

// Example 4: Using MongoDB loader and cache
func UseMongoExample(ctx context.Context, coll *mongo.Collection) {
	// Create MongoDB loader with aggregation
	orderLoader := loader.NewMongoLoader[models.Order](ctx, coll).
		WithAggregate(mongo.Pipeline{
			bson.D{{Key: "$match", Value: bson.M{"amount": bson.M{"$gt": 1000}}}},
			bson.D{{Key: "$sort", Value: bson.M{"amount": -1}}},
		})

	// Create cache for high-value orders
	orderCache := cache.NewRelatedCacheManager[models.Order](orderLoader, 1*time.Minute)

	// Initialize cache
	if err := orderCache.Refresh(); err != nil {
		log.Printf("Failed to initialize order cache: %v", err)
	}

	// Query orders by user
	userOrders := orderCache.GetByForeignKey(1)
	fmt.Printf("Found %d high-value orders for user 1\n", len(userOrders))

	// Query orders by amount condition
	amountCondition := cache.NumberFieldCondition[models.Order, float64]{
		FieldExtractor: func(o models.Order) float64 { return o.Amount },
		Value:          2000,
		Operation:      "gte",
	}

	veryHighValueOrders := orderCache.Query(amountCondition)
	fmt.Printf("Found %d orders with amount >= 2000\n", len(veryHighValueOrders))
}
