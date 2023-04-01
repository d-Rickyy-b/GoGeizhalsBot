package database

import (
	"log"
	"sync"

	"github.com/d-Rickyy-b/gogeizhalsbot/v2/internal/bot/models"
)

var (
	userCache  = make(map[int64]models.User)
	cacheMutex sync.Mutex
)

func GetUserFromCache(userID int64) models.User {
	if cachedUser, ok := userCache[userID]; ok {
		return cachedUser
	}

	user := models.User{ID: userID}
	db.Create(user)

	return user
}

// PopulateCaches populates all the caches from the database
func PopulateCaches() {
	// Currently, we only cache users
	populateUserCache()
}

// populateUserCache loads all users from the database into the cache
func populateUserCache() {
	var users []models.User
	db.Find(&users)

	for _, user := range users {
		userCache[user.ID] = user
	}
}

// CreateUserWithCache checks if a user already exists in cache/db and creates a user in the database and cache if it does not exist
func CreateUserWithCache(user models.User) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
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

func DeleteUserWithCache(userID int64) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	DeleteUser(userID)
	delete(userCache, userID)
}
