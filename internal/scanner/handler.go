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
	pageURL := c.FormValue("url")
	if pageURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "url is required",
		})
	}

	userId, ok := c.Locals("user_id").(int32)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user_id type",
		})
	}

	results, err := h.service.Scan(pageURL, userId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error scanning: " + err.Error())
	}

	var rows []scannerui.ResultRow
	for _, r := range results {
		rows = append(rows, scannerui.ResultRow{
			Link:   r.LinkUrl,
			Status: r.Status,
		})
	}

	return ui.RenderComponent(c, scannerui.ResultsTable(rows, pageURL))
}
