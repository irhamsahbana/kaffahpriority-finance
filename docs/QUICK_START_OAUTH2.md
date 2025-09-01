# Quick Start: Dropbox OAuth2 Setup

## TL;DR - Get Started in 5 Minutes

### 1. Add App Credentials to .env

```bash
# Get these from https://www.dropbox.com/developers/apps
DROPBOX_APP_KEY=your_app_key_here
DROPBOX_APP_SECRET=your_app_secret_here
```

### 2. Generate Authorization URL

```bash
./kaffah-finance dropbox-oauth --action=auth
```

### 3. Authorize in Browser

Open the URL from step 2, authorize your app, copy the authorization code.

### 4. Exchange Code for Tokens

```bash
./kaffah-finance dropbox-oauth --action=exchange --code=YOUR_AUTH_CODE
```

### 5. Add Generated Tokens to .env

```bash
DROPBOX_ACCESS_TOKEN=your_generated_access_token
DROPBOX_REFRESH_TOKEN=your_generated_refresh_token
```

### 6. Test Backup

```bash
./kaffah-finance cronjob --once
```

## What This Solves

- ✅ **No more 4-hour token expiration**: Tokens refresh automatically
- ✅ **Zero code changes**: Existing backup code works unchanged  
- ✅ **Production ready**: Handles token refresh in long-running services
- ✅ **Thread safe**: Multiple backup jobs can run safely

## Key Environment Variables

```bash
# Required for OAuth2
DROPBOX_APP_KEY=ovpvlb81xin84xn              # From Dropbox Console
DROPBOX_APP_SECRET=your_secret               # From Dropbox Console  
DROPBOX_ACCESS_TOKEN=generated_token         # Generated via OAuth
DROPBOX_REFRESH_TOKEN=generated_refresh      # Generated via OAuth

# Optional
DROPBOX_REDIRECT_URL=http://localhost:8080/auth/dropbox/callback  # Defaults shown
```

## Production Deployment

Your backup cronjob now works indefinitely:

```bash
# Schedule daily backups at 23:59:59 Asia/Makassar
./kaffa-finance cronjob --schedule=daily

# Or run as systemd service
sudo systemctl enable kaffah-backup.service
sudo systemctl start kaffah-backup.service
```

## Monitoring

Watch for automatic token refresh in logs:

```bash
tail -f logs/app.log | grep "access token refreshed successfully"
```

## Need Help?

- See [DROPBOX_OAUTH2.md](./DROPBOX_OAUTH2.md) for detailed documentation
- Check [BACKUP_CRONJOB.md](./BACKUP_CRONJOB.md) for backup configuration
