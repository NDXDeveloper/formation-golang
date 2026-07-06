/* ============================================================================
   Section 10.2 : Clean architecture / hexagonale en Go (sans sur-ingénierie)
   Description : Le DOMAINE — un package Go PUR : aucune dépendance technique
                 (pas de tag JSON, pas de database/sql, pas de framework).
                 Il contient l'entité Order, l'erreur du domaine ErrNotFound,
                 le PORT Repository (l'interface déclarée ICI, côté
                 consommateur, implémentée ailleurs par les adaptateurs) et
                 le use case Service.Place qui porte la règle métier. Les
                 dépendances pointent VERS ce package ; lui n'importe rien.
   Fichier source : 02-clean-architecture.md
   ============================================================================ */

package orders

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
)

// Order est une entité du domaine : des données et des règles, rien de plus.
type Order struct {
	ID       string
	Customer string
	Total    int64 // en centimes
}

var ErrNotFound = errors.New("commande introuvable")

// Repository est un PORT : l'interface dont le domaine a besoin pour
// persister. Déclarée ICI (côté consommateur), implémentée ailleurs
// (adaptateur PostgreSQL en production, faux en mémoire dans les tests).
type Repository interface {
	ByID(ctx context.Context, id string) (Order, error)
	Save(ctx context.Context, o Order) error
}

// Service porte la logique métier (le « use case »).
// Il ACCEPTE une interface (le port) et RETOURNE une struct concrète.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) Place(ctx context.Context, customer string, total int64) (Order, error) {
	if total <= 0 {
		return Order{}, errors.New("le montant doit être positif") // règle métier
	}
	o := Order{ID: newID(), Customer: customer, Total: total}
	if err := s.repo.Save(ctx, o); err != nil {
		return Order{}, err
	}
	return o, nil
}

// newID : un générateur trivial pour l'exemple (en vrai : UUID, ULID…).
var seq atomic.Int64

func newID() string { return fmt.Sprintf("ord-%d", seq.Add(1)) }
