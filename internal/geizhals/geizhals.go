package geizhals

import (
	"errors"
	"regexp"
)

var (
	wishlistURLPattern = regexp.MustCompile(`^((?:https?://)?(?:geizhals\.(?:de|at)|cenowarka\.pl|skinflint\.co\.uk)/?(\?cat=WL-|wishlists/)(\d+)).*$`)
	productURLPattern  = regexp.MustCompile(`^((?:https?://)?(?:geizhals\.(?:de|at)|cenowarka\.pl|skinflint\.co\.uk)/([0-9a-zA-Z\-]*a(\d+).html))\??.*$`)
)

var (
	locationPattern       = regexp.MustCompile(`hloc=(de|at|uk|pl)`)
	locationDomainPattern = regexp.MustCompile(`(?:geizhals\.(de|at)|cenowarka\.(pl)|skinflint\.co\.(uk))`)
	geizhalsDomains       = map[string]string{
		"de": "geizhals.de",
		"at": "geizhals.at",
		"eu": "geizhals.eu",
		"uk": "skinflint.co.uk",
		"pl": "cenowarka.pl",
	}
)

var ErrTooManyRetries = errors.New("too many retries")
var ErrInvalidURL = errors.New("invalid URL")

// UpdateEntityPrice returns an updated EntityPrice struct from a given input Entity
func UpdateEntityPrice(entity Entity, location string) (EntityPrice, error) {
	updatedEntity, downloadErr := DownloadEntity(entity.FullURL(location))
	if len(updatedEntity.Prices) > 0 {
		return updatedEntity.Prices[0], downloadErr
	}

	return EntityPrice{
		EntityID: entity.ID,
		Location: location,
		Price:    0,
	}, downloadErr
}
