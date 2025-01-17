package loader

import (
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

// DataLoader defines the interface for loading data
type DataLoader[T any] interface {
	Load() ([]T, error)
}

// GormLoaderOption defines the interface for GORM loader options
type GormLoaderOption interface {
	Apply(*gorm.DB) *gorm.DB
}

// MongoLoaderOption defines the interface for MongoDB loader options
type MongoLoaderOption interface {
	Apply(*mongo.Collection) *mongo.Collection
}

// Loader defines the common interface for all loaders
type Loader[T any] interface {
	DataLoader[T]
	WithDebug(debug bool) Loader[T]
}

// GormDataLoader defines the interface for GORM specific loader operations
type GormDataLoader[T any] interface {
	Loader[T]
	WithCondition(query interface{}, args ...interface{}) GormDataLoader[T]
	WithPreload(preloads ...string) GormDataLoader[T]
	WithPreloadQuery(relation string, query interface{}, args ...interface{}) GormDataLoader[T]
	WithPreloadJoin(relation string, query interface{}, args ...interface{}) GormDataLoader[T]
	WithJoins(joins ...string) GormDataLoader[T]
	WithJoinsModel(model interface{}, foreignKey, referenceKey string) GormDataLoader[T]
}

// MongoDataLoader defines the interface for MongoDB specific loader operations
type MongoDataLoader[T any] interface {
	Loader[T]
	WithFilter(filter interface{}) MongoDataLoader[T]
	WithOptions(opts interface{}) MongoDataLoader[T]
	WithAggregate(pipeline mongo.Pipeline) MongoDataLoader[T]
}

// JoinModel represents a join model configuration
type JoinModel struct {
	Model        interface{}
	ForeignKey   string
	ReferenceKey string
}

// PreloadQuery represents a preload query configuration
type PreloadQuery struct {
	Relation string
	Query    interface{}
	Args     []interface{}
}

// LoaderConfig represents the common configuration for all loaders
type LoaderConfig struct {
	Debug bool
}

// GormLoaderConfig represents the configuration for GORM loader
type GormLoaderConfig struct {
	LoaderConfig
	Model          interface{}
	Conditions     []interface{}
	Preloads       []string
	PreloadQueries map[string]PreloadQuery
	PreloadJoins   map[string][]interface{}
	Joins          []string
	JoinModels     []JoinModel
}

// MongoLoaderConfig represents the configuration for MongoDB loader
type MongoLoaderConfig struct {
	LoaderConfig
	Filter    interface{}
	Options   interface{}
	Pipeline  mongo.Pipeline
	Aggregate bool
}
