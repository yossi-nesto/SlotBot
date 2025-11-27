# SlotBot

Environment Booking Bot for Slack, integrated with Google Calendar.

## Setup

1.  **Google Calendar**:
    -   Create a Service Account in Google Cloud Console.
    -   Download the JSON key and save it as `credentials.json`.
    -   Share your target Google Calendar with the Service Account email (Make changes to events).
    -   Get the Calendar ID (e.g., `primary` or `c_xxxxxxxx@group.calendar.google.com`).

2.  **Slack App**:
    -   Create a new Slack App.
    -   Enable **Slash Commands**:
        -   `/env-book`: Request URL `https://your-domain.com/slack/book`
        -   `/env-bookings`: Request URL `https://your-domain.com/slack/bookings`
    -   Install App to Workspace.
    -   Copy `Signing Secret` and `Bot User OAuth Token`.

3.  **Environment Variables**:
    -   Copy `.env.example` to `.env` and fill in the values.

## Running Locally

You can use the provided `Makefile`:

```bash
make build  # Build binary to bin/slotbot
make run    # Run locally
make test   # Run tests
```

Or standard Go commands:

```bash
go run ./cmd/server
```

## Usage

**Book an environment:**
```
/env-book staging applications PROJ-123
/env-book qa payments PROJ-456 2025-11-27T15:00 30m
```

**List bookings:**
```
/env-bookings
```

**Find next available slot:**
```
/env-next staging auth
/env-next staging auth 2h
```
