# Slack Bot Setup Guide

This guide will walk you through creating and configuring your Slack bot for SlotBot.

## Quick Setup (Using Manifest - Recommended)

1. Go to [https://api.slack.com/apps](https://api.slack.com/apps)
2. Click **"Create New App"**
3. Choose **"From an app manifest"**
4. Select your workspace
5. Choose **"YAML"** tab
6. Copy and paste the contents of `slack-manifest.yml` from this repository
7. Click **"Next"** → **"Create"**
8. **IMPORTANT**: Update the slash command URLs:
   - Go to **"Slash Commands"** in the left sidebar
   - Edit each command and replace `https://YOUR_DOMAIN` with your actual URL (ngrok or production)
9. Go to **"Install App"** → **"Install to Workspace"** → **"Allow"**
10. Copy your **Bot User OAuth Token** (starts with `xoxb-`) and **Signing Secret**
11. Add them to your `.env` file

That's it! Skip to **Step 4** below.

---

## Manual Setup (Alternative)

If you prefer to set up manually, follow these steps:

## Step 1: Create a Slack App

1. Go to [https://api.slack.com/apps](https://api.slack.com/apps)
2. Click **"Create New App"**
3. Choose **"From scratch"**
4. Enter:
   - **App Name**: `SlotBot` (or your preferred name)
   - **Workspace**: Select your workspace
5. Click **"Create App"**

## Step 2: Configure Slash Commands

1. In your app settings, go to **"Slash Commands"** (in the left sidebar under "Features")
2. Click **"Create New Command"**

### Command 1: `/env-book`
- **Command**: `/env-book`
- **Request URL**: `https://YOUR_DOMAIN/slack/book` (you'll update this later with ngrok or your deployment URL)
- **Short Description**: `Book an environment`
- **Usage Hint**: `<env> <service> <jira> [start] [duration]`
- Click **"Save"**

### Command 2: `/env-next`
- **Command**: `/env-next`
- **Request URL**: `https://YOUR_DOMAIN/slack/next`
- **Short Description**: `Find next available slot`
- **Usage Hint**: `<env> <service> [duration]`
- Click **"Save"**

### Command 3: `/env-bookings`
- **Command**: `/env-bookings`
- **Request URL**: `https://YOUR_DOMAIN/slack/bookings`
- **Short Description**: `List current bookings`
- **Usage Hint**: `[env]`
- Click **"Save"**

## Step 3: Get Your Credentials

### Signing Secret
1. Go to **"Basic Information"** (in the left sidebar under "Settings")
2. Scroll down to **"App Credentials"**
3. Copy the **"Signing Secret"**
4. Add to your `.env` file:
   ```
   SLACK_SIGNING_SECRET=your_signing_secret_here
   ```

### Bot Token
1. Go to **"OAuth & Permissions"** (in the left sidebar under "Features")
2. Scroll down to **"Scopes"** → **"Bot Token Scopes"**
3. Add the following scopes:
   - `commands` (for slash commands)
   - `chat:write` (if you want to send messages)
4. Scroll to the top and click **"Install to Workspace"**
5. Click **"Allow"**
6. Copy the **"Bot User OAuth Token"** (starts with `xoxb-`)
7. Add to your `.env` file:
   ```
   SLACK_BOT_TOKEN=xoxb-your-bot-token-here
   ```

## Step 4: Expose Your Local Server (for Testing)

### Option A: Using ngrok (Recommended for local testing)

1. Install ngrok: [https://ngrok.com/download](https://ngrok.com/download)
2. Start your SlotBot server:
   ```bash
   make run
   ```
3. In another terminal, start ngrok:
   ```bash
   ngrok http 8080
   ```
4. Copy the HTTPS URL (e.g., `https://abc123.ngrok.io`)
5. Go back to your Slack App → **"Slash Commands"**
6. Update each command's **Request URL**:
   - `/env-book` → `https://abc123.ngrok.io/slack/book`
   - `/env-next` → `https://abc123.ngrok.io/slack/next`
   - `/env-bookings` → `https://abc123.ngrok.io/slack/bookings`

### Option B: Deploy to Production

Deploy to your preferred platform (GCP Cloud Run, Heroku, etc.) and use the production URL instead of ngrok.

## Step 5: Test Your Bot

1. Go to any channel in your Slack workspace
2. Try the commands:
   ```
   /env-book staging auth PROJ-123
   /env-next staging auth
   /env-bookings
   ```

## Troubleshooting

### "dispatch_failed" error
- Make sure your server is running
- Verify ngrok is forwarding to port 8080
- Check that the Request URLs in Slack match your ngrok URL

### "Invalid signature" error
- Verify `SLACK_SIGNING_SECRET` in your `.env` matches the one in Slack
- Make sure you're using the correct signing secret (not the client secret)

### 3. "Calendar not configured"
- Make sure you have set up OAuth correctly (see `OAUTH_SETUP.md`)
- Check if `oauth_credentials.json` exists in the root directory
- Check the server logs for specific errors

## Next Steps

- Set up Google Calendar OAuth (see `OAUTH_SETUP.md`)
- Run the server locally with `make run`
- Test the commands in Slack

## Complete .env Example

```bash
SLACK_SIGNING_SECRET=abc123def456...
SLACK_BOT_TOKEN=xoxb-123456789...
GCAL_CALENDAR_ID=c_e2461e84e3a7f5dce6fa0d24a2e8f7c809a9aea23843ad3c1798e17dec4c43a0@group.calendar.google.com
DEFAULT_TIMEZONE=America/New_York
PORT=8080

## Next Steps

- Configure your production deployment
- Add monitoring and logging
