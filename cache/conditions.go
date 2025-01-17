package cache

import (
	"strings"
)

// StringFieldCondition represents a condition for string field comparison
type StringFieldCondition[T any] struct {
	FieldExtractor func(T) string
	Value          string
	Operation      string // "eq", "contains", "startsWith", "endsWith"
}

func (c StringFieldCondition[T]) Match(item T) bool {
	fieldValue := c.FieldExtractor(item)
	switch c.Operation {
	case "eq":
		return fieldValue == c.Value
	case "contains":
		return strings.Contains(fieldValue, c.Value)
	case "startsWith":
		return strings.HasPrefix(fieldValue, c.Value)
	case "endsWith":
		return strings.HasSuffix(fieldValue, c.Value)
	default:
		return false
	}
}

// NumberFieldCondition represents a condition for number field comparison
type NumberFieldCondition[T any, N Number] struct {
	FieldExtractor func(T) N
	Value          N
	Operation      string // "eq", "gt", "gte", "lt", "lte"
}

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

func (c NumberFieldCondition[T, N]) Match(item T) bool {
	fieldValue := c.FieldExtractor(item)
	switch c.Operation {
	case "eq":
		return fieldValue == c.Value
	case "gt":
		return fieldValue > c.Value
	case "gte":
		return fieldValue >= c.Value
	case "lt":
		return fieldValue < c.Value
	case "lte":
		return fieldValue <= c.Value
	default:
		return false
	}
}

// CompositeCondition combines multiple conditions with AND/OR logic
type CompositeCondition[T any] struct {
	Conditions []QueryCondition[T]
	Operation  string // "and", "or"
}

func (c CompositeCondition[T]) Match(item T) bool {
	if len(c.Conditions) == 0 {
		return true
	}

	switch c.Operation {
	case "and":
		for _, cond := range c.Conditions {
			if !cond.Match(item) {
				return false
			}
		}
		return true
	case "or":
		for _, cond := range c.Conditions {
			if cond.Match(item) {
				return true
			}
		}
		return false
	default:
		return false
	}
}
