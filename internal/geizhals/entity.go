package geizhals

import (
	"fmt"
	"time"
)

type Entity struct {
	ID         int64 `json:"id"`
	GeizhalsID int64
	UpdatedAt  time.Time
	Prices     []EntityPrice `gorm:"foreignkey:EntityID;constraint:OnDelete:CASCADE;"`
	Name       string        `json:"name"`
	URL        string        `json:"url"`
	Type       EntityType    `json:"type"`
}

// FullURL returns the URL to download the HTML of the entity for the given location.
func (e Entity) FullURL(location string) string {
	domain, ok := geizhalsDomains[location]
	if !ok {
		return ""
	}

	return fmt.Sprintf("https://%s/%s", domain, e.URL)
}

func (e Entity) String() string {
	return fmt.Sprintf("'%s', URL: '%s'", e.Name, e.URL)
}

// GetPrice returns the price (EntityPrice) of the entity for the given location.
// If no price is found for the given location, an empty EntityPrice is returned.
func (e Entity) GetPrice(location string) EntityPrice {
	for _, p := range e.Prices {
		if p.Location == location {
			return p
		}
	}

	return EntityPrice{
		EntityID:  e.ID,
		Location:  location,
		UpdatedAt: time.Time{},
		Price:     0,
		Currency:  CurrencyFromLocation(location),
	}
}

type EntityType int

const (
	Product  EntityType = 1
	Wishlist EntityType = 2
)
