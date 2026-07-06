/* ============================================================================
   Section 10.2 : Clean architecture / hexagonale en Go (sans sur-ingénierie)
   Description : L'adaptateur « DRIVEN » (piloté par le domaine) — il vit
                 dans son propre package, importe le domaine (jamais
                 l'inverse), satisfait le port orders.Repository
                 IMPLICITEMENT (aucun « implements »), et TRADUIT les
                 erreurs d'infrastructure en erreurs du domaine
                 (sql.ErrNoRows → orders.ErrNotFound). Compilé dans ce
                 projet ; l'exécution réelle contre PostgreSQL relève du
                 module 7 (le SQL est identique).
   Fichier source : 02-clean-architecture.md
   ============================================================================ */

package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/exemple/hexagonal/internal/orders"
)

type OrderRepo struct{ db *sql.DB }

func NewOrderRepo(db *sql.DB) *OrderRepo { return &OrderRepo{db: db} }

// OrderRepo satisfait implicitement orders.Repository — aucune déclaration « implements ».
func (r *OrderRepo) ByID(ctx context.Context, id string) (orders.Order, error) {
	var o orders.Order
	err := r.db.QueryRowContext(ctx,
		`SELECT id, customer, total FROM orders WHERE id = $1`, id).
		Scan(&o.ID, &o.Customer, &o.Total)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return orders.Order{}, orders.ErrNotFound // erreur infra → erreur domaine
	case err != nil:
		return orders.Order{}, err
	}
	return o, nil
}

func (r *OrderRepo) Save(ctx context.Context, o orders.Order) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO orders (id, customer, total) VALUES ($1, $2, $3)`,
		o.ID, o.Customer, o.Total)
	return err
}
