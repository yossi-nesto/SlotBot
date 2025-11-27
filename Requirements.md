## Requirments


Project Name

Environment Booking Bot (Slack + Go + Chi + Google Calendar)

Objective

Build a Slack bot using Go and Chi that manages bookings for shared testing environments (staging, QA, demo).
The bot uses Google Calendar as the source of truth and prevents conflicts when reserving environments.

Tech Stack

Language: Go

HTTP router: github.com/go-chi/chi/v5

Slack API:

Slash commands

Signing secret verification

Google Calendar API (Service Account)

No database (Calendar is system of record)

Environments & Calendar
Calendar Rules

One shared Google Calendar for all environments.

Events must strictly follow the title format:

env | service | jira | holder


Example:

staging | applications | OG-12345 | Yossi Gruner

Allowed Environments

Only the following are valid:

staging
qa
demo


Input must be normalized as:

env → lowercase

service → lowercase

jira → uppercase

Jira Validation

All JIRA entries must match:

^[A-Z]+-\d+$


Invalid Jira strings must return a friendly Slack error.

Booking Duration Rules
Rule	Value
Default duration	1 hour
Max allowed	2 hours

Error on violation:

❌ Maximum booking allowed is 2 hours.

Slack Commands
/env-book
/env-book <env> <service> <jira> [startISO] [duration]


Example:

/env-book staging backoffice OG-12345
/env-book qa applications OG-7890 2025-11-27T13:00 1h

Argument Rules
Field	Required	Validation
env	Yes	Must be staging/qa/demo
service	Yes	string
jira	Yes	Regex enforced
start	No	RFC3339
duration	No	Go duration (30m, 2h)

Defaults:

start = now

duration = 1h

Booking Logic
Conflict Definition

A conflict exists if:

existing.Start < new.End
AND
existing.End > new.Start
AND
existing.env == new.env
AND
existing.service == new.service

Booking Flow

Parse arguments

Validate env + jira

Normalize values

Enforce max duration

Query calendar for conflicts

Reject if conflict exists

Create event if clean

Slack Success Response
✅ Booked staging / applications
14:00–15:00
Link: <calendar event>

Slack Conflict Response
❌ Conflict detected

staging / applications already booked
10:00–12:00
OG-54321 • Alice

/env-bookings
/env-bookings
/env-bookings staging

Behavior

Shows all bookings today

Optional env filter

Sorted by start time

Holder Identity Rules

Slack user identity handling:

Use Slack API to retrieve real name or display name

Do NOT trust provided username

All records must show the resolved full name

Google Calendar Safety Controls

Bot-created events must be distinguishable:

Either:

Prefix:

[ENVBOT] staging | service | jira | holder


OR

Description must always contain:

Managed by Env Booking Bot

System behavior:

Bot processes only events that it created.

Manual calendar events are ignored.

Conflicts apply ONLY to ENV bot entries.

Error Handling (Strict)
Condition	Behavior
Invalid env	Slack error
Bad Jira	Slack error
Invalid duration	Slack error
Conflict	Slack error
Calendar failure	Graceful message
Invalid signature	HTTP 401
Parsing error	Usage hint
Logging

Log all commands

Log conflicts

Log API failures

Structured logs (JSON preferred)

Access Control (Future-ready)

Later addition: restrict users by Slack channel / group

Design must allow:

Allowlist users

Admin overrides

Stability & Safety

Strict parsing of calendar titles

Ignore malformed titles

Reject unknown envs

Normalize all input before comparison

Enforce UTC or a single timezone globally

Commands Future-Ready

Include extensible architecture for:

/env-cancel

/env-extend

Admin override

Duration changes

Permission tiers

(These are not required in v1, but code must be modular.)

Architecture
cmd/server/main.go
internal/slack/
    handler.go
    verify.go
internal/calendar/
    client.go
    booking.go
internal/config/
    config.go

Runtime ENV Variables
SLACK_SIGNING_SECRET
SLACK_BOT_TOKEN
GCAL_ENV_CALENDAR_ID
GOOGLE_APPLICATION_CREDENTIALS

Acceptance Criteria

✅ Slash command works
✅ Conflicts blocked
✅ Title enforced
✅ Jira validated
✅ Max duration enforced
✅ Booking view works
✅ Slack name resolved
✅ Calendar is source of truth
✅ Bot-tagged events only
✅ Chi router
✅ No DB
✅ Logging enabled

If you want, I can next generate:

✅ .env.example
✅ Developer setup README
✅ Google service account steps
✅ Infra checklist
✅ Postman tests
✅ Example production deployment