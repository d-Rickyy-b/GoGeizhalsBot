package config

import (
	"bufio"
	"log"
	"net/url"
	"os"
	"strings"
)

// LoadProxies tries to load the list of proxies from the given file.
// If it fails, it returns an empty list.
func LoadProxies(filename string) []*url.URL {
	var proxyList []*url.URL

	file, err := os.Open(filename)
	if err != nil {
		log.Println(err)
		return proxyList
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		line := scanner.Text()
		// Ignore comments
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		parsedProxy, parseErr := url.Parse(line)
		if parseErr != nil {
			log.Printf("Line %s is not a valid proxy url: %s\n", line, parseErr)
			continue
		}

		proxyList = append(proxyList, parsedProxy)
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}

	return proxyList
}
