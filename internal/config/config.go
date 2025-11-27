package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	SlackSigningSecret         string
	SlackBotToken              string
	GoogleCalendarID           string
	GoogleApplicationCredentials string
	DefaultTimezone            *time.Location
	Port                       string
}

func Load() (*Config, error) {
	tzStr := os.Getenv("DEFAULT_TIMEZONE")
	if tzStr == "" {
		tzStr = "UTC"
	}
	loc, err := time.LoadLocation(tzStr)
	if err != nil {
		return nil, fmt.Errorf("invalid DEFAULT_TIMEZONE: %w", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		SlackSigningSecret:           os.Getenv("SLACK_SIGNING_SECRET"),
		SlackBotToken:                os.Getenv("SLACK_BOT_TOKEN"),
		GoogleCalendarID:             os.Getenv("GCAL_CALENDAR_ID"),
		GoogleApplicationCredentials: os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
		DefaultTimezone:              loc,
		Port:                         port,
	}, nil
}
