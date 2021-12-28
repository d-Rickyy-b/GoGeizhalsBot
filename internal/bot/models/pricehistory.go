package models

import (
	"GoGeizhalsBot/internal/geizhals"
	"time"
)

type HistoricPrice struct {
	ID        uint64 `gorm:"primary_key;autoIncrement:true"`
	CreatedAt time.Time
	Price     float64
	EntityID  int64
	Entity    geizhals.Entity `gorm:"foreignkey:EntityID"`
}
