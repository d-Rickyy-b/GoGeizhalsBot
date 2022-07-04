package config

import (
	"io/ioutil"
	"log"
	"regexp"

	"gopkg.in/yaml.v3"
)

type Config struct {
	BotToken      string `yaml:"bot_token"`
	LangDirectory string `yaml:"lang_path"`
	Webhook       struct {
		Enabled     bool   `yaml:"enabled"`
		ListenIP    string `yaml:"listen_ip"`
		ListenPort  int    `yaml:"listen_port"`
		ListenPath  string `yaml:"listen_path"`
		URL         string `yaml:"url"`
		CertPath    string `yaml:"cert_path"`
		CertKeyPath string `yaml:"cert_key_path"`
	} `yaml:"webhook"`
	Proxy struct {
		Enabled       bool   `yaml:"enabled"`
		ProxyListPath string `yaml:"proxy_list_path"`
	} `yaml:"proxy"`
	LogDirectory string `yaml:"log_directory"`
	Prometheus   struct {
		Enabled    bool   `yaml:"enabled"`
		ExportIP   string `yaml:"export_ip"`
		ExportPort int    `yaml:"export_port"`
	} `yaml:"prometheus"`
}

func ReadConfig(configFile string) (Config, error) {
	conf, parseErr := parseConfigFromFile(configFile)
	if parseErr != nil {
		log.Fatalln("Error while parsing yaml file:", parseErr)
	}
	if !validateConfig(conf) {
		log.Fatalln("Invalid config")
	}

	return conf, nil
}

func parseConfigFromFile(configFile string) (Config, error) {
	yamlFileContent, readErr := ioutil.ReadFile(configFile)
	if readErr != nil {
		return Config{}, readErr
	}

	conf, parseErr := parseConfigFromBytes(yamlFileContent)
	if parseErr != nil {
		return Config{}, parseErr
	}
	return conf, nil
}

func parseConfigFromBytes(data []byte) (Config, error) {
	var config Config

	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

func validateConfig(config Config) bool {
	urlRegex := regexp.MustCompile(`https?://.+`)
	// Still matches invalid IP addresses but good enough for detecting completely wrong formats
	IPRegex := regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)

	// Check webhook config
	if config.Webhook.Enabled {
		if config.Webhook.URL == "" || !urlRegex.MatchString(config.Webhook.URL) {
			log.Fatalln("Webhook URL does not match pattern 'http(s)://<hostname>/path'")
			return false
		}
		if config.Webhook.ListenIP == "" || !IPRegex.MatchString(config.Webhook.ListenIP) {
			log.Fatalln("Webhook listen IP is does not match pattern 'x.x.x.x'")
			return false
		}
		if config.Webhook.ListenPort == 0 {
			log.Fatalln("Webhook listen port is not set")
			return false
		}
	}

	if config.BotToken == "" {
		log.Fatalln("Bot token is not set")
		return false
	}
	if config.Proxy.Enabled {
		if config.Proxy.ProxyListPath == "" {
			log.Fatalln("Proxy list path is not set")
			return false
		}
	}

	if config.Prometheus.Enabled {
		if config.Prometheus.ExportIP == "" || !IPRegex.MatchString(config.Prometheus.ExportIP) {
			log.Fatalln("Prometheus export IP does not match pattern 'x.x.x.x'")
			return false
		}
		if config.Prometheus.ExportPort == 0 {
			log.Fatalln("Prometheus export port is not set")
			return false
		}
	}
	return true
}
