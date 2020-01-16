package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/jmoiron/sqlx"
	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/platform/auth"
	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/platform/web"
	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/product"
	"go.opencensus.io/trace"
)

type Products struct {
	db  *sqlx.DB
	log *log.Logger
}

func (p *Products) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := trace.StartSpan(ctx, "handles.Product.List")
	defer span.End()

	list, err := product.List(ctx, p.db)
	if err != nil {
		return fmt.Errorf("error: listing products: %w", err)
	}

	return web.Respond(ctx, w, list, http.StatusOK)
}

func (p *Products) Retrieve(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := trace.StartSpan(ctx, "handles.Product.Retrieve")
	defer span.End()

	id := chi.URLParam(r, "id")

	prod, err := product.Retrieve(ctx, p.db, id)
	if err != nil {
		switch err {
		case product.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case product.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		default:
			return fmt.Errorf("get product %w", err)
		}
	}

	return web.Respond(ctx, w, prod, http.StatusOK)
}

func (p *Products) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := trace.StartSpan(ctx, "handles.Product.Create")
	defer span.End()

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	var np product.NewProduct
	if err := web.Decode(r, &np); err != nil {
		return fmt.Errorf("decoding product %w", err)
	}

	prod, err := product.Create(ctx, p.db, claims, np, time.Now())
	if err != nil {
		return fmt.Errorf("creating product %w", err)
	}

	return web.Respond(ctx, w, prod, http.StatusOK)
}

func (p *Products) AddSale(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := trace.StartSpan(ctx, "handles.Product.AddSale")
	defer span.End()

	var ns product.NewSale
	if err := web.Decode(r, &ns); err != nil {
		return fmt.Errorf("decoding new sale: %w", err)
	}

	productID := chi.URLParam(r, "id")

	sale, err := product.AddSale(ctx, p.db, ns, productID, time.Now())
	if err != nil {
		return fmt.Errorf("adding new sale: %w", err)
	}

	return web.Respond(ctx, w, sale, http.StatusCreated)
}

func (p *Products) ListSales(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := trace.StartSpan(ctx, "handles.Product.ListSales")
	defer span.End()

	id := chi.URLParam(r, "id")

	list, err := product.ListSales(ctx, p.db, id)
	if err != nil {
		return fmt.Errorf("get sales list %w", err)
	}

	return web.Respond(ctx, w, list, http.StatusOK)
}

func (p *Products) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := trace.StartSpan(ctx, "handles.Product.Update")
	defer span.End()

	id := chi.URLParam(r, "id")

	var update product.UpdateProduct
	if err := web.Decode(r, &update); err != nil {
		return fmt.Errorf("decoding product update %w", err)
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	if err := product.Update(ctx, p.db, claims, id, update, time.Now()); err != nil {
		switch err {
		case product.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case product.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		default:
			return fmt.Errorf("updating product %q", id)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (p *Products) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := trace.StartSpan(ctx, "handles.Product.Delete")
	defer span.End()

	id := chi.URLParam(r, "id")

	if err := product.Delete(ctx, p.db, id); err != nil {
		switch err {
		case product.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		default:
			return fmt.Errorf("deleting product %q", id)
		}
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}
