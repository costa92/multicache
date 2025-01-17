package models

// UserV2 represents the user table model with Orders relation
type UserV2 struct {
	ID     uint    `json:"id" gorm:"primaryKey"`
	Name   string  `json:"name"`
	Email  string  `json:"email"`
	Orders []Order `json:"orders"`
}

// GetID implements the Identifiable interface
func (u UserV2) GetID() uint {
	return u.ID
}

// GetTableName returns the table name
func (u UserV2) GetTableName() string {
	return "user_v2"
}
