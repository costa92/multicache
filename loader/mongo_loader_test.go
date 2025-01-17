package loader

import (
	"context"
	"testing"

	"github.com/costa92/multicache/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupTestMongoDB(t *testing.T) (*mongo.Client, context.Context) {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	// Clean up test database
	err = client.Database("testdb").Drop(ctx)
	require.NoError(t, err)

	// Insert test data
	users := []interface{}{
		models.User{ID: 1, Name: "John", Email: "john@example.com"},
		models.User{ID: 2, Name: "Jane", Email: "jane@example.com"},
	}
	_, err = client.Database("testdb").Collection("users").InsertMany(ctx, users)
	require.NoError(t, err)

	orders := []interface{}{
		models.Order{ID: 1, UserID: 1, Amount: 100},
		models.Order{ID: 2, UserID: 1, Amount: 200},
		models.Order{ID: 3, UserID: 2, Amount: 300},
	}
	_, err = client.Database("testdb").Collection("orders").InsertMany(ctx, orders)
	require.NoError(t, err)

	return client, ctx
}

func TestMongoLoader(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping MongoDB integration test")
	}

	client, ctx := setupTestMongoDB(t)
	defer client.Disconnect(ctx)

	t.Run("basic load", func(t *testing.T) {
		coll := client.Database("testdb").Collection("users")
		loader := NewMongoLoader[models.User](ctx, coll)
		users, err := loader.Load()
		require.NoError(t, err)
		assert.Len(t, users, 2)
	})

	t.Run("load with filter", func(t *testing.T) {
		coll := client.Database("testdb").Collection("users")
		loader := NewMongoLoader[models.User](ctx, coll).
			WithFilter(bson.M{"name": "John"})
		users, err := loader.Load()
		require.NoError(t, err)
		assert.Len(t, users, 1)
		assert.Equal(t, "John", users[0].Name)
	})

	t.Run("load with options", func(t *testing.T) {
		coll := client.Database("testdb").Collection("orders")
		loader := NewMongoLoader[models.Order](ctx, coll).
			WithOptions(options.Find().SetSort(bson.M{"amount": -1}).SetLimit(2))
		orders, err := loader.Load()
		require.NoError(t, err)
		assert.Len(t, orders, 2)
		assert.Equal(t, float64(300), orders[0].Amount)
	})
	t.Run("load with aggregation", func(t *testing.T) {
		coll := client.Database("testdb").Collection("orders")
		pipeline := mongo.Pipeline{
			bson.D{{Key: "$match", Value: bson.M{"user_id": 1}}},
			bson.D{{Key: "$sort", Value: bson.M{"amount": -1}}},
		}
		loader := NewMongoLoader[models.Order](ctx, coll).
			WithAggregate(pipeline)
		orders, err := loader.Load()
		require.NoError(t, err)
		assert.Len(t, orders, 2)
		assert.Equal(t, float64(200), orders[0].Amount)
	})
}
