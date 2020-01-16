package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/mid"
	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/platform/auth"
	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/platform/web"
)

func API(shutdown chan os.Signal, db *sqlx.DB, log *log.Logger, authenticator *auth.Authenticator) http.Handler {
	app := web.NewApp(shutdown, log,
		mid.Logger(log),
		mid.Errors(log),
		mid.Metrics(),
		mid.Panics(log),
	)

	{
		c := Check{db: db}
		app.Handle(http.MethodGet, "/v1/health", c.Health)
	}

	{
		u := Users{db: db, authenticator: authenticator}
		app.Handle(http.MethodGet, "/v1/users/token", u.Token)
	}

	{
		p := Products{db: db, log: log}

		app.Handle(http.MethodGet, "/v1/products", p.List)
		app.Handle(http.MethodGet, "/v1/products/{id}", p.Retrieve, mid.Authenticate(authenticator))
		app.Handle(http.MethodPost, "/v1/products", p.Create, mid.Authenticate(authenticator))
		app.Handle(http.MethodPut, "/v1/products/{id}", p.Update, mid.Authenticate(authenticator))
		app.Handle(http.MethodDelete, "/v1/products/{id}", p.Delete, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))

		app.Handle(http.MethodPost, "/v1/products/{id}/sales", p.AddSale, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))
		app.Handle(http.MethodGet, "/v1/products/{id}/sales", p.ListSales, mid.Authenticate(authenticator))
	}

	return app
}
