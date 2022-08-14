package models

import (
	"GoGeizhalsBot/internal/geizhals"
	"fmt"
	"time"
)

type PriceAgent struct {
	CreatedAt            time.Time
	ID                   int64                `json:"id" gorm:"primarykey;autoIncrement:true"`
	Name                 string               `json:"name"`
	UserID               int64                `json:"user_id" gorm:"not null;default:null;index:user_entity_idx,unique"`
	User                 User                 `json:"user" gorm:"foreignkey:UserID"`
	EntityID             int64                `json:"-" gorm:"index:user_entity_idx,unique"`
	Entity               geizhals.Entity      `json:"entity" gorm:"foreignkey:EntityID"`
	Location             string               `json:"location" gorm:"default:de"`
	NotificationID       int64                `json:"-"`
	NotificationSettings NotificationSettings `json:"notificationSettings" gorm:"foreignkey:NotificationID;constraint:OnDelete:CASCADE;"`
	Enabled              bool                 `json:"enabled" gorm:"default:1"`
}

func (pa PriceAgent) String() string {
	return fmt.Sprintf("%d - '%s' (%s) | User: %d", pa.ID, pa.Name, pa.Entity.Name, pa.UserID)
}

type NotificationSettings struct {
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ID              int64   `gorm:"primarykey"`
	NotifyAlways    bool    `json:"notifyAlways"`
	NotifyPriceDrop bool    `json:"notifyPriceDrop"`
	NotifyPriceRise bool    `json:"notifyPriceRise"`
	NotifyAbove     bool    `json:"notifyAbove"`
	NotifyBelow     bool    `json:"notifyBelow"`
	AbovePrice      float64 `json:"abovePrice"`
	BelowPrice      float64 `json:"belowPrice"`
}

func (ns NotificationSettings) String() string {
	var humanReadableSettings string
	switch {
	case ns.NotifyBelow:
		humanReadableSettings = fmt.Sprintf("Unter %.2f €", ns.BelowPrice)
	case ns.NotifyAlways:
		humanReadableSettings = "Immer alarmieren"
	default:
		humanReadableSettings = "Unbekannt"
	}
	return humanReadableSettings
}
