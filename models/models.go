package models

import "time"

// User represents the user table model
type User struct {
	ID    uint   `json:"id" gorm:"primaryKey"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// GetID implements the Identifiable interface
func (u User) GetID() uint {
	return u.ID
}

// GetTableName returns the table name
func (u User) GetTableName() string {
	return "users"
}

// Order represents the order table model
type Order struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

// GetID implements the Identifiable interface
func (o Order) GetID() uint {
	return o.ID
}

// GetUserID implements the ForeignKeyable interface
func (o Order) GetUserID() uint {
	return o.UserID
}

// GetTableName returns the table name
func (o Order) GetTableName() string {
	return "orders"
}
