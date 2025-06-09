package postgres

import (
	"context"
	"errors"

	"github.com/cristianortiz/auctionEngine/internal/user/domain" // Importa el dominio del usuario
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jackc/pgx/v5"
)

// UserRepository implementa la interfaz domain.UserRepository para PostgreSQL.
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository crea una nueva instancia de UserRepository.
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// GetByID obtiene un usuario por su ID desde la base de datos.
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `SELECT id FROM users WHERE id = $1`

	var userID uuid.UUID
	err := r.db.QueryRow(ctx, query, id).Scan(&userID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Usuario no encontrado
			return nil, nil // O un error específico como domain.ErrUserNotFound
		}
		// Otro error de base de datos
		return nil, err
	}

	// Si se encontró el usuario, crea y retorna la entidad User
	user := &domain.User{
		ID: userID,
	}

	return user, nil
}

// Otros métodos del repositorio se agregarán en fases futuras.
