package database

import (
	"GoGeizhalsBot/internal/bot/models"
	"GoGeizhalsBot/internal/geizhals"
	"fmt"
	"log"
)

var globalPriceagents = make(map[int64][]models.PriceAgent)

func CreatePriceAgentForUser(priceAgent models.PriceAgent, userID int64) {
	log.Println("Add priceagent!")
	if priceagents, ok := globalPriceagents[userID]; ok {
		globalPriceagents[userID] = append(priceagents, priceAgent)
		return
	}
	globalPriceagents[userID] = []models.PriceAgent{priceAgent}
}

func DeletePriceAgentForUser(priceAgent models.PriceAgent, userID int64) {
	log.Println("Delete priceagent!")
	if priceagents, ok := globalPriceagents[userID]; ok {
		for i, priceagent := range priceagents {
			if priceagent.ID == priceAgent.ID {
				globalPriceagents[userID] = append(priceagents[:i], priceagents[i+1:]...)
				return
			}
		}
	}
}

func ReplacePriceAgentForUser(priceAgent models.PriceAgent, userID int64) {
	log.Println("Replace priceagent!")
	if priceagents, ok := globalPriceagents[userID]; ok {
		for i, priceagent := range priceagents {
			if priceagent.ID == priceAgent.ID {
				priceagents[i] = priceAgent
				globalPriceagents[userID] = priceagents
				return
			}
		}
	}
}

func GetPriceAgentsForUser(userID int64) ([]models.PriceAgent, error) {
	if priceagents, ok := globalPriceagents[userID]; ok {
		return priceagents, nil
	}
	return []models.PriceAgent{}, fmt.Errorf("no priceagents found for user")
}

func GetProductPriceagentsForUser(userID int64) ([]models.PriceAgent, error) {
	pa, err := GetPriceAgentsForUser(userID)
	if err != nil {
		log.Println(err)
		return []models.PriceAgent{}, err
	}

	var productPriceagents []models.PriceAgent
	for _, pa := range pa {
		if pa.Entity.Type == geizhals.Product {
			productPriceagents = append(productPriceagents, pa)
		}
	}
	return productPriceagents, nil
}

func GetWishlistPriceagentsForUser(userID int64) ([]models.PriceAgent, error) {
	pa, err := GetPriceAgentsForUser(userID)
	if err != nil {
		log.Println(err)
		return []models.PriceAgent{}, err
	}

	var wishlistPriceagents []models.PriceAgent
	for _, pa := range pa {
		if pa.Entity.Type == geizhals.Wishlist {
			wishlistPriceagents = append(wishlistPriceagents, pa)
		}
	}
	return wishlistPriceagents, nil
}

func GetPriceagentForUserByID(userID int64, priceagentID string) (models.PriceAgent, error) {
	priceagents, err := GetPriceAgentsForUser(userID)
	if err != nil {
		log.Println(err)
		return models.PriceAgent{}, err
	}

	for _, pa := range priceagents {
		if pa.ID == priceagentID {
			return pa, nil
		}
	}
	return models.PriceAgent{}, fmt.Errorf("no priceagent found for user")
}

func UpdateNotificationSettings(userID int64, priceagentID string, notifSettings models.NotificationSettings) error {
	priceagent, err := GetPriceagentForUserByID(userID, priceagentID)
	if err != nil {
		log.Println(err)
		return err
	}

	priceagent.NotificationSettings = notifSettings
	ReplacePriceAgentForUser(priceagent, userID)
	return nil
}
