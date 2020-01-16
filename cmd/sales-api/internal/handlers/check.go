package handlers

import (
	"context"
	"net/http"

	"github.com/jmoiron/sqlx"
	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/platform/database"
	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/platform/web"
	"go.opencensus.io/trace"
)

type Check struct {
	db *sqlx.DB
}

func (c *Check) Health(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := trace.StartSpan(ctx, "handles.Check.Health")
	defer span.End()

	var health struct {
		Status string `json:"status"`
	}

	if err := database.StatusCheck(ctx, c.db); err != nil {
		return web.Respond(ctx, w, health, http.StatusInternalServerError)
	}

	health.Status = "ok"
	return web.Respond(ctx, w, health, http.StatusOK)
}
