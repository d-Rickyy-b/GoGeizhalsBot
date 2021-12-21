package models

import "GoGeizhalsBot/internal/geizhals"

type PriceAgent struct {
	ID                   string               `json:"id"`
	Name                 string               `json:"name"`
	Entity               geizhals.Entity      `json:"entity"`
	NotificationSettings NotificationSettings `json:"notificationSettings"`
}

type NotificationSettings struct {
	NotifyAlways    bool    `json:"notifyAlways"`
	NotifyPriceDrop bool    `json:"notifyPriceDrop"`
	NotifyPriceRise bool    `json:"notifyPriceRise"`
	NotifyAbove     bool    `json:"notifyAbove"`
	NotifyBelow     bool    `json:"notifyBelow"`
	AbovePrice      float32 `json:"abovePrice"`
	BelowPrice      float32 `json:"belowPrice"`
}
