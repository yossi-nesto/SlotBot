# Quick OAuth Setup Guide

## Step 1: Get OAuth Credentials

You need to create OAuth credentials. You have two options:

### Option A: Create Your Own (Free Google Cloud Account)

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project (free)
3. Enable **Google Calendar API**:
   - Go to "APIs & Services" → "Library"
   - Search for "Google Calendar API"
   - Click "Enable"
4. Create OAuth credentials:
   - Go to "APIs & Services" → "Credentials"
   - Click "Create Credentials" → "OAuth client ID"
   - If prompted, configure OAuth consent screen:
     - Choose "External"
     - App name: "SlotBot"
     - User support email: your email
     - Developer contact: your email
     - Click "Save and Continue" through the rest
   - Back to "Create OAuth client ID":
     - Application type: **"Desktop app"**
     - Name: "SlotBot"
     - Click "Create"
5. Download the JSON file
6. Rename it to `oauth_credentials.json`
7. Place it in your SlotBot directory

### Option B: Ask Someone to Create It

If you can't create a Google Cloud project, ask a colleague/friend to do steps 1-5 above and send you the `oauth_credentials.json` file.

## Step 2: First-Time Authorization

1. Make sure `oauth_credentials.json` is in your SlotBot directory
2. Start your server:
   ```bash
   make run
   ```
3. The server will print a URL like:
   ```
   🔐 Go to the following link in your browser:
   https://accounts.google.com/o/oauth2/auth?...
   
   After authorization, paste the code here:
   ```
4. Open that URL in your browser
5. Sign in with your Google account
6. Click "Allow" to grant calendar access
7. Copy the authorization code
8. Paste it back in the terminal
9. Done! A `token.json` file will be created

## Step 3: Use Your Calendar

The bot will now use YOUR personal Google Calendar (the one you authorized with).

You can use your existing calendar or create a new one specifically for SlotBot.

## Troubleshooting

### "Access blocked: This app's request is invalid"
- Make sure you configured the OAuth consent screen
- Add your email to "Test users" in the consent screen

### "The caller does not have permission"
- Make sure you granted calendar access when authorizing
- Delete `token.json` and try authorizing again

### Token expired
- Delete `token.json` and restart the server to re-authorize

## Notes

- The `token.json` file contains your access token - keep it secure!
- Both `oauth_credentials.json` and `token.json` are in `.gitignore` (won't be committed)
- You only need to authorize once - the token will be reused
