package geizhals

import (
	"fmt"
	"time"
)

// EntityPrice represents a price for a specific Entity for a given location.
type EntityPrice struct {
	ID        int64
	EntityID  int64  `gorm:"not null;"`
	Location  string `gorm:"not null;"`
	UpdatedAt time.Time
	Price     float64  `gorm:"not null;default:0"`
	Currency  Currency `gorm:"not null;default:1"`
}

func (e EntityPrice) String() string {
	return fmt.Sprintf("%.2f %s", e.Price, e.Currency.String())
}

// Currency represents the currency of an entity price.
type Currency int

const (
	EUR Currency = 1
	PLN Currency = 2
	GBP Currency = 3
)

func (c Currency) String() string {
	switch c {
	case EUR:
		return "€"
	case PLN:
		return "zł"
	case GBP:
		return "£"
	}

	return ""
}

// CurrencyFromLocation returns the currency of the given location.
func CurrencyFromLocation(location string) Currency {
	switch location {
	case "de":
		return EUR
	case "at":
		return EUR
	case "pl":
		return PLN
	case "uk":
		return GBP
	}

	return EUR
}
