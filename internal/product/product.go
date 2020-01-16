package product

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/platform/auth"
	"go.opencensus.io/trace"
)

var (
	ErrNotFound  = errors.New("product not found")
	ErrInvalidID = errors.New("ID is not in its proper form")
	ErrForbidden = errors.New("attempted action is not allowed")
)

func List(ctx context.Context, db *sqlx.DB) ([]Product, error) {
	ctx, span := trace.StartSpan(ctx, "internal.product.List")
	defer span.End()

	var products []Product

	const q = `SELECT
			p.*,
			COALESCE(SUM(s.quantity), 0) AS sold,
			COALESCE(SUM(s.paid), 0) AS revenue
		FROM products AS p
		LEFT JOIN sales AS s ON p.product_id = s.product_id
		GROUP BY p.product_id`
	if err := db.SelectContext(ctx, &products, q); err != nil {
		return nil, fmt.Errorf("selecting products: %w", err)
	}

	return products, nil
}

func Retrieve(ctx context.Context, db *sqlx.DB, id string) (*Product, error) {
	ctx, span := trace.StartSpan(ctx, "internal.product.Retrieve")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidID
	}

	var p Product

	const q = `SELECT
			p.*,
			COALESCE(SUM(s.quantity), 0) AS sold,
			COALESCE(SUM(s.paid), 0) AS revenue
		FROM products AS p
		LEFT JOIN sales AS s ON p.product_id = s.product_id
		WHERE p.product_id = $1
		GROUP BY p.product_id`
	if err := db.GetContext(ctx, &p, q, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("selecting single product %w", err)
	}

	return &p, nil
}

func Create(ctx context.Context, db *sqlx.DB, user auth.Claims, np NewProduct, now time.Time) (*Product, error) {
	ctx, span := trace.StartSpan(ctx, "internal.product.Create")
	defer span.End()

	p := Product{
		ID:          uuid.New().String(),
		Name:        np.Name,
		Cost:        np.Cost,
		Quantity:    np.Quantity,
		UserID:      user.Subject,
		DateCreated: now.UTC(),
		DateUpdated: now.UTC(),
	}

	const q = `
		insert into products
		(product_id, user_id, name, cost, quantity, date_created, date_updated)
		values($1, $2, $3, $4, $5, $6, $7)
		`
	_, err := db.ExecContext(ctx, q,
		p.ID, p.UserID, p.Name,
		p.Cost, p.Quantity,
		p.DateCreated, p.DateUpdated)
	if err != nil {
		return nil, fmt.Errorf("inserting product %w", err)
	}

	return &p, nil
}

func Update(ctx context.Context, db *sqlx.DB, user auth.Claims, id string, update UpdateProduct, now time.Time) error {
	ctx, span := trace.StartSpan(ctx, "internal.product.Update")
	defer span.End()

	p, err := Retrieve(ctx, db, id)
	if err != nil {
		return err
	}

	if !user.HasRole(auth.RoleAdmin) && p.UserID != user.Subject {
		return ErrForbidden
	}

	if update.Name != nil {
		p.Name = *update.Name
	}

	if update.Cost != nil {
		p.Cost = *update.Cost
	}

	if update.Quantity != nil {
		p.Quantity = *update.Quantity
	}

	p.DateUpdated = now

	const q = `update products set name = $2, cost = $3, quantity = $4, date_updated = $5 where product_id = $1`
	_, err = db.ExecContext(ctx, q, p.ID, p.Name, p.Cost, p.Quantity, p.DateUpdated)
	if err != nil {
		return fmt.Errorf("updating product: %w", err)
	}
	return nil
}

func Delete(ctx context.Context, db *sqlx.DB, id string) error {
	ctx, span := trace.StartSpan(ctx, "internal.product.Delete")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidID
	}

	const q = `delete from products where product_id = $1`

	if _, err := db.ExecContext(ctx, q, id); err != nil {
		return fmt.Errorf("deleting product: %w", err)
	}

	return nil
}
