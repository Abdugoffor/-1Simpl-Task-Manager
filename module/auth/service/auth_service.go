package auth_service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	auth_dto "task-manager/module/auth/dto"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailExists        = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotActive      = errors.New("user is not active")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

type AuthService interface {
	Register(ctx context.Context, req *auth_dto.Register) (*auth_dto.Response, error)
	Login(ctx context.Context, req *auth_dto.Login) (*auth_dto.LoginResponse, error)
	Authenticate(ctx context.Context, token string) (*auth_dto.Response, error)
	Logout(ctx context.Context, token string) error
}

type authService struct {
	db *pgxpool.Pool
}

func NewAuthService(db *pgxpool.Pool) AuthService {
	return &authService{
		db: db,
	}
}

func (service *authService) Register(ctx context.Context, req *auth_dto.Register) (*auth_dto.Response, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	{
		if err != nil {
			return nil, err
		}
	}

	var res auth_dto.Response

	err = service.db.QueryRow(ctx, `
		INSERT INTO users(name, email, password_hash, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, email, status, created_at, updated_at
	`, strings.TrimSpace(req.Name), strings.ToLower(strings.TrimSpace(req.Email)), string(passwordHash), auth_dto.StatusNew).Scan(
		&res.ID, &res.Name, &res.Email, &res.Status, &res.CreatedAt, &res.UpdatedAt,
	)

	if err != nil {
		var databaseError *pgconn.PgError

		if errors.As(err, &databaseError) && databaseError.Code == "23505" && databaseError.ConstraintName == "users_email_key" {
			return nil, ErrEmailExists
		}

		return nil, err
	}

	return &res, nil
}

func (service *authService) Login(ctx context.Context, req *auth_dto.Login) (*auth_dto.LoginResponse, error) {
	var user auth_dto.Response
	var passwordHash string

	err := service.db.QueryRow(ctx, `
		SELECT id, name, email, password_hash, status, created_at, updated_at
		FROM users
		WHERE email = $1
	`, strings.ToLower(strings.TrimSpace(req.Email))).Scan(
		&user.ID, &user.Name, &user.Email, &passwordHash, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrInvalidCredentials
	}

	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	{
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, ErrInvalidCredentials
		}

		if err != nil {
			return nil, err
		}
	}

	if user.Status != auth_dto.StatusActive {
		return nil, ErrUserNotActive
	}

	var tokenBytes [32]byte

	_, err = rand.Read(tokenBytes[:])
	{
		if err != nil {
			return nil, err
		}
	}

	token := base64.RawURLEncoding.EncodeToString(tokenBytes[:])
	expiresAt := time.Now().UTC().Add(24 * time.Hour)

	err = service.db.QueryRow(ctx, `
		INSERT INTO user_sessions(token_hash, user_id, expires_at)
		SELECT $1, id, $2 FROM users WHERE id = $3 AND status = $4
		RETURNING expires_at
	`, hashToken(token), expiresAt, user.ID, auth_dto.StatusActive).Scan(&expiresAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotActive
	}

	if err != nil {
		return nil, err
	}

	return &auth_dto.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresAt:   expiresAt,
		User:        user,
	}, nil
}

func (service *authService) Authenticate(ctx context.Context, token string) (*auth_dto.Response, error) {
	if !validToken(token) {
		return nil, ErrInvalidToken
	}

	var res auth_dto.Response

	err := service.db.QueryRow(ctx, `
		SELECT u.id, u.name, u.email, u.status, u.created_at, u.updated_at
		FROM user_sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1 AND s.expires_at > NOW()
	`, hashToken(token)).Scan(
		&res.ID, &res.Name, &res.Email, &res.Status, &res.CreatedAt, &res.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrInvalidToken
	}

	if err != nil {
		return nil, err
	}

	if res.Status != auth_dto.StatusActive {
		return nil, ErrUserNotActive
	}

	return &res, nil
}

func (service *authService) Logout(ctx context.Context, token string) error {
	if !validToken(token) {
		return ErrInvalidToken
	}

	cmd, err := service.db.Exec(ctx, "DELETE FROM user_sessions WHERE token_hash = $1", hashToken(token))
	{
		if err != nil {
			return err
		}
	}

	if cmd.RowsAffected() == 0 {
		return ErrInvalidToken
	}

	return nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))

	return hex.EncodeToString(hash[:])
}

func validToken(token string) bool {
	if len(token) != 43 {
		return false
	}

	decoded, err := base64.RawURLEncoding.Strict().DecodeString(token)

	return err == nil && len(decoded) == 32
}
