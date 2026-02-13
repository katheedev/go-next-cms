package views

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"go-next-cms/internal/models"

	"github.com/a-h/templ"
)

type NavData struct {
	CountryCode string
	Lang        string
	User        *models.User
	T           func(string) string
}

func Render(c templ.Component) (string, error) {
	var b bytes.Buffer
	if err := c.Render(nil, &b); err != nil {
		return "", err
	}
	return b.String(), nil
}

func Layout(title string, nav NavData, content template.HTML) templ.Component {
	return templ.ComponentFunc(func(ctx templ.Context, w templ.Writer) error {
		_, err := fmt.Fprintf(w, `<!doctype html><html><head><meta charset='utf-8'><meta name='viewport' content='width=device-width, initial-scale=1'><script src='https://unpkg.com/htmx.org@1.9.12'></script><link rel='stylesheet' href='/static/css/app.css'><title>%s</title></head><body><header><nav><a href='/%s'>%s</a> | <a href='/%s/deals'>%s</a>`, title, nav.CountryCode, nav.T("home"), nav.CountryCode, nav.T("deals"))
		if nav.User == nil {
			fmt.Fprintf(w, ` | <a href='/account/login'>%s</a> | <a href='/account/register'>%s</a>`, nav.T("login"), nav.T("register"))
		} else {
			fmt.Fprintf(w, ` | <a href='/account/submissions'>%s</a> | <a href='/account/logout'>%s</a>`, nav.T("my_submissions"), nav.T("logout"))
			if nav.User.Role == models.RoleAdmin {
				fmt.Fprintf(w, ` | <a href='/admin'>%s</a>`, nav.T("admin"))
			}
		}
		fmt.Fprintf(w, `</nav></header><main>%s</main></body></html>`, content)
		return err
	})
}

func DealCards(deals []models.Deal, countryCode string) template.HTML {
	var b bytes.Buffer
	for _, d := range deals {
		fmt.Fprintf(&b, "<article class='card'><h3><a href='/%s/deal/%s'>%s</a></h3><p>%s</p><small>%s - %s</small></article>", countryCode, d.Slug, template.HTMLEscapeString(d.Title), template.HTMLEscapeString(d.Description), d.CityName, d.EndAt.Format("2006-01-02"))
	}
	if len(deals) == 0 {
		b.WriteString("<p>No deals found.</p>")
	}
	return template.HTML(b.String())
}

func HomeContent(featured, ending []models.Deal, categories []models.Category, cc string, t func(string) string) template.HTML {
	var b bytes.Buffer
	fmt.Fprintf(&b, "<h1>%s</h1><section><h2>%s</h2>%s</section>", cc, t("featured_deals"), DealCards(featured, cc))
	fmt.Fprintf(&b, "<section><h2>%s</h2><ul>", t("categories"))
	for _, c := range categories {
		fmt.Fprintf(&b, "<li><a href='/%s/category/%s'>%s</a></li>", cc, c.Slug, c.Name)
	}
	b.WriteString("</ul></section>")
	fmt.Fprintf(&b, "<section><h2>%s</h2>%s</section>", t("ending_soon"), DealCards(ending, cc))
	return template.HTML(b.String())
}

func DealDetail(d *models.Deal, tr *models.DealTranslation) template.HTML {
	title := d.Title
	desc := d.Description
	if tr != nil {
		title = tr.Title
		desc = tr.Description
	}
	return template.HTML(fmt.Sprintf("<article><h1>%s</h1><p>%s</p><p>Category: %s | City: %s | Type: %s</p><p>Valid: %s to %s</p></article>", template.HTMLEscapeString(title), template.HTMLEscapeString(desc), d.CategoryName, d.CityName, d.DealTypeName, d.StartAt.Format(time.DateOnly), d.EndAt.Format(time.DateOnly)))
}
