package main

import (
	"context"
	"crypto/rsa"
	"errors"
	_ "expvar"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"contrib.go.opencensus.io/exporter/zipkin"
	"github.com/dgrijalva/jwt-go"
	openzipkin "github.com/openzipkin/zipkin-go"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"
	"gitlab.fenbishuo.com/fenbishuo/service-training/cmd/sales-api/internal/handlers"
	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/platform/auth"
	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/platform/conf"
	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/platform/database"
	"go.opencensus.io/trace"
)

func main() {
	if err := run(); err != nil {
		log.Println("shutting down", "error:", err)
		os.Exit(1)
	}
}

func run() error {
	log := log.New(os.Stdout, "SALES : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	var cfg struct {
		Web struct {
			Address         string        `conf:"default:localhost:8000"`
			Debug           string        `conf:"default:localhost:6060"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:5s"`
			ShutdownTimeout time.Duration `conf:"default:5s"`
		}
		DB struct {
			User       string `conf:"default:postgres"`
			Password   string `conf:"default:postgres,noprint"`
			Host       string `conf:"default:localhost"`
			Name       string `conf:"default:postgres"`
			DisableTLS bool   `conf:"default:true"`
		}
		Auth struct {
			KeyID          string `conf:"default:1"`
			PrivateKeyFile string `conf:"default:private.pem"`
			Algorithm      string `conf:"default:RS256"`
		}
		Trace struct {
			URL         string  `conf:"default:http://localhost:9411/api/v2/spans"`
			Service     string  `conf:"default:sales-api"`
			Probability float64 `conf:"default:1"`
		}
	}

	if err := conf.Parse(os.Args[1:], "SALES", &cfg); err != nil {
		if err == conf.ErrHelpWanted {
			usage, err := conf.Usage("SALES", &cfg)
			if err != nil {
				return fmt.Errorf("generating config usage: %w", err)
			}
			fmt.Println(usage)
			return nil
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	log.Printf("main: started")
	defer log.Println("main: completed")

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}
	log.Printf("main: Config: \n%v\n", out)

	db, err := database.Open(database.Config{
		User:       cfg.DB.User,
		Password:   cfg.DB.Password,
		Host:       cfg.DB.Host,
		Name:       cfg.DB.Name,
		DisableTLS: cfg.DB.DisableTLS,
	})
	if err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}
	defer db.Close()

	closer, err := registerTracer(
		cfg.Trace.Service,
		cfg.Web.Address,
		cfg.Trace.URL,
		cfg.Trace.Probability,
	)
	if err != nil {
		return err
	}
	defer closer()

	go func() {
		log.Println("debug service listening on", cfg.Web.Debug)
		err := http.ListenAndServe(cfg.Web.Debug, http.DefaultServeMux)
		log.Println("debug service closed", err)
	}()

	authenticator, err := createAuth(cfg.Auth.PrivateKeyFile, cfg.Auth.KeyID, cfg.Auth.Algorithm)
	if err != nil {
		return fmt.Errorf("constructing authenticator: %w", err)
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	api := http.Server{
		Addr:         cfg.Web.Address,
		Handler:      handlers.API(shutdown, db, log, authenticator),
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("main: API listening on %s", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		return fmt.Errorf("start server: %w", err)
	case sig := <-shutdown:
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		err := api.Shutdown(ctx)
		if err != nil {
			log.Printf("main: graceful shutdown did not complete in %v: %v", cfg.Web.ShutdownTimeout, err)
			err = api.Close()
		}

		switch {
		case sig == syscall.SIGSTOP:
			return errors.New("integrity issue caused shutdown")
		case err != nil:
			return fmt.Errorf("could not stop server gracefully %w", err)
		}
	}
	return nil
}

func createAuth(privateKeyFile, keyID, algorithm string) (*auth.Authenticator, error) {
	keyContents, err := ioutil.ReadFile(privateKeyFile)
	if err != nil {
		return nil, fmt.Errorf("reading auth private key %w", err)
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(keyContents)
	if err != nil {
		return nil, fmt.Errorf("parsing auth private key %w", err)
	}

	public := auth.NewSimpleKeyLookupFunc(keyID, key.Public().(*rsa.PublicKey))

	return auth.NewAuthenticator(key, keyID, algorithm, public)
}

func registerTracer(service, httpAddr, traceURL string, probability float64) (func() error, error) {
	localEndpoint, err := openzipkin.NewEndpoint(service, httpAddr)
	if err != nil {
		return nil, fmt.Errorf("creating the local zipkinEndpoint %w", err)
	}
	reporter := zipkinHTTP.NewReporter(traceURL)

	trace.RegisterExporter(zipkin.NewExporter(reporter, localEndpoint))
	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.ProbabilitySampler(probability),
	})

	return reporter.Close, nil
}
