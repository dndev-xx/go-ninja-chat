//go:build integration

package outbox_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"

	jobsrepo "github.com/dndev-xx/go-ninja-chat/internal/repositories/jobs"
	"github.com/dndev-xx/go-ninja-chat/internal/services/outbox"
	"github.com/dndev-xx/go-ninja-chat/internal/testingh"
)

var (
	workers    = 10
	idleTime   = 250 * time.Millisecond
	reserveFor = time.Second
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

type OutboxServiceSuite struct {
	testingh.DBSuite
	ctrl      *gomock.Controller
	outboxSvc *outbox.Service
}

func TestOutboxServiceSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &OutboxServiceSuite{DBSuite: testingh.NewDBSuite("TestOutboxServiceSuite")})
}

func (s *OutboxServiceSuite) SetupSuite() {
	s.DBSuite.SetupSuite()

	jobsRepo, err := jobsrepo.New(jobsrepo.NewOptions(s.Database))
	s.Require().NoError(err)

	s.outboxSvc, err = outbox.New(outbox.NewOptions(
		workers,
		idleTime,
		reserveFor,
		jobsRepo,
		s.Database,
	))
	s.Require().NoError(err)
}

func (s *OutboxServiceSuite) TearDownSuite() {
	s.DBSuite.TearDownSuite()
}

func (s *OutboxServiceSuite) SetupTest() {
	s.DBSuite.SetupTest()
	s.ctrl = gomock.NewController(s.T())

	s.Database.Job(s.Ctx).Delete().ExecX(s.Ctx)
	s.Database.FailedJob(s.Ctx).Delete().ExecX(s.Ctx)
}

func (s *OutboxServiceSuite) TearDownTest() {
	s.ctrl.Finish()
	s.DBSuite.TearDownTest()
}

func (s *OutboxServiceSuite) TestMustRegisterJob() {
	job := newJobMock("duplicated_job", nop, time.Second, 1)

	s.NotPanics(func() {
		s.outboxSvc.MustRegisterJob(job)
	})

	s.Panics(func() {
		s.outboxSvc.MustRegisterJob(job)
	})
}

func (s *OutboxServiceSuite) TestPutJob() {
	// Arrange.
	const jobName = "TestPutJob"
	const jobPayload = "{}"
	availableAt := time.Now()

	// Action.
	jobID, err := s.outboxSvc.Put(s.Ctx, jobName, jobPayload, availableAt)
	s.Require().NoError(err)

	// Assert.
	j, err := s.Store.Job.Get(s.Ctx, jobID)
	s.Require().NoError(err)
	s.Equal(jobID, j.ID)
	s.Equal(jobName, j.Name)
	s.Equal(jobPayload, j.Payload)
	s.Equal(0, j.Attempts)
	s.Equal(availableAt.Unix(), j.AvailableAt.Unix())
	s.NotEmpty(j.ReservedUntil)
	s.NotEmpty(j.CreatedAt)
}

func (s *OutboxServiceSuite) TestAllJobsProcessed() {
	// Arrange.
	const jobName = "TestAllJobsProcessed"

	job := newJobMock(jobName, nop, time.Second, 1)
	s.outboxSvc.MustRegisterJob(job)

	const jobsCount = 30
	for i := 0; i < jobsCount; i++ {
		_, err := s.outboxSvc.Put(s.Ctx, jobName, `{messageId:"4242"}`, time.Now())
		s.Require().NoError(err)
	}

	// Action.
	s.runOutboxFor(time.Second)

	// Assert.
	s.Equal(jobsCount, job.ExecutedTimes())
	s.Equal(0, s.Store.Job.Query().CountX(s.Ctx))
	s.Equal(0, s.Store.FailedJob.Query().CountX(s.Ctx))
}

func (s *OutboxServiceSuite) TestDLQ_UnknownJob() {
	// Arrange.
	const jobName = "unknown-job"
	const jobPayload = "{}"
	_, err := s.outboxSvc.Put(s.Ctx, jobName, jobPayload, time.Now())
	s.Require().NoError(err)

	// Action.
	s.runOutboxFor(idleTime)

	// Assert.
	s.Require().Equal(0, s.Store.Job.Query().CountX(s.Ctx))
	s.Require().Equal(1, s.Store.FailedJob.Query().CountX(s.Ctx))

	j, err := s.Store.FailedJob.Query().Only(s.Ctx)
	s.Require().NoError(err)
	s.NotEmpty(j.ID)
	s.Equal(jobName, j.Name)
	s.Equal(jobPayload, j.Payload)
	s.NotEmpty(j.Reason)
	s.NotEmpty(j.CreatedAt)
}

func (s *OutboxServiceSuite) TestDLQ_AfterMaxAttemptsExceeding() {
	// Arrange.
	const jobName = "TestDLQ_AfterMaxAttemptsExceeding"
	const jobPayload = "{}"
	const maxAttempts = 3
	availableAt := time.Now()

	var executedTimes int
	job := newJobMock(jobName, func(ctx context.Context, _ string) error {
		executedTimes++
		if executedTimes == maxAttempts-1 {
			select {
			case <-ctx.Done():
				return ctx.Err() // Check job failing after ExecutionTimeout() exceeding.
			case <-time.After(50 * time.Millisecond):
			}
		}
		return errors.New("unknown")
	}, time.Millisecond, maxAttempts)
	s.outboxSvc.MustRegisterJob(job)

	_, err := s.outboxSvc.Put(s.Ctx, jobName, jobPayload, availableAt)
	s.Require().NoError(err)

	// Action.
	s.runOutboxFor(maxAttempts * time.Second)

	// Assert.
	s.Require().Equal(0, s.Store.Job.Query().CountX(s.Ctx))
	s.Require().Equal(1, s.Store.FailedJob.Query().CountX(s.Ctx))

	j, err := s.Store.FailedJob.Query().Only(s.Ctx)
	s.Require().NoError(err)
	s.NotEmpty(j.ID)
	s.Equal(jobName, j.Name)
	s.Equal(jobPayload, j.Payload)
	s.NotEmpty(j.Reason)
	s.NotEmpty(j.CreatedAt)

	s.Equal(maxAttempts, job.ExecutedTimes())
}

func (s *OutboxServiceSuite) TestIfNoJobsThenWorkersSleepForIdleTime() {
	// Arrange.
	const jobName = "TestIfNoJobsThenWorkersSleepForIdleTime"

	job := newJobMock(jobName, nop, time.Second, 1)
	s.outboxSvc.MustRegisterJob(job)

	// Action.
	cancel, errCh := s.runOutbox()
	defer cancel()

	// Assert.
	time.Sleep(idleTime / 25)

	const jobsCount = 3
	for i := 0; i < jobsCount; i++ {
		_, err := s.outboxSvc.Put(s.Ctx, jobName, fmt.Sprintf(`{messageId:"%d"}`, i), time.Now())
		s.Require().NoError(err)
	}

	s.Require().Equal(jobsCount, s.Store.Job.Query().CountX(s.Ctx)) // Workers fell asleep before the jobs appearing.
	s.Equal(0, job.ExecutedTimes())

	time.Sleep(2 * idleTime)

	s.Equal(0, s.Store.Job.Query().CountX(s.Ctx)) // Workers woke up and processed the jobs.
	s.Equal(0, s.Store.FailedJob.Query().CountX(s.Ctx))
	s.Equal(jobsCount, job.ExecutedTimes())

	cancel()
	s.NoError(<-errCh)
}

func (s *OutboxServiceSuite) runOutboxFor(timeout time.Duration) {
	s.T().Helper()

	cancel, errCh := s.runOutbox()
	defer cancel()

	time.Sleep(timeout)
	cancel()
	s.NoError(<-errCh) // No error expected because of graceful shutdown via cancel ctx.
}

func (s *OutboxServiceSuite) runOutbox() (context.CancelFunc, <-chan error) {
	s.T().Helper()

	ctx, cancel := context.WithCancel(s.Ctx)

	errCh := make(chan error)
	go func() { errCh <- s.outboxSvc.Run(ctx) }()

	return cancel, errCh
}

var nop = func(_ context.Context, _ string) error {
	time.Sleep(10 * time.Millisecond) // Prevent PSQL DDoS.
	return nil
}

type jobMock struct {
	name          string
	handler       func(ctx context.Context, s string) error
	timeout       time.Duration
	maxAttempts   int
	executedTimes int32
}

func newJobMock(
	name string,
	h func(ctx context.Context, s string) error,
	executionTimeout time.Duration,
	maxAttempts int,
) *jobMock {
	return &jobMock{
		name:          name,
		handler:       h,
		timeout:       executionTimeout,
		maxAttempts:   maxAttempts,
		executedTimes: 0,
	}
}

func (j *jobMock) Name() string {
	return j.name
}

func (j *jobMock) Handle(ctx context.Context, payload string) error {
	atomic.AddInt32(&j.executedTimes, 1)
	return j.handler(ctx, payload)
}

func (j *jobMock) ExecutionTimeout() time.Duration {
	return j.timeout
}

func (j *jobMock) MaxAttempts() int {
	return j.maxAttempts
}

// ExecutedTimes returns global (for all different jobs of this type
// processed at different times) execution counter.
func (j *jobMock) ExecutedTimes() int {
	return int(atomic.LoadInt32(&j.executedTimes))
}
