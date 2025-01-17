package cache

// Identifiable represents an entity that has an ID
type Identifiable interface {
	GetID() uint
}

// ForeignKeyable represents an entity that has a foreign key to User
type ForeignKeyable interface {
	Identifiable
	GetUserID() uint
}

// DataLoader defines the interface for loading data
type DataLoader[T any] interface {
	Load() ([]T, error)
}

// QueryCondition defines the interface for query conditions
type QueryCondition[T any] interface {
	Match(item T) bool
}

// Cache defines the interface for cache operations
type Cache[T any] interface {
	Get(id uint) (T, bool)
	GetAll() []T
	Query(condition QueryCondition[T]) []T
	Refresh() error
	Clear()
}

// RelatedCache extends Cache with foreign key operations
type RelatedCache[T ForeignKeyable] interface {
	Cache[T]
	GetByForeignKey(fkID uint) []T
}
