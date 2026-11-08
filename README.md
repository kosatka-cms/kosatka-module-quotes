# Quotes

Quotes and their authors for [Kosatka](https://github.com/kosatka-cms/kosatka): the `/quotes` pages and the random quote block.

The *Quotes* module (`quotes`) keeps a collection of quotes. Each quote
has a text, an optional author and an optional source. Visitors browse
them on the `/quotes` pages, and a *Random quote* block can show a
different quote on every page view. An author can be linked to a content
item, for example a biography, and that item's page then links to the
author's quotes.

## Screenshots

On a site with the engine's demo blog and the built-in theme; a picture
opens the whole page.

<table>
  <tr>
    <td width="50%"><a href="docs/screenshots/quotes.webp"><img src="docs/screenshots/quotes-preview.webp" alt="The quotes page"></a><br>The quotes, newest first (<code>/quotes</code>)</td>
    <td width="50%"><a href="docs/screenshots/author.webp"><img src="docs/screenshots/author-preview.webp" alt="An author's quotes"></a><br>An author's quotes (<code>/quotes/author/&lt;name&gt;</code>)</td>
  </tr>
  <tr>
    <td><img src="docs/screenshots/block.webp" alt="The random quote block" width="338"><br>The random quote block: a new quote on every page view</td>
    <td><a href="docs/screenshots/phone-quotes.webp"><img src="docs/screenshots/phone-quotes-preview.webp" alt="The quotes on a phone" width="48%"></a><br>On a phone</td>
  </tr>
</table>

## Install

A site with modules from outside the engine builds its own binary
(`kosatka.Main`); `app module add` adds this one to it:

```bash
app module add github.com/kosatka-cms/kosatka-module-quotes
go build -o app .
```

Then enable *Quotes* on the site's Modules page: it installs its
tables and starter rows. See [Using modules](https://kosatka-cms.github.io/kosatka/modules/using-modules/).

A new site built with it gets it with `app install -modules quotes,…`
(or `-modules all`).

## What it adds

**Pages.**

| Address | What it shows | Who sees it |
|---|---|---|
| `/quotes` | all quotes, newest first, 50 per page | everyone |
| `/quotes/byauthor` | *By author*: the authors in alphabetical order, each with the number of quotes | everyone |
| `/quotes/author/<name>` | *Quotes: <name>*: one author's quotes, 50 per page | everyone |
| `/quotes/create` | *Add quote* | users with *Manage quotes* |
| `/quotes/edit/<id>` | *Edit quote* | users with *Manage quotes* |
| `/quotes/delete/<id>` | *Delete quote*, a confirmation | users with *Manage quotes* |
| `/quotes/authors/edit/<id>` | *Edit author* | users with *Manage quotes* |

The author's page uses the author's name as written, so renaming an
author changes the page's address (the old one answers 404). An author
with no quotes isn't listed on *By author*. The `/quotes` page doesn't
link to *By author* by itself: it is reachable from the breadcrumbs of
an author's page, or add a menu link to `/quotes/byauthor`
([Menus](https://kosatka-cms.github.io/kosatka/guides/menus/)).

The module has no admin pages: quotes are managed on the pages above.
Users who may manage quotes see *Add quote*, *Edit* and *Delete* links
right on them (and *Edit author* on an author's page), and *Quote* in
their *Create content* menu.

**Block.** *Random quote* (**Structure → Blocks**, *Add system block*).
It has no settings of its own; see [below](#the-random-quote-block).

**Permission.** *Manage quotes*, in the general group of **People →
Permissions** ([Users and permissions](https://kosatka-cms.github.io/kosatka/guides/users-and-permissions/#permissions)):
adding, editing and deleting quotes and editing their authors. Viewing
the quotes needs no permission.

**On the Modules page**, *What the module includes* lists the module's
data: *Quotes* and *Quote authors* with their row counts.

## Add a quote

1. Open `/quotes` and click **Add quote** (or choose *Quote* in the
   *Create content* menu).
2. Fill in the form:

   | Field | What to enter |
   |---|---|
   | *Quote text* | required. A rich text editor: bold, italic, lists, links and images from your files |
   | *Author* | start typing: after two letters, existing authors whose name contains them are suggested. Pick one, or type a new name and press Enter to create that author. Leave it empty for an anonymous quote |
   | *Source (citation)* | optional, plain text: the book, issue, page… |

3. Click **Save**. You return to `/quotes`, where the new quote is at
   the top.

A new name that matches an existing author exactly is attached to that
author rather than creating a second one.

To change a quote, click *Edit* under it on any quotes page. To remove
it, click *Delete* and confirm with **Yes, delete**. Deleting a quote
doesn't delete its author.

## Edit an author

Authors are created from the quote form; they have no page of their own
to create them. To rename one or link it to content:

1. Open the author's page (click the name under any of their quotes)
   and click **Edit author**.
2. Change the *Author name* (required, unique), and in *Linked content
   ID* enter the ID of a content item about the author. The ID is the
   number in the item's edit address, `/node/<id>/edit`.
   Leave it empty to remove the link.
3. Click **Save**. You go to the author's page under its new name.

While an author is linked, the linked item's page shows a *Quotes* link
with the number of quotes, for example *Quotes (12)*, leading to the
author's page. It works for content of any type. If the linked item is
deleted, the link is cleared.

An author without quotes (all deleted) disappears from *By author*, and
its edit form offers **Delete author**. An author with quotes can't be
deleted: delete or reassign the quotes first.

## The random quote block

1. On **Structure → Blocks** (`/admin/config/blocks`), under *Add system
   block*, choose *Random quote* and a region, and press *Add*.
2. On the block's page you can give it a title and choose where it shows
   ([Visibility](https://kosatka-cms.github.io/kosatka/guides/blocks/#visibility)).

The page itself contains only a placeholder. When the visitor scrolls
close to it, the browser asks the site for one random quote and shows
its text and author (with a link to the author's page); the source isn't
shown. So every page view gets a new quote, even if the page is cached.
If the site has no quotes, the block hides itself. The block keeps
working while the site is in maintenance mode.

The block's script is the `randomQuote` behavior (see
[Scripts](https://kosatka-cms.github.io/kosatka/themes/scripts/)).

## Other things the module does

- **robots.txt.** The add, edit and delete pages and the author edit
  pages are disallowed for crawlers
  ([robots.txt](https://kosatka-cms.github.io/kosatka/guides/urls-and-seo/#robotstxt)).
- **Files.** Before an image or document is deleted or renamed in the
  file manager, the check of where it is used also reads quote texts;
  a match is listed as *Quote "…"* with a link to its edit page
  ([Files and images](https://kosatka-cms.github.io/kosatka/guides/files-and-images/)).
- **Demo content.** A site installed with `-demo` and this module gets
  a few sample quotes, the *Random quote* block at the end of the
  *Sidebar* and a *Quotes* link in the footer menu.

## Uninstall

Uninstalling the module ([Using modules](https://kosatka-cms.github.io/kosatka/modules/using-modules/#uninstall-a-module))
drops its two tables, `quotes` and `quotes_authors`, with every quote and
author, removes the *Random quote* block's placements and takes *Manage
quotes* away from the roles. The module has no content types, so there
is no content to delete first.

It also removes the menu links to `/quotes` and the addresses under it
(a link's sub-items move up to its place). Images inserted into quote
texts stay in the uploaders' files.

## Development

Its strings are translated in `translations/<lang>.json`.

The module is built on the engine's module SDK only
(`github.com/kosatka-cms/kosatka/modulekit`, `pkg/*`).

```bash
go test ./...
# with MySQL and PostgreSQL (the engine's docker-compose.test.yml):
KOSATKA_TEST_MYSQL_DSN="root:test@tcp(127.0.0.1:33306)/?parseTime=true" \
KOSATKA_TEST_POSTGRES_DSN="postgres://postgres:test@127.0.0.1:35432/postgres?sslmode=disable" \
  go test -tags dbintegration ./...
```

A schema change: edit `ent/schema`, `go generate ./ent`, then
`go run ./cmd/migrations -name <name>` (both test databases running) -
it writes the next migration for MySQL and PostgreSQL into
`ent/migrations`; `go run ./cmd/migrations -check` fails while the files
are behind the schema. See [Write a module](https://kosatka-cms.github.io/kosatka/modules/write-a-module/).

The README's screenshots: on a site installed with the engine's demo
blog and this module (`app install -demo -modules quotes`),
`python3 scripts/screenshots.py http://127.0.0.1:8080` (the site's
address) takes them again into `docs/screenshots/` - Python 3 and
Google Chrome are all it needs.

## License

MIT
