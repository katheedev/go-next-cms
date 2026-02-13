package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-next-cms/internal/config"
	"go-next-cms/internal/db"
	"go-next-cms/internal/http/handlers"
	"go-next-cms/internal/http/middleware"
	"go-next-cms/internal/i18n"
	"go-next-cms/internal/repo"
	"go-next-cms/internal/service"
	"go-next-cms/internal/storage"

	"github.com/gofiber/fiber/v2"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	r := repo.New(pool)
	bundle, err := i18n.Load("i18n")
	if err != nil {
		log.Fatal(err)
	}
	svc := service.New(r)
	sessions := middleware.SessionStore()
	h := handlers.New(r, svc, bundle, sessions, storage.LocalUploader{Dir: cfg.UploadDir, BasePath: "/static/uploads", MaxSize: cfg.MaxUploadBytes})

	app := fiber.New()
	app.Use(middleware.StructuredLogger())
	app.Use(middleware.AttachUser(sessions, r))
	app.Static("/static", "./static")

	app.Get("/", func(c *fiber.Ctx) error { return c.Redirect("/LK") })
	app.Get("/:countryCode", h.Home)
	app.Get("/:countryCode/deals", h.Deals)
	app.Get("/:countryCode/city/:city", func(c *fiber.Ctx) error { c.Context().QueryArgs().Set("city", c.Params("city")); return h.Deals(c) })
	app.Get("/:countryCode/category/:categorySlug", func(c *fiber.Ctx) error {
		c.Context().QueryArgs().Set("category", c.Params("categorySlug"))
		return h.Deals(c)
	})
	app.Get("/:countryCode/deal/:dealSlug", h.DealDetail)

	account := app.Group("/account", middleware.CSRFMiddleware(sessions))
	account.Get("/register", h.RegisterForm)
	account.Post("/register", h.Register)
	account.Get("/login", h.LoginForm)
	account.Post("/login", h.Login)
	account.Get("/logout", h.Logout)
	account.Get("/submissions", middleware.RequireAuth(), h.SubmissionList)
	account.Get("/submissions/new", middleware.RequireAuth(), h.NewSubmissionForm)
	account.Post("/submissions/new", middleware.RequireAuth(), h.CreateSubmission)
	account.Get("/submissions/:id/edit", middleware.RequireAuth(), h.EditSubmissionForm)
	account.Post("/submissions/:id/edit", middleware.RequireAuth(), h.UpdateSubmission)

	admin := app.Group("/admin", middleware.RequireAuth(), middleware.RequireAdmin(), middleware.CSRFMiddleware(sessions))
	admin.Get("/", h.AdminDashboard)
	admin.Get("/moderation", h.AdminModeration)
	admin.Post("/moderation/:id", h.AdminModerate)
	admin.Get("/users", h.AdminUsers)
	admin.Post("/users/:id/role", h.AdminUserRole)
	admin.Get("/config", h.AdminConfig)
	admin.Post("/config", h.SaveConfig)
	admin.Get("/master", h.AdminMaster)
	admin.Post("/master/country", h.CreateCountry)
	admin.Post("/master/city", h.CreateCity)
	admin.Post("/master/category", h.CreateCategory)
	admin.Post("/master/merchant", h.CreateMerchant)
	admin.Post("/master/dealtype", h.CreateDealType)
	admin.Get("/deals/new", h.AdminNewDealForm)
	admin.Post("/deals/new", h.AdminCreateDeal)

	go func() {
		if err := app.Listen(cfg.Addr()); err != nil {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	_ = app.ShutdownWithTimeout(5 * time.Second)
}
