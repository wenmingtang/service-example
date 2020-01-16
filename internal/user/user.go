package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/platform/auth"
	"go.opencensus.io/trace"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrAuthenticationFailure = errors.New("authentication failed")
)

func Create(ctx context.Context, db *sqlx.DB, n NewUser, now time.Time) (*User, error) {
	ctx, span := trace.StartSpan(ctx, "internal.user.Create")
	defer span.End()

	hash, err := bcrypt.GenerateFromPassword([]byte(n.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("generating password hash %w", err)
	}

	u := User{
		ID:           uuid.New().String(),
		Name:         n.Name,
		Email:        n.Email,
		Roles:        n.Roles,
		PasswordHash: hash,
		DateCreated:  now.UTC(),
		DateUpdated:  now.UTC(),
	}

	const q = `INSERT INTO users
		(user_id, name, email, password_hash, roles, date_created, date_updated)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err = db.ExecContext(
		ctx, q,
		u.ID, u.Name, u.Email, u.PasswordHash,
		u.Roles, u.DateCreated, u.DateUpdated,
	)
	if err != nil {
		return nil, fmt.Errorf("inserting user %w", err)
	}
	return &u, nil
}

func Authenticate(ctx context.Context, db *sqlx.DB, now time.Time, email, password string) (auth.Claims, error) {
	ctx, span := trace.StartSpan(ctx, "internal.user.Authenticate")
	defer span.End()

	const q = `select * from users where email = $1`
	var u User
	if err := db.GetContext(ctx, &u, q, email); err != nil {
		if err == sql.ErrNoRows {
			return auth.Claims{}, ErrAuthenticationFailure
		}

		return auth.Claims{}, fmt.Errorf("selecting single user %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(password)); err != nil {
		return auth.Claims{}, ErrAuthenticationFailure
	}

	claims := auth.NewClaims(u.ID, u.Roles, now, time.Hour)
	return claims, nil
}
