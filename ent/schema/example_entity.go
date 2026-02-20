package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ExampleEntity adalah template dasar ent schema
// Rename nama struct ini sesuai kebutuhan
type ExampleEntity struct {
	ent.Schema
}

// Mixin untuk reusable fields.
// Hanya UUIDMixin di sini supaya uid tetap paling atas,
// mixin lain akan ditambahkan via helper pada Fields/Indexes.
func (ExampleEntity) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDMixin{}, // uid di atas
	}
}

// Fields of the ExampleEntity
// Fields disini hanya untuk field yang spesifik untuk entity ini saja
// Field seperti id, uid sudah ada di mixin
func (ExampleEntity) Fields() []ent.Field {
	fields := []ent.Field{
		// Common name field
		field.String("name").
			NotEmpty().
			MaxLen(255).
			Comment("Entity name/title"),

		// Optional description
		field.Text("description").
			Optional().
			Comment("Entity description"),

		// Status field example (dapat disesuaikan)
		field.Enum("status").
			Values("active", "inactive", "draft").
			Default("active").
			Comment("Entity status"),

		// Contoh field lainnya
		field.Float("price").
			Optional().
			Positive().
			Comment("Price in currency"),

		field.Int("quantity").
			Optional().
			Default(0).
			NonNegative().
			Comment("Available quantity"),

		field.Bool("is_featured").
			Default(false).
			Comment("Whether this entity is featured"),
	}

	// Tambahkan audit + timestamp fields di bawah menggunakan helper agar urutan konsisten.
	return AppendMixinFields(fields,
		AuditMixin{},
		TimeMixin{},
	)
}

// Edges of the ExampleEntity
func (ExampleEntity) Edges() []ent.Edge {
	return []ent.Edge{
		// Contoh relasi One-to-Many
		// edge.To("items", Item.Type).
		// 	Comment("One-to-many relation to items"),

		// Contoh relasi Many-to-One
		// edge.From("category", Category.Type).
		// 	Ref("items").
		// 	Unique().
		// 	Comment("Many-to-one relation to category"),

		// Contoh relasi Many-to-Many
		// edge.To("tags", Tag.Type).
		// 	Comment("Many-to-many relation to tags"),
	}
}

// Indexes untuk optimasi query
// Index untuk field dari mixin ditambahkan via helper
func (ExampleEntity) Indexes() []ent.Index {
	indexes := []ent.Index{
		// Index untuk name search
		index.Fields("name"),

		// Composite index untuk status + name
		index.Fields("status", "name"),

		// Composite index untuk featured items
		index.Fields("is_featured", "status"),
	}

	// Tambahkan index bawaan mixin (mis. deleted_at)
	return AppendMixinIndexes(indexes,
		AuditMixin{},
		TimeMixin{},
	)
}

// Hooks untuk business logic (optional)
// func (ExampleEntity) Hooks() []ent.Hook {
// 	return []ent.Hook{
// 		// Hook untuk validasi sebelum create/update
// 		hook.On(
// 			func(next ent.Mutator) ent.Mutator {
// 				return hook.ExampleEntityFunc(func(ctx context.Context, m *gen.ExampleEntityMutation) (ent.Value, error) {
// 					// Custom validation logic
// 					return next.Mutate(ctx, m)
// 				})
// 			},
// 			ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne,
// 		),
// 	}
// }

// Annotations untuk custom config (optional)
// func (ExampleEntity) Annotations() []schema.Annotation {
// 	return []schema.Annotation{
// 		entsql.Annotation{Table: "example_entities"},
// 	}
// }
