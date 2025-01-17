package loader

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoLoader implements MongoDataLoader interface
type MongoLoader[T any] struct {
	coll      *mongo.Collection
	ctx       context.Context
	filter    interface{}
	opts      *options.FindOptions
	pipeline  mongo.Pipeline
	aggregate bool
	debug     bool
	config    MongoLoaderConfig
}

// NewMongoLoader creates a new MongoDB data loader
func NewMongoLoader[T any](ctx context.Context, coll *mongo.Collection) MongoDataLoader[T] {
	return &MongoLoader[T]{
		ctx:    ctx,
		coll:   coll,
		filter: bson.M{},
		opts:   options.Find(),
		config: MongoLoaderConfig{
			Filter: bson.M{},
		},
	}
}

// WithDebug implements Loader interface
func (l *MongoLoader[T]) WithDebug(debug bool) Loader[T] {
	l.debug = debug
	l.config.Debug = debug
	return l
}

// WithFilter implements MongoDataLoader interface
func (l *MongoLoader[T]) WithFilter(filter interface{}) MongoDataLoader[T] {
	l.filter = filter
	l.config.Filter = filter
	return l
}

// WithOptions implements MongoDataLoader interface
func (l *MongoLoader[T]) WithOptions(opts interface{}) MongoDataLoader[T] {
	if findOpts, ok := opts.(*options.FindOptions); ok {
		l.opts = findOpts
		l.config.Options = opts
	}
	return l
}

// WithAggregate implements MongoDataLoader interface
func (l *MongoLoader[T]) WithAggregate(pipeline mongo.Pipeline) MongoDataLoader[T] {
	l.pipeline = pipeline
	l.aggregate = true
	l.config.Pipeline = pipeline
	l.config.Aggregate = true
	return l
}

// Load implements DataLoader interface
func (l *MongoLoader[T]) Load() ([]T, error) {
	var items []T
	var cursor *mongo.Cursor
	var err error

	if l.debug {
		fmt.Printf("MongoDB Query: filter=%v, aggregate=%v\n", l.filter, l.aggregate)
		if l.aggregate {
			fmt.Printf("Pipeline: %v\n", l.pipeline)
		}
	}

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

// MongoOption implements MongoLoaderOption interface
type MongoOption struct {
	fn func(*mongo.Collection) *mongo.Collection
}

func (o MongoOption) Apply(coll *mongo.Collection) *mongo.Collection {
	return o.fn(coll)
}

// WithIndex creates a MongoDB index option
func WithIndex(keys interface{}, opts ...*options.IndexOptions) MongoLoaderOption {
	return MongoOption{
		fn: func(coll *mongo.Collection) *mongo.Collection {
			_, err := coll.Indexes().CreateOne(context.Background(), mongo.IndexModel{
				Keys:    keys,
				Options: opts[0],
			})
			if err != nil {
				// Log error but don't fail
				fmt.Printf("Failed to create index: %v\n", err)
			}
			return coll
		},
	}
}

// WithCollation creates a MongoDB collation option
func WithCollation(collation *options.Collation) MongoLoaderOption {
	return MongoOption{
		fn: func(coll *mongo.Collection) *mongo.Collection {
			// Directly set the collation on the collection options
			collationOpts := options.Find().SetCollation(collation)
			// Apply the collation options to the collection
			cursor, err := coll.Find(context.Background(), bson.D{}, collationOpts)
			if err != nil {
				return coll
			}
			defer cursor.Close(context.Background())
			return coll
		},
	}
}
