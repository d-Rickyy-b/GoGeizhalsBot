package geizhals

import (
	"GoGeizhalsBot/internal/prometheus"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var wishlistURLPattern = regexp.MustCompile(`^((?:https?://)?(?:geizhals\.(?:de|at|eu)|cenowarka\.pl|skinflint\.co\.uk)(/\?cat=WL(-\d+))).*$`)
var productURLPattern = regexp.MustCompile(`^((?:https?://)?(?:geizhals\.(?:de|at|eu)|cenowarka\.pl|skinflint\.co\.uk)(/[0-9a-zA-Z\-]*a(\d+).html))\??.*$`)

// parsePrice parses a price from a given string, returns 0 if no price could be found.
func parsePrice(priceString string) (float64, error) {
	priceString = strings.ReplaceAll(priceString, ",", ".")
	priceString = strings.ReplaceAll(priceString, "€ ", "")
	price, err := strconv.ParseFloat(priceString, 64)
	if err != nil {
		log.Printf("Can't parse price: '%s' - %s", priceString, err)
	}
	return price, err
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

// downloadHTML downloads the HTML content of the given URL and returns the document and the HTTP status code.
func downloadHTML(entityURL string) (*goquery.Document, int, error) {
	proxyURL := getNextProxy()
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

// parseEntity calls either parseWishlist or parseProduct depending on the entityType.
func parseEntity(url string, entityType EntityType, doc *goquery.Document) (Entity, error) {
	// Then we need to parse products/wishlists differently
	var (
		parseErr error
		entity   Entity
		matches  []string
	)

	switch entityType {
	case Product:
		matches = productURLPattern.FindStringSubmatch(url)
		entity, parseErr = parseProduct(doc)
	case Wishlist:
		matches = wishlistURLPattern.FindStringSubmatch(url)
		entity, parseErr = parseWishlist(doc)
	default:
		log.Printf("Invalid URL '%s'\n", url)
		return Entity{}, fmt.Errorf("invalid URL")
	}
	if parseErr != nil {
		return Entity{}, parseErr
	}

	if len(matches) != 3 {
		return Entity{}, fmt.Errorf("no matches found for URL '%s'", url)
	}
	entityIDString := matches[2]
	cleanedURL := matches[1]
	entityID, err := strconv.Atoi(entityIDString)
	if err != nil {
		return Entity{}, fmt.Errorf("couldn't parse entity ID: %s", entityIDString)
	}

	// Eventually set the correct url
	entity.URL = cleanedURL
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
	price, parseErr := parsePrice(priceString)
	if parseErr != nil {
		return Entity{}, parseErr
	}

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
	price, parseErr := parsePrice(priceString)
	if parseErr != nil {
		return Entity{}, parseErr
	}

	product := Entity{
		Price: price,
		Name:  name,
		Type:  Product,
	}

	return product, nil
}

// UpdateEntity returns an updated Entity struct from a given input Entity
func UpdateEntity(entity Entity) (Entity, error) {
	updatedEntity, downloadErr := DownloadEntity(entity.URL)
	updatedEntity.ID = entity.ID
	return updatedEntity, downloadErr
}

// parseWishlistEntityIDsAndAmounts parses the wishlist entity IDs and the amount of the entity in the wishlist
// from the given HTML document.
func parseWishlistEntityIDsAndAmounts(doc *goquery.Document) (entityIDs []int64, amounts []int64, parseErr error) {
	// get wishlist__item and iterate over all of them
	wishlistItems := doc.Find("div.wishlist__item")
	wishlistItems.Each(func(i int, selection *goquery.Selection) {
		// Extract amount and product ID
		dataID, idExists := selection.Attr("data-id")
		if !idExists {
			return
		}
		ID, convertErr := strconv.ParseInt(dataID, 10, 0)
		if convertErr != nil {
			return
		}

		value, valueExists := selection.Attr("data-count")
		if !valueExists {
			return
		}
		amount, err := strconv.ParseInt(value, 10, 0)
		if err != nil {
			return
		}

		entityIDs = append(entityIDs, ID)
		amounts = append(amounts, amount)
	})
	return entityIDs, amounts, parseErr
}
