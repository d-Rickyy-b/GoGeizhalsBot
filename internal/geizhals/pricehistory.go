package geizhals

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

type priceHistoryRequest struct {
	ID        []int64 `json:"id"`
	Itemcount []int64 `json:"itemcount"`
	Params    struct {
		Days int    `json:"days"`
		Loc  string `json:"loc"`
	} `json:"params"`
}

type PriceHistoryMeta struct {
	LastFormatted string  `json:"last_formatted"`
	Min           float64 `json:"min"`
	Max           float64 `json:"max"`
	CurrentBest   float64 `json:"current_best"`
	FirstTS       float64 `json:"first_ts"`
	LastTS        float64 `json:"last_ts"`
	DownloadedAt  time.Time
}

type PriceHistory struct {
	Meta     PriceHistoryMeta `json:"meta"`
	Response []PriceEntry     `json:"response"`
	Location string           `json:"location"`
}

type PriceEntry struct {
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"ts"`
	Valid     bool      `json:"valid"`
}

var (
	userCache  = make(map[int64]PriceHistory)
	cacheMutex sync.Mutex
)

// UnmarshalJSON implements a custom unmarshaller for the price history response.
// The API response is an array of length 3, containing the timestamp, price and validity.
func (entry *PriceEntry) UnmarshalJSON(p []byte) error {
	// Thanks to https://jhall.io/posts/go-json-tricks-array-as-structs/
	var tmp []json.RawMessage
	if err := json.Unmarshal(p, &tmp); err != nil {
		return fmt.Errorf("unmarshal rawmessage error: %w", err)
	}

	if len(tmp) != 3 {
		return fmt.Errorf("invalid response: %v", tmp)
	}

	// Parse timestamp
	var timestampMillis int64
	if err := json.Unmarshal(tmp[0], &timestampMillis); err != nil {
		return fmt.Errorf("unmarshal timestamp error: %w", err)
	}
	parsedTime := time.Unix(0, timestampMillis*int64(time.Millisecond))
	entry.Timestamp = parsedTime

	// Parse price
	if err := json.Unmarshal(tmp[1], &entry.Price); err != nil {
		return fmt.Errorf("unmarshal price error: %w", err)
	}

	// Parse validity
	var numericBool float64
	if err := json.Unmarshal(tmp[2], &numericBool); err != nil {
		return fmt.Errorf("unmarshal validity error: %w", err)
	}
	entry.Valid = numericBool != 0

	return nil
}

// GetPriceHistory returns the price history for the given entity either from cache or by downloading it.
func GetPriceHistory(entity Entity, location string) (PriceHistory, error) {
	// Check if we already have the price history in cache
	history, isCached := getPriceHistoryFromCache(entity)
	if isCached {
		return history, nil
	}

	return getPriceHistory(entity, location)
}

// getPriceHistoryFromCache returns the price history for the given entity from cache, if it is cached.
func getPriceHistoryFromCache(entity Entity) (PriceHistory, bool) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if priceHistory, ok := userCache[entity.ID]; ok {
		// Check if the price history is still valid
		if time.Since(priceHistory.Meta.DownloadedAt) < 12*time.Hour {
			log.Printf("Using cached price history for '%s'\n", entity.Name)
			return priceHistory, true
		}

		delete(userCache, entity.ID)
	}

	return PriceHistory{}, false
}

// getEntityIDsAndAmounts returns the entity IDs and amounts for the given entity.
// For products, this is just the product ID and 1.
// For wishlists, this is the product IDs and amounts of the products contained in the wishlist.
// For wishlists an HTML download is required to obtain all the entity IDs and amounts.
func getEntityIDsAndAmounts(entity Entity, location string) ([]int64, []int64, error) {
	var entityIDs []int64
	var amounts []int64

	switch entity.Type {
	case Product:
		entityIDs = append(entityIDs, entity.ID)
		amounts = append(amounts, 1)
	case Wishlist:
		html, _, downloadErr := downloadHTML(entity.FullURL(location))
		if downloadErr != nil {
			return nil, nil, downloadErr
		}

		// Fetch Product IDs and Amounts
		var parseErr error
		entityIDs, amounts, parseErr = parseWishlistEntityIDsAndAmounts(html)
		if parseErr != nil {
			log.Println("Error parsing wishlist entities:", parseErr)
			return nil, nil, downloadErr
		}
	}

	return entityIDs, amounts, nil
}

// getPriceHistory downloads the price history for the given entity.
func getPriceHistory(entity Entity, location string) (PriceHistory, error) {
	entityIDs, amounts, err := getEntityIDsAndAmounts(entity, location)
	if err != nil {
		return PriceHistory{}, err
	}

	log.Printf("Downloading price history for '%s'\n", entity.Name)
	pricehistory, downloadErr := DownloadPriceHistory(entityIDs, amounts, location)
	if downloadErr != nil {
		return PriceHistory{}, downloadErr
	}

	if len(pricehistory.Response) == 0 {
		return PriceHistory{}, fmt.Errorf("no price history found for '%s'", entity.Name)
	}
	// Cache the price history
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	userCache[entity.ID] = pricehistory

	return pricehistory, nil
}
