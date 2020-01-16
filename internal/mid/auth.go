package mid

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/platform/auth"
	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/platform/web"
	"go.opencensus.io/trace"
)

var ErrForbidden = web.NewRequestError(
	errors.New("you are not authorized for that action"),
	http.StatusForbidden,
)

func Authenticate(authenticator *auth.Authenticator) web.Middleware {
	f := func(after web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			ctx, span := trace.StartSpan(ctx, "internal.mid.Authenticate")
			defer span.End()

			parts := strings.Split(r.Header.Get("Authorization"), " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				err := errors.New("expected authorization header format: Bearer <token>")
				return web.NewRequestError(err, http.StatusUnauthorized)
			}

			_, span = trace.StartSpan(ctx, "auth.ParseClaims")
			claims, err := authenticator.ParseClaims(parts[1])
			span.End()
			if err != nil {
				return web.NewRequestError(err, http.StatusUnauthorized)
			}

			ctx = context.WithValue(ctx, auth.Key, claims)

			return after(ctx, w, r)
		}

		return h
	}

	return f
}

func HasRole(roles ...string) web.Middleware {
	f := func(after web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			ctx, span := trace.StartSpan(ctx, "internal.mid.HasRole")
			defer span.End()

			claims, ok := ctx.Value(auth.Key).(auth.Claims)
			if !ok {
				return errors.New("claims missing from context: HasRole called without/before Authenticate")
			}
			if !claims.HasRole(roles...) {
				return ErrForbidden
			}

			return after(ctx, w, r)
		}
		return h
	}
	return f
}
