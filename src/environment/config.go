package environment

import (
	"errors"
	"os"
)

type Config struct {
	LogLevel    string
	BotToken    string
	OpenAIToken string
}

func ReadConfig() (Config, error) {
	logLevel, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		logLevel = "INFO"
	}

	botToken, ok := os.LookupEnv("BOT_TOKEN")
	if !ok {
		return Config{}, errors.New("BOT_TOKEN env variable is not set")
	}

	openAIToken, ok := os.LookupEnv("OPENAI_TOKEN")
	if !ok {
		return Config{}, errors.New("OPENAI_TOKEN env variable is not set")
	}

	config := Config{
		LogLevel:    logLevel,
		BotToken:    botToken,
		OpenAIToken: openAIToken,
	}

	return config, nil
}
