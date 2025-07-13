package auth

import (
	"auth/internal/domain/models"
	"auth/internal/domain/sessions"
	"auth/pkg/jwt"
	"auth/pkg/logger"
	"auth/pkg/storage"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type UserRepository interface {
	Create(ctx context.Context, email string, passHash []byte) (userID int64, err error)
	Get(ctx context.Context, email string) (user models.User, err error)
}

type AppRepository interface {
	Get(ctx context.Context, appID int) (app models.App, err error)
}

type RefreshStorage interface {
	Save(ctx context.Context, token string, session sessions.RefreshSession) error
	Get(ctx context.Context, token string) (*sessions.RefreshSession, error)
	Delete(ctx context.Context, token string) error
}

type AuthService struct {
	log                   *slog.Logger
	userRepo              UserRepository
	appRepo               AppRepository
	refreshStorage        RefreshStorage
	accessTTL, refreshTTL time.Duration
}

func New(log *slog.Logger, userRepo UserRepository, appRepo AppRepository, accessTTL, refreshTTL time.Duration) *AuthService {
	return &AuthService{log: log, userRepo: userRepo, appRepo: appRepo, accessTTL: accessTTL, refreshTTL: refreshTTL}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (userID int64, err error) {
	const op = "AuthService.Register"

	log := s.log.With(slog.String("op", op), slog.String("email", email))

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", logger.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	uid, err := s.userRepo.Create(ctx, email, passHash)
	if err != nil {
		log.Error("failed to save user", logger.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return uid, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string, appID int, ip, userAgent string) (access, refresh string, err error) {
	const op = "AuthService.Login"

	log := s.log.With(slog.String("op", op), slog.String("email", email))

	user, err := s.userRepo.Get(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", logger.Err(err))

			return "", "", fmt.Errorf("%s: %w", op, err)
		}

		log.Error("failed to get user", logger.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		log.Info("invalid credentials", logger.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	app, err := s.appRepo.Get(ctx, appID)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	accessToken, err := jwt.GenerateJWT(app.AccessSecret, user.ID, user.Email, app.ID, s.accessTTL)
	if err != nil {
		log.Error("faiiled to generate access token", logger.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	refreshToken := jwt.GenerateRandomToken(32)
	expiresAt := time.Now().Add(s.refreshTTL).UTC()

	session := sessions.RefreshSession{
		UserID:    user.ID,
		AppID:     app.ID,
		IP:        ip,
		UserAgent: userAgent,
		ExpiresAt: expiresAt,
	}

	if err := s.refreshStorage.Save(ctx, refreshToken, session); err != nil {
		log.Error("failed to save refresh token", logger.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged in successfully")

	return accessToken, refreshToken, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (access, refresh string, err error) {
	const op = "AuthService.Refresh"

	log := s.log.With(slog.String("op", op))

	session, err := s.refreshStorage.Get(ctx, refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	app, err := s.appRepo.Get(ctx, session.AppID)
	if err != nil {
		return "", "", fmt.Errorf("%s: app not found", op)
	}

	accessToken, err := jwt.GenerateJWT(app.AccessSecret, session.UserID, session.UserEmail, app.ID, s.accessTTL)
	if err != nil {
		log.Error("faiiled to generate access token", logger.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	newRefresh := jwt.GenerateRandomToken(32)
	newSession := sessions.RefreshSession{
		UserID:    session.UserID,
		UserEmail: session.UserEmail,
		AppID:     app.ID,
		IP:        session.IP,
		UserAgent: session.UserAgent,
		ExpiresAt: time.Now().Add(s.refreshTTL).UTC(),
	}

	if err := s.refreshStorage.Save(ctx, newRefresh, newSession); err != nil {
		return "", "", fmt.Errorf("%s: failed to save refresh token: %w", op, err)
	}

	if err := s.refreshStorage.Delete(ctx, refreshToken); err != nil {
		log.Error("failed to delete old refresh token", logger.Err(err))
	}

	return accessToken, newRefresh, nil
}
