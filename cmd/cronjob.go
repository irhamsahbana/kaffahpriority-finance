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
	"path/filepath"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/rs/zerolog/log"
)

// RunCronjob runs the file upload process depending on file size
func RunCronjob(cmd *flag.FlagSet, args []string) {
	adapter.Adapters.Sync(
		adapter.WithDropboxSDK(),
	)

	// localFilePath := "./x2.zip" // Ganti dengan path file Anda
	// dropboxPath := "/x2.zip"    // Path di Dropbox

	// Execute the backup and get the generated backup file path
	backupFilePath, err := createBackup()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create backup")
		return
	}

	dropboxPath := "/" + filepath.Base(backupFilePath)
	log.Info().Str("dropboxPath", dropboxPath).Msg("Uploading file to Dropbox")

	// dbx := adapter.Adapters.DropboxFiles

	// // Buka file dan cek ukuran file
	// // f, err := os.Open(localFilePath)
	// f, err := os.Open(backupFilePath)
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("Error while opening file")
	// }
	// defer f.Close()

	// // Dapatkan ukuran file
	// fileInfo, err := f.Stat()
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("Error while getting file info")
	// }

	// // Tentukan metode upload berdasarkan ukuran file
	// if fileInfo.Size() < 150*1024*1024 { // Jika kurang dari 150MB
	// 	uploadSmallFile(dbx, f, dropboxPath)
	// } else { // Jika lebih dari atau sama dengan 150MB
	// 	uploadLargeFile(dbx, f, dropboxPath)
	// }
}

// uploadSmallFile uploads small files (<150MB) in a single request
func uploadSmallFile(dbx files.Client, file *os.File, dropboxPath string) {
	// Buat objek UploadArg dengan embedding dari CommitInfo
	uploadArg := &files.UploadArg{
		CommitInfo: files.CommitInfo{
			Path: dropboxPath,
			Mode: &files.WriteMode{Tagged: dropbox.Tagged{Tag: "overwrite"}},
		},
		// Jika Anda ingin menambahkan ContentHash, tentukan nilainya di sini, atau kosongkan jika tidak diperlukan
		ContentHash: "",
	}

	// Gunakan uploadArg sebagai parameter pertama
	_, err := dbx.Upload(uploadArg, file)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upload small file")
	} else {
		log.Info().Msg("Small file successfully uploaded to Dropbox")
	}
}

// uploadLargeFile uploads large files (>=150MB) in chunks
func uploadLargeFile(dbx files.Client, file *os.File, dropboxPath string) {
	const chunkSize = 8 * 1024 * 1024 // 8 MB
	sessionRes, err := dbx.UploadSessionStart(files.NewUploadSessionStartArg(), io.LimitReader(file, chunkSize))
	if err != nil {
		log.Error().Err(err).Msg("Failed to start upload session")
		return
	}

	offset := int64(chunkSize)
	for {
		// Read next chunk
		buf := make([]byte, chunkSize)
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			log.Error().Err(err).Msg("Error reading file chunk")
			return
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
			log.Error().Err(err).Msg("Failed during upload session append")
			return
		}

		// Increase offset after successful append
		offset += int64(n)
	}

	// Finish the upload session with the last chunk
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
		log.Error().Err(err).Msg("Failed to finish upload session")
	} else {
		log.Info().Msg("Large file successfully uploaded to Dropbox in chunks")
	}
}

// createBackup creates a backup archive and returns the path to the final archive
func createBackup() (string, error) {
	// Configuration variables
	dbUser := config.Envs.Postgres.Username
	dbName := config.Envs.Postgres.Database
	dbHost := config.Envs.Postgres.Host
	// backupDir := "/home/kalladig/backups"
	// binariesDir := "/home/kalladig/repositories/binaries/digihub-production"
	backupDir := "d:/backup-test"
	binariesDir := "d:/src/freelance/digihub"
	timestamp := time.Now().In(time.FixedZone("Asia/Makassar", 8*3600)).Format("2006-01-02_15-04-05")

	// Ensure the backup directory exists
	if err := os.MkdirAll(backupDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Define file paths
	sqlFile := filepath.Join(backupDir, fmt.Sprintf("%s_%s.sql", dbName, timestamp))
	dumpFile := filepath.Join(backupDir, fmt.Sprintf("%s_%s.dump", dbName, timestamp))
	storageArchive := filepath.Join(backupDir, fmt.Sprintf("storage_%s.tar.gz", timestamp))
	finalArchive := filepath.Join(backupDir, fmt.Sprintf("%s_backup_%s.tar.gz", dbName, timestamp))

	log.Info().Str("sqlFile", sqlFile).Str("dumpFile", dumpFile).Str("storageArchive", storageArchive).Str("finalArchive", finalArchive).Msg("Backup files")

	// Backup database in SQL format
	if err := execCommand("pg_dump", "-U", dbUser, "-h", dbHost, "-f", sqlFile, dbName); err != nil {
		return "", fmt.Errorf("failed to backup database in SQL format: %w", err)
	}

	// Backup database in custom format
	if err := execCommand("pg_dump", "-U", dbUser, "-h", dbHost, "-Fc", "-f", dumpFile, dbName); err != nil {
		return "", fmt.Errorf("failed to backup database in custom format: %w", err)
	}

	// Compress the storage folder
	if err := execCommand("tar", "-czvf", storageArchive, "-C", binariesDir, "storage"); err != nil {
		return "", fmt.Errorf("failed to compress storage folder: %w", err)
	}

	// Combine all backup files into a single archive
	if err := execCommand("tar", "-czvf", finalArchive, "-C", backupDir, filepath.Base(sqlFile), filepath.Base(dumpFile), filepath.Base(storageArchive)); err != nil {
		return "", fmt.Errorf("failed to create final archive: %w", err)
	}

	// Remove temporary files
	os.Remove(sqlFile)
	os.Remove(dumpFile)
	os.Remove(storageArchive)

	// Log the backup completion
	logFile, err := os.OpenFile(filepath.Join(backupDir, "backup_log.txt"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to open log file: %w", err)
	}
	defer logFile.Close()
	logFile.WriteString(fmt.Sprintf("Backup completed and archived at %s\n", timestamp))

	return finalArchive, nil
}

// execCommand runs a command with the given arguments and logs the output
func execCommand(name string, args ...string) error {
	// log command execution
	log.Info().Str("command", name).Strs("args", args).Msg("Executing command")

	c := exec.Command(name, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
