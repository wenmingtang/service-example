package product_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/platform/auth"
	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/product"
	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/schema"
	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/tests"
)

func TestProducts(t *testing.T) {
	db, teardown := tests.NewUnit(t)
	defer teardown()

	newP := product.NewProduct{
		Name:     "Comic Book",
		Cost:     10,
		Quantity: 55,
	}

	now := time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)
	ctx := context.Background()

	claims := auth.NewClaims(
		"718ffbea-f4a1-4667-8ae3-b349da52675e",
		[]string{auth.RoleAdmin, auth.RoleUser},
		now, time.Hour,
	)

	p0, err := product.Create(ctx, db, claims, newP, now)
	if err != nil {
		t.Fatalf("creating product p0: %s", err)
	}

	p1, err := product.Retrieve(ctx, db, p0.ID)
	if err != nil {
		t.Fatalf("getting product p0: %s", err)
	}

	if diff := cmp.Diff(p1, p0); diff != "" {
		t.Fatalf("fetched != created: \n%s", diff)
	}

	update := product.UpdateProduct{
		Name: tests.StringPointer("Comics"),
		Cost: tests.IntPointer(25),
	}
	updatedTime := time.Date(2019, time.January, 1, 1, 1, 1, 0, time.UTC)

	if err := product.Update(ctx, db, claims, p0.ID, update, updatedTime); err != nil {
		t.Fatalf("updating product p0: %s", err)
	}

	saved, err := product.Retrieve(ctx, db, p0.ID)
	if err != nil {
		t.Fatalf("getting product p0: %s", err)
	}

	want := *p0
	want.Name = "Comics"
	want.Cost = 25
	want.DateUpdated = updatedTime

	if diff := cmp.Diff(want, *saved); diff != "" {
		t.Fatalf("updated record dit not match:\n%s", diff)
	}

	if err := product.Delete(ctx, db, p0.ID); err != nil {
		t.Fatalf("deleting product: %v", err)
	}

	_, err = product.Retrieve(ctx, db, p0.ID)
	if err == nil {
		t.Fatalf("should not be able to retrieve deleted product")
	}
}

func TestProductList(t *testing.T) {
	db, teardown := tests.NewUnit(t)
	defer teardown()

	if err := schema.Seed(db); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	ps, err := product.List(ctx, db)
	if err != nil {
		t.Fatalf("listing products: %s", err)
	}

	if exp, got := 2, len(ps); exp != got {
		t.Fatalf("expected product list size %v, got %v", exp, got)
	}
}
