package scanner

import (
	scannerui "go-deadlink-scanner/internal/templates/scanner"
	"go-deadlink-scanner/internal/ui"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) ScanPage(c *fiber.Ctx) error {
	return ui.RenderComponent(c, scannerui.ScanPage("example.com", nil))
}

func (h *Handler) StartScan(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"userId": c.Locals("user_id"),
	})
}
