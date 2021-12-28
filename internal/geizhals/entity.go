package geizhals

import (
	"fmt"
	"time"
)

type Entity struct {
	ID        int64 `json:"id"`
	ChangedAt time.Time
	Price     float64    `json:"price"`
	Name      string     `json:"name"`
	URL       string     `json:"url"`
	Type      EntityType `json:"type"`
	html      []byte
}

func (e Entity) String() string {
	return fmt.Sprintf("'%s', Price: %.2f €, URL: '%s'", e.Name, e.Price, e.URL)
}

type EntityType int

const (
	Product  EntityType = 1
	Wishlist EntityType = 2
)
