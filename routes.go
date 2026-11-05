package quotes

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/kosatka-cms/kosatka/modulekit"
)

// RegisterQuotesRoutes registers the quotes pages, the random quote API
// and the editors' forms.
func RegisterQuotesRoutes(r modulekit.Mux, store *sessions.CookieStore, qc *QuotesController) {
	// API endpoints (public)
	r.HandleFunc("GET /api/quote/random", qc.LoadRandomQuote)

	// Regular routes
	r.HandleFunc("GET /quotes", qc.List)
	r.HandleFunc("GET /quotes/byauthor", qc.AuthorsList)
	r.HandleFunc("GET /quotes/author/{author...}", qc.ByAuthor)

	// Admin/moderator routes (auth + permission check inside handlers)
	// Use /quotes/edit/{id} and /quotes/delete/{id} to avoid conflict with /quotes/author/{author}
	auth := modulekit.AuthMiddleware(store)
	r.Handle("GET /api/quotes/authors/search", auth(http.HandlerFunc(qc.AuthorsSearch)))
	r.Handle("GET /quotes/authors/edit/{id}", auth(http.HandlerFunc(qc.AuthorEditForm)))
	r.Handle("POST /quotes/authors/edit/{id}", auth(http.HandlerFunc(qc.AuthorEdit)))
	r.Handle("POST /quotes/authors/delete/{id}", auth(http.HandlerFunc(qc.AuthorDelete)))
	r.Handle("GET /quotes/create", auth(http.HandlerFunc(qc.CreateForm)))
	r.Handle("POST /quotes/create", auth(http.HandlerFunc(qc.Create)))
	r.Handle("GET /quotes/edit/{id}", auth(http.HandlerFunc(qc.EditForm)))
	r.Handle("POST /quotes/edit/{id}", auth(http.HandlerFunc(qc.Edit)))
	r.Handle("GET /quotes/delete/{id}", auth(http.HandlerFunc(qc.DeleteConfirm)))
	r.Handle("POST /quotes/delete/{id}", auth(http.HandlerFunc(qc.Delete)))
}
