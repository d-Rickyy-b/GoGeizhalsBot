package bot

import (
	"fmt"
	"log"
	"time"

	"github.com/d-Rickyy-b/gogeizhalsbot/internal/bot/models"
	"github.com/d-Rickyy-b/gogeizhalsbot/internal/database"
	"github.com/d-Rickyy-b/gogeizhalsbot/internal/geizhals"
	"github.com/d-Rickyy-b/gogeizhalsbot/internal/prometheus"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type tempPriceStore struct {
	store map[int64]map[string]float64
}

func (t *tempPriceStore) getPrice(entityID int64, location string) (float64, bool) {
	entity, entityOk := t.store[entityID]
	if !entityOk {
		return 0, false
	}

	price, locationOk := entity[location]
	if !locationOk {
		return 0, false
	}
	return price, true
}

func (t *tempPriceStore) storePrice(entityID int64, location string, price float64) {
	if t.store == nil {
		t.store = make(map[int64]map[string]float64)
	}
	entity, entityOk := t.store[entityID]
	if !entityOk {
		entity = make(map[string]float64)
	}

	entity[location] = price
	t.store[entityID] = entity
}

// updateEntityPrices fetches the current price of all entities and updates the database
func updateEntityPrices() {
	allPriceAgents, fetchErr := database.GetActivePriceAgents()
	if fetchErr != nil {
		log.Println("Error fetching price agents:", fetchErr)
		return
	}

	priceStore := tempPriceStore{}

	for _, priceAgent := range allPriceAgents {
		log.Printf("Updating prices for price agent: %d, '%s'\n", priceAgent.ID, priceAgent.EntityURL())
		var price float64
		var isCached bool
		price, isCached = priceStore.getPrice(priceAgent.EntityID, priceAgent.Location)
		if !isCached {
			updatedPrice, updateErr := geizhals.UpdateEntityPrice(priceAgent.Entity, priceAgent.Location)
			if updateErr != nil {
				log.Println("Error updating entity:", updateErr)
				continue
			}

			if updatedPrice.Price == priceAgent.CurrentPrice() {
				log.Println("Entity price has not changed, skipping update")
				continue
			}

			database.UpdateEntityPrice(updatedPrice)
			price = updatedPrice.Price
		}

		notifyUsers(priceAgent, priceAgent.CurrentPrice(), price)
	}
}

// notifyUsers sends a notification to the users of the price agent if the settings allow it
func notifyUsers(priceAgent models.PriceAgent, oldPrice, updatedPrice float64) {
	settings := priceAgent.NotificationSettings
	diff := updatedPrice - oldPrice

	var change string
	if updatedPrice > oldPrice {
		change = fmt.Sprintf("📈 %s teurer", bold(createPrice(diff, priceAgent.GetCurrency().String())))
	} else {
		change = fmt.Sprintf("📉 %s günstiger", bold(createPrice(diff, priceAgent.GetCurrency().String())))
	}

	var notificationText string
	entityLink := createLink(priceAgent.EntityURL(), priceAgent.Entity.Name)
	entityPrice := bold(createPrice(updatedPrice, priceAgent.GetCurrency().String()))
	if settings.NotifyAlways {
		notificationText = fmt.Sprintf("Der Preis von %s hat sich geändert: %s\n\n%s", entityLink, entityPrice, change)
	} else if settings.NotifyBelow && updatedPrice < settings.BelowPrice {
		notificationText = fmt.Sprintf("Der Preis von %s hat sich geändert: %s\n\n%s", entityLink, entityPrice, change)
	} else {
		log.Println("Price changes don't match the notification settings for user")
		return
	}

	log.Println("Sending notification to user:", priceAgent.UserID)
	prometheus.PriceagentNotifications.Inc()

	markup := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "Zum Preisagenten!", CallbackData: fmt.Sprintf("m03_02_%d", priceAgent.ID)},
			},
		},
	}
	// TODO implement message queueing to avoid hitting telegram api limits (30 msgs/sec)
	_, sendErr := bot.SendMessage(priceAgent.UserID, notificationText, &gotgbot.SendMessageOpts{ParseMode: "HTML", DisableWebPagePreview: true, ReplyMarkup: markup})
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
