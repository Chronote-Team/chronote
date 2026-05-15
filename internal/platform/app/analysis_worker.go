package app

import (
	"context"
	"errors"
	"log"
	"os"
	"strconv"
	"time"

	mediaapp "chronote-refactor/internal/modules/media/app"
	mediainfra "chronote-refactor/internal/modules/media/infra"
	postcardaiapp "chronote-refactor/internal/modules/postcardai/app"
	postcardaiinfra "chronote-refactor/internal/modules/postcardai/infra"
	postcardaiai "chronote-refactor/internal/modules/postcardai/infra/ai"
	postcardsapp "chronote-refactor/internal/modules/postcards/app"
	postcardsinfra "chronote-refactor/internal/modules/postcards/infra"
	platformconfig "chronote-refactor/internal/platform/config"
	platformdb "chronote-refactor/internal/platform/db"
	platforms3 "chronote-refactor/internal/platform/s3"

	"gorm.io/gorm"
)

const (
	defaultWorkerID    = "worker-1"
	defaultIdleSleep   = 2 * time.Second
	defaultErrorSleep  = 5 * time.Second
	defaultWorkerLogOK = 0
)

type analysisJobProcessor interface {
	RunNextAnalysisJob(ctx context.Context, workerID string) (*postcardaiapp.WorkerOutcome, error)
}

type WorkerOptions struct {
	WorkerID   string
	IdleSleep  time.Duration
	ErrorSleep time.Duration
	RunOnce    bool
	Logger     *log.Logger

	sleep func(context.Context, time.Duration) error
}

type AnalysisWorker struct {
	Service *postcardaiapp.Service
}

func NewAnalysisWorker() (*AnalysisWorker, error) {
	cfg, err := platformconfig.Load()
	if err != nil {
		return nil, err
	}
	return NewAnalysisWorkerFromConfig(cfg)
}

func NewAnalysisWorkerFromConfig(cfg *platformconfig.Config) (*AnalysisWorker, error) {
	database, err := platformdb.Open(cfg)
	if err != nil {
		return nil, err
	}
	service, err := NewPostcardAIService(cfg, database, nil)
	if err != nil {
		return nil, err
	}
	return &AnalysisWorker{Service: service}, nil
}

func NewPostcardAIService(cfg *platformconfig.Config, database *gorm.DB, mediaRepo mediaapp.Repository) (*postcardaiapp.Service, error) {
	analysisRepo := postcardaiinfra.NewGormRepository(database)
	postcardRepo := postcardsinfra.NewGormRepository(database)
	if mediaRepo == nil {
		mediaRepo = mediainfra.NewGormRepository(database)
	}

	deps := postcardaiapp.Dependencies{
		Jobs:      analysisRepo,
		Results:   analysisRepo,
		Postcards: postcardsapp.NewSourceAdapter(postcardRepo),
		Media:     mediaapp.NewSourceAdapter(mediaRepo),
		Enabled:   cfg.AI.Enabled,
		Model:     cfg.AI.Model,
	}
	if cfg.AI.Enabled {
		if cfg.AI.Provider == "" || cfg.AI.Provider == "openai" {
			deps.AI = postcardaiai.NewOpenAIResponsesClient(
				cfg.AI.OpenAIAPIKey,
				cfg.AI.Model,
				cfg.AI.Endpoint,
				time.Duration(cfg.AI.Timeout)*time.Second,
			)
		}
		if cfg.S3.Endpoint != "" && cfg.S3.BucketName != "" {
			s3Client, err := platforms3.NewClient(cfg)
			if err != nil {
				return nil, err
			}
			deps.Storage = platforms3.NewPresigner(s3Client, cfg.S3.BucketName)
		}
	}
	return postcardaiapp.NewService(deps), nil
}

func WorkerOptionsFromConfig(cfg *platformconfig.Config) WorkerOptions {
	return NormalizeWorkerOptions(WorkerOptions{
		WorkerID:   cfg.AIWorker.ID,
		IdleSleep:  parseDurationOrDefault(cfg.AIWorker.IdleSleep, defaultIdleSleep),
		ErrorSleep: parseDurationOrDefault(cfg.AIWorker.ErrorSleep, defaultErrorSleep),
		RunOnce:    cfg.AIWorker.RunOnce,
	})
}

func NormalizeWorkerOptions(opts WorkerOptions) WorkerOptions {
	if opts.WorkerID == "" {
		opts.WorkerID = defaultWorkerID
	}
	if opts.IdleSleep <= 0 {
		opts.IdleSleep = defaultIdleSleep
	}
	if opts.ErrorSleep <= 0 {
		opts.ErrorSleep = defaultErrorSleep
	}
	if opts.Logger == nil {
		opts.Logger = log.New(os.Stdout, "chronote-worker ", log.LstdFlags)
	}
	if opts.sleep == nil {
		opts.sleep = sleepContext
	}
	return opts
}

func (w *AnalysisWorker) Run(ctx context.Context, opts WorkerOptions) error {
	if w == nil || w.Service == nil {
		return errors.New("analysis worker service is not configured")
	}
	return RunAnalysisWorker(ctx, w.Service, opts)
}

func RunAnalysisWorker(ctx context.Context, processor analysisJobProcessor, opts WorkerOptions) error {
	if processor == nil {
		return errors.New("analysis job processor is not configured")
	}
	opts = NormalizeWorkerOptions(opts)
	for {
		if err := ctx.Err(); err != nil {
			return nil
		}

		outcome, err := processor.RunNextAnalysisJob(ctx, opts.WorkerID)
		if err != nil {
			logWorkerEvent(opts.Logger, opts.WorkerID, 0, "", "error", err)
			if opts.RunOnce {
				return err
			}
			if sleepErr := opts.sleep(ctx, opts.ErrorSleep); sleepErr != nil {
				return nil
			}
			continue
		}
		if outcome == nil || outcome.JobID == 0 {
			logWorkerEvent(opts.Logger, opts.WorkerID, 0, "", "idle", nil)
			if opts.RunOnce {
				return nil
			}
			if sleepErr := opts.sleep(ctx, opts.IdleSleep); sleepErr != nil {
				return nil
			}
			continue
		}

		logWorkerEvent(opts.Logger, opts.WorkerID, outcome.JobID, string(outcome.Status), "processed", nil)
		if opts.RunOnce {
			return nil
		}
	}
}

func logWorkerEvent(logger *log.Logger, workerID string, jobID uint, status, event string, err error) {
	if logger == nil {
		return
	}
	fields := []any{"event", event, "worker_id", workerID, "job_id", strconv.FormatUint(uint64(jobID), 10)}
	if status != "" {
		fields = append(fields, "status", status)
	}
	if err != nil {
		fields = append(fields, "error", err.Error())
	}
	logger.Println(fields...)
}

func sleepContext(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func parseDurationOrDefault(raw string, fallback time.Duration) time.Duration {
	if raw == "" {
		return fallback
	}
	duration, err := time.ParseDuration(raw)
	if err != nil || duration <= 0 {
		return fallback
	}
	return duration
}
