package user

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	db "go-deadlink-scanner/internal/database/sqlc"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	Queries *db.Queries
}

func NewService(queries *db.Queries) *Service {
	return &Service{Queries: queries}
}

func (s *Service) Register(ctx context.Context, email, password string) (db.User, string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return db.User{}, "", err
	}

	u, err := s.Queries.CreateUser(ctx, db.CreateUserParams{Email: email, Password: string(hash)})
	if err != nil {
		return db.User{}, "", err
	}

	token, err := s.newSessionToken(ctx, u.ID)
	if err != nil {
		return db.User{}, "", err
	}

	return u, token, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (db.User, string, error) {
	user, err := s.Queries.GetUserByEmail(ctx, email)
	if err != nil {
		return db.User{}, "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return db.User{}, "", err
	}

	token, err := s.getOrCreateSessionToken(ctx, user.ID)
	if err != nil {
		return db.User{}, "", err
	}

	return user, token, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	err := s.Queries.DeleteSession(ctx, token)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) SetSession(c *fiber.Ctx, token string) error {
	c.Cookie(&fiber.Cookie{
		Name:     "session_token",
		Value:    token,
		HTTPOnly: true,
		SameSite: "Lax",
		Secure:   false,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		Path:     "/",
	})
	return nil
}

func (s *Service) getOrCreateSessionToken(ctx context.Context, userID int32) (string, error) {
	session, err := s.Queries.GetActiveSessionsByUser(ctx, userID)
	if err == nil {
		return session.SessionToken, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	return s.newSessionToken(ctx, userID)
}

func (s *Service) newSessionToken(ctx context.Context, userID int32) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	expires := time.Now().Add(7 * 24 * time.Hour)
	_, err := s.Queries.CreateSession(ctx, db.CreateSessionParams{
		UserID:       userID,
		SessionToken: token,
		ExpiresAt:    expires,
	})
	if err != nil {
		return "", err
	}
	return token, nil
}
