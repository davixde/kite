package database

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"

	"github.com/kitecloud/kite/kite-service/internal/config"
	"github.com/kitecloud/kite/kite-service/internal/db/s3"
	"github.com/kitecloud/kite/kite-service/internal/logging"
)

type DatabaseBackupOpts struct {
	Operation string
	Name      string
}

func RunBackup(ctx context.Context, db string, opts DatabaseBackupOpts) error {
	if db != "postgres" {
		return fmt.Errorf("database %s is not supported for backup", db)
	}

	if opts.Operation != "create" {
		return fmt.Errorf("operation %s is not supported for backup", opts.Operation)
	}

	cfg, err := config.LoadConfig(".")
	if err != nil {
		return fmt.Errorf("Failed to load server config: %v", err)
	}

	logging.SetupLogger(cfg.Logging)

	s3, err := s3.New(cfg.Database.S3)
	if err != nil {
		slog.With("error", err).Error("Failed to create s3 client")
		return fmt.Errorf("failed to create s3 client: %w", err)
	}

	slog.Info("Creating database backup", "database", db, "operation", opts.Operation)

	tmpFile, err := os.CreateTemp("", "xvault-pg-backup-*.tar")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	slog.Info("Creating database dump", "file", tmpFile.Name())

	pgConfig := cfg.Database.Postgres

	cmd := exec.Command("pg_dump",
		"--host="+pgConfig.Host,
		"--username="+pgConfig.User,
		"--port="+fmt.Sprintf("%d", pgConfig.Port),
		"--dbname="+pgConfig.DBName,
		"--file="+tmpFile.Name(),
		"--format=tar",
	)

	if pgConfig.Password != "" {
		cmd.Env = append(os.Environ(), "PGPASSWORD="+pgConfig.Password)
	}

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create database dump: %w\noutput: %s", err, output)
	}

	stat, err := tmpFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file size: %w", err)
	}

	slog.Info("Successfully created database dump", "file", tmpFile.Name(), "size", stat.Size())

	// Create a temporary gzipped file
	gzipFile, err := os.CreateTemp("", "xvault-pg-backup-*.tar.gz")
	if err != nil {
		return fmt.Errorf("failed to create temporary gzip file: %w", err)
	}
	defer os.Remove(gzipFile.Name())

	// Create gzip writer
	gzipWriter := gzip.NewWriter(gzipFile)

	// Copy data from tmpFile to gzip writer
	if _, err := io.Copy(gzipWriter, tmpFile); err != nil {
		gzipWriter.Close()
		return fmt.Errorf("failed to compress backup file: %w", err)
	}

	if err := gzipWriter.Close(); err != nil {
		return fmt.Errorf("failed to finalize gzip compression: %w", err)
	}

	// Get size of compressed file
	if _, err := gzipFile.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek gzip file: %w", err)
	}

	gzipStat, err := gzipFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get compressed file size: %w", err)
	}

	slog.Info("Storing compressed database backup", "name", opts.Name, "original_size", stat.Size(), "compressed_size", gzipStat.Size())

	err = s3.StoreDBBackup(ctx, db, opts.Name, gzipStat.Size(), gzipFile)
	if err != nil {
		return fmt.Errorf("failed to store db backup: %w", err)
	}

	slog.Info("Successfully stored compressed db backup", "name", opts.Name, "original_size", stat.Size(), "compressed_size", gzipStat.Size())

	return nil
}
