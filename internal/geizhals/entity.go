package geizhals

import "fmt"

type Entity struct {
	Price float32    `json:"price"`
	Name  string     `json:"name"`
	URL   string     `json:"url"`
	Type  EntityType `json:"type"`
	html  []byte
}

func (e Entity) String() string {
	return fmt.Sprintf("'%s', Price: %.2f €, URL: '%s'", e.Name, e.Price, e.URL)
}

type EntityType int

const (
	Product  EntityType = 0
	Wishlist EntityType = 1
)
