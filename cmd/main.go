package main

import (
	"flag"
	"log"
	"net/url"
	"time"

	"github.com/d-Rickyy-b/gogeizhalsbot/internal/bot"
	"github.com/d-Rickyy-b/gogeizhalsbot/internal/config"
	"github.com/d-Rickyy-b/gogeizhalsbot/internal/database"
	"github.com/d-Rickyy-b/gogeizhalsbot/internal/logging"
	"github.com/d-Rickyy-b/gogeizhalsbot/internal/proxy"
)

func main() {
	configFile := flag.String("config", "config.yml", "Path to config file")
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
