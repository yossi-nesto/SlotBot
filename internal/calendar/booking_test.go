package calendar

import (
	"testing"
	"time"

	"github.com/yossigruner/SlotBot/internal/domain"
)

func TestValidateBooking(t *testing.T) {
	tests := []struct {
		name    string
		booking domain.Booking
		wantErr bool
	}{
		{
			name: "Valid booking",
			booking: domain.Booking{
				Env:      "staging",
				Duration: time.Hour,
			},
			wantErr: false,
		},
		{
			name: "Invalid env",
			booking: domain.Booking{
				Env:      "prod",
				Duration: time.Hour,
			},
			wantErr: true,
		},
		{
			name: "Duration too long",
			booking: domain.Booking{
				Env:      "staging",
				Duration: 3 * time.Hour,
			},
			wantErr: true,
		},
		{
			name: "Duration too short",
			booking: domain.Booking{
				Env:      "staging",
				Duration: 1 * time.Minute,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateBooking(tt.booking); (err != nil) != tt.wantErr {
				t.Errorf("ValidateBooking() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckConflict(t *testing.T) {
	now := time.Now()

	existingEvents := []domain.Event{
		{
			Env:       "staging",
			Service:   "auth",
			StartTime: now.Add(1 * time.Hour),
			EndTime:   now.Add(2 * time.Hour),
		},
	}

	tests := []struct {
		name     string
		booking  domain.Booking
		wantConf bool
	}{
		{
			name: "No conflict - different service",
			booking: domain.Booking{
				Env:       "staging",
				Service:   "payments",
				StartTime: now.Add(1 * time.Hour),
				Duration:  time.Hour,
			},
			wantConf: false,
		},
		{
			name: "No conflict - different env",
			booking: domain.Booking{
				Env:       "qa",
				Service:   "auth",
				StartTime: now.Add(1 * time.Hour),
				Duration:  time.Hour,
			},
			wantConf: false,
		},
		{
			name: "No conflict - different time",
			booking: domain.Booking{
				Env:       "staging",
				Service:   "auth",
				StartTime: now.Add(3 * time.Hour),
				Duration:  time.Hour,
			},
			wantConf: false,
		},
		{
			name: "Conflict - exact match",
			booking: domain.Booking{
				Env:       "staging",
				Service:   "auth",
				StartTime: now.Add(1 * time.Hour),
				Duration:  time.Hour,
			},
			wantConf: true,
		},
		{
			name: "Conflict - overlap start",
			booking: domain.Booking{
				Env:       "staging",
				Service:   "auth",
				StartTime: now.Add(30 * time.Minute),
				Duration:  time.Hour,
			},
			wantConf: true,
		},
		{
			name: "Conflict - overlap end",
			booking: domain.Booking{
				Env:       "staging",
				Service:   "auth",
				StartTime: now.Add(90 * time.Minute),
				Duration:  time.Hour,
			},
			wantConf: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckConflict(tt.booking, existingEvents)
			if (got != nil) != tt.wantConf {
				t.Errorf("CheckConflict() conflict = %v, wantConf %v", got, tt.wantConf)
			}
		})
	}
}

func TestFindNextSlot(t *testing.T) {
	now := time.Now()
	// Round now to minute for stable comparison if needed, but logic uses exact time

	existingEvents := []domain.Event{
		{
			Env:       "staging",
			Service:   "auth",
			StartTime: now.Add(30 * time.Minute),
			EndTime:   now.Add(90 * time.Minute),
		},
		{
			Env:       "staging",
			Service:   "auth",
			StartTime: now.Add(2 * time.Hour),
			EndTime:   now.Add(3 * time.Hour),
		},
	}

	tests := []struct {
		name     string
		duration time.Duration
		want     time.Time // Approximate check
	}{
		{
			name:     "Slot available immediately (before first event)",
			duration: 15 * time.Minute,
			want:     now, // Should be roughly now
		},
		{
			name:     "Slot available in gap",
			duration: 30 * time.Minute,
			want:     now.Add(90 * time.Minute), // End of first event
		},
		{
			name:     "Slot available after all events",
			duration: 2 * time.Hour,
			want:     now.Add(3 * time.Hour), // End of last event
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindNextSlot("staging", "auth", tt.duration, existingEvents)

			// Allow small delta for "now" comparisons
			diff := got.Sub(tt.want)
			if diff < 0 {
				diff = -diff
			}
			if diff > time.Second {
				t.Errorf("FindNextSlot() = %v, want %v", got, tt.want)
			}
		})
	}
}
