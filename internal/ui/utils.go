package ui

import (
	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

func IsHX(c *fiber.Ctx) bool { return c.Get("HX-Request") == "true" }

func RenderComponent(c *fiber.Ctx, comp templ.Component) error {
	h := templ.Handler(comp)
	return adaptor.HTTPHandler(h)(c)
}
