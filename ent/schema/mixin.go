package schema

import (
	"time"
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

// TimeMixin untuk timestamp fields
type TimeMixin struct {
	mixin.Schema
}

func (TimeMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Immutable().
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Time("deleted_at").
			Optional().
			Nillable(),
	}
}

func (TimeMixin) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("deleted_at"),
	}
}

// AuditMixin untuk audit trail
type AuditMixin struct {
	mixin.Schema
}

func (AuditMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Int("created_by").Optional(),
		field.String("created_name").Optional().MaxLen(255),
		field.String("created_code").Optional().MaxLen(100),
		field.Int("updated_by").Optional(),
		field.String("updated_name").Optional().MaxLen(255),
		field.String("updated_code").Optional().MaxLen(100),
		field.Int("deleted_by").Optional(),
		field.String("deleted_name").Optional().MaxLen(255),
		field.String("deleted_code").Optional().MaxLen(100),
	}
}

// UUIDMixin untuk UUID field
type UUIDMixin struct {
	mixin.Schema
}

func (UUIDMixin) Fields() []ent.Field {
	return []ent.Field{
		//field.Int("id").
		//	Positive().
		//	Immutable().
		//	StorageKey("id").
		//	Comment("ID of the entity"),

		field.UUID("uid", uuid.UUID{}).
			Default(uuid.New).
			Unique().
			Immutable(),
	}
}

func (UUIDMixin) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id"),
		index.Fields("uid"),
	}
}
