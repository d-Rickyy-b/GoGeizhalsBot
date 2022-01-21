package geizhals

import (
	"math/rand"
	"net/url"
)

var proxies []*url.URL

// InitProxies initializes the proxy list.
func InitProxies(p []*url.URL) {
	// Randomize order of array
	for i := range p {
		j := rand.Intn(i + 1) //nolint:gosec
		p[i], p[j] = p[j], p[i]
	}

	proxies = p
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
