package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

const messageBodyMaxLength = 3000

// Message holds the schema definition for the Message entity.
type Message struct {
	ent.Schema
}

// Fields of the Message.
func (Message) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", types.MessageID{}).Default(types.NewMessageID).Unique().Immutable(),
		field.UUID("chat_id", types.ChatID{}),
		field.UUID("problem_id", types.ProblemID{}),
		field.UUID("author_id", types.UserID{}).Optional().Immutable(),
		field.Bool("is_visible_for_client").Default(false),
		field.Bool("is_visible_for_manager").Default(false),
		field.Text("body").NotEmpty().MaxLen(messageBodyMaxLength).Immutable(),
		field.Time("checked_at").Optional().Nillable(),
		field.Bool("is_blocked").Default(false),
		field.Bool("is_service").Default(false).Immutable(),
		field.Time("created_at").
			Default(time.Now),
	}
}

// Edges of the Message.
func (Message) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("chat", Chat.Type).
			Ref("messages").
			Field("chat_id").
			Unique().
			Required(),
		edge.From("problem", Problem.Type).
			Ref("messages").
			Field("problem_id").
			Unique().
			Required(),
	}
}

func (Message) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("chat_id").Annotations(entsql.IndexType("HASH")),
		index.Fields("chat_id", "created_at"),
		index.Fields("author_id", "is_visible_for_client"),
		index.Fields("created_at").Annotations(entsql.IndexType("BTREE")),
	}
}
