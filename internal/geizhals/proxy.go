package geizhals

import (
	"log"
	"net/url"
)

var proxies []*url.URL

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
