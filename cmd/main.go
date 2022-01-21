package main

import (
	"GoGeizhalsBot/internal/bot"
	"GoGeizhalsBot/internal/config"
	"GoGeizhalsBot/internal/database"
	"GoGeizhalsBot/internal/geizhals"
	"flag"
	"log"
	"time"
)

func main() {
	var configFile = flag.String("config", "config.yml", "Path to config file")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	go bot.UpdatePricesJob(time.Minute * 10)

	database.InitDB()
	database.PopulateCaches()

	botConfig, readConfigErr := config.ReadConfig(*configFile)
	if readConfigErr != nil {
		log.Fatal(readConfigErr)
	}
	log.Println(botConfig)
	proxies := config.LoadProxies("proxies.txt")
	log.Println("Loaded proxies:", len(proxies))

	geizhals.InitProxies(proxies)
	bot.Start(botConfig)
}
