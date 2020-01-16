package product_test

import (
	"context"
	"testing"
	"time"

	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/platform/auth"
	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/product"
	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/tests"
)

func TestSales(t *testing.T) {
	db, teardown := tests.NewUnit(t)
	defer teardown()

	now := time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)

	ctx := context.Background()

	newPuzzles := product.NewProduct{
		Name:     "Puzzles",
		Cost:     25,
		Quantity: 6,
	}

	claims := auth.NewClaims(
		"718ffbea-f4a1-4667-8ae3-b349da52675e",
		[]string{auth.RoleAdmin, auth.RoleUser},
		now, time.Hour,
	)

	puzzles, err := product.Create(ctx, db, claims, newPuzzles, now)
	if err != nil {
		t.Fatalf("creating product: %s", err)
	}

	newToys := product.NewProduct{
		Name:     "Toys",
		Cost:     40,
		Quantity: 3,
	}
	toys, err := product.Create(ctx, db, claims, newToys, now)
	if err != nil {
		t.Fatalf("creating product: %s", err)
	}

	{
		ns := product.NewSale{
			Quantity: 3,
			Paid:     70,
		}

		s, err := product.AddSale(ctx, db, ns, puzzles.ID, now)
		if err != nil {
			t.Fatalf("adding sale: %s", err)
		}

		sales, err := product.ListSales(ctx, db, puzzles.ID)
		if err != nil {
			t.Fatalf("listing sales: %s", err)
		}
		if exp, got := 1, len(sales); exp != got {
			t.Fatalf("expected sale list size %v, got %v", exp, got)
		}

		if exp, got := s.ID, sales[0].ID; exp != got {
			t.Fatalf("expected first sale ID %v, got %v", exp, got)
		}

		sales, err = product.ListSales(ctx, db, toys.ID)
		if err != nil {
			t.Fatalf("listing sales: %s", err)
		}
		if exp, got := 0, len(sales); exp != got {
			t.Fatalf("expected sale list size %v, got %v", exp, got)
		}
	}
}