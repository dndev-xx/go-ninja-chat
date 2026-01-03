package outbox

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	repo "github.com/dndev-xx/go-ninja-chat/internal/repositories/jobs"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

var (
	ErrJobAlreadyRegistered = errors.New("job already registered")
	ErrJobNotFound          = errors.New("job not found")
)

type Outbox struct {
	workers      int
	idleTime     time.Duration
	reserveFor   time.Duration
	repo         JobsRepository
	logger       *zap.Logger
	registry     map[string]Job
	eventStream  eventPublisher
	mu           sync.RWMutex
	wg           sync.WaitGroup
	shutdownChan chan struct{}
	db           transactor
}

type JobsRepository interface {
	CreateJob(ctx context.Context, name, payload string, availableAt time.Time) (types.JobID, error)
	FindAndReserveJob(ctx context.Context, until time.Time) (repo.Job, error)
	CreateFailedJob(ctx context.Context, name, payload, reason string) error
	DeleteJob(ctx context.Context, jobID types.JobID) error
	IncrementAttempts(ctx context.Context, jobID types.JobID) error
}

type transactor interface {
	RunInTx(ctx context.Context, f func(context.Context) error) error
}

type Config struct {
	Workers        int
	IdleTime       time.Duration
	ReserveFor     time.Duration
	Logger         *zap.Logger
	EventPublisher eventPublisher
}

func New(repo JobsRepository, db transactor, cfg Config) *Outbox {
	return &Outbox{
		workers:      cfg.Workers,
		idleTime:     cfg.IdleTime,
		reserveFor:   cfg.ReserveFor,
		repo:         repo,
		logger:       cfg.Logger,
		eventStream:  cfg.EventPublisher,
		registry:     make(map[string]Job),
		shutdownChan: make(chan struct{}),
		db:           db,
	}
}

func (o *Outbox) RegisterJob(job Job) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if _, exists := o.registry[job.Name()]; exists {
		return ErrJobAlreadyRegistered
	}

	o.registry[job.Name()] = job
	return nil
}

func (o *Outbox) MustRegisterJob(job Job) {
	if err := o.RegisterJob(job); err != nil {
		panic(err)
	}
}

func (o *Outbox) Start(ctx context.Context) error {
	o.logger.Info("starting outbox service",
		zap.Int("workers", o.workers),
		zap.Duration("idle_time", o.idleTime),
	)

	for i := 0; i < o.workers; i++ {
		o.wg.Add(1)
		go o.worker(ctx, i)
	}

	return nil
}

func (o *Outbox) Stop() {
	close(o.shutdownChan)
	o.wg.Wait()
}

func (o *Outbox) worker(ctx context.Context, workerID int) {
	defer o.wg.Done()

	o.logger.Debug("worker started", zap.Int("worker_id", workerID))

	for {
		select {
		case <-o.shutdownChan:
			o.logger.Debug("worker stopped by shutdown", zap.Int("worker_id", workerID))
			return
		case <-ctx.Done():
			o.logger.Debug("worker stopped by context", zap.Int("worker_id", workerID))
			return
		default:
			if err := o.processJob(ctx, workerID); err != nil {
				if errors.Is(err, ErrJobNotFound) {
					select {
					case <-o.shutdownChan:
						return
					case <-ctx.Done():
						return
					case <-time.After(o.idleTime):
						continue
					}
				}

				// o.logger.Error("job processing error",
				// 	zap.Int("worker_id", workerID),
				// 	zap.Error(err),
				// )
				time.Sleep(time.Second * 5)
			}
		}
	}
}

func (o *Outbox) processJob(ctx context.Context, workerID int) error {
	reservedUntil := time.Now().Add(o.reserveFor)
	if err := o.db.RunInTx(ctx, func(ctx context.Context) error {
		job, err := o.repo.FindAndReserveJob(ctx, reservedUntil)
		if err != nil {
			if errors.Is(err, ErrJobNotFound) {
				return ErrJobNotFound
			}
			return fmt.Errorf("find and reserve job: %w", err)
		}
		o.logger.Debug("job reserved",
			zap.Int("worker_id", workerID),
			zap.String("job_id", job.ID.String()),
			zap.String("job_name", job.Name),
		)

		o.mu.RLock()
		handler, exists := o.registry[job.Name]
		o.mu.RUnlock()

		if !exists {
			if err := o.repo.CreateFailedJob(ctx, job.Name, job.Payload, "job not registered"); err != nil {
				return fmt.Errorf("create failed job: %w", err)
			}
			if err := o.repo.DeleteJob(ctx, job.ID); err != nil {
				return fmt.Errorf("delete job: %v", err)
			}
			return fmt.Errorf("job %q not registered", job.Name)
		}

		if job.Attempts >= handler.MaxAttempts() {
			if err := o.repo.CreateFailedJob(ctx, job.Name, job.Payload, "max attempts exceeded"); err != nil {
				return fmt.Errorf("create failed job: %w", err)
			}
			if err := o.repo.DeleteJob(ctx, job.ID); err != nil {
				return fmt.Errorf("delete job: %w", err)
			}
			return fmt.Errorf("max attempts exceeded for job %q", job.Name)
		}

		execCtx, cancel := context.WithTimeout(ctx, handler.ExecutionTimeout())
		defer cancel()

		err = handler.Handle(execCtx, job.Payload)
		if err != nil {
			if incErr := o.repo.IncrementAttempts(ctx, job.ID); incErr != nil {
				return fmt.Errorf("increment attempts: %w (original error: %v)", incErr, err)
			}
			return fmt.Errorf("job handler failed: %w", err)
		}

		if err := o.repo.DeleteJob(ctx, job.ID); err != nil {
			return fmt.Errorf("delete job: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("run in tx: %w", err)
	}
	return nil
}

func (o *Outbox) Enqueue(ctx context.Context, name, payload string, availableAt time.Time) (types.JobID, error) {
	return o.repo.CreateJob(ctx, name, payload, availableAt)
}
