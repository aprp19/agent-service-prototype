package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"agent-service-prototype/internal/app/agent-service-prototype/repository"
	"agent-service-prototype/internal/config"
	"agent-service-prototype/pkg/logger"
	"agent-service-prototype/pkg/setup"
)

const (
	StepDownloadBundle  = "DOWNLOAD_BUNDLE"
	StepExtractBundle   = "EXTRACT_BUNDLE"
	StepVerifyChecksum  = "VERIFY_CHECKSUM"
	StepParseManifest   = "PARSE_MANIFEST"
	StepConnectDB       = "CONNECT_DB"
	StepLockDB          = "LOCK_DB"
	StepApplyBaseline   = "APPLY_BASELINE"
	StepApplyMigrations = "APPLY_MIGRATIONS"
	StepPostCheck       = "POST_CHECK"
)

const (
	StatusRunning = "running"
	StatusSuccess = "success"
	StatusFailed  = "failed"
)

type InstallationResult struct {
	Success       bool
	Step          string
	Error         string
	SchemaVersion string
	Duration      time.Duration
}

type RunStatus struct {
	Status     string
	Step       string
	Error      string
	StartedAt  time.Time
	FinishedAt time.Time
}

type Service struct {
	cfg    *config.Config
	repo   *repository.Repository
	mu     sync.Mutex
	status *RunStatus
}

func NewService(cfg *config.Config, repo *repository.Repository) *Service {
	return &Service{
		cfg:  cfg,
		repo: repo,
	}
}

func (s *Service) TryStart() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.status != nil && s.status.Status == StatusRunning {
		return false
	}
	s.status = &RunStatus{
		Status:    StatusRunning,
		Step:      "INITIALIZING",
		StartedAt: time.Now(),
	}
	return true
}

func (s *Service) GetStatus() *RunStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.status == nil {
		return nil
	}
	cp := *s.status
	return &cp
}

func (s *Service) RunInstallation(ctx context.Context) *InstallationResult {
	start := time.Now()
	result := s.doInstallation(ctx)
	now := time.Now()
	result.Duration = now.Sub(start)

	s.mu.Lock()
	if result.Success {
		s.status = &RunStatus{
			Status:     StatusSuccess,
			StartedAt:  start,
			FinishedAt: now,
		}
	} else {
		s.status = &RunStatus{
			Status:     StatusFailed,
			Step:       result.Step,
			Error:      result.Error,
			StartedAt:  start,
			FinishedAt: now,
		}
	}
	s.mu.Unlock()

	return result
}

func (s *Service) doInstallation(ctx context.Context) *InstallationResult {
	db := s.repo.DB()
	dbURL := s.cfg.DatabaseURL()
	bundleURL := s.cfg.HTTP.BundleURL
	if dbURL == "" || bundleURL == "" {
		return &InstallationResult{Step: StepConnectDB, Error: "DB_URL or BUNDLE_URL is not configured"}
	}

	workDir := s.cfg.HTTP.WorkDir
	force, _ := strconv.ParseBool(s.cfg.HTTP.Force)
	skipSmoke, _ := strconv.ParseBool(s.cfg.HTTP.SkipSmoke)
	advisoryKey, _ := strconv.ParseInt(s.cfg.HTTP.AdvisoryLockKey, 10, 64)
	if advisoryKey == 0 {
		advisoryKey = 987654321
	}

	if err := os.MkdirAll(workDir, 0755); err != nil {
		return &InstallationResult{Step: StepDownloadBundle, Error: fmt.Sprintf("failed to create work dir: %v", err)}
	}

	s.updateStep(StepDownloadBundle)
	bundlePath := filepath.Join(workDir, "db-bundle.zip")
	if err := setup.DownloadBundle(ctx, bundleURL, bundlePath); err != nil {
		return &InstallationResult{Step: StepDownloadBundle, Error: err.Error()}
	}

	s.updateStep(StepExtractBundle)
	extractDir := filepath.Join(workDir, "bundle")
	if err := os.RemoveAll(extractDir); err != nil {
		return &InstallationResult{Step: StepExtractBundle, Error: fmt.Sprintf("failed to clean extract dir: %v", err)}
	}
	if err := setup.ExtractZip(bundlePath, extractDir); err != nil {
		return &InstallationResult{Step: StepExtractBundle, Error: err.Error()}
	}

	baseDir, err := setup.ResolveBaseDir(extractDir)
	if err != nil {
		return &InstallationResult{Step: StepParseManifest, Error: err.Error()}
	}

	s.updateStep(StepVerifyChecksum)
	checksums, err := setup.LoadChecksums(baseDir)
	if err != nil {
		return &InstallationResult{Step: StepVerifyChecksum, Error: err.Error()}
	}
	if err := setup.VerifyChecksums(baseDir, checksums); err != nil {
		return &InstallationResult{Step: StepVerifyChecksum, Error: err.Error()}
	}

	s.updateStep(StepParseManifest)
	manifest, err := setup.LoadManifest(baseDir)
	if err != nil {
		return &InstallationResult{Step: StepParseManifest, Error: err.Error()}
	}

	s.updateStep(StepConnectDB)
	conn, err := db.Conn(ctx)
	if err != nil {
		return &InstallationResult{Step: StepConnectDB, Error: fmt.Sprintf("failed to acquire connection: %v", err)}
	}
	defer conn.Close()

	if err := conn.PingContext(ctx); err != nil {
		return &InstallationResult{Step: StepConnectDB, Error: fmt.Sprintf("failed to ping database: %v", err)}
	}

	s.updateStep(StepLockDB)
	if _, err := conn.ExecContext(ctx, fmt.Sprintf("SELECT pg_advisory_lock(%d)", advisoryKey)); err != nil {
		return &InstallationResult{Step: StepLockDB, Error: fmt.Sprintf("failed to acquire advisory lock: %v", err)}
	}
	defer func() {
		if _, err := conn.ExecContext(context.Background(), fmt.Sprintf("SELECT pg_advisory_unlock(%d)", advisoryKey)); err != nil {
			logger.Error().Err(err).Msg("Failed to release advisory lock")
		} else {
			logger.Info().Msg("Advisory lock released")
		}
	}()
	logger.Info().Int64("key", advisoryKey).Msg("Advisory lock acquired")

	s.updateStep(StepApplyBaseline)
	fresh, err := s.repo.IsFreshDB(ctx, conn)
	if err != nil {
		return &InstallationResult{Step: StepApplyBaseline, Error: fmt.Sprintf("failed to detect DB state: %v", err)}
	}

	if fresh {
		logger.Info().Msg("Fresh database detected, applying baseline...")
		baselineSQL, err := os.ReadFile(filepath.Join(baseDir, string(manifest.Baseline)))
		if err != nil {
			return &InstallationResult{Step: StepApplyBaseline, Error: fmt.Sprintf("failed to read baseline: %v", err)}
		}
		if _, err := conn.ExecContext(ctx, string(baselineSQL)); err != nil {
			return &InstallationResult{Step: StepApplyBaseline, Error: fmt.Sprintf("failed to apply baseline: %v", err)}
		}
		if err := s.repo.EnsureMigrationsTable(ctx, conn); err != nil {
			return &InstallationResult{Step: StepApplyBaseline, Error: fmt.Sprintf("failed to ensure migrations table: %v", err)}
		}
		logger.Info().Msg("Baseline applied successfully")
	}

	s.updateStep(StepApplyMigrations)
	var lastVersion string
	for _, mig := range manifest.Migrations {
		applied, err := s.repo.GetMigrationRecord(ctx, conn, mig.Version)
		if err != nil {
			return &InstallationResult{Step: StepApplyMigrations, Error: fmt.Sprintf("failed to check migration %s: %v", mig.Version, err)}
		}

		migrationSQL, err := os.ReadFile(filepath.Join(baseDir, mig.File))
		if err != nil {
			return &InstallationResult{Step: StepApplyMigrations, Error: fmt.Sprintf("failed to read migration %s: %v", mig.Version, err)}
		}
		fileChecksum := setup.SHA256Hex(migrationSQL)

		if applied != nil {
			if applied.Success && !force {
				logger.Info().Str("version", mig.Version).Msg("Migration already applied, skipping")
				lastVersion = mig.Version
				continue
			}
			if applied.Checksum != fileChecksum {
				return &InstallationResult{
					Step:  StepApplyMigrations,
					Error: fmt.Sprintf("checksum mismatch for migration %s: recorded=%s, file=%s", mig.Version, applied.Checksum, fileChecksum),
				}
			}
		}

		logger.Info().Str("version", mig.Version).Str("name", mig.Name).Bool("tx", mig.Transaction).Msg("Applying migration")

		migStart := time.Now()
		var migErr error

		if mig.Transaction {
			migErr = s.repo.ExecInTransaction(ctx, conn, string(migrationSQL))
		} else {
			_, migErr = conn.ExecContext(ctx, string(migrationSQL))
		}
		migDuration := time.Since(migStart)

		rec := repository.MigrationRecord{
			Version:    mig.Version,
			Name:       mig.Name,
			Checksum:   fileChecksum,
			AppliedAt:  time.Now(),
			ExecTimeMs: migDuration.Milliseconds(),
			Success:    migErr == nil,
		}
		if migErr != nil {
			rec.ErrorMsg = migErr.Error()
		}
		if rErr := s.repo.RecordMigration(ctx, conn, rec); rErr != nil {
			logger.Error().Err(rErr).Str("version", mig.Version).Msg("Failed to record migration")
		}
		if migErr != nil {
			return &InstallationResult{Step: StepApplyMigrations, Error: fmt.Sprintf("migration %s failed: %v", mig.Version, migErr)}
		}

		lastVersion = mig.Version
		logger.Info().Str("version", mig.Version).Int64("ms", migDuration.Milliseconds()).Msg("Migration applied")
	}

	if !skipSmoke && manifest.Checks.Smoke != "" {
		s.updateStep(StepPostCheck)
		smokeSQL, err := os.ReadFile(filepath.Join(baseDir, manifest.Checks.Smoke))
		if err != nil {
			return &InstallationResult{Step: StepPostCheck, Error: fmt.Sprintf("failed to read smoke check: %v", err)}
		}
		if _, err := conn.ExecContext(ctx, string(smokeSQL)); err != nil {
			return &InstallationResult{Step: StepPostCheck, Error: fmt.Sprintf("smoke check failed: %v", err)}
		}
		logger.Info().Msg("Smoke check passed")
	}

	return &InstallationResult{Success: true, SchemaVersion: lastVersion}
}

func (s *Service) updateStep(step string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.status != nil {
		s.status.Step = step
	}
}
