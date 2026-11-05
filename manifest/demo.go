package manifest

import (
	"context"
	"embed"
	"fmt"

	"github.com/kosatka-cms/kosatka-module-quotes/ent"
	"github.com/kosatka-cms/kosatka-module-quotes/ent/generated"
	"github.com/kosatka-cms/kosatka/modulekit/demokit"
	"github.com/kosatka-cms/kosatka/modulekit/manifestkit"
)

// demoFS - the module's demo content by language: demo/<lang>.json.
//
//go:embed demo/*.json
var demoFS embed.FS

// demoContent - the quotes (their authors are created by name), the title
// of the random quote block (not shown: the quote speaks for itself) and
// of the footer menu's link to /quotes.
type demoContent struct {
	Block  string      `json:"block"`
	Menu   string      `json:"menu"`
	Quotes []demoQuote `json:"quotes"`
}

type demoQuote struct {
	Text     string `json:"text"`
	Author   string `json:"author"`
	Citation string `json:"citation"`
}

// RegisterSteps registers the module's own ent schema (its tables and
// migration files) and its demo step (manifests.Register).
func RegisterSteps() {
	manifestkit.RegisterSchema(Name, ent.Schema)
	demokit.Register(Name, createDemo)
}

// createDemo adds the quotes, the random quote block at the end of the
// sidebar and a link to the quotes in the footer menu.
func createDemo(ctx context.Context, env demokit.Env) error {
	var content demoContent
	if err := demokit.Content(env, demoFS, &content); err != nil {
		return err
	}
	// The module's own client, in the install transaction.
	client := generated.NewClient(generated.Driver(demokit.Driver(env)))
	authors := map[string]int{}
	for _, q := range content.Quotes {
		id, ok := authors[q.Author]
		if !ok {
			a, err := client.QuoteAuthor.Create().SetName(q.Author).Save(ctx)
			if err != nil {
				return fmt.Errorf("author %q: %w", q.Author, err)
			}
			id, authors[q.Author] = a.ID, a.ID
		}
		if err := client.Quote.Create().
			SetText("<p>" + q.Text + "</p>").
			SetAuthorID(id).
			SetCitation(q.Citation).
			SetCreatedAt(env.Now).
			SetUpdatedAt(env.Now).
			Exec(ctx); err != nil {
			return err
		}
	}
	if err := demokit.PlacePlugin(ctx, env, PluginBlockRandomQuote, nil, demokit.Placement{Region: "sidebar", Title: content.Block, Weight: 30}); err != nil {
		return err
	}
	return demokit.MenuLink(ctx, env, env.FooterMenu, content.Menu, "/quotes", 0, 20)
}
