package quotes

import (
	"github.com/kosatka-cms/kosatka-module-quotes/manifest"
	"github.com/kosatka-cms/kosatka/modulekit"
)

// Wire - the module's service, controller, node page and "Create" hooks
// and routes (modulekit).
func Wire(d modulekit.Deps) modulekit.Parts {
	quotes := NewQuoteService(NewQuoteRepository(d.DB, d.NodeRepo), d.Logger)
	modulekit.AddNodePage(manifest.Name, NodePage(quotes))
	modulekit.AddCreateLinks(manifest.Name, CreateLinks)
	c := NewQuotesController(d.Base, QuotesDeps{QuoteService: quotes})
	return modulekit.Parts{Routes: func(r modulekit.Routes) {
		RegisterQuotesRoutes(r.Mux, r.Store, c)
	}}
}
