package quotes

import (
	"context"
	"time"
)

// QuotesPerPage - the quotes on a page of the quotes list.
const QuotesPerPage = 50

// Quote - a quote: its text (sanitized HTML), its source (Citation, "" -
// none) and its author (nil - none).
type Quote struct {
	ID        int
	Text      string
	Citation  string
	Author    *QuoteAuthor
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Quotes - a list of quotes.
type Quotes []Quote

// QuoteAuthor - an author of quotes, by a unique name; NodeID - the
// material about them, if linked (nil - none).
type QuoteAuthor struct {
	ID     int
	Name   string
	NodeID *int
}

// QuoteAuthorWithCount - an author with the number of their quotes (the
// authors list).
type QuoteAuthorWithCount struct {
	QuoteAuthor
	QuotesCount int
}

// QuotesFilter - which quotes Find returns and TotalCount counts: of an
// author, by ID or name (nil - any), a page of them (Offset, Limit).
type QuotesFilter struct {
	Limit      int
	Offset     int
	AuthorID   *int
	AuthorName *string
}

// QuoteRepository - the store of quotes and their authors.
type QuoteRepository interface {
	Find(ctx context.Context, filter QuotesFilter) (Quotes, error)
	ByID(ctx context.Context, id int) (*Quote, error)
	TotalCount(ctx context.Context, filter QuotesFilter) (int, error)
	Random(ctx context.Context, limit int) (Quotes, error)
	AuthorsList(ctx context.Context) ([]QuoteAuthorWithCount, error)
	Create(ctx context.Context, quote Quote) (*Quote, error)
	Update(ctx context.Context, quote Quote) (*Quote, error)
	Delete(ctx context.Context, id int) error
	SearchAuthorsByName(ctx context.Context, query string, limit int) ([]QuoteAuthor, error)
	AuthorByName(ctx context.Context, name string) (*QuoteAuthor, error)
	AuthorByID(ctx context.Context, id int) (*QuoteAuthor, error)
	// AuthorByNodeID finds the quote author linked to a material (node), if any.
	AuthorByNodeID(ctx context.Context, nodeID int) (*QuoteAuthor, error)
	CreateAuthor(ctx context.Context, name string) (*QuoteAuthor, error)
	// UpdateAuthor sets the author's name and linked material (nodeID == nil clears the link).
	UpdateAuthor(ctx context.Context, authorID int, name string, nodeID *int) error
	// DeleteAuthor deletes the author row (its quotes must be gone).
	DeleteAuthor(ctx context.Context, authorID int) error
}
