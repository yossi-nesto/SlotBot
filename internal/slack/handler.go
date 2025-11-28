package slack

import (
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/yossigruner/SlotBot/internal/calendar"
	"github.com/yossigruner/SlotBot/internal/domain"
)

var jiraRegex = regexp.MustCompile(`^[A-Z]+-\d+$`)

type Handler struct {
	calClient *calendar.Client
}

func NewHandler(calClient *calendar.Client) *Handler {
	return &Handler{calClient: calClient}
}

func (h *Handler) HandleBook(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	text := r.FormValue("text")
	userID := r.FormValue("user_name") // In production, fetch real name via API

	// Parse args: env service jira [start] [duration]
	args := strings.Fields(text)
	if len(args) < 3 {
		respond(w, "Usage: /env-book <env> <service> <jira> [start] [duration]")
		return
	}

	env := args[0]
	service := args[1]
	jira := args[2]

	if !jiraRegex.MatchString(jira) {
		respond(w, "❌ Invalid Jira ticket format. Must be like PROJ-123")
		return
	}

	// Default start: now, duration: 1h
	startTime := time.Now()
	duration := time.Hour

	if len(args) > 3 {
		parsedStart, err := time.Parse(time.RFC3339, args[3]) // Simplistic parsing
		if err == nil {
			startTime = parsedStart
		} else {
			// Try parsing as duration if start parsing fails?
			// For strictness, let's assume position 4 is start if present
			// But requirements say start is optional.
			// Let's stick to strict positional for v1 or try to be smart.
			// Requirements: [startISO] [duration]
			// Let's try to parse args[3] as time. If fail, try as duration?
			// Actually, let's stick to the plan: start is ISO, duration is Go duration.

			// If args[3] is not a time, maybe it's a duration?
			// But the requirement says: /env-book <env> <service> <jira> [startISO] [duration]
			// So if 4 args, check if it's time or duration.
			d, dErr := time.ParseDuration(args[3])
			if dErr == nil {
				duration = d
			} else {
				// If not duration, assume it's start time
				// If parsing fails here, it's an error
				// But wait, if user provides ONLY duration?
				// The command structure implies order.
				// Let's assume strict order for now to keep it simple as per requirements.
				// Actually, let's be flexible:
				// If args[3] parses as time -> start.
				// If args[3] parses as duration -> duration (and start is now).
			}
		}
	}

	if len(args) > 4 {
		d, err := time.ParseDuration(args[4])
		if err == nil {
			duration = d
		}
	}

	booking := domain.Booking{
		Env:        env,
		Service:    service,
		JiraTicket: jira,
		StartTime:  startTime,
		Duration:   duration,
		User:       userID,
	}

	if err := calendar.ValidateBooking(booking); err != nil {
		respond(w, fmt.Sprintf("❌ Validation error: %v", err))
		return
	}

	// Check if calendar client is available
	if h.calClient == nil {
		respond(w, "❌ Calendar not configured. Please set up Google Calendar credentials (see OAUTH_SETUP.md)")
		return
	}

	// Check conflicts
	// We need to fetch events around the booking time
	// Let's fetch -1 day to +1 day to be safe, or just the booking range
	// Ideally, fetch overlapping range.
	events, err := h.calClient.ListEvents(r.Context(), startTime.Add(-24*time.Hour), startTime.Add(24*time.Hour))
	if err != nil {
		respond(w, "❌ Failed to check calendar")
		return
	}

	if conflict := calendar.CheckConflict(booking, events); conflict != nil {
		slog.Info("Booking conflict detected", "env", booking.Env, "service", booking.Service, "conflict_with", conflict.Title)
		respond(w, fmt.Sprintf("❌ Conflict detected!\n%s | %s already booked\n%s - %s\n%s",
			conflict.Env, conflict.Service,
			conflict.StartTime.Format("15:04"), conflict.EndTime.Format("15:04"),
			conflict.Title))
		return
	}

	link, err := h.calClient.CreateEvent(r.Context(), booking)
	if err != nil {
		slog.Error("Failed to create calendar event", "error", err)
		respond(w, "❌ Failed to create calendar event")
		return
	}

	slog.Info("Booking created", "env", booking.Env, "service", booking.Service, "user", booking.User)
	respond(w, fmt.Sprintf("✅ Booked %s / %s\n%s - %s\nLink: %s",
		booking.Env, booking.Service,
		booking.StartTime.Format("15:04"), booking.StartTime.Add(booking.Duration).Format("15:04"),
		link))
}

func (h *Handler) HandleNextSlot(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	text := r.FormValue("text")
	args := strings.Fields(text)
	if len(args) < 2 {
		respond(w, "Usage: /env-next <env> <service> [duration]")
		return
	}

	env := args[0]
	service := args[1]
	duration := time.Hour // Default

	if len(args) > 2 {
		d, err := time.ParseDuration(args[2])
		if err == nil {
			duration = d
		}
	}

	// Check if calendar client is available
	if h.calClient == nil {
		respond(w, "❌ Calendar not configured. Please set up Google Calendar credentials (see OAUTH_SETUP.md)")
		return
	}

	// Fetch events for next 7 days
	now := time.Now()
	events, err := h.calClient.ListEvents(r.Context(), now, now.Add(7*24*time.Hour))
	if err != nil {
		respond(w, "❌ Failed to check calendar")
		return
	}

	nextSlot := calendar.FindNextSlot(env, service, duration, events)

	respond(w, fmt.Sprintf("🔍 Next available slot for %s / %s (%s):\n👉 %s",
		env, service, duration, nextSlot.Format("Mon, 02 Jan 15:04")))
}

func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	text := r.FormValue("text")
	args := strings.Fields(text)

	// Optional env filter
	envFilter := ""
	if len(args) > 0 {
		envFilter = strings.ToLower(args[0])
	}

	// Check if calendar client is available
	if h.calClient == nil {
		respond(w, "❌ Calendar not configured. Please set up Google Calendar credentials (see OAUTH_SETUP.md)")
		return
	}

	// Fetch today's events
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	events, err := h.calClient.ListEvents(r.Context(), startOfDay, endOfDay)
	if err != nil {
		slog.Error("Failed to list calendar events", "error", err)
		respond(w, "❌ Failed to check calendar")
		return
	}

	if len(events) == 0 {
		respond(w, "📅 No bookings for today")
		return
	}

	// Filter by env if specified
	var filteredEvents []domain.Event
	for _, event := range events {
		if envFilter == "" || strings.EqualFold(event.Env, envFilter) {
			filteredEvents = append(filteredEvents, event)
		}
	}

	if len(filteredEvents) == 0 {
		if envFilter != "" {
			respond(w, fmt.Sprintf("📅 No bookings for %s today", envFilter))
		} else {
			respond(w, "📅 No bookings for today")
		}
		return
	}

	// Build response
	var response strings.Builder
	response.WriteString(fmt.Sprintf("📅 Bookings for today (%d):\n\n", len(filteredEvents)))

	for _, event := range filteredEvents {
		response.WriteString(fmt.Sprintf("• %s | %s\n", event.Env, event.Service))
		response.WriteString(fmt.Sprintf("  %s - %s\n",
			event.StartTime.Format("15:04"),
			event.EndTime.Format("15:04")))
		response.WriteString(fmt.Sprintf("  %s\n\n", event.Title))
	}

	respond(w, response.String())
}

func respond(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	// Slack expects a JSON response or plain text.
	// For simple responses, plain text body is fine if response_type is ephemeral (default).
	// But let's send JSON to be safe and future proof.
	fmt.Fprintf(w, `{"text": "%s"}`, message)
}
