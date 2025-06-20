//go:build integration

package outbox_test

import (
	"context"
	"errors"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// TestComplex apologizes for its content.
func (s *OutboxServiceSuite) TestComplex() {
	// Arrange.
	const (
		jobSuccessfulFromFirstTime = "successful-from-first-time-job" // job1
		jobSuccessfulFromThirdTime = "successful-from-third-time-job" // job2

		jobFailedAfterSecondTime = "failed-after-second-time-job" // job3
		jobFailedAfterFiveTime   = "failed-after-five-time-job"   // job4

		jobTmpTimeoutedAndSuccessfulAfter = "temporary-timeouted-and-successful-after-job" // job5
		jobTmpTimeoutedAndFailedAfter     = "temporary-timeouted-and-failed-after-job"     // job6

		jobUnknown = "unknown-job"
	)

	executedTimes := newJobInstancesExecutedTimes()

	job1 := newJobMock(jobSuccessfulFromFirstTime, nop, time.Second, 10)
	s.outboxSvc.MustRegisterJob(job1)

	job2 := newJobMock(jobSuccessfulFromThirdTime, func(_ context.Context, payloadAsIndex string) error {
		k := jobSuccessfulFromThirdTime + payloadAsIndex
		if executedTimes.Inc(k) == 3 {
			return nil
		}
		return errors.New("sorry I'm failed")
	}, time.Second, 4)
	s.outboxSvc.MustRegisterJob(job2)

	job3 := newJobMock(jobFailedAfterSecondTime, func(_ context.Context, _ string) error {
		return errors.New("sorry I'm failed")
	}, time.Second, 2)
	s.outboxSvc.MustRegisterJob(job3)

	job4 := newJobMock(jobFailedAfterFiveTime, func(_ context.Context, _ string) error {
		return errors.New("sorry I'm failed")
	}, time.Second, 5)
	s.outboxSvc.MustRegisterJob(job4)

	job5 := newJobMock(jobTmpTimeoutedAndSuccessfulAfter, func(ctx context.Context, payloadAsIndex string) error {
		k := jobTmpTimeoutedAndSuccessfulAfter + payloadAsIndex
		if executedTimes.Inc(k) == 1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(50 * time.Millisecond):
			}
		}
		return nil
	}, time.Millisecond, 2)
	s.outboxSvc.MustRegisterJob(job5)

	job6 := newJobMock(jobTmpTimeoutedAndFailedAfter, func(ctx context.Context, payloadAsIndex string) error {
		k := jobTmpTimeoutedAndFailedAfter + payloadAsIndex
		if executedTimes.Inc(k) == 1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(50 * time.Millisecond):
			}
		}
		return errors.New("sorry I'm failed")
	}, time.Millisecond, 2)
	s.outboxSvc.MustRegisterJob(job6)

	// Action.
	cancel, errCh := s.runOutbox()
	defer cancel()

	jobCounts := map[string]int{
		jobSuccessfulFromFirstTime: 50,
		jobSuccessfulFromThirdTime: 4,

		jobFailedAfterSecondTime: 1,
		jobFailedAfterFiveTime:   2,

		jobTmpTimeoutedAndSuccessfulAfter: 3,
		jobTmpTimeoutedAndFailedAfter:     3,

		jobUnknown: 30,
	}

	wg, ctx := errgroup.WithContext(s.Ctx)

	for jobName, jobCount := range jobCounts {
		jobName, jobCount := jobName, jobCount
		wg.Go(func() error {
			for i := 1; i <= jobCount; i++ {
				var randAvailableAt time.Time
				// Random choice between immediate and delayed job.
				if rand.Float64() >= 0.5 { //nolint:gosec // for test math/rand is OK
					randAvailableAt = time.Now()
				} else {
					randAvailableAt = time.Now().Add(4 * idleTime)
				}

				if _, err := s.outboxSvc.Put(ctx, jobName, strconv.Itoa(i), randAvailableAt); err != nil {
					return err
				}
			}
			return nil
		})
	}
	err := wg.Wait()
	s.Require().NoError(err)

	// Assert.
	time.Sleep(10 * time.Second)

	cancel()
	s.NoError(<-errCh)

	{
		s.Equal(jobCounts[jobSuccessfulFromFirstTime]*1, job1.ExecutedTimes())
		s.Equal(jobCounts[jobSuccessfulFromThirdTime]*3, job2.ExecutedTimes())

		s.Equal(jobCounts[jobFailedAfterSecondTime]*2, job3.ExecutedTimes())
		s.Equal(jobCounts[jobFailedAfterFiveTime]*5, job4.ExecutedTimes())

		s.Equal(jobCounts[jobTmpTimeoutedAndSuccessfulAfter]*2, job5.ExecutedTimes())
		s.Equal(jobCounts[jobTmpTimeoutedAndFailedAfter]*2, job6.ExecutedTimes())
	}
	{
		s.Equal(0, s.Store.Job.Query().CountX(s.Ctx))

		failedJobsTotal := jobCounts[jobFailedAfterSecondTime] +
			jobCounts[jobFailedAfterFiveTime] +
			jobCounts[jobTmpTimeoutedAndFailedAfter] +
			jobCounts[jobUnknown]
		s.Equal(failedJobsTotal, s.Store.FailedJob.Query().CountX(s.Ctx))
	}
}

type jobInstanceKey = string

// jobInstancesExecutedTimes stores job-instance-isolated counters.
type jobInstancesExecutedTimes struct {
	counters map[jobInstanceKey]int
	mu       sync.RWMutex
}

func newJobInstancesExecutedTimes() *jobInstancesExecutedTimes {
	return &jobInstancesExecutedTimes{
		counters: map[jobInstanceKey]int{},
		mu:       sync.RWMutex{},
	}
}

// Inc increments counter and returns its new value.
func (j *jobInstancesExecutedTimes) Inc(k jobInstanceKey) int {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.counters[k]++
	return j.counters[k]
}
