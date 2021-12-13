package config

import (
	"bufio"
	"log"
	"os"
	"strings"
)

// LoadProxies tries to load the list of proxies from the given file.
// If it fails, it returns an empty list.
func LoadProxies(filename string) []string {
	var proxyList []string
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
		// TODO check for correct syntax
		proxyList = append(proxyList, line)
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}

	return proxyList
}
