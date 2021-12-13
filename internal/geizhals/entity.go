package geizhals

type Entity struct {
	Price float64    `json:"price"`
	Name  string     `json:"name"`
	URL   string     `json:"url"`
	Type  EntityType `json:"type"`
	html  []byte
}

type EntityType int

const (
	Product  EntityType = 0
	Wishlist EntityType = 1
)
