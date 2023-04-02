package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"os/user"
	"time"

	"github.com/d-Rickyy-b/gogeizhalsbot/v2/internal/bot"
	"github.com/d-Rickyy-b/gogeizhalsbot/v2/internal/config"
	"github.com/d-Rickyy-b/gogeizhalsbot/v2/internal/database"
	"github.com/d-Rickyy-b/gogeizhalsbot/v2/internal/logging"
	"github.com/d-Rickyy-b/gogeizhalsbot/v2/internal/proxy"
)

func main() {
	configFile := flag.String("config", "config.yml", "Path to config file")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	printInfo()

	botConfig, readConfigErr := config.ReadConfig(*configFile)
	if readConfigErr != nil {
		log.Fatal(readConfigErr)
	}

	database.InitDB()
	database.PopulateCaches()

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

func printInfo() {
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	log.Println("Working directory: ", path)

	systemUser, err := user.Current()
	if err != nil {
		log.Fatalf(err.Error())
	}

	log.Printf("Username: %s\n", systemUser.Username)
}
