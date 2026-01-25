package cmd

import (
	"bytes"
	"codebase-app/internal/adapter"
	"codebase-app/internal/infrastructure/config"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/go-co-op/gocron/v2"
	"github.com/rs/zerolog/log"
)

// RunCronjob starts the gocron scheduler for automated backup jobs
func RunCronjob(cmd *flag.FlagSet, args []string) {

	// Parse command line arguments for schedule configuration
	schedule := cmd.String("schedule", "daily", "Backup schedule: daily, hourly, weekly, or custom cron expression")
	runOnce := cmd.Bool("once", false, "Run backup once and exit (for testing)")
	if err := cmd.Parse(args); err != nil {
		log.Error().Err(err).Msg("failed to parse args")
		return
	}

	log.Info().Str("schedule", *schedule).Bool("runOnce", *runOnce).Msg("starting backup scheduler")

	// Initialize Dropbox adapter
	adapter.Adapters.Sync(
		adapter.WithDropboxSDK(),
	)

	// If run once flag is set, execute backup immediately and exit
	if *runOnce {
		log.Info().Msg("running backup once")
		if err := executeBackupJob(); err != nil {
			log.Error().Err(err).Msg("backup job failed")
			os.Exit(1)
		}
		log.Info().Msg("backup job completed successfully")
		return
	}

	// Create scheduler with Asia/Makassar timezone
	makassarTZ, err := time.LoadLocation("Asia/Makassar")
	if err != nil {
		log.Error().Err(err).Msg("failed to load Asia/Makassar timezone")
		return
	}

	s, err := gocron.NewScheduler(gocron.WithLocation(makassarTZ))
	if err != nil {
		log.Error().Err(err).Msg("failed to create scheduler")
		return
	}
	defer func() {
		if err := s.Shutdown(); err != nil {
			log.Error().Err(err).Msg("failed to shutdown scheduler")
		}
	}()

	// Schedule backup job based on the specified schedule
	var job gocron.Job
	switch *schedule {
	case "daily":
		job, err = s.NewJob(
			gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(23, 59, 59))), // Daily at 23:59:59 Asia/Makassar
			gocron.NewTask(executeBackupJob),
			gocron.WithName("daily-backup"),
			gocron.WithTags("backup", "database", "dropbox"),
		)
	case "hourly":
		job, err = s.NewJob(
			gocron.DurationJob(time.Hour), // Every hour
			gocron.NewTask(executeBackupJob),
			gocron.WithName("hourly-backup"),
			gocron.WithTags("backup", "database", "dropbox"),
		)
	case "weekly":
		job, err = s.NewJob(
			gocron.WeeklyJob(1, gocron.NewWeekdays(time.Sunday), gocron.NewAtTimes(gocron.NewAtTime(23, 59, 59))), // Weekly on Sunday at 23:59:59
			gocron.NewTask(executeBackupJob),
			gocron.WithName("weekly-backup"),
			gocron.WithTags("backup", "database", "dropbox"),
		)
	default:
		// Try to parse as cron expression
		job, err = s.NewJob(
			gocron.CronJob(*schedule, false), // Custom cron expression
			gocron.NewTask(executeBackupJob),
			gocron.WithName("custom-backup"),
			gocron.WithTags("backup", "database", "dropbox"),
		)
	}

	if err != nil {
		log.Error().Err(err).Str("schedule", *schedule).Msg("failed to create backup job")
		return
	}

	log.Info().Str("jobID", job.ID().String()).Str("schedule", *schedule).Str("timezone", "Asia/Makassar").Msg("backup job scheduled successfully")

	// Start the scheduler
	s.Start()
	log.Info().Msg("scheduler started, waiting for jobs")

	// Set up signal handling for graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Wait for interrupt signal
	<-c
	log.Info().Msg("received shutdown signal, stopping scheduler")
}

// executeBackupJob performs the actual backup operation
func executeBackupJob() error {
	log.Info().Msg("starting backup job execution")

	// Clean up any leftover backup files from previous runs
	cleanupOldBackupFiles()

	// Execute the backup and get the generated backup file path
	backupFilePath, err := createBackup()
	if err != nil {
		log.Error().Err(err).Msg("failed to create backup")
		return err
	}

	dropboxPath := "/" + filepath.Base(backupFilePath)
	log.Info().Str("localPath", backupFilePath).Str("dropboxPath", dropboxPath).Msg("uploading file to Dropbox")

	dbx := adapter.Adapters.DropboxFiles

	// Open file and check file size
	f, err := os.Open(backupFilePath)
	if err != nil {
		log.Error().Err(err).Msg("failed to open backup file")
		return err
	}
	defer f.Close()

	// Get file size
	fileInfo, err := f.Stat()
	if err != nil {
		log.Error().Err(err).Msg("failed to get file info")
		return err
	}

	log.Info().Int64("fileSize", fileInfo.Size()).Msg("backup file size")

	// Determine upload method based on file size
	if fileInfo.Size() < 150*1024*1024 { // If less than 150MB
		err = uploadSmallFile(dbx, f, dropboxPath)
	} else { // If 150MB or larger
		err = uploadLargeFile(dbx, f, dropboxPath)
	}

	if err != nil {
		log.Error().Err(err).Msg("failed to upload backup to Dropbox")
		return err
	}

	// Cleanup local backup file after successful upload
	log.Info().Str("file", backupFilePath).Msg("cleaning up local backup file after successful upload")
	if err := os.Remove(backupFilePath); err != nil {
		log.Warn().Err(err).Str("file", backupFilePath).Msg("failed to cleanup local backup file")
		// Don't return error as upload was successful
	} else {
		log.Info().Str("file", backupFilePath).Msg("local backup file successfully deleted")
	}

	log.Info().Msg("backup job execution completed successfully")
	return nil
}

// uploadSmallFile uploads small files (<150MB) in a single request
func uploadSmallFile(dbx *adapter.DynamicDropboxClient, file *os.File, dropboxPath string) error {

	// Reset file pointer to beginning
	if _, err := file.Seek(0, 0); err != nil {
		log.Error().Err(err).Msg("failed to reset file pointer")
		return err
	}

	// Create upload argument
	uploadArg := &files.UploadArg{
		CommitInfo: files.CommitInfo{
			Path: dropboxPath,
			Mode: &files.WriteMode{Tagged: dropbox.Tagged{Tag: files.WriteModeAdd}},
		},
		ContentHash: "",
	}

	// Upload file
	_, err := dbx.Upload(uploadArg, file)
	if err != nil {
		log.Error().Err(err).Msg("failed to upload small file")
		return err
	}

	log.Info().Str("path", dropboxPath).Msg("small file successfully uploaded to Dropbox")
	return nil
}

// uploadLargeFile uploads large files (>=150MB) in chunks
func uploadLargeFile(dbx *adapter.DynamicDropboxClient, file *os.File, dropboxPath string) error {
	const chunkSize = 8 * 1024 * 1024 // 8 MB

	// Reset file pointer to beginning
	if _, err := file.Seek(0, 0); err != nil {
		log.Error().Err(err).Msg("failed to reset file pointer")
		return err
	}

	// Start upload session with first chunk
	sessionRes, err := dbx.UploadSessionStart(files.NewUploadSessionStartArg(), io.LimitReader(file, chunkSize))
	if err != nil {
		log.Error().Err(err).Msg("failed to start upload session")
		return err
	}

	log.Info().Str("sessionId", sessionRes.SessionId).Msg("upload session started")

	offset := int64(chunkSize)
	chunkNumber := 1

	// Upload remaining chunks
	for {
		// Read next chunk
		buf := make([]byte, chunkSize)
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			log.Error().Err(err).Msg("error reading file chunk")
			return err
		}
		if n == 0 {
			break // End of file
		}

		reader := bytes.NewReader(buf[:n])
		cursor := files.UploadSessionCursor{
			SessionId: sessionRes.SessionId,
			Offset:    uint64(offset),
		}

		// Upload chunk
		err = dbx.UploadSessionAppendV2(&files.UploadSessionAppendArg{Cursor: &cursor}, reader)
		if err != nil {
			log.Error().Err(err).Int("chunkNumber", chunkNumber).Msg("failed during upload session append")
			return err
		}

		log.Info().Int("chunkNumber", chunkNumber).Int64("offset", offset).Msg("chunk uploaded successfully")

		// Increase offset after successful append
		offset += int64(n)
		chunkNumber++
	}

	// Finish the upload session
	cursor := files.UploadSessionCursor{
		SessionId: sessionRes.SessionId,
		Offset:    uint64(offset),
	}
	commitInfo := files.NewCommitInfo(dropboxPath)
	commitInfo.Mode = &files.WriteMode{Tagged: dropbox.Tagged{Tag: "overwrite"}}

	_, err = dbx.UploadSessionFinish(&files.UploadSessionFinishArg{
		Cursor: &cursor,
		Commit: commitInfo,
	}, nil)

	if err != nil {
		log.Error().Err(err).Msg("failed to finish upload session")
		return err
	}

	log.Info().Str("path", dropboxPath).Int("totalChunks", chunkNumber).Msg("large file successfully uploaded to Dropbox in chunks")
	return nil
}

// createBackup creates a backup archive and returns the path to the final archive
func createBackup() (string, error) {

	// Configuration variables
	dbUser := config.Envs.Postgres.Username
	dbName := config.Envs.Postgres.Database
	dbHost := config.Envs.Postgres.Host
	dbPort := config.Envs.Postgres.Port
	dbPassword := config.Envs.Postgres.Password

	// Use environment-appropriate paths
	var backupDir, binariesDir string
	backupDir = "./backups"
	binariesDir = "./"

	timestamp := time.Now().In(time.FixedZone("Asia/Makassar", 8*3600)).Format("2006-01-02_15-04-05")

	// Ensure the backup directory exists
	if err := os.MkdirAll(backupDir, os.ModePerm); err != nil {
		log.Error().Err(err).Str("dir", backupDir).Msg("failed to create backup directory")
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Define file paths
	sqlFile := filepath.Join(backupDir, fmt.Sprintf("%s_%s.sql", dbName, timestamp))
	dumpFile := filepath.Join(backupDir, fmt.Sprintf("%s_%s.dump", dbName, timestamp))
	storageArchive := filepath.Join(backupDir, fmt.Sprintf("storage_%s.tar.gz", timestamp))
	finalArchive := filepath.Join(backupDir, fmt.Sprintf("%s_backup_%s.tar.gz", dbName, timestamp))

	log.Info().Str("sqlFile", sqlFile).Str("dumpFile", dumpFile).Str("storageArchive", storageArchive).Str("finalArchive", finalArchive).Msg("backup file paths")

	// Set PGPASSWORD environment variable for authentication
	os.Setenv("PGPASSWORD", dbPassword)
	defer os.Unsetenv("PGPASSWORD")

	// Backup database in SQL format
	log.Info().Msg("creating SQL backup")
	if err := execCommand("pg_dump", "-U", dbUser, "-h", dbHost, "-p", dbPort, "-f", sqlFile, dbName); err != nil {
		log.Error().Err(err).Msg("failed to backup database in SQL format")
		return "", fmt.Errorf("failed to backup database in SQL format: %w", err)
	}

	// Backup database in custom format
	log.Info().Msg("creating custom format backup")
	if err := execCommand("pg_dump", "-U", dbUser, "-h", dbHost, "-p", dbPort, "-Fc", "-f", dumpFile, dbName); err != nil {
		log.Error().Err(err).Msg("failed to backup database in custom format")
		return "", fmt.Errorf("failed to backup database in custom format: %w", err)
	}

	// Compress the storage folder if it exists
	storagePath := filepath.Join(binariesDir, "storage")
	if _, err := os.Stat(storagePath); err == nil {
		log.Info().Msg("compressing storage folder")
		if err := execCommand("tar", "-czvf", storageArchive, "-C", binariesDir, "storage"); err != nil {
			log.Warn().Err(err).Msg("failed to compress storage folder, continuing without it")
			// Don't return error, continue without storage backup
			storageArchive = ""
		}
	} else {
		log.Info().Msg("storage folder not found, skipping storage backup")
		storageArchive = ""
	}

	// Combine all backup files into a single archive
	log.Info().Msg("creating final archive")
	archiveArgs := []string{"-czvf", finalArchive, "-C", backupDir, filepath.Base(sqlFile), filepath.Base(dumpFile)}
	if storageArchive != "" {
		archiveArgs = append(archiveArgs, filepath.Base(storageArchive))
	}

	if err := execCommand("tar", archiveArgs...); err != nil {
		log.Error().Err(err).Msg("failed to create final archive")
		return "", fmt.Errorf("failed to create final archive: %w", err)
	}

	// Remove temporary files
	log.Info().Msg("cleaning up temporary files")
	os.Remove(sqlFile)
	os.Remove(dumpFile)
	if storageArchive != "" {
		os.Remove(storageArchive)
	}

	// Log the backup completion
	logFile, err := os.OpenFile(filepath.Join(backupDir, "backup_log.txt"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Warn().Err(err).Msg("failed to open log file")
	} else {
		defer logFile.Close()
		logEntry := fmt.Sprintf("[%s] Backup completed and archived at %s\n", timestamp, finalArchive)
		if _, err := logFile.WriteString(logEntry); err != nil {
			log.Warn().Err(err).Msg("failed to write backup log")
		}
	}

	log.Info().Str("archive", finalArchive).Msg("backup archive created successfully")
	return finalArchive, nil
}

// execCommand runs a command with the given arguments and logs the output
func execCommand(name string, args ...string) error {

	// Log command execution (without sensitive information)
	logArgs := make([]string, len(args))
	copy(logArgs, args)
	// Hide potential sensitive database connection parameters
	for i := range logArgs {
		if i > 0 && (args[i-1] == "-U" || args[i-1] == "-h" || args[i-1] == "-p") {
			logArgs[i] = "****"
		}
	}

	log.Info().Str("command", name).Strs("args", logArgs).Msg("executing command")

	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Error().Err(err).Str("command", name).Msg("command execution failed")
		return err
	}

	log.Info().Str("command", name).Msg("command executed successfully")
	return nil
}

// cleanupOldBackupFiles removes any leftover backup files from previous runs
func cleanupOldBackupFiles() {
	backupDir := "./backups"

	// Check if backup directory exists
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return // No backup directory, nothing to clean
	}

	// Read directory contents
	files, err := os.ReadDir(backupDir)
	if err != nil {
		log.Warn().Err(err).Str("dir", backupDir).Msg("failed to read backup directory")
		return
	}

	// Remove backup files (but keep backup_log.txt)
	cleanedCount := 0
	for _, file := range files {
		if file.IsDir() {
			continue // Skip directories
		}

		filename := file.Name()
		// Skip log files and hidden files
		if filename == "backup_log.txt" || filename[0] == '.' {
			continue
		}

		// Remove backup files (typically .tar.gz, .sql, .dump files)
		filePath := filepath.Join(backupDir, filename)
		if err := os.Remove(filePath); err != nil {
			log.Warn().Err(err).Str("file", filePath).Msg("failed to remove old backup file")
		} else {
			log.Info().Str("file", filePath).Msg("removed old backup file")
			cleanedCount++
		}
	}

	if cleanedCount > 0 {
		log.Info().Int("filesRemoved", cleanedCount).Msg("cleanup completed")
	} else {
		log.Info().Msg("no old backup files found to clean")
	}
}
