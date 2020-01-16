package mid

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"

	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/platform/web"
	"go.opencensus.io/trace"
)

func Panics(log *log.Logger) web.Middleware {
	f := func(after web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {
			ctx, span := trace.StartSpan(ctx, "internal.mid.Panics")
			defer span.End()

			v, ok := ctx.Value(web.KeyValues).(*web.Values)
			if !ok {
				return web.NewShutdownError("web value missing from context")
			}

			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("panic: %v", r)

					log.Printf("%s :\n%s", v.TraceID, debug.Stack())
				}
			}()

			return after(ctx, w, r)
		}

		return h
	}

	return f
}
