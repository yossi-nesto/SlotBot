package domain

import (
	"time"
)

// Booking represents a parsed booking request
type Booking struct {
	Env        string
	Service    string
	JiraTicket string
	StartTime  time.Time
	Duration   time.Duration
	User       string // Slack user ID or Name
}

// Event represents a calendar event for conflict checking
type Event struct {
	Title     string
	StartTime time.Time
	EndTime   time.Time
	Env       string
	Service   string
}
