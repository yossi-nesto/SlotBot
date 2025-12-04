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

var jiraRegex = regexp.MustCompile(`^[A-Za-z]+-\d+$`)

type Handler struct {
	calClient  *calendar.Client
	CalendarID string
}

func NewHandler(calClient *calendar.Client, calendarID string) *Handler {
	return &Handler{
		calClient:  calClient,
		CalendarID: calendarID,
	}
}

// HandleUnified is the main handler that routes to subcommands
func (h *Handler) HandleUnified(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	text := r.FormValue("text")
	args := strings.Fields(text)

	if len(args) == 0 {
		helpText := `📚 *SlotBot - Environment Booking Manager*

*Available Commands:*

*1️⃣ Book an Environment*
` + "`/slot book <env> <service> <jira> [start] [duration]`" + `
Book a testing environment for your team.
• *env*: Environment name (e.g., staging, qa, demo)
• *service*: Service name (e.g., api, web, db)
• *jira*: Jira ticket (e.g., PROJ-123 or og-1234)
• *start*: (Optional) Start time (ISO, HH:MM, or relative like '2h')
• *duration*: (Optional) Duration (default: 1h)

*Examples:*
` + "`/slot book staging api OG-1234`" + `
` + "`/slot book qa web OG-456 14:30`" + ` (Start at 14:30 today)
` + "`/slot book demo db OG-789 2h`" + ` (Start in 2 hours)

*2️⃣ Find Next Available Slot*
` + "`/slot next <env> <service> [duration]`" + `
Find the next available time slot for an environment.

*Examples:*
` + "`/slot next staging api`" + `
` + "`/slot next qa web 2h`" + `

*3️⃣ List Today's Bookings*
` + "`/slot list [env]`" + `
View all bookings for today, optionally filtered by environment.

*Examples:*
` + "`/slot list`" + `
` + "`/slot list staging`" + `

*4️⃣ Show Active Bookings*
` + "`/slot current [env]`" + ` (or ` + "`/slot now`" + `)
View what is currently booked right now.

*Examples:*
` + "`/slot current`" + `
` + "`/slot now staging`" + `

*5️⃣ Open Calendar*
` + "`/slot open`" + `
Open Google Calendar in your browser.

*6️⃣ Add Calendar*
` + "`/slot add`" + `
Add the SlotBot calendar to your Google Calendar list.

💡 *Tip:* All bookings are automatically rounded to 15-minute intervals (:00, :15, :30, :45)`
		respond(w, helpText)
		return
	}

	subcommand := strings.ToLower(args[0])
	remainingArgs := args[1:]

	switch subcommand {
	case "book":
		h.handleBookSubcommand(w, r, remainingArgs)
	case "next":
		h.handleNextSubcommand(w, r, remainingArgs)
	case "list":
		h.handleListSubcommand(w, r, remainingArgs)
	case "current", "now":
		h.handleCurrentSubcommand(w, r, remainingArgs)
	case "open":
		h.handleOpenSubcommand(w, r, remainingArgs)
	case "add":
		h.handleAddSubcommand(w, r, remainingArgs)
	default:
		respond(w, fmt.Sprintf("❌ Unknown subcommand: %s\n\nAvailable commands:\n• `book` - Book an environment\n• `next` - Find next available slot\n• `list` - List today's bookings\n• `current` - Show active bookings\n• `open` - Open calendar\n• `add` - Add calendar to your list", subcommand))
	}
}

func (h *Handler) handleOpenSubcommand(w http.ResponseWriter, r *http.Request, args []string) {
	// Generic link to open Google Calendar
	respond(w, "📅 *Open Google Calendar*\n<https://calendar.google.com/calendar/r|Click here to view your calendar>")
}

func (h *Handler) handleAddSubcommand(w http.ResponseWriter, r *http.Request, args []string) {
	// Get calendar ID from struct
	calID := h.CalendarID
	if calID == "" {
		calID = "primary" // Fallback
	}

	// Construct the URL to ADD the calendar
	// https://calendar.google.com/calendar/u/0/r?cid=<CALENDAR_ID>
	url := fmt.Sprintf("https://calendar.google.com/calendar/u/0/r?cid=%s", calID)

	respond(w, fmt.Sprintf("➕ *Add SlotBot Calendar*\n<%s|Click here to add this calendar to your list>", url))
}

func (h *Handler) handleCurrentSubcommand(w http.ResponseWriter, r *http.Request, args []string) {
	envFilter := ""
	if len(args) > 0 {
		envFilter = strings.ToLower(args[0])
	}

	if h.calClient == nil {
		respond(w, "❌ Calendar not configured. Please set up Google Calendar credentials (see SERVICE_ACCOUNT_SETUP.md)")
		return
	}

	// Check a window around "now" to be safe, e.g. last 2 hours to next 2 hours
	// But actually ListEvents filters by start/end.
	// We want events where Start <= Now < End.
	// So we need events that started before now and end after now.
	// Safest is to fetch today's events and filter in memory.
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	events, err := h.calClient.ListEvents(r.Context(), startOfDay, endOfDay)
	if err != nil {
		slog.Error("Failed to list calendar events", "error", err)
		respond(w, "❌ Failed to check calendar")
		return
	}

	var activeEvents []domain.Event
	for _, event := range events {
		// Filter by env if specified
		if envFilter != "" && !strings.EqualFold(event.Env, envFilter) {
			continue
		}

		// Check if currently active
		if event.StartTime.Before(now) && event.EndTime.After(now) {
			activeEvents = append(activeEvents, event)
		}
	}

	if len(activeEvents) == 0 {
		if envFilter != "" {
			respond(w, fmt.Sprintf("🟢 No active bookings for %s right now", envFilter))
		} else {
			respond(w, "🟢 No active bookings right now")
		}
		return
	}

	var response strings.Builder
	response.WriteString(fmt.Sprintf("🔴 Currently Active Bookings (%d):\n\n", len(activeEvents)))
	response.WriteString("```\n")
	response.WriteString(formatEventsTable(activeEvents))
	response.WriteString("```")

	respond(w, response.String())
}

func (h *Handler) handleBookSubcommand(w http.ResponseWriter, r *http.Request, args []string) {
	userID := r.FormValue("user_name")

	if len(args) < 3 {
		respond(w, "Usage: `/slot book <env> <service> <jira> [start] [duration]`")
		return
	}

	env := args[0]
	service := args[1]
	jira := args[2]

	if !jiraRegex.MatchString(jira) {
		respond(w, "❌ Invalid Jira ticket format. Must be like PROJ-123 or OG-1234")
		return
	}

	// Default start: now, duration: 1h
	startTime := time.Now()
	duration := time.Hour

	if len(args) > 3 {
		// Try parsing as ISO time
		parsedStart, err := time.Parse(time.RFC3339, args[3])
		if err == nil {
			startTime = parsedStart
		} else {
			// Try parsing as HH:MM (Kitchen/24h format)
			// We assume today if only time is given
			parsedTime, timeErr := time.Parse("15:04", args[3])
			if timeErr == nil {
				now := time.Now()
				startTime = time.Date(now.Year(), now.Month(), now.Day(), parsedTime.Hour(), parsedTime.Minute(), 0, 0, now.Location())
				// If time has passed today, maybe they mean tomorrow?
				// For now, let's keep it simple: today.
			} else {
				// Try parsing as duration (relative time)
				d, dErr := time.ParseDuration(args[3])
				if dErr == nil {
					// args[3] is start offset (e.g. "2h" -> in 2 hours)
					startTime = startTime.Add(d)
				}
			}
		}

		// If we successfully parsed a start time (ISO or HH:MM or relative), check for duration in next arg
		if len(args) > 4 {
			d, err := time.ParseDuration(args[4])
			if err == nil {
				duration = d
			}
		} else if len(args) == 4 {
			// If args[3] was a duration (e.g., "/slot book env svc jira 2h"), then it's the duration
			// This case is only hit if args[3] was NOT an ISO time or HH:MM.
			// The `startTime = startTime.Add(d)` above handles the "start in X duration" case.
			// If args[3] was just a duration and no start time was specified, then startTime remains time.Now()
			// and args[3] is the duration.
			_, isoErr := time.Parse(time.RFC3339, args[3])
			_, hhmmErr := time.Parse("15:04", args[3])
			if isoErr != nil && hhmmErr != nil { // Only if args[3] was not ISO or HH:MM
				d, dErr := time.ParseDuration(args[3])
				if dErr == nil {
					duration = d
				}
			}
		}
	}

	// Round start time to nearest 15-minute interval
	startTime = roundToQuarterHour(startTime)

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

	if h.calClient == nil {
		respond(w, "❌ Calendar not configured. Please set up Google Calendar credentials (see SERVICE_ACCOUNT_SETUP.md)")
		return
	}

	events, err := h.calClient.ListEvents(r.Context(), startTime.Add(-24*time.Hour), startTime.Add(24*time.Hour))
	if err != nil {
		respond(w, "❌ Failed to check calendar")
		return
	}

	if conflict := calendar.CheckConflict(booking, events); conflict != nil {
		slog.Info("Booking conflict detected", "env", booking.Env, "service", booking.Service, "conflict_with", conflict.Title)

		// Find next available slot
		// We need to fetch more events to find the next slot reliably
		searchStart := time.Now()
		extendedEvents, err := h.calClient.ListEvents(r.Context(), searchStart, searchStart.Add(7*24*time.Hour))
		if err != nil {
			slog.Error("Failed to list events for next slot search", "error", err)
			// Fallback to simple conflict message if we can't search
			respond(w, fmt.Sprintf("❌ Conflict detected!\n```\n%s```", formatEventsTable([]domain.Event{*conflict})))
			return
		}

		nextSlot := calendar.FindNextSlot(booking.Env, booking.Service, booking.Duration, extendedEvents)

		// Format the next slot suggestion
		suggestion := fmt.Sprintf("/slot book %s %s %s %s %s",
			booking.Env,
			booking.Service,
			booking.JiraTicket,
			nextSlot.Format(time.RFC3339),
			booking.Duration)

		respond(w, fmt.Sprintf("❌ Conflict detected!\n```\n%s```\n👉 *Next available slot:*\n%s\nTo book it, copy and paste:\n`%s`",
			formatEventsTable([]domain.Event{*conflict}),
			nextSlot.Format("Mon, 02 Jan 15:04"),
			suggestion))
		return
	}

	link, err := h.calClient.CreateEvent(r.Context(), booking)
	if err != nil {
		slog.Error("Failed to create calendar event", "error", err)
		respond(w, "❌ Failed to create calendar event")
		return
	}

	slog.Info("Booking created", "env", booking.Env, "service", booking.Service, "user", booking.User)

	// Create event object for display
	newEvent := domain.Event{
		Env:       booking.Env,
		Service:   booking.Service,
		Title:     fmt.Sprintf("%s | %s | %s | %s", booking.Env, booking.Service, booking.JiraTicket, booking.User),
		StartTime: booking.StartTime,
		EndTime:   booking.StartTime.Add(booking.Duration),
	}

	respond(w, fmt.Sprintf("✅ Booked!\n```\n%s```\nLink: <%s|Open in Calendar>",
		formatEventsTable([]domain.Event{newEvent}),
		link))
}

func (h *Handler) handleNextSubcommand(w http.ResponseWriter, r *http.Request, args []string) {
	if len(args) < 2 {
		respond(w, "Usage: `/slot next <env> <service> [duration]`")
		return
	}

	env := args[0]
	service := args[1]
	duration := time.Hour

	if len(args) > 2 {
		d, err := time.ParseDuration(args[2])
		if err == nil {
			duration = d
		}
	}

	if h.calClient == nil {
		respond(w, "❌ Calendar not configured. Please set up Google Calendar credentials (see SERVICE_ACCOUNT_SETUP.md)")
		return
	}

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

func (h *Handler) handleListSubcommand(w http.ResponseWriter, r *http.Request, args []string) {
	envFilter := ""
	if len(args) > 0 {
		envFilter = strings.ToLower(args[0])
	}

	if h.calClient == nil {
		respond(w, "❌ Calendar not configured. Please set up Google Calendar credentials (see SERVICE_ACCOUNT_SETUP.md)")
		return
	}

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

	var response strings.Builder
	response.WriteString(fmt.Sprintf("📅 Bookings for today (%d):\n\n", len(filteredEvents)))
	response.WriteString("```\n")
	response.WriteString(formatEventsTable(filteredEvents))
	response.WriteString("```")

	respond(w, response.String())
}

// formatEventsTable creates an ASCII table for the events
func formatEventsTable(events []domain.Event) string {
	if len(events) == 0 {
		return ""
	}

	// Calculate column widths
	// Columns: Time | Env | Service | Title
	timeWidth := 13 // "15:04 - 15:04"
	envWidth := 3   // "Env"
	svcWidth := 7   // "Service"
	titleWidth := 5 // "Title"

	for _, e := range events {
		if len(e.Env) > envWidth {
			envWidth = len(e.Env)
		}
		if len(e.Service) > svcWidth {
			svcWidth = len(e.Service)
		}
		// Truncate title if too long? User asked to show all.
		// So we just take the full length.
		if len(e.Title) > titleWidth {
			titleWidth = len(e.Title)
		}
	}

	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("%-*s | %-*s | %-*s | %-*s\n",
		timeWidth, "Time",
		envWidth, "Env",
		svcWidth, "Service",
		titleWidth, "Title"))

	// Separator
	sb.WriteString(strings.Repeat("-", timeWidth) + "-+-" +
		strings.Repeat("-", envWidth) + "-+-" +
		strings.Repeat("-", svcWidth) + "-+-" +
		strings.Repeat("-", titleWidth) + "\n")

	// Rows
	for _, e := range events {
		timeStr := fmt.Sprintf("%s - %s", e.StartTime.Format("15:04"), e.EndTime.Format("15:04"))
		sb.WriteString(fmt.Sprintf("%-*s | %-*s | %-*s | %-*s\n",
			timeWidth, timeStr,
			envWidth, e.Env,
			svcWidth, e.Service,
			titleWidth, e.Title))
	}

	return sb.String()
}

func respond(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"text": "%s"}`, message)
}

// roundToQuarterHour rounds a time to the nearest 15-minute interval
// (00, 15, 30, or 45 minutes)
func roundToQuarterHour(t time.Time) time.Time {
	minutes := t.Minute()

	// Round to nearest 15-minute mark
	roundedMinutes := ((minutes + 7) / 15) * 15

	// Handle overflow to next hour
	if roundedMinutes >= 60 {
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour()+1, 0, 0, 0, t.Location())
	}

	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), roundedMinutes, 0, 0, t.Location())
}
