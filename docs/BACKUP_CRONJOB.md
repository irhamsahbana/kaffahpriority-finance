# Database Backup Scheduler Documentation

## Overview

The backup scheduler automatically creates comprehensive database backups and uploads them to Dropbox using the **gocron** library for robust scheduling. It provides flexible scheduling options and can run as a standalone service or traditional cron job.

## Features

### Advanced Scheduling with gocron

- **Built-in Scheduler**: Uses gocron v2 for reliable job scheduling
- **Multiple Schedule Types**: Daily, hourly, weekly, or custom cron expressions
- **Graceful Shutdown**: Handles system signals properly
- **Job Management**: Named jobs with tags for better monitoring
- **Service Mode**: Can run as a long-running service

### Comprehensive Backup

- **PostgreSQL SQL Format**: Human-readable SQL dump
- **PostgreSQL Custom Format**: Compressed binary format for faster restore
- **Storage Files**: Includes application storage directory (if exists)
- **Single Archive**: Combines all backups into one compressed file

### Smart Upload Strategy

- **Small Files (<150MB)**: Single upload for efficiency
- **Large Files (≥150MB)**: Chunked upload with progress tracking
- **Automatic Detection**: File size determines upload method

### Robust Error Handling

- Comprehensive logging with structured error messages
- Graceful handling of missing storage directories
- Cleanup of local files after successful upload
- Masking of sensitive database credentials in logs

### Environment Awareness

- **Production & Development**: Uses `./backups` and `./` paths
- **Configuration**: Reads from existing config structure

## Prerequisites

### System Requirements

- PostgreSQL client tools (`pg_dump`)
- `tar` command for archiving
- Go application built and configured
- **gocron v2** (automatically included via go.mod)

### Environment Variables

```bash
# Required Dropbox configuration
DROPBOX_ACCESS_TOKEN=your_dropbox_access_token

# Required PostgreSQL configuration
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=your_db_user
POSTGRES_PASSWORD=your_db_password
POSTGRES_DB=your_database_name
POSTGRES_ENV=production  # or development

# Optional: Storage paths (auto-detected if not set)
```

### Dropbox Setup

1. Create a Dropbox app at <https:/>/www.dropbox.com/developers/apps>
2. Generate an access token
3. Set the `DROPBOX_ACCESS_TOKEN` environment variable

## Installation

### 1. Build the Application

```bash
cd /path/to/kaffah-priority-finance
go build -o kaffah-finance cmd/bin/main.go
```

### 2. Test the Backup

```bash
# Test manually first
./kaffah-finance cronjob --once
```

### 3. Using the Setup Script

```bash
# Run the automated setup script
./scripts/setup-backup-cronjob.sh
```

## Usage

### Command Line Options

#### Service Mode (Recommended)

```bash
# Run with built-in scheduler
./kaffah-finance cronjob --schedule=daily    # Daily at 2:00 AM
./kaffah-finance cronjob --schedule=hourly   # Every hour
./kaffah-finance cronjob --schedule=weekly   # Weekly on Sunday at 3:00 AM
./kaffah-finance cronjob --schedule="0 */6 * * *"  # Custom cron expression
```

#### One-time Execution

```bash
# Run backup once and exit (for testing)
./kaffah-finance cronjob --once
```

#### Custom Configuration

```bash
# Run with specific config
./kaffah-finance -config_path=/etc/kaffah -config_filename=production.env cronjob --schedule=daily
```

### Service Deployment

#### 1. Systemd Service (Linux - Recommended)

Create `/etc/systemd/system/kaffah-backup.service`:

```ini
[Unit]
Description=Kaffah Priority Finance Backup Scheduler
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/path/to/kaffah-finance
ExecStart=/path/to/kaffah-finance cronjob --schedule=daily
Restart=always
RestartSec=10
EnvironmentFile=/path/to/.env

[Install]
WantedBy=multi-user.target
```

Then:

```bash
sudo systemctl enable kaffah-backup.service
sudo systemctl start kaffah-backup.service
sudo systemctl status kaffah-backup.service
```

#### 2. Docker Service

```dockerfile
# Add to your docker-compose.yml
kaffah-backup:
  image: your-kaffah-image
  command: ["./kaffah-finance", "cronjob", "--schedule=daily"]
  environment:
    - DROPBOX_ACCESS_TOKEN=${DROPBOX_ACCESS_TOKEN}
    - POSTGRES_HOST=${POSTGRES_HOST}
    # ... other env vars
  restart: unless-stopped
  depends_on:
    - postgres
```

#### 3. Traditional Cron (Alternative)

```bash
# Add to crontab (crontab -e)
0 2 * * * /path/to/kaffah-finance cronjob --once >> /var/log/kaffah-backup.log 2>&1
```

### Schedule Options

#### Built-in Schedules

- `--schedule=daily`: Every day at 2:00 AM
- `--schedule=hourly`: Every hour
- `--schedule=weekly`: Every Sunday at 3:00 AM

#### Custom Cron Expressions

```bash
# Every 6 hours
--schedule="0 */6 * * *"

# Every weekday at 1 AM
--schedule="0 1 * * 1-5"

# Every first day of month at 3 AM
--schedule="0 3 1 * *"

# Every 30 minutes
--schedule="*/30 * * * *"
```

## Backup Process Flow

1. **Initialize**: Sets up Dropbox adapter and gocron scheduler
2. **Schedule Jobs**: Configures backup jobs based on schedule option
3. **Execute Backup**:
   - Dumps PostgreSQL database (SQL + custom formats)
   - Archives storage directory (if exists)
   - Combines into single compressed archive
4. **Upload to Dropbox**:
   - Determines file size
   - Uses appropriate upload method
   - Monitors upload progress
5. **Cleanup**: Removes local backup files after successful upload
6. **Logging**: Records backup completion with timestamp

## File Naming Convention

Backup files use the following naming pattern:

```bash
{database_name}_backup_{timestamp}.tar.gz
```

Example: `kaffah_finance_backup_2024-01-15_02-00-05.tar.gz`

Dropbox path: `/backups/{filename}`

## Monitoring and Management

### Service Status

```bash
# Check systemd service status
sudo systemctl status kaffah-backup.service

# View service logs
sudo journalctl -u kaffah-backup.service -f

# Restart service
sudo systemctl restart kaffah-backup.service
```

### Log Monitoring

```bash
# Application logs (structured JSON)
tail -f /var/log/kaffah-backup.log

# Check recent successful backups
grep "backup job execution completed successfully" /var/log/kaffah-backup.log

# Check for errors
grep "ERROR" /var/log/kaffah-backup.log

# Check scheduler status
grep "scheduler started" /var/log/kaffah-backup.log
```

### Key Log Messages

- `starting backup scheduler` - Scheduler service started
- `backup job scheduled successfully` - Job added to scheduler
- `starting backup job execution` - Backup job began
- `backup job execution completed successfully` - Backup completed
- `received shutdown signal` - Graceful shutdown initiated
- `failed to create backup` - Backup creation error
- `failed to upload backup to Dropbox` - Upload error
- `scheduler started, waiting for jobs` - Service ready

## Troubleshooting

### Common Issues

#### Database Connection Errors

```bash
# Check PostgreSQL connectivity
pg_dump -U $POSTGRES_USER -h $POSTGRES_HOST -l

# Verify environment variables
env | grep POSTGRES
```

#### Dropbox Upload Failures

```bash
# Verify Dropbox token
# Check token permissions and expiration
# Ensure /backups folder exists in Dropbox
```

#### Service Not Starting

```bash
# Check service configuration
sudo systemctl status kaffah-backup.service

# Check application logs
sudo journalctl -u kaffah-backup.service --no-pager

# Verify binary and permissions
ls -la /path/to/kaffah-finance
sudo -u www-data /path/to/kaffah-finance cronjob --once
```

#### Jobs Not Running

```bash
# Check if scheduler is active
grep "scheduler started" /var/log/kaffah-backup.log

# Verify schedule configuration
grep "backup job scheduled" /var/log/kaffah-backup.log

# Test manual execution
./kaffah-finance cronjob --once
```

#### Storage Directory Not Found

```bash
# This is normal and handled gracefully
# Check logs for: "storage folder not found, skipping storage backup"
# Backup will continue without storage files
```

#### Permission Issues

```bash
# Ensure backup directory is writable
chmod 755 ./backups  # or /home/backups in production

# Check service user permissions
sudo -u www-data ./kaffah-finance cronjob --once  # if running as www-data
```

### Recovery Testing

Regularly test backup restoration:

```bash
# Extract backup
tar -xzf kaffah_finance_backup_2024-01-15_02-00-05.tar.gz

# Restore database (SQL format)
psql -U user -d database < kaffah_finance_2024-01-15_02-00-05.sql

# Restore database (custom format)
pg_restore -U user -d database kaffah_finance_2024-01-15_02-00-05.dump
```

## Security Considerations

### Database Credentials

- Uses `PGPASSWORD` environment variable
- Masks sensitive parameters in logs
- Credentials are not stored in backup files

### Dropbox Security

- Use app-specific access tokens
- Regularly rotate access tokens
- Monitor Dropbox access logs

### File Permissions

- Backup files have restricted permissions
- Local files are cleaned up after upload
- Log files should have appropriate permissions

## Performance Considerations

### Backup Size Optimization

- Database dumps are compressed
- Storage files use tar.gz compression
- Regular cleanup of old Dropbox backups recommended

### Network Efficiency

- Small files: Single upload
- Large files: 8MB chunks for reliability
- Upload progress tracking for monitoring

### Resource Usage

- Temporary disk space needed (2x backup size)
- CPU usage during compression
- Network bandwidth during upload

## Maintenance

### Regular Tasks

1. **Monitor backup logs** for failures
2. **Test restore procedures** monthly
3. **Clean old backups** from Dropbox
4. **Verify backup integrity** regularly
5. **Update Dropbox tokens** before expiration

### Health Checks

```bash
# Create health check script for monitoring systems
#!/bin/bash
# /usr/local/bin/kaffah-backup-health.sh

LAST_BACKUP=$(find /backups -name "*.tar.gz" -mtime -1 2>/dev/null | wc -l)
SERVICE_STATUS=$(systemctl is-active kaffah-backup.service 2>/dev/null)

if [ "$SERVICE_STATUS" != "active" ]; then
    echo "CRITICAL: Backup service is not running"
    exit 2
fi

if [ $LAST_BACKUP -eq 0 ]; then
    echo "WARNING: No backup found in last 24 hours"
    exit 1
fi

echo "OK: Backup service running, recent backup found"
exit 0
```

### Alerts

Configure monitoring alerts for:

- Backup service failures
- Missing daily backups
- Large backup size increases
- Dropbox upload failures
- Service restart loops

### Performance Monitoring

```bash
# Monitor backup job duration
grep -E "starting backup job|backup job execution completed" /var/log/kaffah-backup.log | \
  awk '{print $1, $2, $NF}' | \
  while read start_time start_date start_msg; do
    if [[ $start_msg == *"starting"* ]]; then
      echo "Backup started at $start_time"
    fi
  done
```

## Migration from Traditional Cron

If you were using traditional system cron:

1. **Remove old cron entries**:

   ```bash
   crontab -e  # Remove old backup entries
   ```

2. **Set up service mode**:

   ```bash
   # Test new scheduler
   ./kaffah-finance cronjob --once

   # Deploy as service
   sudo systemctl enable kaffah-backup.service
   sudo systemctl start kaffah-backup.service
   ```

3. **Monitor transition**:

   ```bash
   # Watch logs for successful operation
   sudo journalctl -u kaffah-backup.service -f
   ```

## Advantages of gocron Implementation

### Reliability

- **Process Management**: Runs as a single process with job scheduling
- **Error Recovery**: Built-in error handling and job retry capabilities
- **Graceful Shutdown**: Proper cleanup on system signals

### Flexibility

- **Dynamic Scheduling**: Can be reconfigured without system cron
- **Multiple Formats**: Supports both built-in schedules and cron expressions
- **Job Metadata**: Named jobs with tags for better organization

### Monitoring

- **Structured Logging**: Detailed job execution logs
- **Job Status**: Better visibility into schedule and execution status
- **Service Integration**: Works well with systemd and container orchestration

## Integration with Monitoring

### Prometheus Metrics (Future Enhancement)

Consider adding metrics collection for:

- Backup job success/failure rates
- Backup duration times
- Backup file sizes
- Upload speeds

### Backup Retention

Configure Dropbox retention policy or implement cleanup:

```bash
# Example: Keep last 30 days of backups
# (Implementation depends on Dropbox API usage)
```
