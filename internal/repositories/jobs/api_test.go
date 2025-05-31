//go:build integration

package jobsrepo_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"golang.org/x/sync/errgroup"

	jobsrepo "github.com/dndev-xx/go-ninja-chat/internal/repositories/jobs"
	"github.com/dndev-xx/go-ninja-chat/internal/store"
	"github.com/dndev-xx/go-ninja-chat/internal/testingh"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

var (
	name            = "job_name"
	payload         = "job_payload"
	reason          = "any reason"
	availableAt     = time.Now()
	reservationTime = func() time.Time { return time.Now().Add(time.Minute) }
)

type JobsRepoSuite struct {
	testingh.DBSuite
	repo *jobsrepo.Repo
}

func TestJobsRepoSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &JobsRepoSuite{DBSuite: testingh.NewDBSuite("TestJobsRepoSuite")})
}

func (s *JobsRepoSuite) SetupSuite() {
	s.DBSuite.SetupSuite()

	var err error

	s.repo, err = jobsrepo.New(jobsrepo.NewOptions(s.Database))
	s.Require().NoError(err)
}

func (s *JobsRepoSuite) SetupTest() {
	s.DBSuite.SetupTest()
	s.Database.Job(s.Ctx).Delete().ExecX(s.Ctx)
	s.Database.FailedJob(s.Ctx).Delete().ExecX(s.Ctx)
}

func (s *JobsRepoSuite) Test_FindAndReserveJob_JobFoundAndReserved() {
	// Arrange.
	jobExpected, err := s.Database.Job(s.Ctx).Create().
		SetName(name).
		SetPayload(payload).
		SetAvailableAt(availableAt).
		Save(s.Ctx)
	s.Require().NoError(err)
	s.Require().NotEmpty(jobExpected.ID)

	// Action.
	job, err := s.repo.FindAndReserveJob(s.Ctx, reservationTime())

	// Assert.
	s.Require().NoError(err)
	s.Equal(jobExpected.ID, job.ID)

	s.Run("job processing increases attempts", func() {
		s.Equal(1, job.Attempts)
	})
}

func (s *JobsRepoSuite) Test_FindAndReserveJob_SkipReservedJob() {
	// Arrange.
	const jobs = 3

	expected := make([]types.JobID, jobs)
	for i := 0; i < jobs; i++ {
		jobID, err := s.repo.CreateJob(s.Ctx, name, payload, availableAt)
		s.Require().NoError(err)
		s.NotEmpty(jobID)
		expected[i] = jobID
	}

	// Action.
	actual := make([]types.JobID, jobs)
	wg, ctx := errgroup.WithContext(s.Ctx)
	for i := 0; i < jobs; i++ {
		i := i
		wg.Go(func() error {
			job, err := s.repo.FindAndReserveJob(ctx, reservationTime())
			if err != nil {
				return err
			}
			actual[i] = job.ID
			return nil
		})
	}
	err := wg.Wait()
	s.Require().NoError(err)

	wg, ctx = errgroup.WithContext(s.Ctx) // Because wg.Wait() cancel context.
	for i := 0; i < jobs; i++ {
		wg.Go(func() error {
			_, err := s.repo.FindAndReserveJob(ctx, reservationTime())
			if nil == err || errors.Is(err, jobsrepo.ErrNoJobs) {
				return nil
			}
			return err
		})
	}
	err = wg.Wait()
	s.Require().NoError(err)

	// Assert.
	s.ElementsMatch(expected, actual)
}

func (s *JobsRepoSuite) Test_FindAndReserveJob_SkipDelayedJob() {
	{
		// Arrange.
		jobID, err := s.repo.CreateJob(s.Ctx, name, payload, time.Now().Add(2*time.Second))
		s.Require().NoError(err)
		s.Require().NotEmpty(jobID)

		// Action.
		job, err := s.repo.FindAndReserveJob(s.Ctx, reservationTime())

		// Assert.
		s.Require().ErrorIs(err, jobsrepo.ErrNoJobs)
		s.Empty(job)
	}

	{
		// Arrange.
		time.Sleep(3 * time.Second)

		// Action.
		job, err := s.repo.FindAndReserveJob(s.Ctx, reservationTime())

		// Assert.
		s.Require().NoError(err)
		s.Require().NotEmpty(job)
	}
}

func (s *JobsRepoSuite) Test_FindAndReserveJob_JobNotFound() {
	// Action.
	job, err := s.repo.FindAndReserveJob(s.Ctx, reservationTime())

	// Assert.
	s.Require().ErrorIs(err, jobsrepo.ErrNoJobs)
	s.Empty(job.ID)
}

func (s *JobsRepoSuite) Test_CreateJob() {
	// Action.
	jobID, err := s.repo.CreateJob(s.Ctx, name, payload, availableAt)

	// Assert.
	s.Require().NoError(err)
	s.Require().NotEmpty(jobID)

	// Checking if job was created.
	job, err := s.Database.Job(s.Ctx).Get(s.Ctx, jobID)
	s.Require().NoError(err)
	s.Require().NotNil(job)
	s.Equal(jobID, job.ID)
	s.Equal(name, job.Name)
	s.Equal(payload, job.Payload)
	s.Equal(availableAt.Unix(), job.AvailableAt.Unix())
}

func (s *JobsRepoSuite) Test_CreateJob_Multiple() {
	// Arrange.
	const jobs = 3

	// Action.
	for i := 0; i < jobs; i++ {
		jobID, err := s.repo.CreateJob(s.Ctx, name, payload, availableAt)
		s.Require().NoError(err)
		s.Require().NotEmpty(jobID)
	}

	// Assert.
	count, err := s.Database.Job(s.Ctx).Query().Count(s.Ctx)
	s.Require().NoError(err)
	s.Equal(jobs, count)
}

func (s *JobsRepoSuite) Test_CreateFailedJob() {
	err := s.repo.CreateFailedJob(s.Ctx, name, payload, reason)

	// Assert.
	s.Require().NoError(err)

	// Checking if failed job was created.
	fJob, err := s.Database.FailedJob(s.Ctx).Query().Only(s.Ctx)
	s.Require().NoError(err)
	s.Require().NotNil(fJob)
	s.NotEmpty(fJob.ID.String())
	s.Equal(name, fJob.Name)
	s.Equal(payload, fJob.Payload)
	s.Equal(reason, fJob.Reason)
}

func (s *JobsRepoSuite) Test_CreateFailedJob_Multiple() {
	// Arrange.
	const fJobs = 3

	// Action.
	for i := 0; i < fJobs; i++ {
		err := s.repo.CreateFailedJob(s.Ctx, name, payload, reason)
		s.Require().NoError(err)
	}

	// Assert.
	count, err := s.Database.FailedJob(s.Ctx).Query().Count(s.Ctx)
	s.Require().NoError(err)
	s.Equal(fJobs, count)
}

func (s *JobsRepoSuite) Test_DeleteJob() {
	// Arrange.
	jobExpected, err := s.Database.Job(s.Ctx).Create().
		SetName(name).
		SetPayload(payload).
		SetAvailableAt(availableAt).
		Save(s.Ctx)
	s.Require().NoError(err)
	s.Require().NotEmpty(jobExpected.ID)

	// Action.
	err = s.repo.DeleteJob(s.Ctx, jobExpected.ID)

	// Assert.
	s.Require().NoError(err)

	// Checking if failed job was deleted.
	job, err := s.Database.Job(s.Ctx).Get(s.Ctx, jobExpected.ID)
	s.True(store.IsNotFound(err))
	s.Nil(job)
}

func (s *JobsRepoSuite) Test_DeleteJob_NoJobs() {
	// Action.
	err := s.repo.DeleteJob(s.Ctx, types.NewJobID())

	// Assert.
	s.Require().Error(err)
}
