package quotes

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/kosatka-cms/kosatka/modulekit"
)

// QuoteService - the quotes and their authors.
type QuoteService interface {
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
	AuthorByNodeID(ctx context.Context, nodeID int) (*QuoteAuthor, error)
	CreateAuthor(ctx context.Context, name string) (*QuoteAuthor, error)
	UpdateAuthor(ctx context.Context, authorID int, name string, nodeID *int) error
	DeleteAuthor(ctx context.Context, authorID int) error
}

// DefaultQuoteService - QuoteService on a QuoteRepository.
type DefaultQuoteService struct {
	repository QuoteRepository
	logger     *slog.Logger
}

// NewQuoteService - the quotes service on r.
func NewQuoteService(r QuoteRepository, baseLogger *slog.Logger) QuoteService {
	return &DefaultQuoteService{
		repository: r,
		logger:     baseLogger.With("component", "quote.service"),
	}
}

// Find - the quotes matching filter, newest first, with their authors.
func (s *DefaultQuoteService) Find(ctx context.Context, filter QuotesFilter) (Quotes, error) {
	quotes, err := s.repository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("finding quotes: %w", err)
	}

	return quotes, nil
}

// ByID - the quote with its author.
func (s *DefaultQuoteService) ByID(ctx context.Context, id int) (*Quote, error) {
	quote, err := s.repository.ByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting quote by id (id %v): %w", id, err)
	}

	return quote, nil
}

// TotalCount - the number of quotes matching filter.
func (s *DefaultQuoteService) TotalCount(ctx context.Context, filter QuotesFilter) (int, error) {
	total, err := s.repository.TotalCount(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("quotes total count: %w", err)
	}

	return total, nil
}

// Random - up to limit random quotes with their authors.
func (s *DefaultQuoteService) Random(ctx context.Context, limit int) (Quotes, error) {
	quotes, err := s.repository.Random(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("getting random quotes (limit %v): %w", limit, err)
	}

	return quotes, nil
}

// AuthorsList - the authors that have quotes, by name, with their counts.
func (s *DefaultQuoteService) AuthorsList(ctx context.Context) ([]QuoteAuthorWithCount, error) {
	authors, err := s.repository.AuthorsList(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting authors list: %w", err)
	}
	return authors, nil
}

// Create saves a new quote (its author by ID, if any) and returns it as
// stored.
func (s *DefaultQuoteService) Create(ctx context.Context, quote Quote) (*Quote, error) {
	created, err := s.repository.Create(ctx, quote)
	if err != nil {
		return nil, fmt.Errorf("creating quote: %w", err)
	}
	return created, nil
}

// Update saves the quote's text, source and author (nil - none) and
// returns it as stored.
func (s *DefaultQuoteService) Update(ctx context.Context, quote Quote) (*Quote, error) {
	updated, err := s.repository.Update(ctx, quote)
	if err != nil {
		return nil, fmt.Errorf("updating quote (id %v): %w", quote.ID, err)
	}
	return updated, nil
}

// Delete deletes the quote.
func (s *DefaultQuoteService) Delete(ctx context.Context, id int) error {
	err := s.repository.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("deleting quote (id %v): %w", id, err)
	}
	return nil
}

// SearchAuthorsByName - up to limit authors (0 - 10) whose name contains
// query, ignoring case, by name.
func (s *DefaultQuoteService) SearchAuthorsByName(ctx context.Context, query string, limit int) ([]QuoteAuthor, error) {
	authors, err := s.repository.SearchAuthorsByName(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("searching authors (query %v): %w", query, err)
	}
	return authors, nil
}

// AuthorByName - the author with exactly the name; none - a not-found
// error (modulekit.IsNotFound).
func (s *DefaultQuoteService) AuthorByName(ctx context.Context, name string) (*QuoteAuthor, error) {
	author, err := s.repository.AuthorByName(ctx, name)
	if err != nil {
		// Logged by the handler (an unknown author is its 404).
		return nil, fmt.Errorf("author by name (name %v): %w", name, err)
	}
	return author, nil
}

// CreateAuthor - the author with the name: the existing one, else a new
// one (an empty name - nil).
func (s *DefaultQuoteService) CreateAuthor(ctx context.Context, name string) (*QuoteAuthor, error) {
	author, err := s.repository.CreateAuthor(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("creating author (name %v): %w", name, err)
	}
	return author, nil
}

// AuthorByID - the author with their linked material; none - a not-found
// error.
func (s *DefaultQuoteService) AuthorByID(ctx context.Context, id int) (*QuoteAuthor, error) {
	author, err := s.repository.AuthorByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting author by id (id %v): %w", id, err)
	}
	return author, nil
}

// AuthorByNodeID - the quote author linked to the material (node); none -
// a not-found error.
func (s *DefaultQuoteService) AuthorByNodeID(ctx context.Context, nodeID int) (*QuoteAuthor, error) {
	return s.repository.AuthorByNodeID(ctx, nodeID)
}

// The author form's refusals (UpdateAuthor).
var (
	// ErrAuthorNameTaken - another author has the name.
	ErrAuthorNameTaken = errors.New("another author has the name")
	// ErrAuthorNodeMissing - no material with the linked ID.
	ErrAuthorNodeMissing = errors.New("no material with the ID")
	// ErrAuthorHasQuotes - the author still has quotes (DeleteAuthor).
	ErrAuthorHasQuotes = errors.New("the author has quotes")
)

// DeleteAuthor deletes an author without quotes; one with quotes is
// refused (ErrAuthorHasQuotes): its quotes go first.
func (s *DefaultQuoteService) DeleteAuthor(ctx context.Context, authorID int) error {
	n, err := s.repository.TotalCount(ctx, QuotesFilter{AuthorID: &authorID})
	if err != nil {
		return fmt.Errorf("counting the quotes of author %d: %w", authorID, err)
	}
	if n > 0 {
		return fmt.Errorf("author %d (%d quotes): %w", authorID, n, ErrAuthorHasQuotes)
	}
	if err := s.repository.DeleteAuthor(ctx, authorID); err != nil {
		return fmt.Errorf("deleting author %d: %w", authorID, err)
	}
	return nil
}

// UpdateAuthor renames the author and links its material (nodeID nil -
// none): a name another author has (ErrAuthorNameTaken) or a material
// that doesn't exist (ErrAuthorNodeMissing) is refused.
func (s *DefaultQuoteService) UpdateAuthor(ctx context.Context, authorID int, name string, nodeID *int) error {
	switch other, err := s.repository.AuthorByName(ctx, name); {
	case err == nil && other.ID != authorID:
		return fmt.Errorf("renaming author %d to %q: %w", authorID, name, ErrAuthorNameTaken)
	case err != nil && !modulekit.IsNotFound(err):
		return fmt.Errorf("the author named %q: %w", name, err)
	}
	if err := s.repository.UpdateAuthor(ctx, authorID, name, nodeID); err != nil {
		return fmt.Errorf("updating author (author_id %v): %w", authorID, err)
	}
	return nil
}
