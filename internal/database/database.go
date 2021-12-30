package database

import (
	"GoGeizhalsBot/internal/bot/models"
	"GoGeizhalsBot/internal/geizhals"
	"fmt"
	"log"

	"gorm.io/gorm/logger"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func InitDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		panic(fmt.Errorf("failed to connect database: %w", err))
	}
	db.Raw("PRAGMA foreign_keys = ON;")

	// Migrate the schema
	migrateError := db.AutoMigrate(&models.User{}, &models.NotificationSettings{}, &models.PriceAgent{},
		&geizhals.Entity{}, &models.HistoricPrice{})
	if migrateError != nil {
		panic("failed to migrate database")
	}

	log.Println("All fine")
}

func CreatePriceAgentForUser(priceAgent *models.PriceAgent) error {
	log.Println("Add priceagent to database!")

	if priceAgent.UserID == 0 {
		return fmt.Errorf("UserID mustn't be 0")
	}

	tx := db.Create(priceAgent)
	if tx.Error != nil {
		log.Println(tx.Error)
		return tx.Error
	}
	return nil
}

func CreateUser(user models.User) error {
	tx := db.Create(&user)
	return tx.Error
}

func DeletePriceAgentForUser(priceAgent models.PriceAgent) error {
	log.Println("Delete priceagent!")

	if priceAgent.UserID == 0 {
		return fmt.Errorf("UserID mustn't be 0")
	}

	// First remove notification settings, then delete priceagent
	// TODO this should be done in a transaction
	tx := db.Model(&models.NotificationSettings{}).Delete(priceAgent.NotificationSettings)
	if tx.Error != nil {
		log.Println(tx.Error)
		return tx.Error
	}

	tx = db.Delete(&priceAgent)
	if tx.Error != nil {
		log.Println(tx.Error)
		return tx.Error
	}
	return nil
}

func GetProductPriceagentsForUser(userID int64) ([]models.PriceAgent, error) {
	var priceagents []models.PriceAgent
	var query = &models.PriceAgent{UserID: userID}

	// tx := db.Debug().Joins("JOIN entities on price_agents.entity_id = entities.id").Where(query).Where("entities.type = 1").Find(&priceagents)
	// tx := db.Debug().Model(&models.PriceAgent{}).Where(query).Joins("JOIN entities on price_agents.entity_id = entities.id").Where(&geizhals.Entity{Type: geizhals.Product}).Find(&priceagents)
	tx := db.Joins("JOIN entities on price_agents.entity_id = entities.id").Where(query).Where("entities.type = ?", geizhals.Product).Find(&priceagents)
	if tx.Error != nil {
		log.Println(tx.Error)
		return []models.PriceAgent{}, tx.Error
	}

	return priceagents, nil
}

func GetWishlistPriceagentsForUser(userID int64) ([]models.PriceAgent, error) {
	var priceagents []models.PriceAgent

	var query = &models.PriceAgent{UserID: userID}
	tx := db.Joins("JOIN entities on price_agents.entity_id = entities.id").Where(query).Where("entities.type = ?", geizhals.Wishlist).Find(&priceagents)
	if tx.Error != nil {
		log.Println(tx.Error)
		return []models.PriceAgent{}, tx.Error
	}

	return priceagents, nil
}

func GetPriceagentForUserByID(userID int64, priceagentID int64) (models.PriceAgent, error) {
	var priceagent models.PriceAgent
	tx := db.Preload("Entity").Preload("NotificationSettings").Where("user_id = ?", userID).Where("id = ?", priceagentID).First(&priceagent)
	if tx.Error != nil {
		log.Println(tx.Error)
		return models.PriceAgent{}, tx.Error
	}
	return priceagent, nil
}

func UpdateNotificationSettings(userID int64, priceagentID int64, notifSettings models.NotificationSettings) error {
	var priceagent models.PriceAgent
	tx := db.Preload("NotificationSettings").Where("user_id = ?", userID).Where("id = ?", priceagentID).First(&priceagent)
	if tx.Error != nil {
		log.Println(tx.Error)
		return tx.Error
	}

	var notifSettingsMap = map[string]interface{}{
		"notify_always":     notifSettings.NotifyAlways,
		"notify_above":      notifSettings.NotifyAbove,
		"notify_below":      notifSettings.NotifyBelow,
		"notify_price_rise": notifSettings.NotifyPriceRise,
		"notify_price_drop": notifSettings.NotifyPriceDrop,
		"above_price":       notifSettings.AbovePrice,
		"below_price":       notifSettings.BelowPrice,
	}
	notifSettings.ID = priceagent.NotificationSettings.ID
	tx = db.Model(&models.NotificationSettings{}).Where("id = ?", notifSettings.ID).Updates(notifSettingsMap)
	if tx.Error != nil {
		log.Println(tx.Error)
		return tx.Error
	}
	return nil
}

func GetAllEntities() ([]geizhals.Entity, error) {
	var entities []geizhals.Entity
	tx := db.Find(&entities)
	if tx.Error != nil {
		log.Println(tx.Error)
		return []geizhals.Entity{}, tx.Error
	}
	return entities, nil
}

// GetPriceAgentsForEntity returns all priceagents for a given entity
func GetPriceAgentsForEntity(entityID int64) ([]models.PriceAgent, error) {
	var priceagents []models.PriceAgent
	tx := db.Preload("Entity").Preload("User").Preload("NotificationSettings").Where("entity_id = ?", entityID).Find(&priceagents)
	if tx.Error != nil {
		log.Println(tx.Error)
		return []models.PriceAgent{}, tx.Error
	}
	return priceagents, nil
}

func UpdateEntity(entity geizhals.Entity) {
	tx := db.Model(&geizhals.Entity{}).Where("id = ?", entity.ID).Updates(entity)
	if tx.Error != nil {
		log.Println(tx.Error)
		return
	}
}

func AddHistoricPrice(price models.HistoricPrice) error {
	// Only add price if the last price is different
	var lastHistoricPrice models.HistoricPrice
	lookupTx := db.Model(&models.HistoricPrice{}).Where("entity_id = ?", price.EntityID).Order("created_at desc").First(&lastHistoricPrice)
	if lookupTx.Error != nil {
		// For the first time, there is no entry in the database, so First will return an error
		log.Println(lookupTx.Error)
	}
	if lastHistoricPrice.Price == price.Price {
		log.Println("Price is the same as last price, not adding")
		return nil
	}

	// If prices differ, add it to the database
	tx := db.Create(&price)
	if tx.Error != nil {
		log.Println(tx.Error)
		return tx.Error
	}
	return nil
}
