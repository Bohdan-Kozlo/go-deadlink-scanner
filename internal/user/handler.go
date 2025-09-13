package user

import (
	"bytes"

	"github.com/a-h/templ"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	userui "go-deadlink-scanner/internal/templates/user"
)

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler { return &Handler{Service: service} }

func (h *Handler) RegisterPage(c *fiber.Ctx) error {
	return renderComponent(c, userui.RegisterPage("", nil))
}
func (h *Handler) LoginPage(c *fiber.Ctx) error { return renderComponent(c, userui.LoginPage("", nil)) }

func (h *Handler) Register(c *fiber.Ctx) error {
	form := User{Email: c.FormValue("email"), Pass: c.FormValue("password")}
	validate := validator.New()
	if err := validate.Struct(form); err != nil {
		errs := []string{err.Error()}
		if isHX(c) {
			return renderComponent(c, userui.RegisterForm(form.Email, errs))
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	u, token, err := h.Service.Register(c.Context(), form.Email, form.Pass)
	if err != nil {
		if isHX(c) {
			return renderComponent(c, userui.RegisterForm(form.Email, []string{"Error already registered"}))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.Service.SetSession(c, u.ID, token); err != nil {
		if isHX(c) {
			return renderComponent(c, userui.RegisterForm(form.Email, []string{"Problem with session"}))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "session set failed"})
	}

	if isHX(c) {
		return renderComponent(c, userui.RegisterSuccess(u.Email))
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"user_id": u.ID, "email": u.Email, "session_token": token})
}

func (h *Handler) Login(c *fiber.Ctx) error {
	form := User{Email: c.FormValue("email"), Pass: c.FormValue("password")}
	validate := validator.New()
	if err := validate.Struct(form); err != nil {
		errs := []string{err.Error()}
		if isHX(c) {
			return renderComponent(c, userui.LoginForm(form.Email, errs))
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	u, token, err := h.Service.Login(c.Context(), form.Email, form.Pass)
	if err != nil {
		if isHX(c) {
			return renderComponent(c, userui.LoginForm(form.Email, []string{"Invalid credentials"}))
		}
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
	}

	if err := h.Service.SetSession(c, u.ID, token); err != nil {
		if isHX(c) {
			return renderComponent(c, userui.LoginForm(form.Email, []string{"Problem with session"}))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "session set failed"})
	}

	c.Cookie(&fiber.Cookie{Name: "session_id", Value: token, HTTPOnly: true, SameSite: "Lax"})

	if isHX(c) {
		return renderComponent(c, userui.LoginSuccess(u.Email))
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"user_id": u.ID, "email": u.Email, "session_token": token})
}

func isHX(c *fiber.Ctx) bool { return c.Get("HX-Request") == "true" }

func renderComponent(c *fiber.Ctx, comp templ.Component) error {
	var buf bytes.Buffer
	if err := comp.Render(c.Context(), &buf); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("render error")
	}
	c.Type("html", "utf-8")
	return c.Send(buf.Bytes())
}
