// Package schema - the quotes module's own ent schema: its tables, apart
// from the core's. An author's material is a core node: Node here is the
// module's stand-in for the core's table (migrations.Sets).
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/kosatka-cms/kosatka/modulekit/schemakit"
)

// Quote holds the schema definition for the Quote entity.
type Quote struct {
	ent.Schema
}

// Annotations of the Quote.
func (Quote) Annotations() []schema.Annotation {
	return []schema.Annotation{
		schemakit.Table("quotes"),
	}
}

// Fields of the Quote.
func (Quote) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").SchemaType(schemakit.MySQLType("int")),
		field.Text("text").SchemaType(schemakit.MySQLType("mediumtext")).NotEmpty(),
		field.Int("author_id").SchemaType(schemakit.MySQLType("int")).Optional(),
		field.Text("citation").SchemaType(schemakit.MySQLType("mediumtext")).Optional(),
		field.Time("created_at").Default(time.Now).Annotations(entsql.DefaultExpr("CURRENT_TIMESTAMP")),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).Annotations(entsql.DefaultExpr("CURRENT_TIMESTAMP")),
	}
}

// Edges of the Quote.
func (Quote) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("author", QuoteAuthor.Type).Field("author_id").Unique(),
	}
}

// QuoteAuthor holds the schema definition for the QuoteAuthor entity.
type QuoteAuthor struct {
	ent.Schema
}

// Annotations of the QuoteAuthor.
func (QuoteAuthor) Annotations() []schema.Annotation {
	return []schema.Annotation{
		schemakit.Table("quotes_authors"),
	}
}

// Fields of the QuoteAuthor.
func (QuoteAuthor) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").SchemaType(schemakit.MySQLType("int")),
		field.String("name").MaxLen(255).NotEmpty().Unique(),
		field.Int("node_id").SchemaType(schemakit.MySQLType("int unsigned")).Optional().Nillable(),
	}
}

// Edges of the QuoteAuthor.
func (QuoteAuthor) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("quotes", Quote.Type).Ref("author"),
		edge.To("node", Node.Type).Field("node_id").Unique(),
	}
}

// Indexes of the QuoteAuthor.
func (QuoteAuthor) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("node_id"),
	}
}

// Indexes of the Quote.
func (Quote) Indexes() []ent.Index {
	return []ent.Index{
		// Foreign key columns: PostgreSQL doesn't index them itself (MySQL
		// does), and deleting the row they refer to looks them up.
		index.Fields("author_id"),
	}
}
