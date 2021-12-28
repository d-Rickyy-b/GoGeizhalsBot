package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID          int64        `json:"id" gorm:"unique;primaryKey"`
	CreatedAt   time.Time    `json:"-"`
	Username    string       `json:"username"`
	FirstName   string       `json:"first_name"`
	LastName    string       `json:"last_name"`
	LangCode    string       `json:"language_code"`
	PriceAgents []PriceAgent `json:"-"`
}
