package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"taskflow/internal/repository/postgres"
	"taskflow/internal/service"
	"taskflow/internal/transport/httpapi"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	dsn := getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/taskflow")

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatal("db unreachable: ", err)
	}

	repo := postgres.NewTaskRepo(pool)
	svc := service.NewTaskService(repo)
	h := httpapi.NewHandler(svc)

	addr := ":" + getenv("PORT", "8080")
	log.Println("listening on", addr)
	log.Fatal(http.ListenAndServe(addr, h.Routes()))
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
