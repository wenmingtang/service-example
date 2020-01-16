package tests

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/platform/auth"
	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/platform/database"
	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/platform/database/databasetest"
	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/schema"
	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/user"
)

const (
	AdminID = "5cf37266-3473-4006-984f-9325122678b7"
	UserID  = "45b5fbd3-755f-4379-8f07-a58d4a30fa2f"
)

func NewUnit(t *testing.T) (*sqlx.DB, func()) {
	t.Helper()

	c := databasetest.StartContainer(t)

	db, err := database.Open(database.Config{
		User:       "postgres",
		Password:   "postgres",
		Host:       c.Host,
		Name:       "postgres",
		DisableTLS: true,
	})
	if err != nil {
		t.Fatalf("opening database connection: %v", err)
	}

	t.Log("waiting for database to be ready")

	var pingError error
	maxAttempts := 20
	for attempts := 1; attempts <= maxAttempts; attempts++ {
		pingError = db.Ping()
		if pingError == nil {
			break
		}
		time.Sleep(time.Duration(attempts) * 100 * time.Millisecond)
	}

	if pingError != nil {
		databasetest.DumpContainerLogs(t, c)
		databasetest.StopContainer(t, c)
		t.Fatalf("waiting for database to be ready: %v", pingError)
	}

	if err := schema.Migrate(db); err != nil {
		databasetest.StopContainer(t, c)
		t.Fatalf("migrating: %s", err)
	}

	teardown := func() {
		t.Helper()
		db.Close()
		databasetest.StopContainer(t, c)
	}
	return db, teardown
}

type Test struct {
	DB            *sqlx.DB
	Log           *log.Logger
	Authenticator *auth.Authenticator

	t       *testing.T
	cleanup func()
}

func New(t *testing.T) *Test {
	t.Helper()

	db, cleanup := NewUnit(t)

	if err := schema.Seed(db); err != nil {
		t.Fatal(err)
	}

	logger := log.New(os.Stdout, "TEST : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	kid := "4754d86b-7a6d-4df5-9c65-224741361492"
	kf := auth.NewSimpleKeyLookupFunc(kid, key.Public().(*rsa.PublicKey))
	authenticator, err := auth.NewAuthenticator(key, kid, "RS256", kf)
	if err != nil {
		t.Fatal(err)
	}

	return &Test{
		DB:            db,
		Log:           logger,
		Authenticator: authenticator,
		t:             t,
		cleanup:       cleanup,
	}
}

func (test *Test) Teardown() {
	test.cleanup()
}

func (test *Test) Token(email, pass string) string {
	test.t.Helper()

	claims, err := user.Authenticate(context.Background(), test.DB, time.Now(), email, pass)
	if err != nil {
		test.t.Fatal(err)
	}

	tkn, err := test.Authenticator.GenerateToken(claims)
	if err != nil {
		test.t.Fatal(err)
	}
	return tkn
}

func StringPointer(s string) *string {
	return &s
}

func IntPointer(i int) *int {
	return &i
}
