/*
@repo: https://github.com/jackc/pgx
@docs: https://pkg.go.dev/github.com/jackc/pgx/v5
@docs: https://pkg.go.dev/github.com/jackc/pgx/v5#section-documentation
@docs: https://github.com/jackc/pgx/wiki/Getting-started-with-pgx
*/

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// postgres://username:password@host:port/<db-name>
	defaultDatabaseDSN = "postgres://admin:admin@localhost:5432/any_business"
)

func main() {
	// oneSimple()
	// twoConnectionPoolSimple()
	twoConnectionPoolComplexConfigs()
}

// WARNING: this is not safe for concurrent programs, only script or testings
func oneSimple() {
	dsn := getDbDSN()
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer closerDBWithCtx(conn, ctx)

	// one row
	var name string
	err = conn.QueryRow(ctx, "SELECT name FROM auth.account WHERE name = 'user_test_1'").Scan(&name)
	if err != nil {
		log.Fatalf("QueryRow failed: %v\n", err)
	}
	log.Printf("Result of QueryRow: %s\n", name)

	rows, err := conn.Query(ctx, "SELECT name FROM auth.account LIMIT 1000")
	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}
	defer rows.Close()

	i := 0
	names := make([]string, 0)
	for rows.Next() {
		i++
		var name string
		err := rows.Scan(&name)
		if err != nil {
			log.Fatalf("Scan failed at row %d: %v", i, err)
		}

		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Result of Query: %v\n", names)
	for _, name := range names {
		log.Printf("- %s\n", name)
	}
}

func twoConnectionPoolSimple() {
	dsn := getDbDSN()
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Unable to create connection pool @ %s: %v\n", dsn, err)
	}
	defer pool.Close()

	var name string
	err = pool.QueryRow(ctx, "SELECT name FROM auth.account WHERE name = 'user_test_1'").Scan(&name)
	if err != nil {
		log.Fatalf("QueryRow failed: %v\n", err)
	}
	log.Printf("Result of QueryRow: %s\n", name)

}

func twoConnectionPoolComplexConfigs() {
	dsn := getDbDSN()
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatal(err)
	}
	// Pool tuning
	cfg.MinConns = 2
	cfg.MaxConns = 20
	cfg.HealthCheckPeriod = 30 * time.Second
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		log.Fatalf("Unable to create connection pool @ %s: %v\n", dsn, err)
	}
	defer pool.Close()

	// Ping
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		log.Fatalf("ping: %v", err)
	}
	fmt.Println("Connected with Postgres!")

	var name string
	err = pool.QueryRow(ctx, "SELECT name FROM auth.account WHERE name = 'user_test_1'").Scan(&name)
	if err != nil {
		log.Fatalf("QueryRow failed: %v\n", err)
	}
	log.Printf("Result of QueryRow: %s\n", name)

	// MULTI ROW
	rows, err := pool.Query(ctx, "SELECT name FROM auth.account LIMIT 1000")
	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}
	defer rows.Close()

	i := 0
	names := make([]string, 0)
	for rows.Next() {
		i++
		var name string
		err := rows.Scan(&name)
		if err != nil {
			log.Fatalf("Scan failed at row %d: %v", i, err)
		}

		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Result of Query: %v\n", names)
	for _, name := range names {
		log.Printf("- %s\n", name)
	}

	// Get full Account
	type Account struct {
		ID       uuid.UUID
		Name     string
		Email    string
		Disabled bool
		Created  time.Time
		Updated  time.Time
	}

	name = "user_test_1"
	var account Account
	err = pool.QueryRow(ctx, "SELECT * FROM auth.account WHERE name = $1", name).Scan(
		&account.ID,
		&account.Name,
		&account.Email,
		&account.Disabled,
		&account.Created,
		&account.Updated,
	)
	if err != nil {
		log.Fatalf("QueryRow failed: %v\n", err)
	}
	log.Printf("Result of QueryRow: %v\n", account)

	// many accounts |
	rows, err = pool.Query(ctx, "SELECT * FROM auth.account LIMIT 1000")
	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}
	defer rows.Close()

	i = 0
	var accounts []Account
	for rows.Next() {
		i++
		var account Account
		err := rows.Scan(
			&account.ID,
			&account.Name,
			&account.Email,
			&account.Disabled,
			&account.Created,
			&account.Updated,
		)
		if err != nil {
			log.Fatalf("Scan failed at row %d: %v", i, err)
		}
		accounts = append(accounts, account)
	}
	for _, account := range accounts {
		log.Printf("%v\n", account)
	}
}

func getDbDSN() string {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = defaultDatabaseDSN
	}
	return dsn
}

func closer(obj io.Closer) {
	err := obj.Close()
	if err != nil {
		log.Printf("Closer of type %T encountered an error while closing: %v\n", obj, err)
	}
}
func closerDBWithCtx(conn *pgx.Conn, ctx context.Context) {
	err := conn.Close(ctx)
	if err != nil {
		log.Printf("Closer of type %T encountered an error while closing: %v\n", conn, err)
	}
}
