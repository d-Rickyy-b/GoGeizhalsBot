package models

import (
	"errors"
	"strconv"
	"strings"
)

type Menu struct {
	ID           string
	SubMenu      string
	PriceAgentID int64
	Extra        string
}

// NewMenu returns a menu struct from a given menuData string. The menu follows the format <menuID>_<submenuID>_<priceagentID>[_<extraData>]
func NewMenu(menuData string) (*Menu, error) {
	if len(menuData) > 64 {
		return nil, errors.New("length of menu data exceeding 64 bytes")
	}
	components := strings.Split(menuData, "_")
	if len(components) != 3 && len(components) != 4 {
		return nil, errors.New("unable to parse menu data")
	}

	priceagentID, parseErr := strconv.Atoi(components[2])
	if parseErr != nil {
		return nil, parseErr
	}

	menu := &Menu{
		ID:           components[0],
		SubMenu:      components[1],
		PriceAgentID: int64(priceagentID),
	}

	if len(components) == 4 {
		menu.Extra = components[3]
	}
	return menu, nil
}
