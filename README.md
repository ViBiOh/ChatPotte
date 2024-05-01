# ChatPotte

[![Build](https://github.com/ViBiOh/ChatPotte/workflows/Build/badge.svg)](https://github.com/ViBiOh/ChatPotte/actions)

## Getting started

Golang binary is built with static link. You can download it directly from the [GitHub Release page](https://github.com/ViBiOh/ChatPotte/releases) or build it by yourself by cloning this repo and running `make`.

A Docker image is available for `amd64`, `arm` and `arm64` platforms on Docker Hub: [vibioh/ChatPotte](https://hub.docker.com/r/vibioh/ChatPotte/tags).

You can configure app by passing CLI args or environment variables (cf. [Usage](#usage) section). CLI override environment variables.

## Usage

The application can be configured by passing CLI args described below or their equivalent as environment variable. CLI values take precedence over environments variables.

Be careful when using the CLI values, if someone list the processes on the system, they will appear in plain-text. Pass secrets by environment variables: it's less easily visible.

```bash
Usage of discord:
  --applicationID     string  [discord] Application ID ${DISCORD_APPLICATION_ID}
  --clientID          string  [discord] Client ID ${DISCORD_CLIENT_ID}
  --clientSecret      string  [discord] Client Secret ${DISCORD_CLIENT_SECRET}
  --commands          string  [commands] Configuration of commands, as JSON string ${DISCORD_COMMANDS}
  --loggerJson                [logger] Log format as JSON ${DISCORD_LOGGER_JSON} (default false)
  --loggerLevel       string  [logger] Logger level ${DISCORD_LOGGER_LEVEL} (default "INFO")
  --loggerLevelKey    string  [logger] Key for level in JSON ${DISCORD_LOGGER_LEVEL_KEY} (default "level")
  --loggerMessageKey  string  [logger] Key for message in JSON ${DISCORD_LOGGER_MESSAGE_KEY} (default "msg")
  --loggerTimeKey     string  [logger] Key for timestamp in JSON ${DISCORD_LOGGER_TIME_KEY} (default "time")
  --publicKey         string  [discord] Public Key ${DISCORD_PUBLIC_KEY}
```
