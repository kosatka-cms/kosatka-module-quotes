package quotes

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/kosatka-cms/kosatka-module-quotes/ent/generated"
	"github.com/kosatka-cms/kosatka-module-quotes/ent/generated/quote"
	"github.com/kosatka-cms/kosatka-module-quotes/ent/generated/quoteauthor"
	"github.com/kosatka-cms/kosatka/modulekit"
)

// NewQuoteRepository - the quotes on db, in the site's dialect, on the
// module's own ent schema (ent/); an author's material comes from the
// core's nodes.
func NewQuoteRepository(db *sql.DB, nodes modulekit.NodesRepository) *EntQuoteRepository {
	return &EntQuoteRepository{db: db, nodes: nodes}
}

// EntQuoteRepository provides database access to quotes.
type EntQuoteRepository struct {
	db    *sql.DB
	nodes modulekit.NodesRepository
}

// client - the module's ent client: the context's transaction (InTx),
// else the database.
func (r EntQuoteRepository) client(ctx context.Context) *generated.Client {
	return generated.NewClient(generated.Driver(modulekit.Driver(ctx, r.db)))
}

// Find returns the quotes matching filter, newest first, with their
// authors' IDs and names.
func (r EntQuoteRepository) Find(ctx context.Context, filter QuotesFilter) (Quotes, error) {
	query := r.client(ctx).Quote.Query()

	query = query.
		Limit(filter.Limit).
		Offset(filter.Offset).
		Select(
			quote.FieldID,
			quote.FieldText,
			quote.FieldCitation,
			quote.FieldCreatedAt,
		).
		Order(generated.Desc(quote.FieldCreatedAt)).
		WithAuthor(func(aq *generated.QuoteAuthorQuery) {
			aq.Select(
				quoteauthor.FieldID,
				quoteauthor.FieldName,
			)
		})

	if filter.AuthorID != nil {
		query = query.Where(quote.AuthorID(*filter.AuthorID))
	}

	if filter.AuthorName != nil {
		query = query.Where(quote.HasAuthorWith(quoteauthor.Name(*filter.AuthorName)))
	}

	res, err := query.All(ctx)

	if err != nil {
		return nil, err
	}

	return ToQuotesEntity(res), nil
}

// TotalCount returns the number of quotes matching filter.
func (r EntQuoteRepository) TotalCount(ctx context.Context, filter QuotesFilter) (int, error) {
	query := r.client(ctx).Quote.Query()

	if filter.AuthorID != nil {
		query = query.Where(quote.AuthorID(*filter.AuthorID))
	}

	if filter.AuthorName != nil {
		query = query.Where(quote.HasAuthorWith(quoteauthor.Name(*filter.AuthorName)))
	}

	totalCount, err := query.Count(ctx)
	if err != nil {
		return 0, err
	}

	return totalCount, nil
}

// ByID returns the quote with its author's ID and name.
func (r EntQuoteRepository) ByID(ctx context.Context, id int) (*Quote, error) {
	res, err := r.client(ctx).Quote.Query().
		Where(quote.ID(id)).
		WithAuthor(func(aq *generated.QuoteAuthorQuery) {
			aq.Select(
				quoteauthor.FieldID,
				quoteauthor.FieldName,
			)
		}).
		Only(ctx)

	if err != nil {
		return nil, err
	}

	return ToQuoteEntity(res), nil
}

// Random returns up to limit quotes in random order, with their authors.
func (r EntQuoteRepository) Random(ctx context.Context, limit int) (Quotes, error) {
	res, err := r.client(ctx).Quote.Query().
		Order(modulekit.OrderRandom).
		Limit(limit).
		Select(
			quote.FieldID,
			quote.FieldText,
			quote.FieldCitation,
		).
		WithAuthor(func(aq *generated.QuoteAuthorQuery) {
			aq.Select(
				quoteauthor.FieldID,
				quoteauthor.FieldName,
			)
		}).
		All(ctx)

	if err != nil {
		return nil, err
	}

	return ToQuotesEntity(res), nil
}

// AuthorsList returns the authors that have quotes, by name, with their
// quote counts.
func (r EntQuoteRepository) AuthorsList(ctx context.Context) ([]QuoteAuthorWithCount, error) {
	authorsEnt, err := r.client(ctx).QuoteAuthor.Query().
		WithQuotes(func(q *generated.QuoteQuery) {
			q.Select(quote.FieldID) // IDs only, for counting
		}).
		Order(generated.Asc(quoteauthor.FieldName)).
		All(ctx)

	if err != nil {
		return nil, err
	}

	authors := make([]QuoteAuthorWithCount, 0, len(authorsEnt))
	for _, a := range authorsEnt {
		quotesCount := 0
		if a.Edges.Quotes != nil {
			quotesCount = len(a.Edges.Quotes)
		}
		if quotesCount == 0 {
			continue
		}

		authors = append(authors, QuoteAuthorWithCount{
			QuoteAuthor: QuoteAuthor{
				ID:   a.ID,
				Name: a.Name,
			},
			QuotesCount: quotesCount,
		})
	}

	return authors, nil
}

// Create saves a new quote (its author by ID, if any) and returns it as
// read back.
func (r EntQuoteRepository) Create(ctx context.Context, q Quote) (*Quote, error) {
	create := r.client(ctx).Quote.Create().
		SetText(q.Text).
		SetNillableCitation(optionalString(q.Citation))

	if q.Author != nil {
		create = create.SetAuthorID(q.Author.ID)
	}

	res, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}

	return r.ByID(ctx, res.ID)
}

// Update saves the quote's text, source and author (nil - cleared) and
// returns it as read back.
func (r EntQuoteRepository) Update(ctx context.Context, q Quote) (*Quote, error) {
	update := r.client(ctx).Quote.UpdateOneID(q.ID).
		SetText(q.Text).
		SetNillableCitation(optionalString(q.Citation))

	if q.Author != nil {
		update = update.SetAuthorID(q.Author.ID)
	} else {
		update = update.ClearAuthorID()
	}

	_, err := update.Save(ctx)
	if err != nil {
		return nil, err
	}

	return r.ByID(ctx, q.ID)
}

// Delete deletes the quote.
func (r EntQuoteRepository) Delete(ctx context.Context, id int) error {
	return r.client(ctx).Quote.DeleteOneID(id).Exec(ctx)
}

// DeleteAuthor - see QuoteRepository.DeleteAuthor.
func (r EntQuoteRepository) DeleteAuthor(ctx context.Context, authorID int) error {
	return r.client(ctx).QuoteAuthor.DeleteOneID(authorID).Exec(ctx)
}

// UpdateAuthor - see QuoteRepository.UpdateAuthor; a material that
// doesn't exist is ErrAuthorNodeMissing.
func (r EntQuoteRepository) UpdateAuthor(ctx context.Context, authorID int, name string, nodeID *int) error {
	update := r.client(ctx).QuoteAuthor.UpdateOneID(authorID).SetName(name)
	if nodeID != nil {
		// The material is the core's: a missing one would fail the foreign
		// key.
		if _, err := r.nodes.ByID(ctx, *nodeID); err != nil {
			if modulekit.IsNotFound(err) {
				return fmt.Errorf("material %d: %w", *nodeID, ErrAuthorNodeMissing)
			}
			return fmt.Errorf("material %d: %w", *nodeID, err)
		}
		update = update.SetNodeID(*nodeID)
	} else {
		update = update.ClearNodeID()
	}
	return update.Exec(ctx)
}

// author - an author row.
func author(res *generated.QuoteAuthor) *QuoteAuthor {
	return &QuoteAuthor{ID: res.ID, Name: res.Name, NodeID: res.NodeID}
}

// AuthorByName returns the author with exactly the name (trimmed); none -
// modulekit.ErrEntityNotFound.
func (r EntQuoteRepository) AuthorByName(ctx context.Context, name string) (*QuoteAuthor, error) {
	name = trimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%w", modulekit.ErrEntityNotFound)
	}

	res, err := r.client(ctx).QuoteAuthor.Query().
		Where(quoteauthor.Name(name)).
		Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return nil, fmt.Errorf("%w", modulekit.ErrEntityNotFound)
		}
		return nil, err
	}

	return author(res), nil
}

// AuthorByID returns the author by ID; none - modulekit.ErrEntityNotFound.
func (r EntQuoteRepository) AuthorByID(ctx context.Context, id int) (*QuoteAuthor, error) {
	res, err := r.client(ctx).QuoteAuthor.Query().
		Where(quoteauthor.ID(id)).
		Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return nil, fmt.Errorf("%w", modulekit.ErrEntityNotFound)
		}
		return nil, err
	}

	return author(res), nil
}

// AuthorByNodeID - see QuoteRepository.AuthorByNodeID; none -
// modulekit.ErrEntityNotFound.
func (r EntQuoteRepository) AuthorByNodeID(ctx context.Context, nodeID int) (*QuoteAuthor, error) {
	res, err := r.client(ctx).QuoteAuthor.Query().
		Where(quoteauthor.NodeID(nodeID)).
		Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return nil, fmt.Errorf("%w", modulekit.ErrEntityNotFound)
		}
		return nil, err
	}

	return author(res), nil
}

// SearchAuthorsByName returns up to limit authors (0 - 10) whose name
// contains query, ignoring case, by name.
func (r EntQuoteRepository) SearchAuthorsByName(ctx context.Context, query string, limit int) ([]QuoteAuthor, error) {
	if limit <= 0 {
		limit = 10
	}

	authorsEnt, err := r.client(ctx).QuoteAuthor.Query().
		Where(quoteauthor.NameContainsFold(query)).
		Order(generated.Asc(quoteauthor.FieldName)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}

	authors := make([]QuoteAuthor, 0, len(authorsEnt))
	for _, a := range authorsEnt {
		authors = append(authors, QuoteAuthor{
			ID:   a.ID,
			Name: a.Name,
		})
	}

	return authors, nil
}

// CreateAuthor returns the author with the name (trimmed): the existing
// one, else a new one; an empty name - nil and no error.
func (r EntQuoteRepository) CreateAuthor(ctx context.Context, name string) (*QuoteAuthor, error) {
	name = trimSpace(name)
	if name == "" {
		return nil, nil
	}

	// Try to find existing author by exact name
	existing, err := r.client(ctx).QuoteAuthor.Query().
		Where(quoteauthor.Name(name)).
		Only(ctx)
	switch {
	case err == nil:
		return &QuoteAuthor{ID: existing.ID, Name: existing.Name}, nil
	case !generated.IsNotFound(err):
		return nil, fmt.Errorf("looking for quote author %q: %w", name, err)
	}

	created, err := r.client(ctx).QuoteAuthor.Create().
		SetName(name).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return &QuoteAuthor{ID: created.ID, Name: created.Name}, nil
}

func optionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func trimSpace(s string) string {
	return strings.TrimSpace(s)
}
