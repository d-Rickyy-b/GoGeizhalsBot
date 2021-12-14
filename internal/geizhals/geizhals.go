package geizhals

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"

	"github.com/PuerkitoBio/goquery"
)

var wishlistURLPattern = regexp.MustCompile(`^((?:https?://)?geizhals\.(?:de|at|eu)/\?cat=WL-([0-9]+))$`)
var productURLPattern = regexp.MustCompile(`^((?:https?://)?geizhals\.(?:de|at|eu)/[0-9a-zA-Z\-]*a([0-9]+).html)\?.*$`)

var proxies []*url.URL

// IsValidURL checks if the given URL is a valid Geizhals URL.
func IsValidURL(url string) bool {
	return wishlistURLPattern.MatchString(url) || productURLPattern.MatchString(url)
}

// DownloadEntity retrieves the metadata (name, price) for a given entity hosted on Geizhals.
func DownloadEntity(url string) (Entity, error) {
	matchWishlist := wishlistURLPattern.MatchString(url)
	matchProduct := productURLPattern.MatchString(url)

	var err error
	var e Entity
	switch {
	case matchProduct:
		e, err = downloadProduct(url)
	case matchWishlist:
		e, err = downloadWishlist(url)
	default:
		return Entity{}, fmt.Errorf("invalid URL")
	}

	if err != nil {
		log.Fatal(err)
	}
	return e, nil
}

// InitProxies initializes the proxy list.
func InitProxies(p []string) {
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

// downloadWishlist downloads and parses the wishlist metadata for a given Geizhals wishlist URL.
func downloadWishlist(wishlistURL string) (Entity, error) {
	proxyURL := getNextProxy()
	if proxyURL != nil {
		http.DefaultTransport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
		log.Println("Using proxy: ", proxyURL)
	}

	resp, err := http.Get(wishlistURL)
	if err != nil {
		log.Println(err)
		return Entity{}, fmt.Errorf("invalid URL")
	}
	// Cleanup when this function ends
	defer resp.Body.Close()
	// Read & parse response data

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: Add correct selectors for parsing data
	// Print content of <title></title>
	doc.Find("title").Each(func(i int, s *goquery.Selection) {
		fmt.Printf("Title of the page: %s\n", s.Text())
	})

	wishlist := Entity{}
	return wishlist, nil
}

// downloadProduct downloads and parses the product metadata for a given Geizhals product URL.
func downloadProduct(productURL string) (Entity, error) {
	proxyURL := getNextProxy()
	if proxyURL != nil {
		http.DefaultTransport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
		log.Println("Using proxy: ", proxyURL)
	}

	resp, err := http.Get(productURL)
	if err != nil {
		print(err)
		return Entity{}, fmt.Errorf("invalid URL")
	}
	// Cleanup when this function ends
	defer resp.Body.Close()
	// Read & parse response data

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: Add correct selectors for parsing data
	// Print content of <title></title>
	doc.Find("title").Each(func(i int, s *goquery.Selection) {
		fmt.Printf("Title of the page: %s\n", s.Text())
	})

	product := Entity{}
	return product, nil
}
