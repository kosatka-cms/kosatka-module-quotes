package quotes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kosatka-cms/kosatka/modulekit"
	"github.com/kosatka-cms/kosatka/modulekit/testkit"
)

type fakeRandom struct {
	QuoteService
}

func (fakeRandom) Random(context.Context, int) (Quotes, error) {
	return Quotes{{Text: "Great power.", Author: &QuoteAuthor{Name: "Ben"}}}, nil
}

// TestLoadRandomQuote - the random quote API renders the module's quote
// partial (the block's placeholder fills itself from it).
func TestLoadRandomQuote(t *testing.T) {
	testkit.UseEngineRoot(t)
	c := &QuotesController{
		BaseController: modulekit.NewBaseController(modulekit.BaseDeps{Templates: testkit.Templates(), Logger: testkit.Logger()}),
		quoteService:   fakeRandom{},
	}
	rec := httptest.NewRecorder()
	c.LoadRandomQuote(rec, httptest.NewRequest(http.MethodGet, "/api/quote/random", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Great power.") || !strings.Contains(rec.Body.String(), `"found":true`) {
		t.Errorf("%d %s", rec.Code, rec.Body.String())
	}
}
