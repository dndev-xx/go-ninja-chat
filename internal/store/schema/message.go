package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"time"
	"github.com/google/uuid"
)

// Message holds the schema definition for the Message entity.
type Message struct {
	ent.Schema
}

// Fields of the Message.
func (Message) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("chat_id", uuid.UUID{}),
		field.UUID("problem_id", uuid.UUID{}),
		field.UUID("author_id", uuid.UUID{}),
		field.Bool("is_visible_for_client").
			Default(false),
		field.Bool("is_visible_for_manager").
			Default(true),
		field.Text("body"),
		field.Time("checked_at").
			Optional().
			Nillable(),
		field.Bool("is_blocked").
			Default(false),
		field.Bool("is_service").
			Default(false),
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