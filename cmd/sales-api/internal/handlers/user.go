package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/platform/auth"
	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/platform/web"
	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/user"
	"go.opencensus.io/trace"
)

type Users struct {
	db            *sqlx.DB
	authenticator *auth.Authenticator
}

func (u *Users) Token(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := trace.StartSpan(ctx, "handles.User.Token")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	email, pass, ok := r.BasicAuth()
	if !ok {
		err := errors.New("must provide email and password in Basic auth")
		return web.NewRequestError(err, http.StatusUnauthorized)
	}

	claims, err := user.Authenticate(ctx, u.db, v.Start, email, pass)
	if err != nil {
		switch err {
		case user.ErrAuthenticationFailure:
			return web.NewRequestError(err, http.StatusUnauthorized)
		default:
			return fmt.Errorf("authenticating %w", err)
		}
	}

	var tkn struct {
		Token string `json:"token"`
	}

	tkn.Token, err = u.authenticator.GenerateToken(claims)
	if err != nil {
		return fmt.Errorf("generating token %w", err)
	}
	return web.Respond(ctx, w, tkn, http.StatusOK)
}
