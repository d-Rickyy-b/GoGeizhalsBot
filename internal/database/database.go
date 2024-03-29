package database

import (
	"fmt"
	"log"

	"github.com/d-Rickyy-b/gogeizhalsbot/v2/internal/bot/models"
	"github.com/d-Rickyy-b/gogeizhalsbot/v2/internal/geizhals"

	"gorm.io/gorm/logger"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func InitDB() {
	var err error

	db, err = gorm.Open(sqlite.Open("users.db"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		panic(fmt.Errorf("failed to connect database: %w", err))
	}

	db.Raw("PRAGMA foreign_keys = ON;")

	// Migrate the schema
	migrateError := db.AutoMigrate(&models.User{}, &models.NotificationSettings{}, &models.PriceAgent{},
		&geizhals.Entity{}, &geizhals.EntityPrice{})
	if migrateError != nil {
		log.Println("Couldn't migrate database!", migrateError.Error())
		panic("failed to migrate database")
	}

	deleteError := DeleteDisabledPriceagents()
	if deleteError != nil {
		log.Println("Couldn't delete pending price agents!", deleteError.Error())
	}

	log.Println("Database init complete")
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

func GetPriceAgentCountForUser(userID int64) int64 {
	var count int64
	db.Model(&models.PriceAgent{}).Where("user_id = ?", userID).Count(&count)

	return count
}

func GetPriceAgentCount() int64 {
	var count int64
	db.Model(&models.PriceAgent{}).Count(&count)

	return count
}

func GetPriceAgentProductCount() int64 {
	var count int64
	db.Model(&models.PriceAgent{}).Joins("JOIN entities on price_agents.entity_id = entities.id").Where("entities.type = ?", geizhals.Product).Count(&count)

	return count
}

func GetPriceAgentWishlistCount() int64 {
	var count int64
	db.Model(&models.PriceAgent{}).Joins("JOIN entities on price_agents.entity_id = entities.id").Where("entities.type = ?", geizhals.Wishlist).Count(&count)

	return count
}

func GetUserCount() int64 {
	var count int64
	db.Model(&models.User{}).Count(&count)

	return count
}

func CreateUser(user models.User) error {
	tx := db.Create(&user)
	return tx.Error
}

func GetDarkmode(userID int64) bool {
	var user models.User

	tx := db.Where("id = ?", userID).First(&user)
	if tx.Error != nil {
		log.Println(tx.Error)
		return true
	}

	return user.DarkMode
}

func UpdateDarkMode(userID int64, darkMode bool) {
	tx := db.Model(&models.User{}).Where("id = ?", userID).Update("dark_mode", darkMode)
	if tx.Error != nil {
		log.Println(tx.Error)
	}
}

func GetAllUsers() []models.User {
	var users []models.User
	db.Find(&users)

	return users
}

func GetActivePriceAgents() ([]models.PriceAgent, error) {
	var priceagents []models.PriceAgent

	tx := db.Preload("Entity").Preload("Entity.Prices").Preload("NotificationSettings").Preload("User").Where("enabled = true").Find(&priceagents)
	if tx.Error != nil {
		log.Println(tx.Error)
		return []models.PriceAgent{}, tx.Error
	}

	return priceagents, nil
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
	query := &models.PriceAgent{UserID: userID, Enabled: true}

	tx := db.Joins("JOIN entities on price_agents.entity_id = entities.id").Where(query).Where("entities.type = ?", geizhals.Product).Find(&priceagents)
	if tx.Error != nil {
		log.Println(tx.Error)
		return []models.PriceAgent{}, tx.Error
	}

	return priceagents, nil
}

func GetWishlistPriceagentsForUser(userID int64) ([]models.PriceAgent, error) {
	var priceagents []models.PriceAgent
	query := &models.PriceAgent{UserID: userID, Enabled: true}

	tx := db.Joins("JOIN entities on price_agents.entity_id = entities.id").Where(query).Where("entities.type = ?", geizhals.Wishlist).Find(&priceagents)
	if tx.Error != nil {
		log.Println(tx.Error)
		return []models.PriceAgent{}, tx.Error
	}

	return priceagents, nil
}

func GetPriceagentForUserByID(userID int64, priceagentID int64) (models.PriceAgent, error) {
	var priceagent models.PriceAgent
	tx := db.Preload("Entity").Preload("Entity.Prices").Preload("NotificationSettings").Where("user_id = ?", userID).Where("id = ?", priceagentID).First(&priceagent)
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

	notifSettingsMap := map[string]interface{}{
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

func UpdatePriceagent(priceagent models.PriceAgent) error {
	tx := db.Model(&priceagent).Updates(priceagent)
	if tx.Error != nil {
		log.Println(tx.Error)
		return tx.Error
	}

	return nil
}

func DeleteDisabledPriceagents() error {
	tx := db.Where("enabled = 0").Delete(&models.PriceAgent{})
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

func GetAllEntitiesWithPriceagents() ([]geizhals.Entity, error) {
	var entities []geizhals.Entity
	tx := db.Model(&geizhals.Entity{}).Joins("JOIN price_agents on price_agents.entity_id = entities.id").
		// Joins("JOIN entity_prices on entity_prices.entity_id = entities.id").
		Where("price_agents.enabled = 1").
		// Where("price_agents.location = entity_prices.location").
		Find(&entities)

	if tx.Error != nil {
		log.Println(tx.Error)
		return []geizhals.Entity{}, tx.Error
	}

	return entities, nil
}

// HasUserPriceAgentForEntity checks if a user already has a priceagent for a given entity
func HasUserPriceAgentForEntity(userID int64, entityID int64) (bool, error) {
	var priceagent models.PriceAgent
	query := &models.PriceAgent{UserID: userID, EntityID: entityID, Enabled: true}

	tx := db.Where(query).Limit(1).Find(&priceagent)
	if tx.Error != nil {
		log.Println(tx.Error)
		return false, tx.Error
	}

	exists := tx.RowsAffected > 0

	return exists, nil
}

// GetPriceAgentsForEntity returns all priceagents for a given entity
func GetPriceAgentsForEntity(entityID int64) ([]models.PriceAgent, error) {
	var priceagents []models.PriceAgent

	tx := db.Preload("Entity").Preload("User").Preload("NotificationSettings").Where("entity_id = ?", entityID).Where("enabled = 1").Find(&priceagents)
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

func UpdateEntityPrice(price geizhals.EntityPrice) {
	tx := db.Model(&geizhals.EntityPrice{}).Where("entity_id = ?", price.EntityID).Where("location = ?", price.Location).Updates(price)
	if tx.RowsAffected == 0 {
		tx = db.Create(&price)
	}

	if tx.Error != nil {
		log.Println(tx.Error)
		return
	}
}

// DeleteUser deletes a user and their PriceAgents from the database
func DeleteUser(userID int64) {
	// Start a new transaction
	_ = db.Transaction(func(tx *gorm.DB) error {
		// delete all the things
		var notifSettings []models.NotificationSettings
		if err := tx.Model(&models.NotificationSettings{}).Joins("JOIN price_agents on price_agents.notification_id = notification_settings.id").Where("price_agents.user_id = ?", userID).Find(&notifSettings); err.Error != nil {
			// returning any error will roll back
			return err.Error
		}

		if len(notifSettings) > 0 {
			if err := tx.Delete(notifSettings); err.Error != nil {
				// returning any error will roll back
				return err.Error
			}
		}

		if err := tx.Model(&models.PriceAgent{}).Where("user_id = ?", userID).Delete(&models.PriceAgent{}); err.Error != nil {
			// returning any error will roll back
			return err.Error
		}
		if err := tx.Model(&models.User{}).Where("id = ?", userID).Delete(&models.User{}); err.Error != nil {
			// returning any error will roll back
			return err.Error
		}

		// return nil will commit the whole transaction
		return nil
	})
}
