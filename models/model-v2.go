package models

type UserV2 struct {
	ID     uint    `json:"id"`
	Name   string  `json:"name"`
	Email  string  `json:"email"`
	Orders []Order `json:"orders" gorm:"foreignKey:UserID"`
}

func (u *UserV2) GetID() uint {
	return u.ID
}
