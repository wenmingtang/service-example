package product

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opencensus.io/trace"
)

func AddSale(ctx context.Context, db *sqlx.DB, ns NewSale, productID string, now time.Time) (*Sale, error) {
	ctx, span := trace.StartSpan(ctx, "internal.product.AddSale")
	defer span.End()

	s := Sale{
		ID:          uuid.New().String(),
		ProductID:   productID,
		Quantity:    ns.Quantity,
		Paid:        ns.Paid,
		DateCreated: now,
	}

	const q = `insert into sales (sale_id, product_id, quantity, paid, date_created) values ($1, $2, $3, $4, $5)`

	_, err := db.ExecContext(ctx, q, s.ID, s.ProductID, s.Quantity, s.Paid, s.DateCreated)
	if err != nil {
		return nil, fmt.Errorf("inserting sale: %w", err)
	}

	return &s, nil
}

func ListSales(ctx context.Context, db *sqlx.DB, productID string) ([]Sale, error) {
	ctx, span := trace.StartSpan(ctx, "internal.product.ListSales")
	defer span.End()

	var sales []Sale

	const q = `select * from sales where product_id = $1`

	if err := db.SelectContext(ctx, &sales, q, productID); err != nil {
		return nil, fmt.Errorf("selecting sales: %w", err)
	}
	return sales, nil
}
