package main

import (
	"GoGeizhalsBot/internal/bot"
	"GoGeizhalsBot/internal/config"
	"GoGeizhalsBot/internal/database"
	"GoGeizhalsBot/internal/geizhals"
	"log"
	"time"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	go bot.UpdatePricesJob(time.Minute * 2)

	database.InitDB()
	database.PopulateCaches()

	proxies := config.LoadProxies("proxies.txt")
	log.Println("Loaded proxies:", len(proxies))

	geizhals.InitProxies(proxies)
	bot.Start()
}
