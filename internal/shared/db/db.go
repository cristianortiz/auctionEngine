package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var (
	dbPool *pgxpool.Pool
	once   sync.Once
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
func GetPostgresDBPool(ctx context.Context) (*pgxpool.Pool, error) {
	var err error
	once.Do(func() {

		databaseURL := BuildPostgresDSN()

		// Configura el pool
		config, configErr := pgxpool.ParseConfig(databaseURL)
		if configErr != nil {
			err = fmt.Errorf("failed to parse database config: %w", configErr)
			return
		}

		// config.MaxConns = 10
		// config.MinConns = 2
		// config.MaxConnLifetime = time.Hour
		// config.HealthCheckPeriod = time.Minute

		//connects using pool
		pool, connectErr := pgxpool.NewWithConfig(ctx, config)
		if connectErr != nil {
			err = fmt.Errorf("unable to connect to DB: %w", connectErr)
			return
		}
		dbPool = pool
	})

	if err != nil {
		return nil, err
	}

	// Verifica la conexión del pool
	if dbPool == nil {
		return nil, errors.New("database pool was not initialized")
	}
	if pingErr := dbPool.Ping(ctx); pingErr != nil {
		return nil, fmt.Errorf("database pool ping failed: %w", pingErr)
	}

	return dbPool, nil
}

// func GetPostgresDB() *pgx.Conn {
// 	once.Do(func() {
// 		// Carga las variables de entorno desde .env si existe
// 		_ = godotenv.Load()

// 		host := os.Getenv("DB_HOST")
// 		port := os.Getenv("DB_PORT")
// 		user := os.Getenv("DB_USER")
// 		password := os.Getenv("DB_PASSWORD")
// 		dbname := os.Getenv("DB_NAME")
// 		sslmode := os.Getenv("DB_SSLMODE")

// 		connStr := fmt.Sprintf(
// 			"postgres://%s:%s@%s:%s/%s?sslmode=%s",
// 			user, password, host, port, dbname, sslmode,
// 		)

// 		var err error
// 		db, err = pgx.Connect(context.Background(), connStr)
// 		if err != nil {
// 			panic("failed to connect to database: " + err.Error())
// 		}
// 	})
// 	return db
// }
