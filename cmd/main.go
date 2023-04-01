package main

import (
	"GoGeizhalsBot/internal/bot"
	"GoGeizhalsBot/internal/config"
	"GoGeizhalsBot/internal/database"
	"GoGeizhalsBot/internal/logging"
	"GoGeizhalsBot/internal/proxy"
	"flag"
	"log"
	"net/url"
	"time"
)

func main() {
	var configFile = flag.String("config", "config.yml", "Path to config file")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	database.InitDB()
	database.PopulateCaches()

	botConfig, readConfigErr := config.ReadConfig(*configFile)
	if readConfigErr != nil {
		log.Fatal(readConfigErr)
	}

	logging.SetupLogging(botConfig.LogDirectory)

	var proxies []*url.URL
	if botConfig.Proxy.Enabled {
		proxies = config.LoadProxies(botConfig.Proxy.ProxyListPath)
		log.Println("Loaded proxies:", len(proxies))
	}

	updateInterval := time.Duration(botConfig.UpdateIntervalMinutes) * time.Minute
	go bot.UpdatePricesJob(updateInterval)

	proxy.InitProxies(proxies)
	bot.Start(botConfig)
}
