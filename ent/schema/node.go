package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/kosatka-cms/kosatka/modulekit/schemakit"
)

// Node - the core's nodes table as the module's stand-in: its id only (of
// the core's column type), for the author's edge to a material - the
// foreign key, ON DELETE SET NULL. A table named as a core one isn't the
// module's: its set leaves it out (migrations.Sets) and the module reads
// nodes through the core, never through this.
type Node struct {
	ent.Schema
}

// Annotations of the Node.
func (Node) Annotations() []schema.Annotation {
	return []schema.Annotation{schemakit.Table("nodes")}
}

// Fields of the Node.
func (Node) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").SchemaType(schemakit.MySQLType("int unsigned")),
	}
}
