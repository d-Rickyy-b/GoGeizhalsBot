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

func (pa PriceAgent) EntityURL() string {
	return pa.Entity.FullURL(pa.Location)
}

func (pa PriceAgent) CurrentEntityPrice() geizhals.EntityPrice {
	return pa.Entity.GetPrice(pa.Location)
}

func (pa PriceAgent) CurrentPrice() float64 {
	return pa.Entity.GetPrice(pa.Location).Price
}

func (pa PriceAgent) GetCurrency() geizhals.Currency {
	return pa.Entity.GetPrice(pa.Location).Currency
}

type NotificationSettings struct {
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ID              int64   `gorm:"primarykey"`
	NotifyAlways    bool    `json:"notifyAlways" gorm:"default:true"`
	NotifyPriceDrop bool    `json:"notifyPriceDrop" gorm:"default:false"`
	NotifyPriceRise bool    `json:"notifyPriceRise" gorm:"default:false"`
	NotifyAbove     bool    `json:"notifyAbove" gorm:"default:false"`
	NotifyBelow     bool    `json:"notifyBelow" gorm:"default:false"`
	AbovePrice      float64 `json:"abovePrice" gorm:"default:0"`
	BelowPrice      float64 `json:"belowPrice" gorm:"default:0"`
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
