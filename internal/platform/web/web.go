package web

import (
	"context"
	"log"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"go.opencensus.io/trace"
)

type ctxKey int

const KeyValues ctxKey = 1

type Values struct {
	StatusCode int
	Start      time.Time
	TraceID    string
}

type Handler func(context.Context, http.ResponseWriter, *http.Request) error

type App struct {
	log      *log.Logger
	mux      *chi.Mux
	mw       []Middleware
	och      *ochttp.Handler
	shutdown chan os.Signal
}

func NewApp(shutdown chan os.Signal, log *log.Logger, mw ...Middleware) *App {
	app := App{
		log:      log,
		mux:      chi.NewRouter(),
		mw:       mw,
		shutdown: shutdown,
	}

	app.och = &ochttp.Handler{
		Handler:     app.mux,
		Propagation: &tracecontext.HTTPFormat{},
	}

	return &app
}

func (a *App) Handle(method, url string, h Handler, mw ...Middleware) {
	h = wrapMiddleware(mw, h)

	h = wrapMiddleware(a.mw, h)

	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx, span := trace.StartSpan(r.Context(), "internal.platform.web")
		defer span.End()

		v := Values{
			TraceID: span.SpanContext().TraceID.String(),
			Start:   time.Now(),
		}
		ctx = context.WithValue(r.Context(), KeyValues, &v)

		if err := h(ctx, w, r); err != nil {
			a.log.Printf("Unhandled error: %+v", err)

			if IsShutdown(err) {
				a.SignalShutdown()
			}
		}
	}

	a.mux.MethodFunc(method, url, fn)
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}

func (a *App) SignalShutdown() {
	a.log.Println("error returned from handler indicated integrity issue, shutting down service")
	a.shutdown <- syscall.SIGSTOP
}
