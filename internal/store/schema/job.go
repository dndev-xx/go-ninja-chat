package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"time"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

// jobMaxAttempts is some limit as protection from endless retries of outbox jobs.
const jobMaxAttempts = 30

type Job struct {
	ent.Schema
}

func (Job) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", types.JobID{}).Default(types.NewJobID).Unique().Immutable(),
		field.String("name").NotEmpty().Immutable(),
		field.Bytes("payload").NotEmpty().Immutable(),
		field.Int("attempts").Default(0).Range(0, jobMaxAttempts),
		field.Time("available_at").Immutable(),
		field.Time("reserved_until").Optional(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (Job) Indexes() []ent.Index {
	return []ent.Index{
		// Index for finding available jobs (not reserved or reservation expired)
		index.Fields("available_at", "reserved_until"),
		// Index for finding jobs by name if needed
		index.Fields("name"),
	}
}

type FailedJob struct {
	ent.Schema
}

func (FailedJob) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", types.FailedJobID{}).Default(types.NewFailedJobID).Unique().Immutable(),
		field.String("name").NotEmpty().Immutable(),
		field.Bytes("payload").NotEmpty().Immutable(),
		field.String("reason").NotEmpty().Immutable(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}