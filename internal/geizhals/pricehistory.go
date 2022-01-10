package geizhals

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type priceHistoryRequest struct {
	ID     int64 `json:"id"`
	Params struct {
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
}

type PriceEntry struct {
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"ts"`
	Valid     bool      `json:"valid"`
}

var userCache = make(map[int64]PriceHistory)

// UnmarshalJSON implements a custom unmarshaller for the price history response.
// The API response is an array of length 3, containing the timestamp, price and validity.
func (entry *PriceEntry) UnmarshalJSON(p []byte) error {
	// Thanks to https://jhall.io/posts/go-json-tricks-array-as-structs/
	var tmp []json.RawMessage
	if err := json.Unmarshal(p, &tmp); err != nil {
		return err
	}

	if len(tmp) != 3 {
		return fmt.Errorf("invalid response: %v", tmp)
	}

	// Parse timestamp
	var timestampMillis int64
	if err := json.Unmarshal(tmp[0], &timestampMillis); err != nil {
		return err
	}
	parsedTime := time.Unix(0, timestampMillis*int64(time.Millisecond))
	entry.Timestamp = parsedTime

	// Parse price
	if err := json.Unmarshal(tmp[1], &entry.Price); err != nil {
		return err
	}

	// Parse validity
	var numericBool float64
	if err := json.Unmarshal(tmp[2], &numericBool); err != nil {
		return err
	}
	entry.Valid = numericBool != 0

	return nil
}

// DownloadPriceHistory downloads the price history for the given entity.
func DownloadPriceHistory(entity Entity) (PriceHistory, error) {
	if entity.Type != Product {
		return PriceHistory{}, fmt.Errorf("can only fetch pricehistory for products as of now")
	}

	// Check if we already have the price history in cache
	if priceHistory, ok := userCache[entity.ID]; ok {
		// Check if the price history is still valid
		if time.Since(priceHistory.Meta.DownloadedAt) < 12*time.Hour {
			return priceHistory, nil
		}
		delete(userCache, entity.ID)
	}

	priceHistoryAPI := "https://geizhals.de/api/gh0/price_history"
	proxyURL := getNextProxy()
	httpClient := &http.Client{}
	if proxyURL != nil {
		httpClient.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
		log.Println("Using proxy: ", proxyURL)
	}

	// Currently, this requests only supports geizhals.de (coming from loc = "de").
	requestBody := priceHistoryRequest{
		ID: entity.ID,
		Params: struct {
			Days int    `json:"days"`
			Loc  string `json:"loc"`
		}{Days: 9999, Loc: "de"},
	}

	result, marshalErr := json.Marshal(requestBody)
	if marshalErr != nil {
		return PriceHistory{}, fmt.Errorf("error while marshalling request: %w", marshalErr)
	}

	resp, getErr := httpClient.Post(priceHistoryAPI, "application/json", bytes.NewBuffer(result)) //nolint:gosec
	if getErr != nil {
		log.Println(getErr)
		return PriceHistory{}, fmt.Errorf("error while downloading content from Geizhals: %w", getErr)
	}
	// Cleanup when this function ends
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Received status code %d - returning...\n", resp.StatusCode)
		return PriceHistory{}, fmt.Errorf("error for http request")
	}

	var pricehistory PriceHistory
	unmarshalErr := json.NewDecoder(resp.Body).Decode(&pricehistory)
	if unmarshalErr != nil {
		return PriceHistory{}, fmt.Errorf("error while unmarshalling response: %w", unmarshalErr)
	}
	pricehistory.Meta.DownloadedAt = time.Now()

	// Cache the price history
	userCache[entity.ID] = pricehistory

	return pricehistory, nil
}
