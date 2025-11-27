package calendar

import (
	"fmt"
	"strings"
	"time"

	"github.com/yossigruner/SlotBot/internal/domain"
)

func CheckConflict(newBooking domain.Booking, existingEvents []domain.Event) *domain.Event {
	newStart := newBooking.StartTime
	newEnd := newBooking.StartTime.Add(newBooking.Duration)

	for _, event := range existingEvents {
		// Check if env and service match
		if !strings.EqualFold(event.Env, newBooking.Env) ||
			!strings.EqualFold(event.Service, newBooking.Service) {
			continue
		}

		// Check for time overlap
		// Overlap exists if (StartA < EndB) and (EndA > StartB)
		if event.StartTime.Before(newEnd) && event.EndTime.After(newStart) {
			return &event
		}
	}

	return nil
}

func ValidateBooking(b domain.Booking) error {
	validEnvs := map[string]bool{
		"staging": true,
		"qa":      true,
		"demo":    true,
	}

	if !validEnvs[strings.ToLower(b.Env)] {
		return fmt.Errorf("invalid environment: %s. Must be staging, qa, or demo", b.Env)
	}

	if b.Duration > 2*time.Hour {
		return fmt.Errorf("maximum booking duration is 2 hours")
	}

	if b.Duration < 5*time.Minute {
		return fmt.Errorf("minimum booking duration is 5 minutes")
	}

	return nil
}

func FindNextSlot(env, service string, duration time.Duration, existingEvents []domain.Event) time.Time {
	// We assume existingEvents are sorted by StartTime (Calendar API does this)
	// Filter events for this env/service
	var relevantEvents []domain.Event
	for _, e := range existingEvents {
		if strings.EqualFold(e.Env, env) &&
			strings.EqualFold(e.Service, service) {
			relevantEvents = append(relevantEvents, e)
		}
	}

	// Start looking from now (rounded up to next 15 mins for cleanliness, or just now)
	// Let's say "now"
	searchStart := time.Now()

	// Check gap before first event
	if len(relevantEvents) == 0 {
		return searchStart
	}

	// If first event starts after (now + duration), then we can book now
	if relevantEvents[0].StartTime.Sub(searchStart) >= duration {
		return searchStart
	}

	// Check gaps between events
	for i := 0; i < len(relevantEvents)-1; i++ {
		currentEnd := relevantEvents[i].EndTime
		nextStart := relevantEvents[i+1].StartTime

		// If current event ends in the past, treat it as "now" (or ignore)
		if currentEnd.Before(searchStart) {
			currentEnd = searchStart
		}

		// If gap is big enough
		if nextStart.Sub(currentEnd) >= duration {
			return currentEnd
		}
	}

	// Check after last event
	lastEventEnd := relevantEvents[len(relevantEvents)-1].EndTime
	if lastEventEnd.Before(searchStart) {
		return searchStart
	}
	return lastEventEnd
}
