package mid

import (
	"context"
	"expvar"
	"net/http"
	"runtime"

	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/platform/web"
	"go.opencensus.io/trace"
)

var m = struct {
	gr  *expvar.Int
	req *expvar.Int
	err *expvar.Int
}{
	gr:  expvar.NewInt("goroutines"),
	req: expvar.NewInt("requests"),
	err: expvar.NewInt("errors"),
}

func Metrics() web.Middleware {
	f := func(before web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			ctx, span := trace.StartSpan(ctx, "internal.mid.Metrics")
			defer span.End()

			err := before(ctx, w, r)

			m.req.Add(1)

			if m.req.Value()%100 == 0 {
				m.gr.Set(int64(runtime.NumGoroutine()))
			}

			if err != nil {
				m.err.Add(1)
			}

			return err
		}
		return h
	}
	return f
}
