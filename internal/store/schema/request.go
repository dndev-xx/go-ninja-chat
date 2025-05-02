package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)


type Request struct {
	ent.Schema
}

func (Request) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", types.RequestID{}).Default(types.NewRequestID).Unique().Immutable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("deleted_at").
			Default(time.Now).
			Immutable(),
	}
}

func (Request) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("created_at").Annotations(entsql.IndexType("BTREE")),
		index.Fields("deleted_at").Annotations(entsql.IndexType("BTREE")),
	}
}
