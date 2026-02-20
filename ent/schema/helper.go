package schema

import "entgo.io/ent"

func AppendMixinFields(base []ent.Field, mixins ...ent.Mixin) []ent.Field {
	for _, m := range mixins {
		if f, ok := m.(interface{ Fields() []ent.Field }); ok {
			base = append(base, f.Fields()...)
		}
	}
	return base
}

func AppendMixinIndexes(base []ent.Index, mixins ...ent.Mixin) []ent.Index {
	for _, m := range mixins {
		if idx, ok := m.(interface{ Indexes() []ent.Index }); ok {
			base = append(base, idx.Indexes()...)
		}
	}
	return base
}
