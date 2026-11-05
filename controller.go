package quotes

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/kosatka-cms/kosatka-module-quotes/manifest"
	"github.com/kosatka-cms/kosatka/modulekit"
	"github.com/kosatka-cms/kosatka/pkg/htmlsanitize"
	"github.com/kosatka-cms/kosatka/pkg/httputil"
	"github.com/kosatka-cms/kosatka/pkg/i18n"
	"github.com/kosatka-cms/kosatka/pkg/metadescription"
	"github.com/kosatka-cms/kosatka/pkg/pagination"
)

// QuotesController - the quotes pages: the list, the authors, an author's
// quotes, the editors' forms and the random quote API.
type QuotesController struct {
	*modulekit.BaseController
	quoteService QuoteService
}

// QuotesDeps - what NewQuotesController needs besides the BaseDeps.
type QuotesDeps struct {
	QuoteService QuoteService
}

// NewQuotesController - the quotes controller on the base dependencies
// and the quotes service (d).
func NewQuotesController(deps modulekit.BaseDeps, d QuotesDeps) *QuotesController {
	base := modulekit.NewBaseController(deps)
	base.SetLogger(deps.Logger.With("component", "quotes.handler"))
	return &QuotesController{
		BaseController: base,
		quoteService:   d.QuoteService,
	}
}

// List renders the quotes list (GET /quotes): a page of all quotes, newest
// first, with the edit links for editors.
func (c *QuotesController) List(w http.ResponseWriter, r *http.Request) {
	data, err := c.listData(r)
	if err != nil {
		c.LogFailure(err, "error getting quotes list data", "path", r.URL.Path)
		httputil.InternalServerError(w)
		return
	}

	title, _ := data["title"].(string)
	c.SetPathTrail(data, "/quotes", title)

	c.RenderOrFail(w, r, "quotes/list.html", data)
}

// AuthorsList renders the authors list (GET /quotes/byauthor): the authors
// that have quotes, by name, with their counts.
func (c *QuotesController) AuthorsList(w http.ResponseWriter, r *http.Request) {
	dbCtx, dbCancel := c.DBContext(r)
	defer dbCancel()

	authors, err := c.quoteService.AuthorsList(dbCtx)
	if err != nil {
		c.LogFailure(err, "error fetching authors list")
		httputil.InternalServerError(w)
		return
	}

	pageTitle := i18n.T(r.Context(), "By author")

	data := c.GetCommonData(r)
	newData := map[string]any{
		"authors": authors,
		"pageSEO": modulekit.PageSEOForList(pageTitle, r.URL.Path),
		"title":   pageTitle,
	}

	maps.Copy(data, newData)
	c.SetChildTrail(data, quotesTrailParents(r.Context()), pageTitle)

	c.RenderOrFail(w, r, "quotes/authors.html", data)
}

// ByAuthor renders the quotes of an author (GET /quotes/author/{author...}),
// the author by name: a page of them; an unknown author is a 404.
func (c *QuotesController) ByAuthor(w http.ResponseWriter, r *http.Request) {
	authorName := strings.TrimSpace(r.PathValue("author"))
	if authorName == "" {
		c.Logger().Warn("empty author name")
		c.RenderNotFound(w, r)
		return
	}

	data, err := c.listDataByAuthorName(w, r, authorName)
	if err != nil {
		c.RenderLoadFailure(w, r, err, "error getting quotes by author data", "author_name", authorName)
		return
	}

	if author, ok := data["author"].(*QuoteAuthor); ok && author != nil {
		c.SetChildTrail(data, quoteAuthorsTrailParents(r.Context()), author.Name)
	}

	c.RenderOrFail(w, r, "quotes/by_author.html", data)
}

// listData - the quotes page's data: a page of every quote.
func (c *QuotesController) listData(r *http.Request) (map[string]any, error) {
	dbCtx, dbCancel := c.DBContext(r)
	defer dbCancel()

	page := c.GetPageNumber(r)
	itemsPerPage := QuotesPerPage

	filter := QuotesFilter{}

	totalCount, err := c.quoteService.TotalCount(dbCtx, filter)
	if err != nil {
		return nil, fmt.Errorf("counting the quotes: %w", err)
	}

	pager := pagination.New(
		page,
		totalCount,
		itemsPerPage,
		r.URL.Path,
		r.URL.Query(),
	)

	filter.Offset = pager.Offset
	filter.Limit = pager.ItemsPerPage

	quotes, err := c.quoteService.Find(dbCtx, filter)
	if err != nil {
		return nil, fmt.Errorf("listing the quotes: %w", err)
	}

	pageTitle := i18n.T(r.Context(), "Quotes")

	user, _ := c.GetCurrentUser(r)
	canEditQuotes := c.AuthService().HasPermission(r.Context(), user, manifest.PermManage)

	data := c.GetCommonData(r)
	newData := map[string]any{
		"title":         pageTitle,
		"quotes":        quotes,
		"pagination":    pager,
		"pageSEO":       modulekit.PageSEOForPaginatedList(pageTitle, r.URL.Path, page, r.URL.Query()),
		"canEditQuotes": canEditQuotes,
	}

	maps.Copy(data, newData)

	return data, nil
}

func (c *QuotesController) listDataByAuthorName(
	w http.ResponseWriter,
	r *http.Request,
	authorName string,
) (map[string]any, error) {
	dbCtx, dbCancel := c.DBContext(r)
	defer dbCancel()

	author, err := c.quoteService.AuthorByName(dbCtx, authorName)
	if err != nil {
		return nil, err
	}

	page := c.GetPageNumber(r)
	itemsPerPage := QuotesPerPage

	filter := QuotesFilter{
		AuthorID: &author.ID,
	}

	totalCount, err := c.quoteService.TotalCount(dbCtx, filter)
	if err != nil {
		return nil, fmt.Errorf("getting quotes total count: %w", err)
	}

	pager := pagination.New(
		page,
		totalCount,
		itemsPerPage,
		r.URL.Path,
		r.URL.Query(),
	)

	filter.Offset = pager.Offset
	filter.Limit = pager.ItemsPerPage

	quotes, err := c.quoteService.Find(dbCtx, filter)
	if err != nil {
		return nil, fmt.Errorf("getting quotes list: %w", err)
	}

	pageTitle := i18n.T(r.Context(), "Quotes: @author", "@author", author.Name)

	user, _ := c.GetCurrentUser(r)
	canEditQuotes := c.AuthService().HasPermission(r.Context(), user, manifest.PermManage)

	pageSEO := modulekit.PageSEOForPaginatedList(pageTitle, r.URL.Path, page, r.URL.Query())
	pageSEO.Description = metadescription.Generate(
		i18n.T(r.Context(), "@author - quotes.", "@author", author.Name),
		modulekit.MaxMetaDesc,
	)

	data := c.GetCommonData(r)
	newData := map[string]any{
		"title":         pageTitle,
		"author":        author,
		"quotes":        quotes,
		"pagination":    pager,
		"pageSEO":       pageSEO,
		"canEditQuotes": canEditQuotes,
	}

	maps.Copy(data, newData)

	return data, nil
}

func (c *QuotesController) requireQuoteEditor(w http.ResponseWriter, r *http.Request) (*modulekit.User, bool) {
	user, err := c.GetCurrentUser(r)
	if err != nil || user == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return nil, false
	}
	if !c.AuthService().HasPermission(r.Context(), user, manifest.PermManage) {
		c.RenderForbidden(w, r)
		return nil, false
	}
	return user, true
}

// CreateForm renders the form of a new quote (GET /quotes/create) to the
// quotes' editors.
func (c *QuotesController) CreateForm(w http.ResponseWriter, r *http.Request) {
	if _, ok := c.requireQuoteEditor(w, r); !ok {
		return
	}

	authors, err := c.quoteService.AuthorsList(r.Context())
	if err != nil {
		c.LogFailure(err, "error fetching authors for form")
		httputil.InternalServerError(w)
		return
	}

	data := c.GetCommonData(r)
	data["quote"] = (*Quote)(nil)
	data["authors"] = authors
	data["initialAuthorJSON"] = "null"
	pageTitle := i18n.T(r.Context(), "Add quote")
	data["pageSEO"] = modulekit.PageSEOForList(pageTitle, r.URL.Path)
	data["title"] = pageTitle
	c.SetChildTrail(data, quotesTrailParents(r.Context()), pageTitle)

	c.RenderOrFail(w, r, "quotes/form.html", data)
}

// Create saves a new quote from its form (POST /quotes/create): its
// author chosen by ID or found or created by name; then the quotes list.
func (c *QuotesController) Create(w http.ResponseWriter, r *http.Request) {
	if _, ok := c.requireQuoteEditor(w, r); !ok {
		return
	}

	dbCtx, dbCancel := c.DBContext(r)
	defer dbCancel()

	// Rich text from the editor, output unescaped: sanitized on save.
	text := strings.TrimSpace(htmlsanitize.SanitizeAdmin(r.FormValue("text")))
	if text == "" {
		c.RenderBadRequest(w, r, i18n.T(r.Context(), "Quote text is required"))
		return
	}

	quote := Quote{
		Text:     text,
		Citation: strings.TrimSpace(r.FormValue("citation")),
	}

	authorIDStr := strings.TrimSpace(r.FormValue("author_id"))
	authorName := strings.TrimSpace(r.FormValue("author_name"))

	if authorIDStr != "" {
		authorID, err := strconv.Atoi(authorIDStr)
		if err == nil && authorID > 0 {
			quote.Author = &QuoteAuthor{ID: authorID}
		}
	} else if authorName != "" {
		author, err := c.quoteService.CreateAuthor(dbCtx, authorName)
		if err != nil {
			c.LogFailure(err, "error creating author", "name", authorName)
			httputil.InternalServerError(w)
			return
		}
		if author != nil {
			quote.Author = author
		}
	}

	created, err := c.quoteService.Create(dbCtx, quote)
	if err != nil {
		c.LogFailure(err, "error creating quote")
		httputil.InternalServerError(w)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/quotes?created=%d", created.ID), http.StatusFound)
}

// EditForm renders the quote's edit form (GET /quotes/edit/{id}) to the
// quotes' editors.
func (c *QuotesController) EditForm(w http.ResponseWriter, r *http.Request) {
	if _, ok := c.requireQuoteEditor(w, r); !ok {
		return
	}

	id, ok := c.PathID(r)
	if !ok {
		c.RenderNotFound(w, r)
		return
	}

	dbCtx, dbCancel := c.DBContext(r)
	defer dbCancel()

	quote, err := c.quoteService.ByID(dbCtx, id)
	if err != nil {
		c.RenderLoadFailure(w, r, err, "loading the quote", "quoteID", id)
		return
	}

	authors, err := c.quoteService.AuthorsList(dbCtx)
	if err != nil {
		c.LogFailure(err, "error fetching authors for form")
		httputil.InternalServerError(w)
		return
	}

	initialAuthorJSON := "null"
	if quote.Author != nil {
		b, _ := json.Marshal(map[string]any{"id": quote.Author.ID, "value": quote.Author.Name})
		initialAuthorJSON = string(b)
	}

	data := c.GetCommonData(r)
	data["quote"] = quote
	data["authors"] = authors
	data["initialAuthorJSON"] = initialAuthorJSON
	pageTitle := i18n.T(r.Context(), "Edit quote")
	data["pageSEO"] = modulekit.PageSEOForList(pageTitle, r.URL.Path)
	data["title"] = pageTitle
	c.SetChildTrail(data, quotesTrailParents(r.Context()), pageTitle)

	c.RenderOrFail(w, r, "quotes/form.html", data)
}

// Edit saves the quote from its form (POST /quotes/edit/{id}): its text,
// source and author (chosen by ID, by name, or none); then the quotes list.
func (c *QuotesController) Edit(w http.ResponseWriter, r *http.Request) {
	if _, ok := c.requireQuoteEditor(w, r); !ok {
		return
	}

	id, ok := c.PathID(r)
	if !ok {
		c.RenderNotFound(w, r)
		return
	}

	dbCtx, dbCancel := c.DBContext(r)
	defer dbCancel()

	quote, err := c.quoteService.ByID(dbCtx, id)
	if err != nil {
		c.RenderLoadFailure(w, r, err, "loading the quote", "quoteID", id)
		return
	}

	// Rich text from the editor, output unescaped: sanitized on save.
	text := strings.TrimSpace(htmlsanitize.SanitizeAdmin(r.FormValue("text")))
	if text == "" {
		c.RenderBadRequest(w, r, i18n.T(r.Context(), "Quote text is required"))
		return
	}

	quote.Text = text
	quote.Citation = strings.TrimSpace(r.FormValue("citation"))

	authorIDStr := strings.TrimSpace(r.FormValue("author_id"))
	authorName := strings.TrimSpace(r.FormValue("author_name"))

	quote.Author = nil
	if authorIDStr != "" {
		authorID, err := strconv.Atoi(authorIDStr)
		if err == nil && authorID > 0 {
			quote.Author = &QuoteAuthor{ID: authorID}
		}
	} else if authorName != "" {
		author, err := c.quoteService.CreateAuthor(dbCtx, authorName)
		if err != nil {
			c.LogFailure(err, "error creating author", "name", authorName)
			httputil.InternalServerError(w)
			return
		}
		if author != nil {
			quote.Author = author
		}
	}

	_, err = c.quoteService.Update(dbCtx, *quote)
	if err != nil {
		c.LogFailure(err, "error updating quote", "id", id)
		httputil.InternalServerError(w)
		return
	}

	http.Redirect(w, r, "/quotes", http.StatusFound)
}

// DeleteConfirm renders the confirmation of deleting the quote (GET
// /quotes/delete/{id}) to the quotes' editors.
func (c *QuotesController) DeleteConfirm(w http.ResponseWriter, r *http.Request) {
	if _, ok := c.requireQuoteEditor(w, r); !ok {
		return
	}

	id, ok := c.PathID(r)
	if !ok {
		c.RenderNotFound(w, r)
		return
	}

	dbCtx, dbCancel := c.DBContext(r)
	defer dbCancel()

	quote, err := c.quoteService.ByID(dbCtx, id)
	if err != nil {
		c.RenderLoadFailure(w, r, err, "loading the quote", "quoteID", id)
		return
	}

	data := c.GetCommonData(r)
	data["quote"] = quote
	pageTitle := i18n.T(r.Context(), "Delete quote")
	data["pageSEO"] = modulekit.PageSEOForList(pageTitle, r.URL.Path)
	data["title"] = pageTitle
	c.SetChildTrail(data, quotesTrailParents(r.Context()), pageTitle)

	c.RenderOrFail(w, r, "quotes/delete.html", data)
}

// Delete deletes the quote (POST /quotes/delete/{id}); then the quotes
// list.
func (c *QuotesController) Delete(w http.ResponseWriter, r *http.Request) {
	if _, ok := c.requireQuoteEditor(w, r); !ok {
		return
	}

	id, ok := c.PathID(r)
	if !ok {
		c.RenderNotFound(w, r)
		return
	}

	dbCtx, dbCancel := c.DBContext(r)
	defer dbCancel()

	err := c.quoteService.Delete(dbCtx, id)
	if err != nil {
		c.LogFailure(err, "error deleting quote", "id", id)
		httputil.InternalServerError(w)
		return
	}

	http.Redirect(w, r, "/quotes", http.StatusFound)
}

// AuthorDelete deletes an author without quotes (POST
// /quotes/authors/delete/{id}); one with quotes is a bad request.
func (c *QuotesController) AuthorDelete(w http.ResponseWriter, r *http.Request) {
	if _, ok := c.requireQuoteEditor(w, r); !ok {
		return
	}
	id, ok := c.PathID(r)
	if !ok {
		c.RenderNotFound(w, r)
		return
	}
	dbCtx, dbCancel := c.DBContext(r)
	defer dbCancel()
	if err := c.quoteService.DeleteAuthor(dbCtx, id); err != nil {
		if errors.Is(err, ErrAuthorHasQuotes) {
			c.RenderBadRequest(w, r, i18n.T(r.Context(), "The author has quotes: delete them first"))
			return
		}
		c.RenderLoadFailure(w, r, err, "deleting the quote author", "author_id", id)
		return
	}
	http.Redirect(w, r, "/quotes/byauthor", http.StatusFound)
}

// AuthorsSearch answers the quote form's author autocomplete (GET
// /api/quotes/authors/search?q=…): up to 10 authors whose name has q, as
// JSON suggestions (a q under 2 characters - none).
func (c *QuotesController) AuthorsSearch(w http.ResponseWriter, r *http.Request) {
	if _, ok := c.requireQuoteEditor(w, r); !ok {
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(query) < 2 {
		httputil.SendJSON(w, http.StatusOK, map[string]any{"suggestions": []any{}})
		return
	}

	dbCtx, dbCancel := c.DBContext(r)
	defer dbCancel()

	authors, err := c.quoteService.SearchAuthorsByName(dbCtx, query, 10)
	if err != nil {
		c.LogFailure(err, "error searching authors", "query", query)
		httputil.JSONError(w, "Error searching authors", http.StatusInternalServerError)
		return
	}

	type Suggestion struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	suggestions := make([]Suggestion, len(authors))
	for i, a := range authors {
		suggestions[i] = Suggestion{ID: a.ID, Name: a.Name}
	}

	response := map[string]any{
		"suggestions": suggestions,
	}
	httputil.JSON(w, response, http.StatusOK)
}

// AuthorEditForm displays the quote author edit form (name and linked node_id).
func (c *QuotesController) AuthorEditForm(w http.ResponseWriter, r *http.Request) {
	if _, ok := c.requireQuoteEditor(w, r); !ok {
		return
	}

	id, ok := c.PathID(r)
	if !ok {
		c.RenderNotFound(w, r)
		return
	}

	dbCtx, dbCancel := c.DBContext(r)
	defer dbCancel()

	author, err := c.quoteService.AuthorByID(dbCtx, id)
	if err != nil {
		c.RenderLoadFailure(w, r, err, "loading the quote's author", "authorID", id)
		return
	}

	// An author without quotes can be deleted from the form.
	quotes, err := c.quoteService.TotalCount(dbCtx, QuotesFilter{AuthorID: &id})
	if err != nil {
		c.RenderLoadFailure(w, r, err, "counting the author's quotes", "author_id", id)
		return
	}

	data := c.GetCommonData(r)
	data["author"] = author
	data["authorQuotes"] = quotes
	pageTitle := i18n.T(r.Context(), "Edit author")
	data["pageSEO"] = modulekit.PageSEOForList(pageTitle, r.URL.Path)
	data["title"] = pageTitle
	c.SetChildTrail(data, quoteAuthorsTrailParents(r.Context()), pageTitle)

	c.RenderOrFail(w, r, "quotes/author_form.html", data)
}

// AuthorEdit saves the quote author's name and linked material (node_id).
func (c *QuotesController) AuthorEdit(w http.ResponseWriter, r *http.Request) {
	if _, ok := c.requireQuoteEditor(w, r); !ok {
		return
	}

	id, ok := c.PathID(r)
	if !ok {
		c.RenderNotFound(w, r)
		return
	}

	dbCtx, dbCancel := c.DBContext(r)
	defer dbCancel()

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		c.RenderBadRequest(w, r, i18n.T(r.Context(), "Author name is required"))
		return
	}

	var nodeID *int
	nodeIDStr := strings.TrimSpace(r.FormValue("node_id"))
	if nodeIDStr != "" {
		parsedID, err := strconv.Atoi(nodeIDStr)
		if err != nil || parsedID <= 0 {
			c.RenderBadRequest(w, r, i18n.T(r.Context(), "Invalid content ID"))
			return
		}
		nodeID = &parsedID
	}

	if err := c.quoteService.UpdateAuthor(dbCtx, id, name, nodeID); err != nil {
		switch {
		case errors.Is(err, ErrAuthorNameTaken):
			c.RenderBadRequest(w, r, i18n.T(r.Context(), "Another author already has this name"))
		case errors.Is(err, ErrAuthorNodeMissing):
			c.RenderBadRequest(w, r, i18n.T(r.Context(), "No content with this ID"))
		default:
			c.LogFailure(err, "updating the quote author", "author_id", id)
			httputil.InternalServerError(w)
		}
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/quotes/author/%s", url.PathEscape(name)), http.StatusFound)
}

// LoadRandomQuote handles AJAX requests for a random quote.
func (c *QuotesController) LoadRandomQuote(w http.ResponseWriter, r *http.Request) {
	dbCtx, dbCancel := c.DBContext(r)
	defer dbCancel()

	quotes, err := c.quoteService.Random(dbCtx, 1)
	if err != nil {
		c.LogFailure(err, "error fetching random quote")
		httputil.InternalServerError(w)
		return
	}

	// No quotes: empty response.
	if len(quotes) == 0 {
		httputil.SendJSON(w, http.StatusOK, map[string]any{"html": "", "found": false})
		return
	}

	data := map[string]any{
		"randomQuote":             quotes[0],
		modulekit.RequestThemeKey: modulekit.RequestTheme(r.Context()),
	}

	var buf bytes.Buffer
	err = c.TemplateService().RenderNamed(&buf, "quote-content", data, "blocks/random_quote.html")
	if err != nil {
		c.LogFailure(err, "error rendering random quote template")
		httputil.InternalServerError(w)
		return
	}

	httputil.SendJSON(w, http.StatusOK, map[string]any{"html": buf.String(), "found": true})
}

// quotesTrailParents - the quotes list (a page's trail parent).
func quotesTrailParents(ctx context.Context) []modulekit.BreadcrumbItem {
	return []modulekit.BreadcrumbItem{{URL: "/quotes", Label: i18n.T(ctx, "Quotes")}}
}

// quoteAuthorsTrailParents - the quotes list, then the authors list.
func quoteAuthorsTrailParents(ctx context.Context) []modulekit.BreadcrumbItem {
	return append(quotesTrailParents(ctx), modulekit.BreadcrumbItem{URL: "/quotes/byauthor", Label: i18n.T(ctx, "By author")})
}
