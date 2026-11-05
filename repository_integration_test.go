//go:build dbintegration

package quotes

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/kosatka-cms/kosatka/modulekit"
	"github.com/kosatka-cms/kosatka/modulekit/testkit"
)

// TestQuoteRepository - random quotes come from the stored ones; an
// author is found by the node it is linked to, with that node's title
// and address (the core's node, through its repository); the node gone,
// the author stays without it (the module's foreign key to the core's
// nodes: ON DELETE SET NULL).
func TestQuoteRepository(t *testing.T) {
	testkit.SchemaDB(t, func(t *testing.T, db testkit.DB) {
		ctx := context.Background()
		for _, q := range []string{
			`INSERT INTO nodes (id, type, title, slug) VALUES (7, 'article', 'Mark Twain', 'article/mark-twain')`,
			`INSERT INTO quotes_authors (id, name, node_id) VALUES (1, 'Author', 7)`,
			`INSERT INTO quotes (text, author_id) VALUES ('one', 1), ('two', 1), ('three', 1)`,
		} {
			if _, err := db.DB.Exec(q); err != nil {
				t.Fatal(err)
			}
		}
		repo := NewQuoteRepository(db.DB, testkit.Deps(db).NodeRepo)
		quotes, err := repo.Random(ctx, 2)
		if err != nil || len(quotes) != 2 {
			t.Errorf("random quotes: %+v %v", quotes, err)
		}
		if n, err := repo.TotalCount(ctx, QuotesFilter{AuthorID: ptr(1)}); err != nil || n != 3 {
			t.Errorf("author's quotes: %d %v", n, err)
		}
		a, err := repo.AuthorByNodeID(ctx, 7)
		if err != nil || a.Name != "Author" || a.NodeID == nil || *a.NodeID != 7 {
			t.Fatalf("author by node: %+v %v", a, err)
		}
		if _, err := db.DB.Exec(`DELETE FROM nodes WHERE id = 7`); err != nil {
			t.Fatalf("deleting the author's node: %v", err)
		}
		if a, err := repo.AuthorByID(ctx, 1); err != nil || a.NodeID != nil {
			t.Errorf("the author without the node: %+v %v", a, err)
		}
		// A missing quote is the engine's missing record (a 404).
		if _, err := repo.ByID(ctx, 987654); !modulekit.IsNotFound(err) {
			t.Errorf("a missing quote: %v", err)
		}
	})
}

func ptr(i int) *int { return &i }

// TestUpdateAuthor - an author is renamed and linked to a material; a name
// another author has and a material that doesn't exist are refused with
// their errors (not a failed write), the author left as it was.
func TestUpdateAuthor(t *testing.T) {
	testkit.SchemaDB(t, func(t *testing.T, db testkit.DB) {
		ctx := context.Background()
		for _, q := range []string{
			`INSERT INTO nodes (id, type, title, slug) VALUES (7, 'article', 'Mark Twain', 'article/mark-twain')`,
			`INSERT INTO quotes_authors (id, name) VALUES (1, 'Author'), (2, 'Other')`,
		} {
			if _, err := db.DB.Exec(q); err != nil {
				t.Fatal(err)
			}
		}
		repo := NewQuoteRepository(db.DB, testkit.Deps(db).NodeRepo)
		s := NewQuoteService(repo, slog.New(slog.DiscardHandler))
		if err := s.UpdateAuthor(ctx, 1, "Other", nil); !errors.Is(err, ErrAuthorNameTaken) {
			t.Errorf("a taken name: %v", err)
		}
		if err := s.UpdateAuthor(ctx, 1, "Author", ptr(99)); !errors.Is(err, ErrAuthorNodeMissing) {
			t.Errorf("a missing material: %v", err)
		}
		if err := s.UpdateAuthor(ctx, 1, "Samuel Clemens", ptr(7)); err != nil {
			t.Fatal(err)
		}
		if a, err := repo.AuthorByNodeID(ctx, 7); err != nil || a.ID != 1 || a.Name != "Samuel Clemens" {
			t.Errorf("renamed and linked: %+v %v", a, err)
		}
	})
}

// TestDeleteAuthor - an author with quotes isn't deleted
// (ErrAuthorHasQuotes); without them, it is.
func TestDeleteAuthor(t *testing.T) {
	testkit.SchemaDB(t, func(t *testing.T, db testkit.DB) {
		ctx := context.Background()
		for _, q := range []string{
			`INSERT INTO quotes_authors (id, name) VALUES (1, 'Author'), (2, 'Unquoted')`,
			`INSERT INTO quotes (text, author_id) VALUES ('one', 1)`,
		} {
			if _, err := db.DB.Exec(q); err != nil {
				t.Fatal(err)
			}
		}
		repo := NewQuoteRepository(db.DB, testkit.Deps(db).NodeRepo)
		s := NewQuoteService(repo, slog.New(slog.DiscardHandler))
		if err := s.DeleteAuthor(ctx, 1); !errors.Is(err, ErrAuthorHasQuotes) {
			t.Errorf("an author with quotes: %v", err)
		}
		if err := s.DeleteAuthor(ctx, 2); err != nil {
			t.Fatal(err)
		}
		var n int
		if err := db.DB.QueryRow(`SELECT COUNT(*) FROM quotes_authors`).Scan(&n); err != nil || n != 1 {
			t.Errorf("authors left: %d %v", n, err)
		}
	})
}
