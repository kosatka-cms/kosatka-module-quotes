package quotes

import (
	"context"
	"net/url"
	"strconv"

	"github.com/kosatka-cms/kosatka-module-quotes/manifest"
	"github.com/kosatka-cms/kosatka/modulekit"
	"github.com/kosatka-cms/kosatka/pkg/i18n"
)

// NodePage - the "Quotes (N)" link of a node a quote author is linked to
// (quotes_authors.node_id), whatever its content type (AddNodePage).
func NodePage(svc QuoteService) func(ctx context.Context, node modulekit.Node) modulekit.NodePageParts {
	return func(ctx context.Context, node modulekit.Node) modulekit.NodePageParts {
		author, err := svc.AuthorByNodeID(ctx, node.ID)
		if err != nil || author == nil {
			return modulekit.NodePageParts{}
		}
		label := i18n.T(ctx, "Quotes")
		if count, err := svc.TotalCount(ctx, QuotesFilter{AuthorID: &author.ID}); err == nil && count > 0 {
			label += " (" + strconv.Itoa(count) + ")"
		}
		return modulekit.NodePageParts{Links: []modulekit.NodeLink{{URL: "/quotes/author/" + url.PathEscape(author.Name), Label: label}}}
	}
}

// CreateLinks - "Quote" in the "Create content" menu of those who manage
// quotes (AddCreateLinks).
func CreateLinks(ctx context.Context, user *modulekit.User, auth modulekit.AuthService) []modulekit.CreatableMaterialLink {
	if !auth.HasPermission(ctx, user, manifest.PermManage) {
		return nil
	}
	return []modulekit.CreatableMaterialLink{{URL: "/quotes/create", Label: i18n.T(ctx, "Quote")}}
}
