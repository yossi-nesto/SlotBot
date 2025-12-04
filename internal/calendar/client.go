package calendar

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/yossigruner/SlotBot/internal/config"
	"github.com/yossigruner/SlotBot/internal/domain"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Client struct {
	srv        *calendar.Service
	calendarID string
	timezone   *time.Location
}

func NewClient(ctx context.Context, cfg *config.Config) (*Client, error) {
	var opts []option.ClientOption

	// Check if using Service Account (service-account.json exists)
	if _, err := os.Stat("service-account.json"); err == nil {
		slog.Info("Using service account authentication")
		opts = append(opts, option.WithCredentialsFile("service-account.json"))
	} else {
		return nil, fmt.Errorf("service-account.json not found - see SERVICE_ACCOUNT_SETUP.md for setup instructions")
	}

	srv, err := calendar.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Calendar client: %w", err)
	}

	// If using OAuth and no calendar ID specified, use "primary"
	calendarID := cfg.GoogleCalendarID
	if calendarID == "" {
		calendarID = "primary"
		slog.Info("Using primary calendar")
	}

	return &Client{
		srv:        srv,
		calendarID: calendarID,
		timezone:   cfg.DefaultTimezone,
	}, nil
}

func (c *Client) ListEvents(ctx context.Context, start, end time.Time) ([]domain.Event, error) {
	events, err := c.srv.Events.List(c.calendarID).
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(start.Format(time.RFC3339)).
		TimeMax(end.Format(time.RFC3339)).
		OrderBy("startTime").
		Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve next ten of the user's events: %w", err)
	}

	var domainEvents []domain.Event
	for _, item := range events.Items {
		// Parse title: env | service | jira | holder
		parts := strings.Split(item.Summary, "|")
		if len(parts) < 2 {
			continue // Skip malformed events
		}

		env := strings.TrimSpace(strings.ToLower(parts[0]))
		service := strings.TrimSpace(strings.ToLower(parts[1]))

		startTime, _ := time.Parse(time.RFC3339, item.Start.DateTime)
		endTime, _ := time.Parse(time.RFC3339, item.End.DateTime)

		domainEvents = append(domainEvents, domain.Event{
			Title:     item.Summary,
			StartTime: startTime,
			EndTime:   endTime,
			Env:       env,
			Service:   service,
		})
	}

	return domainEvents, nil
}

func (c *Client) CreateEvent(ctx context.Context, booking domain.Booking) (string, error) {
	summary := fmt.Sprintf("%s | %s | %s | %s",
		strings.ToLower(booking.Env),
		strings.ToLower(booking.Service),
		strings.ToUpper(booking.JiraTicket),
		booking.User,
	)

	description := "Managed by Env Booking Bot"

	event := &calendar.Event{
		Summary:     summary,
		Description: description,
		Start: &calendar.EventDateTime{
			DateTime: booking.StartTime.Format(time.RFC3339),
			TimeZone: c.timezone.String(),
		},
		End: &calendar.EventDateTime{
			DateTime: booking.StartTime.Add(booking.Duration).Format(time.RFC3339),
			TimeZone: c.timezone.String(),
		},
	}

	createdEvent, err := c.srv.Events.Insert(c.calendarID, event).Do()
	if err != nil {
		return "", fmt.Errorf("unable to create event: %w", err)
	}

	return createdEvent.HtmlLink, nil
}
