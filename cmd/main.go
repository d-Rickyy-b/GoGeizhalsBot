package main

import (
	"GoGeizhalsBot/internal/bot"
	"GoGeizhalsBot/internal/config"
	"GoGeizhalsBot/internal/database"
	"GoGeizhalsBot/internal/geizhals"
	"flag"
	"log"
	"net/url"
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

	var proxies []*url.URL
	if botConfig.Proxy.Enabled {
		proxies = config.LoadProxies(botConfig.Proxy.ProxyListPath)
		log.Println("Loaded proxies:", len(proxies))
	}

	geizhals.InitProxies(proxies)
	bot.Start(botConfig)
}
