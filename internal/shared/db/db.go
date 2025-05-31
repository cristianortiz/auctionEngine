package db

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

var (
	db   *pgx.Conn
	once sync.Once
)

func BuildPostgresDSN() string {
	_ = godotenv.Load()
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, dbname, sslmode,
	)
}

// GetDB returns a singleton *pgx.Conn instance using pgx driver and environment variables.
func GetPostgresDB() *pgx.Conn {
	once.Do(func() {
		// Carga las variables de entorno desde .env si existe
		_ = godotenv.Load()

		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")
		sslmode := os.Getenv("DB_SSLMODE")

		connStr := fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=%s",
			user, password, host, port, dbname, sslmode,
		)

		var err error
		db, err = pgx.Connect(context.Background(), connStr)
		if err != nil {
			panic("failed to connect to database: " + err.Error())
		}
	})
	return db
}
