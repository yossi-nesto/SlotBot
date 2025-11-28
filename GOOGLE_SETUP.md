# Google Calendar Setup Guide

This guide will walk you through setting up Google Calendar API access for SlotBot.

## Step 1: Create a Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Click **"Select a project"** → **"New Project"**
3. Enter:
   - **Project name**: `SlotBot` (or your preferred name)
   - **Organization**: (optional)
4. Click **"Create"**

## Step 2: Enable Google Calendar API

1. In your project, go to **"APIs & Services"** → **"Library"**
2. Search for **"Google Calendar API"**
3. Click on it and click **"Enable"**

## Step 3: Create a Service Account

1. Go to **"APIs & Services"** → **"Credentials"**
2. Click **"Create Credentials"** → **"Service Account"**
3. Enter:
   - **Service account name**: `slotbot-calendar`
   - **Service account ID**: (auto-generated)
   - **Description**: `SlotBot calendar access`
4. Click **"Create and Continue"**
5. For **"Grant this service account access to project"**:
   - Skip this (click **"Continue"**)
6. For **"Grant users access to this service account"**:
   - Skip this (click **"Done"**)

## Step 4: Create Service Account Key

1. In the **"Credentials"** page, find your service account in the list
2. Click on the service account email
3. Go to the **"Keys"** tab
4. Click **"Add Key"** → **"Create new key"**
5. Choose **"JSON"**
6. Click **"Create"**
7. A JSON file will download - this is your `credentials.json`
8. Move this file to your SlotBot project directory:
   ```bash
   mv ~/Downloads/slotbot-*.json /path/to/SlotBot/credentials.json
   ```

## Step 5: Share Calendar with Service Account

1. Open the JSON file and find the `client_email` field (looks like `slotbot-calendar@PROJECT_ID.iam.gserviceaccount.com`)
2. Go to [Google Calendar](https://calendar.google.com/)
3. Find your calendar in the left sidebar (the one you want to use for SlotBot)
4. Click the three dots next to it → **"Settings and sharing"**
5. Scroll down to **"Share with specific people or groups"**
6. Click **"Add people and groups"**
7. Paste the service account email
8. Set permissions to **"Make changes to events"**
9. Click **"Send"**

## Step 6: Get Calendar ID

1. In the same **"Settings and sharing"** page
2. Scroll down to **"Integrate calendar"**
3. Copy the **"Calendar ID"** (looks like `c_abc123...@group.calendar.google.com` or just your email)
4. Add to your `.env` file:
   ```
   GCAL_CALENDAR_ID=your_calendar_id_here
   ```

## Step 7: Update .env File

Your `.env` should now have:

```bash
GOOGLE_APPLICATION_CREDENTIALS=credentials.json
GCAL_CALENDAR_ID=c_e2461e84e3a7f5dce6fa0d24a2e8f7c809a9aea23843ad3c1798e17dec4c43a0@group.calendar.google.com
```

## Step 8: Test the Connection

Run your bot and try creating a booking. Check your Google Calendar to see if the event appears!

## Troubleshooting

### "Failed to create calendar client" error
- Verify `credentials.json` exists in the project directory
- Check that the JSON file is valid
- Ensure the Calendar API is enabled in your Google Cloud project

### "Failed to check calendar" error
- Verify the calendar is shared with the service account email
- Check that the service account has "Make changes to events" permission
- Confirm the `GCAL_CALENDAR_ID` is correct

### Events not appearing in calendar
- Double-check the calendar ID
- Verify the service account has write permissions
- Check the server logs for errors

## Security Notes

⚠️ **IMPORTANT**: Never commit `credentials.json` to git! It's already in `.gitignore`.

For production deployments:
- Store credentials as environment variables or use secret management (GCP Secret Manager, AWS Secrets Manager, etc.)
- Rotate service account keys periodically
- Use least-privilege permissions
