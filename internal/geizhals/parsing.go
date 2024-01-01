package geizhals

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Price struct {
	Price    float64
	Currency Currency
}

// parsePrice parses a price from a given string, returns 0 if no price could be found.
func parsePrice(priceString string) (Price, error) {
	var currency Currency

	switch {
	case strings.Contains(priceString, "€"):
		currency = EUR
	case strings.Contains(priceString, "£"):
		currency = GBP
	case strings.Contains(priceString, "PLN"):
		currency = PLN
	default:
		return Price{}, fmt.Errorf("could not parse price")
	}

	priceString = strings.ReplaceAll(priceString, ",", ".")
	priceString = strings.ReplaceAll(priceString, "€ ", "")
	priceString = strings.ReplaceAll(priceString, "£ ", "")
	priceString = strings.ReplaceAll(priceString, "PLN ", "")

	price, err := strconv.ParseFloat(priceString, 64)
	if err != nil {
		log.Printf("Can't parse price: '%s' - %s", priceString, err)
		return Price{}, fmt.Errorf("could not parse price: %w", err)
	}

	return Price{Price: price, Currency: currency}, nil
}

// parseEntity calls either parseWishlist or parseProduct depending on the entityType.
func parseEntity(ghURL EntityURL, doc *goquery.Document) (Entity, error) {
	// Then we need to parse products/wishlists differently
	var (
		parseErr error
		name     string
		price    Price
	)

	entity := Entity{
		ID:   ghURL.EntityID,
		URL:  ghURL.Path,
		Type: ghURL.Type,
	}

	switch ghURL.Type {
	case Product:
		name, price, parseErr = parseProduct(doc)
	case Wishlist:
		name, price, parseErr = parseWishlist(doc)
	default:
		log.Printf("Invalid entityType '%v'\n", ghURL.Type)
		return entity, fmt.Errorf("invalid entityType")
	}

	if parseErr != nil {
		return entity, parseErr
	}

	entity.Name = name
	entity.Prices = []EntityPrice{{
		EntityID: entity.ID,
		Price:    price.Price,
		Currency: price.Currency,
		Location: ghURL.Location,
	}}

	return entity, nil
}

// parseWishlist parses the geizhals wishlist page and returns an Entity struct.
func parseWishlist(doc *goquery.Document) (string, Price, error) {
	// Parse name from html
	name := doc.Find("div.wishlist h1.wishlist__headline > span").Text()
	name = strings.TrimSpace(name)

	// Parse price from html
	priceString := doc.Find("div.wishlist_sum_area span.gh_price span.gh_price > span.gh_price").Text()

	price, parseErr := parsePrice(priceString)
	if parseErr != nil {
		return "", Price{}, parseErr
	}

	return name, price, nil
}

// parseProduct parses the geizhals product page and returns an Entity struct.
func parseProduct(doc *goquery.Document) (string, Price, error) {
	// parse name from html
	name := doc.Find("div.variant__header h1").Text()
	name = strings.TrimSpace(name)

	// Parse price from html
	priceString := doc.Find("div#offer__price-0 span.gh_price").Text()

	price, parseErr := parsePrice(priceString)
	if parseErr != nil {
		return "", Price{}, parseErr
	}

	return name, price, nil
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

		entityID, convertErr := strconv.ParseInt(dataID, 10, 0)
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

		entityIDs = append(entityIDs, entityID)
		amounts = append(amounts, amount)
	})

	return entityIDs, amounts, parseErr
}
