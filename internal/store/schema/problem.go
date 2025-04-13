package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"time"
	"github.com/google/uuid"
)

// Problem holds the schema definition for the Problem entity.
type Problem struct {
	ent.Schema
}

// Fields of the Problem.
func (Problem) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("chat_id", uuid.UUID{}),
		field.UUID("manager_id", uuid.UUID{}).
			Optional(),
		field.Time("resolved_at").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now),
	}
}

// Edges of the Problem.
func (Problem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("chat", Chat.Type).
			Ref("problems").
			Field("chat_id").
			Unique().
			Required(),
		edge.To("messages", Message.Type),
	}
}