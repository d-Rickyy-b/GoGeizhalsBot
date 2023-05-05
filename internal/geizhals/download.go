package geizhals

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/d-Rickyy-b/gogeizhalsbot/v2/internal/config"

	"github.com/d-Rickyy-b/gogeizhalsbot/v2/internal/prometheus"
	"github.com/d-Rickyy-b/gogeizhalsbot/v2/internal/proxy"

	"github.com/PuerkitoBio/goquery"
)

const priceHistoryURL = "https://geizhals.de/api/gh0/price_history"

// DownloadEntity retrieves the metadata (name, price) for a given entity hosted on Geizhals.
func DownloadEntity(url string) (Entity, error) {
	ghURL, parseErr := parseGeizhalsURL(url)
	if parseErr != nil {
		log.Printf("Error while parsing URL: %s - %s\n", url, parseErr)
		return Entity{}, parseErr
	}

	return downloadEntity(ghURL)
}

// downloadEntity retrieves the metadata (name, price) for a given entity hosted on Geizhals.
func downloadEntity(url EntityURL) (Entity, error) {
	var (
		doc         *goquery.Document
		statusCode  int
		downloadErr error
	)

	maxTries := maxTries()

	// execute function downloadHTML() maximum 3 times to avoid 429 Too Many Requests
	for tries := 0; tries < maxTries; tries++ {
		// First we download the html content of the given URL
		doc, statusCode, downloadErr = downloadHTML(url.CleanURL)
		if downloadErr == nil {
			break
		}

		if statusCode == http.StatusTooManyRequests {
			log.Printf("Too many requests, trying again (%d/%d)!\n", tries+1, maxTries)
			continue
		}

		return Entity{}, downloadErr
	}

	if downloadErr != nil {
		return Entity{}, downloadErr
	}

	return parseEntity(url, doc)
}

// downloadHTML downloads the HTML content of the given URL and returns the document and the HTTP status code.
func downloadHTML(entityURL string) (*goquery.Document, int, error) {
	proxyURL := proxy.GetNextProxy()
	httpClient := &http.Client{Timeout: time.Second * 10}

	if proxyURL != nil {
		log.Println("Using proxy: ", proxyURL)
		httpClient.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	}

	prometheus.GeizhalsHTTPRequests.Inc()

	resp, getErr := httpClient.Get(entityURL)
	if getErr != nil {
		log.Println(getErr)
		prometheus.HTTPErrors.Inc()

		return nil, 0, fmt.Errorf("error while downloading content from Geizhals: %w", getErr)
	}
	// Cleanup when this function ends
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		prometheus.HTTPRequests429.Inc()
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("Received status code %d - returning...\n", resp.StatusCode)
		return nil, resp.StatusCode, fmt.Errorf("error for http request")
	}

	// Read & parse response data
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("error while parsing body: %w", err)
	}

	return doc, resp.StatusCode, nil
}

func DownloadPriceHistory(entityIDs, amounts []int64, location string) (PriceHistory, error) {
	var downloadErr error
	var resp *http.Response

	maxTries := maxTries()

	// execute function downloadHTML() maximum 3 times to avoid 429 Too Many Requests
	for tries := 0; tries < maxTries; tries++ {
		resp, downloadErr = downloadPriceHistory(entityIDs, amounts, location)

		if resp.StatusCode == http.StatusTooManyRequests {
			log.Printf("Too many requests, trying again (%d/%d)!", tries+1, maxTries)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("Received status code %d - returning...\n", resp.StatusCode)
			return PriceHistory{}, fmt.Errorf("error for http request")
		}

		break
	}
	// Cleanup when this function ends
	defer resp.Body.Close()

	if downloadErr != nil {
		log.Println(downloadErr)
		return PriceHistory{}, fmt.Errorf("error while downloading content from Geizhals: %w", downloadErr)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		log.Printf("Too many requests, returning...\n")
		return PriceHistory{}, ErrTooManyRetries
	}

	var pricehistory PriceHistory
	unmarshalErr := json.NewDecoder(resp.Body).Decode(&pricehistory)
	if unmarshalErr != nil {
		return PriceHistory{}, fmt.Errorf("error while unmarshalling response: %w", unmarshalErr)
	}
	pricehistory.Meta.DownloadedAt = time.Now()

	return pricehistory, nil
}

func downloadPriceHistory(entityIDs, amounts []int64, location string) (*http.Response, error) {
	proxyURL := proxy.GetNextProxy()
	httpClient := &http.Client{}
	if proxyURL != nil {
		httpClient.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
		log.Println("Using proxy: ", proxyURL)
	}

	// Currently, this requests only supports geizhals.de (coming from loc = "de").
	requestBody := priceHistoryRequest{
		ID:        entityIDs,
		Itemcount: amounts,
		Params: struct {
			Days int    `json:"days"`
			Loc  string `json:"loc"`
		}{Days: 9999, Loc: location},
	}

	result, marshalErr := json.Marshal(requestBody)
	if marshalErr != nil {
		return nil, fmt.Errorf("error while marshalling request: %w", marshalErr)
	}

	prometheus.GeizhalsHTTPRequests.Inc()
	resp, downloadErr := httpClient.Post(priceHistoryURL, "application/json", bytes.NewBuffer(result))
	if downloadErr != nil {
		prometheus.HTTPErrors.Inc()
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		prometheus.HTTPRequests429.Inc()
	}
	return resp, downloadErr
}

// maxTries returns the maximum number of tries for http requests from the config.
func maxTries() int {
	conf, err := config.GetConfig()
	if err != nil {
		log.Println("Error while reading config file: ", err)
		return 3
	}

	return conf.HTTPMaxTries
}
