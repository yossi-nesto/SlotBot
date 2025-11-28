# SlotBot

Environment Booking Bot for Slack, integrated with Google Calendar.

## Setup

1.  **Slack App**:
    -   Create a new Slack App.
    -   Enable **Slash Commands**:
        -   `/env-book`: Request URL `https://your-domain.com/slack/book`
        -   `/env-bookings`: Request URL `https://your-domain.com/slack/bookings`
        -   `/env-next`: Request URL `https://your-domain.com/slack/next`
    -   Install App to Workspace.
    -   Copy `Signing Secret` and `Bot User OAuth Token`.

2.  **Google Calendar (OAuth)**:
    -   Follow instructions in `OAUTH_SETUP.md`.
    -   Place `oauth_credentials.json` in the project root.

3.  **Environment Variables**:
    -   Copy `.env.example` to `.env` and fill in the values.
    -   `SLACK_SIGNING_SECRET` & `SLACK_BOT_TOKEN`: From Slack App.
    -   `GCAL_CALENDAR_ID`: Your target calendar ID (or "primary").

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

## Running with Docker

**Build and run:**
```bash
docker-compose up --build
```

**Run in background:**
```bash
docker-compose up -d
```

**View logs:**
```bash
docker-compose logs -f
```

**Stop:**
```bash
docker-compose down
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
