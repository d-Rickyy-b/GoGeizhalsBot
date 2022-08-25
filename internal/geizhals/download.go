package geizhals

import (
	"GoGeizhalsBot/internal/prometheus"
	"GoGeizhalsBot/internal/proxy"
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

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

	// execute function downloadHTML() maximum 3 times to avoid 429 Too Many Requests
	for retries := 0; retries < 3; retries++ {
		// First we download the html content of the given URL
		doc, statusCode, downloadErr = downloadHTML(url.CleanURL)
		if downloadErr == nil {
			break
		}

		if statusCode == http.StatusTooManyRequests {
			log.Println("Too many requests, trying again!")
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
	httpClient := &http.Client{}
	if proxyURL != nil {
		httpClient.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
		log.Println("Using proxy: ", proxyURL)
	}

	prometheus.GeizhalsHTTPRequests.Inc()
	resp, getErr := httpClient.Get(entityURL)
	if getErr != nil {
		log.Println(getErr)
		prometheus.HttpErrors.Inc()
		return nil, 0, fmt.Errorf("error while downloading content from Geizhals: %w", getErr)
	}
	// Cleanup when this function ends
	defer resp.Body.Close()

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
