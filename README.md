# GoGeizhalsBot
[![build](https://github.com/d-Rickyy-b/GoGeizhalsBot/actions/workflows/release_build.yml/badge.svg)](https://github.com/d-Rickyy-b/GoGeizhalsBot/actions/workflows/release_build.yml)
[![Docker Image Version (latest semver)](https://img.shields.io/docker/v/0rickyy0/gogeizhalsbot?label=docker&sort=semver)](https://hub.docker.com/repository/docker/0rickyy0/gogeizhalsbot)
[![Go Reference](https://pkg.go.dev/badge/github.com/d-Rickyy-b/gogeizhalsbot.svg)](https://pkg.go.dev/github.com/d-Rickyy-b/gogeizhalsbot)

This repository holds a rewrite of my [Python-GeizhalsBot](https://raw.githubusercontent.com/d-Rickyy-b/Python-GeizhalsBot) Telegram bot in the Go programming language. 
GoGeizhalsbot is a Telegram bot that notifies you via chat messages about changes of the price of a [geizhals.de](https://geizhals.de) product or wishlist.

![chat examples](https://raw.githubusercontent.com/d-Rickyy-b/GoGeizhalsBot/master/docs/example.png)

## Configuration
The software searches for a config.yml file in the current working directory.
Check [config.sample.yml](https://raw.githubusercontent.com/d-Rickyy-b/GoGeizhalsBot/master/config.sample.yml) for an example.
This section gives a quick overview of the possible configuration items.

### Config root
On the root level of the config file you can set the bot token.

| Field     | Type   | Function                        |
|-----------|--------|---------------------------------|
| bot_token | string | The bot token to run the bot on |

### Webhook config
Long polling isn't all too bad, but using webhooks cuts out the need of constantly contacting the Telegram server for new updates.
The `webhook` key of the config allows you to set up a webhook for the bot.

| Field         | Type   | Function                                                               |
|---------------|--------|------------------------------------------------------------------------|
| enabled       | bool   | Specifies if the bot should use webhooks                               |
| listen_ip     | string | The IP for the webhook to listen on                                    |
| listen_port   | int    | The port for the webhook to listen on                                  |
| listen_path   | string | The path to listen on (e.g. "/" or "/bot")                             |
| url           | string | The publicly reachable url which Telegram calls (passed in SetWebhook) |
| cert_path     | string | Path to a certificate file to use                                      |
| cert_key_path | string | Path to the certificate key file                                       |

### Proxy config
Contacting the geizhals.de website regularly will get your IP address banned in no time.
Therefore, it is recommended to use proxy servers to circumvent this. Use the `proxy` key to configure your proxy servers.
 
| Field           | Type   | Function                                                                 |
|-----------------|--------|--------------------------------------------------------------------------|
| enabled         | bool   | Specifies if the bot should use proxies for the connection to Geizhals   |
| proxy_list_path | string | Path to a file that contains a newline separated list of proxy addresses |

### Prometheus config
Monitoring your services is always a good idea. 
Prometheus is a time series database that allows you to collect metrics over time and render them in cool graphs e.g. with tools like Grafana.
This configuration can be found on the `prometheus` key.

| Field       | Type   | Function                                         |
|-------------|--------|--------------------------------------------------|
| enabled     | bool   | Specifies if the prometheus interface is active  |
| export_ip   | string | The IP adress to run the export http server on   |
| export_port | int    | The port number to run the export http server on |