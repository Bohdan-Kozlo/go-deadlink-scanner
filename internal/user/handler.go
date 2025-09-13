package user

import (
	"go-deadlink-scanner/internal/ui"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	userui "go-deadlink-scanner/internal/templates/user"
)

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler { return &Handler{Service: service} }

func (h *Handler) RegisterPage(c *fiber.Ctx) error {
	return ui.RenderComponent(c, userui.RegisterPage("", nil))
}
func (h *Handler) LoginPage(c *fiber.Ctx) error {
	return ui.RenderComponent(c, userui.LoginPage("", nil))
}

func (h *Handler) Register(c *fiber.Ctx) error {
	form := User{Email: c.FormValue("email"), Pass: c.FormValue("password")}
	validate := validator.New()
	if err := validate.Struct(form); err != nil {
		errs := []string{err.Error()}
		if ui.IsHX(c) {
			return ui.RenderComponent(c, userui.RegisterForm(form.Email, errs))
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	u, token, err := h.Service.Register(c.Context(), form.Email, form.Pass)
	if err != nil {
		if ui.IsHX(c) {
			return ui.RenderComponent(c, userui.RegisterForm(form.Email, []string{"Error already registered"}))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.Service.SetSession(c, u.ID, token); err != nil {
		if ui.IsHX(c) {
			return ui.RenderComponent(c, userui.RegisterForm(form.Email, []string{"Problem with session"}))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "session set failed"})
	}

	if ui.IsHX(c) {
		return ui.RenderComponent(c, userui.RegisterSuccess(u.Email))
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"user_id": u.ID, "email": u.Email, "session_token": token})
}

func (h *Handler) Login(c *fiber.Ctx) error {
	form := User{Email: c.FormValue("email"), Pass: c.FormValue("password")}
	validate := validator.New()
	if err := validate.Struct(form); err != nil {
		errs := []string{err.Error()}
		if ui.IsHX(c) {
			return ui.RenderComponent(c, userui.LoginForm(form.Email, errs))
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	u, token, err := h.Service.Login(c.Context(), form.Email, form.Pass)
	if err != nil {
		if ui.IsHX(c) {
			return ui.RenderComponent(c, userui.LoginForm(form.Email, []string{"Invalid credentials"}))
		}
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
	}

	if err := h.Service.SetSession(c, u.ID, token); err != nil {
		if ui.IsHX(c) {
			return ui.RenderComponent(c, userui.LoginForm(form.Email, []string{"Problem with session"}))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "session set failed"})
	}

	if ui.IsHX(c) {
		return ui.RenderComponent(c, userui.LoginSuccess(u.Email))
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"user_id": u.ID, "email": u.Email, "session_token": token})
}
