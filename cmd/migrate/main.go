package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"taskflow/migrations"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: migrate up|down|status")
	}

	dsn := getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/taskflow")
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	switch os.Args[1] {
	case "up":
		err = goose.UpContext(ctx, db, ".")
	case "down":
		err = goose.DownContext(ctx, db, ".")
	case "status":
		err = goose.StatusContext(ctx, db, ".")
	default:
		log.Fatalf("unknown command %q", os.Args[1])
	}
	if err != nil {
		log.Fatal(err)
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
