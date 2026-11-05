// Package manifest declares the quotes module (its registry entry): what
// it adds to the core.
package manifest

import (
	"github.com/kosatka-cms/kosatka/modulekit/manifestkit"
	"github.com/kosatka-cms/kosatka/pkg/i18n"
)

// Name - the module's machine name.
const Name = "quotes"

// PermManage - add, edit and delete quotes and their authors.
const PermManage = "manage quotes"

// PluginBlockRandomQuote - lazy-loaded on the client (GET
// /api/quote/random) and rendered fresh on every view.
const PluginBlockRandomQuote manifestkit.PluginBlockID = "random_quote"

// FileSourceQuote - a quote's text in the files manager's "used in".
const FileSourceQuote = "quote"

// Module - the module's registry entry.
func Module() manifestkit.Module {
	return manifestkit.Module{
		Name: Name, Label: i18n.N("Quotes"), Description: i18n.N("Quotes and their authors: /quotes pages and the random quote block."),
		Category:     manifestkit.CategoryContent,
		PluginBlocks: []manifestkit.ModulePluginBlock{{ID: PluginBlockRandomQuote, Label: i18n.N("Random quote")}},
		Permissions:  []manifestkit.ModulePermission{{Key: PermManage, Label: i18n.N("Manage quotes")}},
		CSS:          []string{"quotes"},
		JS:           []string{"quotes"},
		// The random quote block loads it on every page, the maintenance
		// page too.
		MaintenancePaths: []string{"/api/quote/random"},
		RobotsDisallow:   []string{"/quotes/create", "/quotes/edit/", "/quotes/delete/", "/quotes/authors/edit/"},
		Paths:            []string{"/quotes"},
		Tables:           []string{"quotes", "quotes_authors"},
		Data:             []manifestkit.ModuleData{{Label: i18n.N("Quotes"), Table: "quotes"}, {Label: i18n.N("Quote authors"), Table: "quotes_authors"}},
		FileSources: []manifestkit.ModuleFileSource{{Source: FileSourceQuote, Table: "quotes", IDColumn: "id", Columns: []string{"text"},
			Label: i18n.N("Quote \"@title\""), TitleColumn: "text", EditPath: "/quotes/edit/"}},
	}
}
