package quotes

import (
	"github.com/kosatka-cms/kosatka-module-quotes/ent/generated"
)

// ToQuoteEntity converts a generated quote to its domain entity, with its
// author's ID and name when the author was loaded.
func ToQuoteEntity(q *generated.Quote) *Quote {
	if q == nil {
		return nil
	}

	quote := &Quote{
		ID:        q.ID,
		Text:      q.Text,
		Citation:  q.Citation,
		CreatedAt: q.CreatedAt,
		UpdatedAt: q.UpdatedAt,
	}

	if q.Edges.Author != nil {
		quote.Author = &QuoteAuthor{
			ID:   q.Edges.Author.ID,
			Name: q.Edges.Author.Name,
		}
	}

	return quote
}

// ToQuotesEntity converts generated quotes to domain entities.
func ToQuotesEntity(quotes []*generated.Quote) Quotes {
	result := make(Quotes, 0, len(quotes))

	for _, q := range quotes {
		result = append(result, *ToQuoteEntity(q))
	}

	return result
}
