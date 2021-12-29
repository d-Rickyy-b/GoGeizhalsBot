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

var proxies []*url.URL

// IsValidURL checks if the given URL is a valid Geizhals URL.
func IsValidURL(url string) bool {
	return wishlistURLPattern.MatchString(url) || productURLPattern.MatchString(url)
}

// parsePrice parses a price
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
	matchWishlist := wishlistURLPattern.MatchString(url)
	matchProduct := productURLPattern.MatchString(url)
	if !matchWishlist && !matchProduct {
		return Entity{}, fmt.Errorf("invalid URL: %s", url)
	}

	// First we download the html content of the given URL
	doc, downloadErr := downloadHTML(url)
	if downloadErr != nil {
		return Entity{}, downloadErr
	}

	// Then we need to parse products/wishlists differently
	var (
		parseErr error
		entity   Entity
		matches  [][]string
	)

	switch {
	case matchProduct:
		matches = productURLPattern.FindAllStringSubmatch(url, -1)
		entity, parseErr = parseProduct(doc)
	case matchWishlist:
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

// InitProxies initializes the proxy list.
func InitProxies(p []string) {
	// TODO scramble proxies
	for _, proxy := range p {
		parsedProxy, parseErr := url.Parse(proxy)
		if parseErr != nil {
			log.Println("Can't parse proxy!", parseErr)
			continue
		}
		proxies = append(proxies, parsedProxy)
	}
}

// getNextProxy returns the next proxy from the list. Proxies are cycled so that
// a maximum time between first and second use of the same proxy passes.
func getNextProxy() *url.URL {
	if len(proxies) == 0 {
		return nil
	}

	// Get next proxy, dequeue and enqueue again for round-robin
	proxy := proxies[0]
	proxies = proxies[1:]
	proxies = append(proxies, proxy)
	return proxy
}

func downloadHTML(entityURL string) (*goquery.Document, error) {
	// TODO try (at max.) three different proxies if there's a connection error
	proxyURL := getNextProxy()
	if proxyURL != nil {
		http.DefaultTransport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
		log.Println("Using proxy: ", proxyURL)
	}

	resp, getErr := http.Get(entityURL)
	if getErr != nil {
		log.Println(getErr)
		return nil, fmt.Errorf("error while downloading content from Geizhals: %w", getErr)
	}
	// Cleanup when this function ends
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Received status code %d - returning...\n", resp.StatusCode)
		return nil, fmt.Errorf("error for http request")
	}

	// Read & parse response data
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error while parsing body: %w", err)
	}

	return doc, nil
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
