package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/halora-land/halora-be/internal/models"
)

// HashPassword returns a bcrypt hash of the plaintext password.
func HashPassword(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(h), nil
}

// CheckPassword verifies a plaintext password against a bcrypt hash.
// Returns nil on match, non-nil on mismatch/error.
func CheckPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// SessionClaims is the JWT claims structure for local-issued session tokens.
type SessionClaims struct {
	UserID int32 `json:"uid"`
	jwt.RegisteredClaims
}

// IssueSessionToken creates a signed HS256 JWT carrying the user ID.
// The token expires after 24 hours.
func IssueSessionToken(userID int32, secret string) (string, error) {
	if secret == "" {
		return "", errors.New("JWT secret is empty")
	}
	claims := SessionClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// VerifySessionToken parses and validates a session JWT, returning the user ID.
func VerifySessionToken(tokenStr, secret string) (int32, error) {
	if secret == "" {
		return 0, errors.New("JWT secret is empty")
	}
	claims := &SessionClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return 0, fmt.Errorf("parse session token: %w", err)
	}
	if !token.Valid {
		return 0, errors.New("invalid session token")
	}
	return claims.UserID, nil
}

// LocalLogin verifies email/password against the local DB and returns the user.
func (v *Verifier) LocalLogin(ctx context.Context, email, password string) (*AuthUser, error) {
	var (
		u           models.User
		passwordHash *string
	)
	err := v.pool.QueryRow(ctx, `
		SELECT id, "namaLengkap", email, role, "accountType", "isDemo", "passwordHash"
		FROM users WHERE email = $1`, email).
		Scan(&u.ID, &u.NamaLengkap, &u.Email, &u.Role, &u.AccountType, &u.IsDemo, &passwordHash)
	if err == pgx.ErrNoRows {
		return nil, errors.New("email atau password salah")
	}
	if err != nil {
		return nil, err
	}
	if passwordHash == nil || *passwordHash == "" {
		return nil, errors.New("akun ini tidak memiliki password, silakan daftar ulang")
	}
	if err := CheckPassword(*passwordHash, password); err != nil {
		return nil, errors.New("email atau password salah")
	}
	return &AuthUser{
		UserID:      u.ID,
		Email:       u.Email,
		Role:        u.Role,
		NamaLengkap: u.NamaLengkap,
		AccountType: u.AccountType,
		IsDemo:      u.IsDemo,
	}, nil
}

// LocalRegister creates a new user in the local DB with a hashed password.
// For now all users are created as ADMIN (per-product decision).
func (v *Verifier) LocalRegister(ctx context.Context, namaLengkap, email, password string) (*AuthUser, error) {
	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	var u models.User
	err = v.pool.QueryRow(ctx, `
		INSERT INTO users ("namaLengkap", email, role, "accountType", "isDemo", "passwordHash")
		VALUES ($1, $2, 'ADMIN', 'free', false, $3)
		RETURNING id, "namaLengkap", email, role, "accountType", "isDemo"`,
		namaLengkap, email, hash).
		Scan(&u.ID, &u.NamaLengkap, &u.Email, &u.Role, &u.AccountType, &u.IsDemo)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, errors.New("email sudah terdaftar")
		}
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &AuthUser{
		UserID:      u.ID,
		Email:       u.Email,
		Role:        u.Role,
		NamaLengkap: u.NamaLengkap,
		AccountType: u.AccountType,
		IsDemo:      u.IsDemo,
	}, nil
}

// LocalListUsers returns all users, newest first.
func (v *Verifier) LocalListUsers(ctx context.Context) ([]models.User, error) {
	rows, err := v.pool.Query(ctx, `
		SELECT id, "namaLengkap", email, role, "accountType", "isDemo", "createdAt", "updatedAt"
		FROM users ORDER BY "createdAt" DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.NamaLengkap, &u.Email, &u.Role, &u.AccountType, &u.IsDemo, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// LocalUpdatePassword sets a new password hash for the given user ID.
func (v *Verifier) LocalUpdatePassword(ctx context.Context, userID int32, password string) error {
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	_, err = v.pool.Exec(ctx, `UPDATE users SET "passwordHash" = $1 WHERE id = $2`, hash, userID)
	return err
}

// LocalUpdateProfile updates the namaLengkap and email for the given user ID.
func (v *Verifier) LocalUpdateProfile(ctx context.Context, userID int32, namaLengkap, email string) (*AuthUser, error) {
	var u models.User
	err := v.pool.QueryRow(ctx, `
		UPDATE users SET "namaLengkap" = $1, email = $2, "updatedAt" = NOW()
		WHERE id = $3
		RETURNING id, "namaLengkap", email, role, "accountType", "isDemo"`,
		namaLengkap, email, userID).
		Scan(&u.ID, &u.NamaLengkap, &u.Email, &u.Role, &u.AccountType, &u.IsDemo)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, errors.New("email sudah terdaftar")
		}
		return nil, fmt.Errorf("update profile: %w", err)
	}
	return &AuthUser{
		UserID:      u.ID,
		Email:       u.Email,
		Role:        u.Role,
		NamaLengkap: u.NamaLengkap,
		AccountType: u.AccountType,
		IsDemo:      u.IsDemo,
	}, nil
}

// isUniqueViolation checks for a PostgreSQL unique constraint violation.
func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "unique constraint")
}
