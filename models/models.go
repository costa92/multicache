package models

import "time"

// User represents the user table model
type User struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	// Orders []Order `json:"orders"`
}

// func (u User) GetTableName() string {
// 	return "users"
// }

// Order represents the order table model
type Order struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

func (o Order) GetTableName() string {
	return "orders"
}

// GetID implements the Identifiable interface
func (u User) GetID() uint {
	return u.ID
}

// GetID implements the Identifiable interface
func (o Order) GetID() uint {
	return o.ID
}

// GetUserID implements the ForeignKeyable interface
func (o Order) GetUserID() uint {
	return o.UserID
}
