package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

// Problem holds the schema definition for the Problem entity.
type Problem struct {
	ent.Schema
}

// Fields of the Problem.
func (Problem) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", types.ProblemID{}).Default(types.NewProblemID).Unique().Immutable(),
		field.UUID("chat_id", types.ChatID{}),
		field.UUID("manager_id", types.UserID{}).Optional(),
		field.Time("resolved_at").Optional(),
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
