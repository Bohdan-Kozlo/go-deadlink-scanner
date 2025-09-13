package user

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	db "go-deadlink-scanner/internal/database/sqlc"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	Queries      *db.Queries
	SessionStore *session.Store
}

func NewService(queries *db.Queries, store *session.Store) *Service {
	return &Service{Queries: queries, SessionStore: store}
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

	token, err := s.newSessionToken(ctx, user.ID)
	if err != nil {
		return db.User{}, "", err
	}

	return user, token, nil
}

func (s *Service) SetSession(c *fiber.Ctx, userID int32, token string) error {
	sess, err := s.SessionStore.Get(c)
	if err != nil {
		return err
	}
	sess.Set("user_id", userID)
	sess.Set("session_token", token)
	sess.SetExpiry(7 * 24 * time.Hour)

	return sess.Save()
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
