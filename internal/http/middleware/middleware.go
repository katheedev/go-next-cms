package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"

	"go-next-cms/internal/models"
	"go-next-cms/internal/repo"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/session"
)

func StructuredLogger() fiber.Handler {
	return logger.New(logger.Config{Format: "{\"time\":\"${time}\",\"status\":${status},\"latency\":${latency},\"method\":\"${method}\",\"path\":\"${path}\"}\n"})
}

func SessionStore() *session.Store {
	return session.New(session.Config{CookieHTTPOnly: true, CookieSameSite: "lax"})
}

func AttachUser(store *session.Store, repo *repo.Repository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		s, _ := store.Get(c)
		idRaw := s.Get("uid")
		if idRaw == nil {
			return c.Next()
		}
		id, _ := strconv.ParseInt(idRaw.(string), 10, 64)
		u, err := repo.UserByID(c.Context(), id)
		if err == nil {
			c.Locals("user", u)
		}
		return c.Next()
	}
}

func RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		u := c.Locals("user")
		if u == nil {
			return c.Redirect("/account/login")
		}
		return c.Next()
	}
}

func RequireAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		u, ok := c.Locals("user").(*models.User)
		if !ok || u.Role != models.RoleAdmin {
			return c.SendStatus(fiber.StatusForbidden)
		}
		return c.Next()
	}
}

func CSRFMiddleware(store *session.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		s, _ := store.Get(c)
		tok := s.Get("csrf")
		if tok == nil {
			b := make([]byte, 16)
			_, _ = rand.Read(b)
			v := hex.EncodeToString(b)
			s.Set("csrf", v)
			s.Save()
			c.Locals("csrf", v)
		} else {
			c.Locals("csrf", tok)
		}
		if c.Method() == fiber.MethodPost {
			if c.FormValue("csrf") != c.Locals("csrf") {
				return c.Status(fiber.StatusBadRequest).SendString("CSRF mismatch")
			}
		}
		return c.Next()
	}
}
