package geizhals

import (
	"regexp"
)

var wishlistURLPattern = regexp.MustCompile(`^((?:https?://)?(?:geizhals\.(?:de|at)|cenowarka\.pl|skinflint\.co\.uk)/(\?cat=WL(-\d+))).*$`)
var productURLPattern = regexp.MustCompile(`^((?:https?://)?(?:geizhals\.(?:de|at)|cenowarka\.pl|skinflint\.co\.uk)/([0-9a-zA-Z\-]*a(\d+).html))\??.*$`)

var locationPattern = regexp.MustCompile(`hloc=(de|at|uk|pl)`)
var locationDomainPattern = regexp.MustCompile(`(?:geizhals\.(de|at)|cenowarka\.(pl)|skinflint\.co\.(uk))`)
var geizhalsDomains = map[string]string{
	"de": "geizhals.de",
	"at": "geizhals.at",
	"eu": "geizhals.eu",
	"uk": "skinflint.co.uk",
	"pl": "cenowarka.pl",
}

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
