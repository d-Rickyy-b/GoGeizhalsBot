package database

import (
	"GoGeizhalsBot/internal/bot/models"
	"log"
)

var userCache = make(map[int64]models.User)

func GetUserFromCache(userID int64) models.User {
	if cachedUser, ok := userCache[userID]; ok {
		return cachedUser
	}

	user := models.User{ID: userID}
	db.Create(user)
	return user
}

func PopulateCaches() {
	populateUserCache()
}

func populateUserCache() {
	var users []models.User
	db.Find(&users)

	for _, user := range users {
		userCache[user.ID] = user
	}
}

func CreateUserWithCache(user models.User) {
	if _, ok := userCache[user.ID]; ok {
		// User already exists in cache
		return
	}

	// Create user in database and add user to cache
	createErr := CreateUser(user)
	if createErr != nil {
		log.Println(createErr)
	}
	userCache[user.ID] = user
}
