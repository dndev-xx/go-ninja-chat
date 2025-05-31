package jobsrepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dndev-xx/go-ninja-chat/internal/store"
	"github.com/dndev-xx/go-ninja-chat/internal/store/job"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

var ErrNoJobs = errors.New("no jobs found")

type Job struct {
	ID       types.JobID
	Name     string
	Payload  string
	Attempts int
}

func (r *Repo) FindAndReserveJob(ctx context.Context, until time.Time) (Job, error) {
	// FIXME: Избегая гонки на уровне строчек БД, сделать следующее:
	// FIXME: - выбрать не зарезервированную другим воркером джобу, чьё время выполнения уже настало;
	// FIXME: - увеличить счётчик попыток выполнения на 1;
	// FIXME: - зарезервировать джобу до `until`;
	// FIXME: - вернуть в ответ необходимую инфу о джобе.
	cur := time.Now()
	curJob, err := r.Db.Job(ctx).Query().
		Where(
			job.AvailableAtLTE(cur),
			job.Or(
				job.ReservedUntilIsNil(),
				job.ReservedUntilLT(cur),
			),
		).
		Order(store.Asc(job.FieldAvailableAt)).
		First(ctx)
	if err != nil {
		if store.IsNotFound(err) {
			return Job{}, ErrNoJobs
		}
		return Job{}, fmt.Errorf("failed to query available jobs: %w", err)
	}

	updated, err := r.Db.Job(ctx).UpdateOneID(curJob.ID).
		SetAttempts(curJob.Attempts + 1).
		SetReservedUntil(until).
		Where(
			job.Or(
				job.ReservedUntilIsNil(),
				job.ReservedUntilLT(time.Now()),
			),
		).
		Save(ctx)
	if err != nil {
		return Job{}, fmt.Errorf("failed to reserve job: %w", err)
	}

	return Job{
		ID:       updated.ID,
		Name:     updated.Name,
		Payload:  string(updated.Payload),
		Attempts: updated.Attempts,
	}, nil
}

func (r *Repo) CreateJob(ctx context.Context, name, payload string, availableAt time.Time) (types.JobID, error) {
	job, err := r.Db.Job(ctx).Create().
		SetName(name).
		SetPayload([]byte(payload)).
		SetAvailableAt(availableAt).
		Save(ctx)
	if err != nil {
		return types.JobIDNil, nil
	}
	return job.ID, nil
}

func (r *Repo) CreateFailedJob(ctx context.Context, name, payload, reason string) error {
	_, err := r.Db.FailedJob(ctx).Create().
		SetName(name).
		SetPayload([]byte(payload)).
		SetReason(reason).
		Save(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repo) DeleteJob(ctx context.Context, jobID types.JobID) error {
	err := r.Db.Job(ctx).
		DeleteOneID(jobID).
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repo) IncrementAttempts(ctx context.Context, jobID types.JobID) error {
	job, err := r.Db.Job(ctx).Get(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to find job: %w", err)
	}
	_, err = r.Db.Job(ctx).UpdateOneID(jobID).
		SetAttempts(job.Attempts + 1).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to reserve job: %w", err)
	}
	return nil
}
