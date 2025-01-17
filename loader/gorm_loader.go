package loader

import (
	"fmt"

	"gorm.io/gorm"
)

// GormLoader implements DataLoader interface for GORM
type GormLoader[T any] struct {
	db             *gorm.DB
	model          T
	condition      interface{}
	preloads       []string
	joins          []string
	joinsModel     []joinModel
	preloadJoins   map[string][]interface{}
	preloadQueries map[string][]interface{}
	debug          bool
}

type joinModel struct {
	model        interface{}
	foreignKey   string
	referenceKey string
}

// NewGormLoader creates a new GORM data loader
func NewGormLoader[T any](db *gorm.DB, model T) *GormLoader[T] {
	return &GormLoader[T]{
		db:             db,
		model:          model,
		preloadJoins:   make(map[string][]interface{}),
		preloadQueries: make(map[string][]interface{}),
	}
}

// WithCondition adds a query condition
func (l *GormLoader[T]) WithCondition(query interface{}, args ...interface{}) *GormLoader[T] {
	l.condition = append([]interface{}{query}, args...)
	return l
}

// WithPreload adds preload relations
func (l *GormLoader[T]) WithPreload(preloads ...string) *GormLoader[T] {
	l.preloads = append(l.preloads, preloads...)
	return l
}

// WithPreloadQuery adds preload relations with conditions
func (l *GormLoader[T]) WithPreloadQuery(relation string, query interface{}, args ...interface{}) *GormLoader[T] {
	if l.preloadQueries == nil {
		l.preloadQueries = make(map[string][]interface{})
	}
	l.preloadQueries[relation] = append([]interface{}{query}, args...)
	return l
}

// WithPreloadJoin adds preload relations with joins and conditions
func (l *GormLoader[T]) WithPreloadJoin(relation string, query interface{}, args ...interface{}) *GormLoader[T] {
	l.preloadJoins[relation] = append([]interface{}{query}, args...)
	return l
}

// WithJoins adds join clauses
func (l *GormLoader[T]) WithJoins(joins ...string) *GormLoader[T] {
	l.joins = append(l.joins, joins...)
	return l
}

// WithJoinsModel adds join clauses using model and fields
func (l *GormLoader[T]) WithJoinsModel(model interface{}, foreignKey, referenceKey string) *GormLoader[T] {
	l.joinsModel = append(l.joinsModel, joinModel{model: model, foreignKey: foreignKey, referenceKey: referenceKey})
	return l
}

// WithDebug enables debug mode for the loader
func (l *GormLoader[T]) WithDebug(debug bool) *GormLoader[T] {
	l.debug = debug
	return l
}

// Load implements DataLoader interface
func (l *GormLoader[T]) Load() ([]T, error) {
	var items []T
	query := l.db.Model(&l.model) // Ensure the model is set for the query

	// Add joins if any
	for _, join := range l.joins {
		query = query.Joins(join)
	}

	// Add joins using models and fields
	for _, jm := range l.joinsModel {
		stmt := &gorm.Statement{DB: l.db}
		if err := stmt.Parse(jm.model); err != nil {
			return nil, fmt.Errorf("failed to parse model: %w", err)
		}
		tableName := stmt.Schema.Table
		query = query.Joins(fmt.Sprintf("JOIN %s ON %s = %s", tableName, jm.foreignKey, jm.referenceKey))
	}

	// Add preloads with joins and conditions
	for relation, joinConditions := range l.preloadJoins {
		if len(joinConditions) > 0 {
			query = query.Joins(fmt.Sprintf("JOIN %s ON %v", relation, joinConditions[0]), joinConditions[1:]...)
		}
	}

	// Add preloads with conditions
	for relation, conditions := range l.preloadQueries {
		if len(conditions) > 0 {
			query = query.Preload(relation, func(db *gorm.DB) *gorm.DB {
				return db.Where(conditions[0], conditions[1:]...)
			})
		}
	}

	// Add regular preloads
	for _, preload := range l.preloads {
		query = query.Preload(preload)
	}

	// Add conditions if any
	if l.condition != nil {
		conditions, ok := l.condition.([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid condition format")
		}

		if len(conditions) > 0 {
			query = query.Where(conditions[0], conditions[1:]...)
		}
	}

	// Enable debug mode if requested
	if l.debug {
		query = query.Debug()
	}

	result := query.Find(&items)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to load data: %w", result.Error)
	}

	return items, nil
}
