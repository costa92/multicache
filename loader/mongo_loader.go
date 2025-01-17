package loader

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoLoader implements DataLoader interface for MongoDB
type MongoLoader[T any] struct {
	coll      *mongo.Collection
	ctx       context.Context
	filter    interface{}
	opts      *options.FindOptions
	pipeline  mongo.Pipeline
	aggregate bool
}

// NewMongoLoader creates a new MongoDB data loader
func NewMongoLoader[T any](ctx context.Context, coll *mongo.Collection) *MongoLoader[T] {
	return &MongoLoader[T]{
		ctx:    ctx,
		coll:   coll,
		filter: bson.M{},
		opts:   options.Find(),
	}
}

// WithFilter adds a filter condition
func (l *MongoLoader[T]) WithFilter(filter interface{}) *MongoLoader[T] {
	l.filter = filter
	return l
}

// WithOptions adds find options
func (l *MongoLoader[T]) WithOptions(opts *options.FindOptions) *MongoLoader[T] {
	l.opts = opts
	return l
}

// WithAggregate sets the aggregation pipeline
func (l *MongoLoader[T]) WithAggregate(pipeline mongo.Pipeline) *MongoLoader[T] {
	l.pipeline = pipeline
	l.aggregate = true
	return l
}

// Load implements DataLoader interface
func (l *MongoLoader[T]) Load() ([]T, error) {
	var items []T
	var cursor *mongo.Cursor
	var err error

	if l.aggregate {
		cursor, err = l.coll.Aggregate(l.ctx, l.pipeline)
	} else {
		cursor, err = l.coll.Find(l.ctx, l.filter, l.opts)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer cursor.Close(l.ctx)

	if err := cursor.All(l.ctx, &items); err != nil {
		return nil, fmt.Errorf("failed to decode results: %w", err)
	}

	return items, nil
}
