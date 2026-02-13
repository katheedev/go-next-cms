package handlers

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"go-next-cms/internal/auth"
	"go-next-cms/internal/i18n"
	"go-next-cms/internal/models"
	"go-next-cms/internal/repo"
	"go-next-cms/internal/service"
	"go-next-cms/internal/storage"
	"go-next-cms/internal/views"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type Handler struct {
	Repo     *repo.Repository
	Service  *service.Service
	I18n     *i18n.Bundle
	Sessions *session.Store
	Uploader storage.LocalUploader
}

func New(r *repo.Repository, s *service.Service, b *i18n.Bundle, sessions *session.Store, up storage.LocalUploader) *Handler {
	return &Handler{Repo: r, Service: s, I18n: b, Sessions: sessions, Uploader: up}
}

func (h *Handler) lang(c *fiber.Ctx) string {
	lang := c.Query("lang", c.Cookies("lang", "en"))
	lang = service.AllowedLang(lang)
	c.Cookie(&fiber.Cookie{Name: "lang", Value: lang, Path: "/", SameSite: "Lax"})
	return lang
}

func (h *Handler) render(c *fiber.Ctx, title, country string, body template.HTML) error {
	lang := h.lang(c)
	var u *models.User
	if v := c.Locals("user"); v != nil {
		u = v.(*models.User)
	}
	html, err := views.Render(views.Layout(title, views.NavData{CountryCode: country, Lang: lang, User: u, T: func(k string) string { return h.I18n.T(lang, k) }}, body))
	if err != nil {
		return err
	}
	return c.Type("html").SendString(html)
}

func (h *Handler) Home(c *fiber.Ctx) error {
	cc := strings.ToUpper(c.Params("countryCode"))
	featured, _ := h.Repo.FeaturedDeals(c.Context(), cc, 6)
	ending, _ := h.Repo.EndingSoonDeals(c.Context(), cc, 6)
	cats, _ := h.Repo.Categories(c.Context())
	lang := h.lang(c)
	return h.render(c, "Home", cc, views.HomeContent(featured, ending, cats, cc, func(k string) string { return h.I18n.T(lang, k) }))
}

func (h *Handler) Deals(c *fiber.Ctx) error {
	cc := strings.ToUpper(c.Params("countryCode"))
	page, _ := strconv.Atoi(c.Query("page", "1"))
	dt, _ := strconv.ParseInt(c.Query("deal_type", "0"), 10, 64)
	mID, _ := strconv.ParseInt(c.Query("merchant", "0"), 10, 64)
	deals, _ := h.Repo.ListDeals(c.Context(), models.DealFilter{CountryCode: cc, CitySlug: c.Query("city"), CategorySlug: c.Query("category"), DealTypeID: dt, MerchantID: mID, Search: c.Query("q"), EndingSoon: c.Query("ending_soon") == "1", Page: page, PageSize: 10})
	partial := views.DealCards(deals, cc)
	if c.Get("HX-Request") == "true" {
		return c.SendString(string(partial))
	}
	body := template.HTML(fmt.Sprintf("<h1>Deals</h1><form hx-get='/%s/deals' hx-target='#results'><input name='q' placeholder='search'><button>Search</button></form><div id='results'>%s</div>", cc, partial))
	return h.render(c, "Deals", cc, body)
}

func (h *Handler) DealDetail(c *fiber.Ctx) error {
	cc := strings.ToUpper(c.Params("countryCode"))
	d, err := h.Repo.DealBySlug(c.Context(), cc, c.Params("dealSlug"))
	if err != nil {
		return c.SendStatus(404)
	}
	tr, _ := h.Repo.DealTranslation(c.Context(), d.ID, h.lang(c))
	return h.render(c, d.Title, cc, views.DealDetail(d, tr))
}

func (h *Handler) RegisterForm(c *fiber.Ctx) error {
	body := template.HTML("<h1>Register</h1><form method='post'><input name='name' placeholder='name'><input name='email' type='email'><input name='password' type='password'><input type='hidden' name='csrf' value='" + c.Locals("csrf").(string) + "'><button>Create account</button></form>")
	return h.render(c, "Register", "LK", body)
}

func (h *Handler) Register(c *fiber.Ctx) error {
	hash, err := auth.HashPassword(c.FormValue("password"))
	if err != nil {
		return err
	}
	u := &models.User{Name: c.FormValue("name"), Email: strings.ToLower(c.FormValue("email")), PasswordHash: hash, Role: models.RoleSubmitter}
	if err := h.Repo.CreateUser(c.Context(), u); err != nil {
		return c.Status(400).SendString(err.Error())
	}
	return c.Redirect("/account/login")
}

func (h *Handler) LoginForm(c *fiber.Ctx) error {
	body := template.HTML("<h1>Login</h1><form method='post'><input name='email' type='email'><input name='password' type='password'><input type='hidden' name='csrf' value='" + c.Locals("csrf").(string) + "'><button>Login</button></form>")
	return h.render(c, "Login", "LK", body)
}

func (h *Handler) Login(c *fiber.Ctx) error {
	u, err := h.Repo.UserByEmail(c.Context(), c.FormValue("email"))
	if err != nil || auth.CheckPassword(u.PasswordHash, c.FormValue("password")) != nil {
		return c.Status(400).SendString("invalid credentials")
	}
	sess, _ := h.Sessions.Get(c)
	sess.Set("uid", fmt.Sprintf("%d", u.ID))
	sess.Save()
	return c.Redirect("/LK")
}

func (h *Handler) Logout(c *fiber.Ctx) error {
	s, _ := h.Sessions.Get(c)
	s.Destroy()
	return c.Redirect("/LK")
}

func (h *Handler) SubmissionList(c *fiber.Ctx) error {
	u := c.Locals("user").(*models.User)
	items, _ := h.Repo.SubmissionDeals(c.Context(), u.ID)
	body := template.HTML("<h1>My submissions</h1><a href='/account/submissions/new'>New</a>" + string(views.DealCards(items, "LK")))
	return h.render(c, "My submissions", "LK", body)
}

func (h *Handler) NewSubmissionForm(c *fiber.Ctx) error {
	countries, _ := h.Repo.Countries(c.Context())
	cities, _ := h.Repo.CitiesByCountry(c.Context(), countries[0].ID)
	cats, _ := h.Repo.Categories(c.Context())
	dts, _ := h.Repo.DealTypes(c.Context())
	var cityOpts, catOpts, dtOpts strings.Builder
	for _, x := range cities {
		fmt.Fprintf(&cityOpts, "<option value='%d'>%s</option>", x.ID, x.Name)
	}
	for _, x := range cats {
		fmt.Fprintf(&catOpts, "<option value='%d'>%s</option>", x.ID, x.Name)
	}
	for _, x := range dts {
		fmt.Fprintf(&dtOpts, "<option value='%d'>%s</option>", x.ID, x.Name)
	}
	body := template.HTML(fmt.Sprintf("<h1>Submit Deal</h1><form method='post' enctype='multipart/form-data'><input name='title'><textarea name='description'></textarea><select name='city_id'>%s</select><select name='category_id'>%s</select><select name='deal_type_id'>%s</select><input name='start_at' type='date'><input name='end_at' type='date'><input type='file' name='image'><input name='title_si' placeholder='Sinhala title'><textarea name='description_si'></textarea><input name='title_ta' placeholder='Tamil title'><textarea name='description_ta'></textarea><input type='hidden' name='csrf' value='%s'><button>Submit</button></form>", cityOpts.String(), catOpts.String(), dtOpts.String(), c.Locals("csrf")))
	return h.render(c, "New submission", "LK", body)
}

func (h *Handler) CreateSubmission(c *fiber.Ctx) error {
	u := c.Locals("user").(*models.User)
	country, _ := h.Repo.CountryByCode(c.Context(), "LK")
	slug := strings.ToLower(strings.ReplaceAll(c.FormValue("title"), " ", "-"))
	if exists, _ := h.Repo.SlugExists(c.Context(), country.ID, slug); exists {
		slug += "-" + strconv.FormatInt(u.ID, 10)
	}
	cityID, _ := strconv.ParseInt(c.FormValue("city_id"), 10, 64)
	catID, _ := strconv.ParseInt(c.FormValue("category_id"), 10, 64)
	dtID, _ := strconv.ParseInt(c.FormValue("deal_type_id"), 10, 64)
	start, _ := service.ParseDate(c.FormValue("start_at"))
	end, _ := service.ParseDate(c.FormValue("end_at"))
	img, err := h.Uploader.Save(c, "image")
	if err != nil {
		return c.Status(400).SendString(err.Error())
	}
	d := &models.Deal{Title: c.FormValue("title"), Slug: slug, Description: c.FormValue("description"), CountryID: country.ID, CityID: cityID, CategoryID: catID, DealTypeID: dtID, StartAt: start, EndAt: end, ImageURL: img, Status: models.DealPending, CreatedByUserID: u.ID}
	trs := []models.DealTranslation{{Lang: "en", Title: d.Title, Description: d.Description}, {Lang: "si", Title: c.FormValue("title_si"), Description: c.FormValue("description_si")}, {Lang: "ta", Title: c.FormValue("title_ta"), Description: c.FormValue("description_ta")}}
	if err := h.Repo.CreateDeal(c.Context(), d, trs); err != nil {
		return err
	}
	return c.Redirect("/account/submissions")
}

func (h *Handler) EditSubmissionForm(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	d, err := h.Repo.DealByID(c.Context(), id)
	if err != nil {
		return c.SendStatus(404)
	}
	u := c.Locals("user").(*models.User)
	if !service.DealEditable(d, u.ID, u.Role == models.RoleAdmin) {
		return c.SendStatus(403)
	}
	body := template.HTML(fmt.Sprintf("<h1>Edit</h1><form method='post' enctype='multipart/form-data'><input name='title' value='%s'><textarea name='description'>%s</textarea><input name='start_at' type='date' value='%s'><input name='end_at' type='date' value='%s'><input type='file' name='image'><input type='hidden' name='csrf' value='%s'><button>Save</button></form>", d.Title, d.Description, d.StartAt.Format("2006-01-02"), d.EndAt.Format("2006-01-02"), c.Locals("csrf")))
	return h.render(c, "Edit submission", "LK", body)
}

func (h *Handler) UpdateSubmission(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	d, err := h.Repo.DealByID(c.Context(), id)
	if err != nil {
		return c.SendStatus(404)
	}
	u := c.Locals("user").(*models.User)
	if !service.DealEditable(d, u.ID, u.Role == models.RoleAdmin) {
		return c.SendStatus(403)
	}
	if img, err := h.Uploader.Save(c, "image"); err == nil && img != "" {
		d.ImageURL = img
	}
	d.Title = c.FormValue("title")
	d.Description = c.FormValue("description")
	d.StartAt, _ = service.ParseDate(c.FormValue("start_at"))
	d.EndAt, _ = service.ParseDate(c.FormValue("end_at"))
	if err := h.Repo.UpdateSubmission(c.Context(), d); err != nil {
		return err
	}
	return c.Redirect("/account/submissions")
}

func (h *Handler) AdminDashboard(c *fiber.Ctx) error {
	p, pub, _ := h.Repo.DashboardCounts(c.Context())
	body := template.HTML(fmt.Sprintf("<h1>Admin</h1><p>Pending: %d, Published: %d</p><ul><li><a href='/admin/moderation'>Moderation</a></li><li><a href='/admin/deals/new'>Create Deal</a></li><li><a href='/admin/users'>Users</a></li><li><a href='/admin/config'>Config</a></li><li><a href='/admin/master'>Master Data</a></li></ul>", p, pub))
	return h.render(c, "Admin", "LK", body)
}

func (h *Handler) AdminModeration(c *fiber.Ctx) error {
	items, _ := h.Repo.PendingDeals(c.Context())
	var b strings.Builder
	b.WriteString("<h1>Moderation Queue</h1>")
	for _, d := range items {
		fmt.Fprintf(&b, "<div><b>%s</b><form method='post' action='/admin/moderation/%d'><input type='hidden' name='csrf' value='%s'><button name='action' value='approve'>Approve</button><input name='reason' placeholder='reason'><button name='action' value='reject'>Reject</button></form></div>", d.Title, d.ID, c.Locals("csrf"))
	}
	return h.render(c, "Moderation", "LK", template.HTML(b.String()))
}

func (h *Handler) AdminModerate(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	a := c.FormValue("action")
	if a == "approve" {
		_ = h.Repo.ModerateDeal(c.Context(), id, models.DealApproved, nil)
		_ = h.Repo.UpdateDealStatusDirect(c.Context(), id, models.DealPublished)
	}
	if a == "reject" {
		r := c.FormValue("reason")
		_ = h.Repo.ModerateDeal(c.Context(), id, models.DealRejected, &r)
	}
	return c.Redirect("/admin/moderation")
}

func (h *Handler) AdminUsers(c *fiber.Ctx) error {
	users, _ := h.Repo.Users(c.Context())
	var b strings.Builder
	b.WriteString("<h1>Users</h1>")
	for _, u := range users {
		fmt.Fprintf(&b, "<div>%s (%s) <form method='post' action='/admin/users/%d/role'><select name='role'><option value='submitter'>submitter</option><option value='admin'>admin</option></select><input type='hidden' name='csrf' value='%s'><button>Update</button></form></div>", u.Email, u.Role, u.ID, c.Locals("csrf"))
	}
	return h.render(c, "Users", "LK", template.HTML(b.String()))
}

func (h *Handler) AdminUserRole(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	_ = h.Repo.UpdateUserRole(c.Context(), id, models.UserRole(c.FormValue("role")))
	return c.Redirect("/admin/users")
}

func (h *Handler) AdminConfig(c *fiber.Ctx) error {
	cfg, _ := h.Repo.AdminConfigs(c.Context())
	var b strings.Builder
	b.WriteString("<h1>Config</h1><form method='post'><input name='key'><textarea name='value'>{}</textarea><input type='hidden' name='csrf' value='" + c.Locals("csrf").(string) + "'><button>Save</button></form>")
	for _, x := range cfg {
		fmt.Fprintf(&b, "<pre>%s = %s</pre>", x.Key, string(x.Value))
	}
	return h.render(c, "Config", "LK", template.HTML(b.String()))
}

func (h *Handler) SaveConfig(c *fiber.Ctx) error {
	_ = h.Repo.SetAdminConfig(c.Context(), c.FormValue("key"), c.FormValue("value"))
	return c.Redirect("/admin/config")
}

func (h *Handler) AdminMaster(c *fiber.Ctx) error {
	countries, _ := h.Repo.Countries(c.Context())
	cats, _ := h.Repo.Categories(c.Context())
	mers, _ := h.Repo.Merchants(c.Context())
	dts, _ := h.Repo.DealTypes(c.Context())
	body := "<h1>Master data</h1>"
	body += "<h2>Create Country</h2><form method='post' action='/admin/master/country'><input name='code'><input name='name'><input name='default_language' value='en'><input type='hidden' name='csrf' value='" + c.Locals("csrf").(string) + "'><button>Save</button></form>"
	body += "<h2>Create City</h2><form method='post' action='/admin/master/city'><input name='country_id'><input name='name'><input name='slug'><input type='hidden' name='csrf' value='" + c.Locals("csrf").(string) + "'><button>Save</button></form>"
	body += "<h2>Create Category</h2><form method='post' action='/admin/master/category'><input name='name'><input name='slug'><input type='hidden' name='csrf' value='" + c.Locals("csrf").(string) + "'><button>Save</button></form>"
	body += "<h2>Create Merchant</h2><form method='post' action='/admin/master/merchant'><input name='name'><input name='slug'><input name='contact'><input type='hidden' name='csrf' value='" + c.Locals("csrf").(string) + "'><button>Save</button></form>"
	body += "<h2>Create Deal Type</h2><form method='post' action='/admin/master/dealtype'><input name='code'><input name='name'><input type='hidden' name='csrf' value='" + c.Locals("csrf").(string) + "'><button>Save</button></form>"
	body += fmt.Sprintf("<pre>countries:%d categories:%d merchants:%d dealtypes:%d</pre>", len(countries), len(cats), len(mers), len(dts))
	return h.render(c, "Master", "LK", template.HTML(body))
}

func (h *Handler) CreateCountry(c *fiber.Ctx) error {
	_ = h.Repo.CreateCountry(c.Context(), &models.Country{Code: strings.ToUpper(c.FormValue("code")), Name: c.FormValue("name"), DefaultLanguage: c.FormValue("default_language")})
	return c.Redirect("/admin/master")
}
func (h *Handler) CreateCity(c *fiber.Ctx) error {
	cid, _ := strconv.ParseInt(c.FormValue("country_id"), 10, 64)
	_ = h.Repo.CreateCity(c.Context(), &models.City{CountryID: cid, Name: c.FormValue("name"), Slug: c.FormValue("slug")})
	return c.Redirect("/admin/master")
}
func (h *Handler) CreateCategory(c *fiber.Ctx) error {
	_ = h.Repo.CreateCategory(c.Context(), &models.Category{Name: c.FormValue("name"), Slug: c.FormValue("slug")})
	return c.Redirect("/admin/master")
}
func (h *Handler) CreateMerchant(c *fiber.Ctx) error {
	_ = h.Repo.CreateMerchant(c.Context(), &models.Merchant{Name: c.FormValue("name"), Slug: c.FormValue("slug"), Contact: c.FormValue("contact")})
	return c.Redirect("/admin/master")
}
func (h *Handler) CreateDealType(c *fiber.Ctx) error {
	_ = h.Repo.CreateDealType(c.Context(), &models.DealType{Code: c.FormValue("code"), Name: c.FormValue("name")})
	return c.Redirect("/admin/master")
}

func (h *Handler) AdminNewDealForm(c *fiber.Ctx) error {
	return h.NewSubmissionForm(c)
}
func (h *Handler) AdminCreateDeal(c *fiber.Ctx) error {
	if err := h.CreateSubmission(c); err != nil {
		return err
	}
	return nil
}
