package geizhals

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var wishlistURLPattern = regexp.MustCompile(`^((?:https?://)?geizhals\.(?:de|at|eu)/\?cat=WL-(\d+))$`)
var productURLPattern = regexp.MustCompile(`^((?:https?://)?geizhals\.(?:de|at|eu)/[0-9a-zA-Z\-]*a(\d+).html)\??.*$`)

// IsValidURL checks if the given URL is a valid Geizhals URL.
func IsValidURL(url string) bool {
	return wishlistURLPattern.MatchString(url) || productURLPattern.MatchString(url)
}

// parsePrice parses a price from a given string, returns 0 if no price could be found.
func parsePrice(priceString string) float64 {
	priceString = strings.ReplaceAll(priceString, ",", ".")
	priceString = strings.ReplaceAll(priceString, "€ ", "")
	price, err := strconv.ParseFloat(priceString, 64)
	if err != nil {
		log.Printf("Can't parse price: '%s' - %s", priceString, err)
		return 0
	}
	return price
}

// DownloadEntity retrieves the metadata (name, price) for a given entity hosted on Geizhals.
func DownloadEntity(url string) (Entity, error) {
	var entityType EntityType

	switch {
	case wishlistURLPattern.MatchString(url):
		entityType = Wishlist
	case productURLPattern.MatchString(url):
		entityType = Product
	default:
		return Entity{}, fmt.Errorf("invalid URL: %s", url)
	}

	var doc *goquery.Document
	var statusCode int
	var downloadErr error

	// execute function downloadHTML() maximum 3 times to avoid 429 Too Many Requests
	for retries := 0; retries < 3; retries++ {
		// First we download the html content of the given URL
		doc, statusCode, downloadErr = downloadHTML(url)
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

	return parseEntity(url, entityType, doc)
}

func downloadHTML(entityURL string) (*goquery.Document, int, error) {
	// TODO try (at max.) three different proxies if there's a connection error
	proxyURL := getNextProxy()
	httpClient := &http.Client{}
	if proxyURL != nil {
		httpClient.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
		log.Println("Using proxy: ", proxyURL)
	}

	resp, getErr := httpClient.Get(entityURL)
	if getErr != nil {
		log.Println(getErr)
		return nil, resp.StatusCode, fmt.Errorf("error while downloading content from Geizhals: %w", getErr)
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

// parseEntity calls either parseWishlist or parseProduct depending on the entityType.
func parseEntity(url string, entityType EntityType, doc *goquery.Document) (Entity, error) {
	// Then we need to parse products/wishlists differently
	var (
		parseErr error
		entity   Entity
		matches  [][]string
	)

	switch entityType {
	case Product:
		matches = productURLPattern.FindAllStringSubmatch(url, -1)
		entity, parseErr = parseProduct(doc)
	case Wishlist:
		matches = wishlistURLPattern.FindAllStringSubmatch(url, -1)
		entity, parseErr = parseWishlist(doc)
	default:
		log.Printf("Invalid URL '%s'\n", url)
		return Entity{}, fmt.Errorf("invalid URL")
	}
	if parseErr != nil {
		return Entity{}, parseErr
	}

	entityIDString := matches[0][2]
	entityID, err := strconv.Atoi(entityIDString)
	if err != nil {
		return Entity{}, fmt.Errorf("couldn't parse entity ID: %s", entityIDString)
	}

	// Eventually set the correct url
	entity.URL = url
	entity.ID = int64(entityID)
	return entity, nil
}

// parseWishlist parses the geizhals wishlist page and returns an Entity struct.
func parseWishlist(doc *goquery.Document) (Entity, error) {
	// Parse name from html
	nameSelection := doc.Find("div.wishlist h1.wishlist__headline > span")
	name := nameSelection.Text()
	name = strings.TrimSpace(name)

	// Parse price from html
	priceSelection := doc.Find("div.wishlist_sum_area span.gh_price span.gh_price > span.gh_price")
	priceString := priceSelection.Text()
	price := parsePrice(priceString)

	wishlist := Entity{
		Price: price,
		Name:  name,
		Type:  Wishlist,
	}
	return wishlist, nil
}

// parseProduct parses the geizhals product page and returns an Entity struct.
func parseProduct(doc *goquery.Document) (Entity, error) {
	// parse name from html
	nameSelection := doc.Find("div.variant__header h1[itemprop='name']")
	name := nameSelection.Text()
	name = strings.TrimSpace(name)

	// Parse price from html
	priceSelection := doc.Find("div#offer__price-0 span.gh_price")
	priceString := priceSelection.Text()
	price := parsePrice(priceString)

	product := Entity{
		Price: price,
		Name:  name,
		Type:  Product,
	}

	return product, nil
}

// UpdateEntity returns an updated Entity struct from a given input Entity
func UpdateEntity(entity Entity) (Entity, error) {
	return DownloadEntity(entity.URL)
}
