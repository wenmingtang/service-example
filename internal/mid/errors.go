package mid

import (
	"context"
	"log"
	"net/http"

	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/platform/web"
	"go.opencensus.io/trace"
)

func Errors(log *log.Logger) web.Middleware {
	f := func(before web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			ctx, span := trace.StartSpan(ctx, "internal.mid.Errors")
			defer span.End()

			v, ok := ctx.Value(web.KeyValues).(*web.Values)
			if !ok {
				return web.NewShutdownError("web value missing from context")
			}

			if err := before(ctx, w, r); err != nil {
				log.Printf("%s : ERROR: %+v", v.TraceID, err)

				if err := web.RespondError(ctx, w, err); err != nil {
					return err
				}

				if ok := web.IsShutdown(err); ok {
					return err
				}
			}
			return nil
		}
		return h
	}
	return f
}
