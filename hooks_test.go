package quotes

import (
	"context"
	"testing"

	"github.com/kosatka-cms/kosatka/modulekit"
)

type fakeQuotes struct {
	QuoteService
	author *QuoteAuthor
	count  int
}

func (f fakeQuotes) AuthorByNodeID(context.Context, int) (*QuoteAuthor, error) { return f.author, nil }

func (f fakeQuotes) TotalCount(context.Context, QuotesFilter) (int, error) { return f.count, nil }

// TestNodePage - a node with a linked quote author gets "Quotes (N)" to
// the author's quotes; without one, nothing (the core skips the hook while
// the module is off).
func TestNodePage(t *testing.T) {
	ctx := context.Background()
	links := NodePage(fakeQuotes{author: &QuoteAuthor{ID: 2, Name: "Stan Lee"}, count: 12})(ctx, modulekit.Node{ID: 5}).Links
	if len(links) != 1 || links[0].URL != "/quotes/author/Stan%20Lee" || links[0].Label != "Quotes (12)" {
		t.Errorf("links: %+v", links)
	}
	if links := NodePage(fakeQuotes{})(ctx, modulekit.Node{ID: 5}).Links; len(links) != 0 {
		t.Errorf("no author: %+v", links)
	}
}
