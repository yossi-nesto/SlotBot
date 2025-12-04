# Service Account Setup Guide

## Overview

Service accounts are ideal for server-to-server applications like SlotBot. Unlike OAuth, they don't require interactive user authorization - perfect for automated systems.

## Step 1: Create a Service Account

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Select your project (or create a new one)
3. Navigate to **IAM & Admin** → **Service Accounts**
4. Click **Create Service Account**
5. Fill in the details:
   - **Service account name**: `slotbot-calendar`
   - **Service account ID**: Will auto-generate (e.g., `slotbot-calendar@your-project.iam.gserviceaccount.com`)
   - **Description**: "Service account for SlotBot calendar access"
6. Click **Create and Continue**
7. Skip the optional steps (no roles needed for Calendar API)
8. Click **Done**

## Step 2: Enable Google Calendar API

1. In Google Cloud Console, go to **APIs & Services** → **Library**
2. Search for **"Google Calendar API"**
3. Click on it and click **Enable**

## Step 3: Download Credentials

1. Go back to **IAM & Admin** → **Service Accounts**
2. Find your newly created service account
3. Click on the service account email
4. Go to the **Keys** tab
5. Click **Add Key** → **Create new key**
6. Select **JSON** format
7. Click **Create**
8. The JSON file will download automatically
9. **Rename** the file to `service-account.json`
10. **Move** it to your SlotBot project directory: `/Users/yossigruner/work/gits/personal/SlotBot/`

## Step 4: Share Your Calendar

This is the crucial step! The service account needs permission to access your calendar.

1. Open [Google Calendar](https://calendar.google.com/)
2. Find the calendar you want to use (or create a new one for SlotBot)
3. Click the three dots next to the calendar name → **Settings and sharing**
4. Scroll down to **Share with specific people**
5. Click **Add people**
6. **Enter the service account email** (from Step 1, looks like `slotbot-calendar@your-project.iam.gserviceaccount.com`)
7. Set permission to **"Make changes to events"**
8. Click **Send**

## Step 5: Configure SlotBot (Optional)

If you want to use a specific calendar (not your primary one):

1. Get the Calendar ID:
   - In Google Calendar settings for your calendar
   - Scroll to **Integrate calendar**
   - Copy the **Calendar ID** (looks like `abc123@group.calendar.google.com` or your email)

2. Set it in your `.env` file:
   ```bash
   GOOGLE_CALENDAR_ID=your-calendar-id@group.calendar.google.com
   ```

If you don't set this, SlotBot will use your primary calendar.

## Step 6: Test It

1. Make sure `service-account.json` is in your SlotBot directory
2. Run:
   ```bash
   make run
   ```
3. You should see:
   ```
   INF Using service account authentication
   INF Starting server port=8080
   ```
4. No authorization prompt should appear!

## Troubleshooting

### "Failed to create calendar client"

**Check the error message:**

- **"invalid character '\\n' in string literal"**: Your JSON file is corrupted. Re-download it from GCP.
- **"The caller does not have permission"**: You forgot to share the calendar with the service account email (Step 4).
- **"service-account.json not found"**: Make sure the file is in the project root directory.

### "Access Not Configured"

- Make sure you enabled the Google Calendar API (Step 2)

### Calendar operations fail

- Verify the service account email has "Make changes to events" permission
- Check that you're using the correct Calendar ID in your `.env` file

## Security Notes

- The `service-account.json` file contains sensitive credentials - keep it secure!
- It's already in `.gitignore` and won't be committed to version control
- Never share this file publicly or commit it to GitHub

## Comparison: Service Account vs OAuth

| Feature | Service Account | OAuth |
|---------|----------------|-------|
| **Setup** | Create account, share calendar | Create OAuth app, authorize |
| **Authorization** | None needed | Interactive browser flow |
| **Best for** | Server applications | User-facing apps |
| **Calendar access** | Only calendars shared with it | User's own calendars |
| **Token expiry** | Never expires | Refresh tokens can expire |

For SlotBot (a server application), **Service Account is recommended**.
