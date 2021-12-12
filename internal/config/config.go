package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

type Webhook struct {
	Enabled bool   `json:"enabled"`
	Url     string `json:"url"`
	Listen  string `json:"listen"`
}

type Config struct {
	Token   string  `json:"token"`
	Webhook Webhook `json:"webhook"`
}

func ReadConfig(configFile string) (Config, error) {
	configFile = path.Clean(configFile)
	jsonFile, err := os.Open(configFile)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
		return Config{}, err
	}

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	var config Config
	byteValue, _ := ioutil.ReadAll(jsonFile)

	unmarshalError := json.Unmarshal(byteValue, &config)
	if unmarshalError != nil {
		fmt.Println(unmarshalError)
		return Config{}, unmarshalError
	}

	return config, nil
}
