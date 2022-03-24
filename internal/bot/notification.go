package bot

import (
	"GoGeizhalsBot/internal/bot/models"
	"GoGeizhalsBot/internal/database"
	"GoGeizhalsBot/internal/geizhals"
	"GoGeizhalsBot/internal/prometheus"
	"fmt"
	"log"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

// updateEntityPrices fetches the current price of all entities and updates the database
func updateEntityPrices() {
	allEntities, fetchEntitiesErr := database.GetAllEntities()
	if fetchEntitiesErr != nil {
		log.Println("Error fetching entites:", fetchEntitiesErr)
		return
	}

	// Iterate over all price agents.
	// For each price agent, update prices and store updated prices in the entity in the database.
	// Also update price history with the new prices.
	for _, entity := range allEntities {
		log.Println("Updating prices for:", entity.Name)

		// If there are two price agents with the same entity, we currently fetch it twice
		updatedEntity, updateErr := geizhals.UpdateEntity(entity)
		if updateErr != nil {
			log.Println("Error updating entity:", updateErr)
			continue
		}

		if updatedEntity.Price == entity.Price {
			log.Println("Entity price has not changed, skipping update")
			continue
		}

		database.UpdateEntity(updatedEntity)

		// fetch all priceagents for this entity
		priceAgents, fetchPriceAgentsErr := database.GetPriceAgentsForEntity(updatedEntity.ID)
		if fetchPriceAgentsErr != nil {
			log.Println("Error fetching price agents for entity:", fetchPriceAgentsErr)
			continue
		}

		for _, priceAgent := range priceAgents {
			notifyUsers(priceAgent, entity, updatedEntity)
		}
	}
}

// notifyUsers sends a notification to the users of the price agent if the settings allow it
func notifyUsers(priceAgent models.PriceAgent, oldEntity, updatedEntity geizhals.Entity) {
	settings := priceAgent.NotificationSettings
	user := priceAgent.User
	diff := updatedEntity.Price - oldEntity.Price

	var change string
	if updatedEntity.Price > oldEntity.Price {
		change = fmt.Sprintf("📈 %s teurer", bold(createPrice(diff)))
	} else {
		change = fmt.Sprintf("📉 %s günstiger", bold(createPrice(diff)))
	}

	var notificationText string
	entityLink := createLink(updatedEntity.URL, updatedEntity.Name)
	entityPrice := bold(createPrice(updatedEntity.Price))
	if settings.NotifyAlways {
		notificationText = fmt.Sprintf("Der Preis von %s hat sich geändert: %s\n\n%s", entityLink, entityPrice, change)
	} else if settings.NotifyBelow && updatedEntity.Price < settings.BelowPrice {
		notificationText = fmt.Sprintf("Der Preis von %s hat sich geändert: %s\n\n%s", entityLink, entityPrice, change)
	} else if settings.NotifyAbove && updatedEntity.Price > settings.AbovePrice {
		notificationText = "Hi, preis über Grenze!"
	} else if settings.NotifyPriceDrop && updatedEntity.Price < oldEntity.Price {
		notificationText = "Hi, preis gefallen!"
	} else if settings.NotifyPriceRise && updatedEntity.Price > oldEntity.Price {
		notificationText = "Hi, preis gestiegen!"
	} else {
		log.Println("Price changes don't match the notification settings for user")
		return
	}

	log.Println("Sending notification to user:", user.ID)
	prometheus.PriceagentNotifications.Inc()

	markup := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "Zum Preisagenten!", CallbackData: fmt.Sprintf("m03_00_%d", priceAgent.ID)},
			},
		},
	}
	// TODO implement message queueing to avoid hitting telegram api limits (30 msgs/sec)
	_, sendErr := bot.SendMessage(user.ID, notificationText, &gotgbot.SendMessageOpts{ParseMode: "HTML", DisableWebPagePreview: true, ReplyMarkup: markup})
	if sendErr != nil {
		log.Println("Error sending message:", sendErr)
		return
	}
}

// UpdatePricesJob is a job that updates prices of all price agents at a given interval.
func UpdatePricesJob(updateFrequency time.Duration) {
	// Align method execution at certain intervals - e.g. every 5 minutes at :05, :10, :15, etc. similar to cron.
	for {
		sleepDuration := calculateSleep(updateFrequency)
		log.Println("Sleeping for:", sleepDuration)
		time.Sleep(sleepDuration)

		updateEntityPrices()
	}
}

// calculateSleep calculates the duration to sleep before the next update.
func calculateSleep(updateFrequency time.Duration) time.Duration {
	delta := time.Now().Unix() % int64(updateFrequency.Seconds())
	initialDelay := updateFrequency - (time.Second * time.Duration(delta))
	return initialDelay
}
