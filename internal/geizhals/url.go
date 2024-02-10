package geizhals

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
)

type EntityURL struct {
	SubmittedURL string
	CleanURL     string
	Path         string
	Location     string
	Locations    []string
	EntityID     int64
	Type         EntityType
}

// parseGeizhalsURL parses the given URL and returns a EntityURL struct.
func parseGeizhalsURL(rawurl string) (EntityURL, error) {
	var matches []string
	var location string
	entityType := Product

	matches = productURLPattern.FindStringSubmatch(rawurl)
	if len(matches) != 4 {
		matches = wishlistURLPattern.FindStringSubmatch(rawurl)
		entityType = Wishlist
	}

	if len(matches) != 4 {
		return EntityURL{}, ErrInvalidURL
	}

	entityIDString := matches[3]
	// Prepend a minus sign to the entityID if it's a wishlist
	if entityType == Wishlist && !strings.HasPrefix(entityIDString, "-") {
		entityIDString = "-" + entityIDString

	}

	entityID, err := strconv.Atoi(entityIDString)
	if err != nil {
		return EntityURL{}, fmt.Errorf("couldn't parse entity ID: %s", entityIDString)
	}

	// Pick location from the domain name
	locationTLDMatches := locationDomainPattern.FindStringSubmatch(rawurl)
	for i, locationTLDMatch := range locationTLDMatches {
		if locationTLDMatch == "" || i == 0 {
			continue
		}
		location = locationTLDMatch

		break
	}

	if location == "" {
		log.Println("No location found for URL:", rawurl)
		return EntityURL{}, errors.New("couldn't parse location")
	}

	ghURL := EntityURL{
		SubmittedURL: rawurl,
		CleanURL:     matches[1],
		Path:         matches[2],
		EntityID:     int64(entityID),
		Location:     location,
		Type:         entityType,
	}

	return ghURL, nil
}

func LocationFromURL(rawurl string) (string, error) {
	locationTLDMatches := locationDomainPattern.FindStringSubmatch(rawurl)
	for i, locationTLDMatch := range locationTLDMatches {
		if locationTLDMatch == "" || i == 0 {
			continue
		}

		return locationTLDMatch, nil
	}

	return "", errors.New("couldn't parse location")
}
