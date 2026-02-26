package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()

	// Connection string — adjust host/port/user/dbname as needed.
	// DoltgreSQL defaults: port 5432, user "doltgres", no password, db "doltgres"
	connStr := getEnv("DOLTGRESQL_CONN", "postgres://postgres:password@localhost:5432/postgres")

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close(ctx)

	fmt.Println("Connected to DoltgreSQL.")

	// ------------------------------------------------------------------
	// Setup: create a table and insert some rows so we have data to query
	// ------------------------------------------------------------------
	setupSQL := []string{
		`DROP TABLE IF EXISTS users`,
		`CREATE TABLE users (
			id   INT PRIMARY KEY,
			name TEXT NOT NULL
		)`,
		`INSERT INTO users VALUES (1, 'alice'), (2, 'bob'), (3, 'carol'), (4, 'dave')`,
	}

	for _, sql := range setupSQL {
		if _, err := conn.Exec(ctx, sql); err != nil {
			log.Fatalf("setup error (%s): %v", sql, err)
		}
	}
	fmt.Println("Table created and seeded.")

	// ------------------------------------------------------------------
	// Parameterized query using ANY($1) with an array argument
	// ------------------------------------------------------------------
	// We want:  SELECT id, name FROM users WHERE id = ANY($1)
	// and pass in a []int32 slice as the array.
	targetIDs := []int32{1, 3}

	rows, err := conn.Query(ctx,
		`SELECT id, name FROM users WHERE id = ANY($1)`,
		targetIDs, // pgx serializes a Go slice as a Postgres array
	)
	if err != nil {
		log.Fatalf("query error: %v", err)
	}
	defer rows.Close()

	fmt.Printf("\nRows matching id = ANY(%v):\n", targetIDs)
	for rows.Next() {
		var id int32
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			log.Fatalf("scan error: %v", err)
		}
		fmt.Printf("  id=%d  name=%s\n", id, name)
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("rows error: %v", err)
	}

	fmt.Println("\nDone.")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
