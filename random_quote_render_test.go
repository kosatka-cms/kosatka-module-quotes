package quotes

import (
	"strings"
	"testing"

	"github.com/kosatka-cms/kosatka/modulekit/testkit"
)

// TestRandomQuoteWithoutAuthor - a quote with no author renders without
// the author line instead of failing (nil Author).
func TestRandomQuoteWithoutAuthor(t *testing.T) {
	testkit.UseEngineRoot(t)
	for _, q := range []*Quote{
		{Text: "No author."},
		{Text: "With author.", Author: &QuoteAuthor{Name: "Someone"}},
	} {
		out := testkit.RenderNamed(t, "quote-content", map[string]any{"randomQuote": q}, "blocks/random_quote.html")
		hasAuthor := strings.Contains(out, "quotes-author")
		if hasAuthor != (q.Author != nil) {
			t.Errorf("%q: author line = %v", q.Text, hasAuthor)
		}
	}
}
