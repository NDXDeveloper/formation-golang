/* ============================================================================
   Section 13.3 : Tests d'intégration avec Testcontainers
   Description : Un dépôt d'utilisateurs minimal sur PostgreSQL (pgxpool). Le
                 code testé — Create et Get — que le test d'intégration exerce
                 contre une VRAIE base Postgres démarrée en conteneur, plutôt
                 que contre un mock.
   Fichier source : 03-tests-integration.md
   ============================================================================ */

package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("store: utilisateur introuvable")

type User struct {
	ID    int
	Email string
}

type Store struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

// Migrate crée le schéma (dans un vrai projet : golang-migrate/goose).
func (s *Store) Migrate(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS users (
		id    SERIAL PRIMARY KEY,
		email TEXT NOT NULL UNIQUE)`)
	return err
}

func (s *Store) Create(ctx context.Context, email string) (User, error) {
	var u User
	err := s.pool.QueryRow(ctx,
		`INSERT INTO users (email) VALUES ($1) RETURNING id, email`, email).
		Scan(&u.ID, &u.Email)
	return u, err
}

func (s *Store) Get(ctx context.Context, id int) (User, error) {
	var u User
	err := s.pool.QueryRow(ctx,
		`SELECT id, email FROM users WHERE id = $1`, id).Scan(&u.ID, &u.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	return u, err
}
