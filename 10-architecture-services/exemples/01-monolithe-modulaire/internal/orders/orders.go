/* ============================================================================
   Section 10.1 : Monolithe modulaire vs microservices
   Description : Le module « commandes » et LE point le plus idiomatique du
                 cours : l'interface catalogGetter est définie PAR LE
                 CONSOMMATEUR — orders ne dépend que de la poignée de
                 méthodes qu'il utilise réellement (ici : Get). Trois
                 bénéfices d'un coup : couplage minimal, test trivial (on
                 injecte un faux), et la COUTURE D'EXTRACTION déjà en place
                 (voir cmd/server : un client réseau satisfera la même
                 interface sans qu'orders change d'une ligne).
   Fichier source : 01-monolithe-vs-microservices.md
   ============================================================================ */

package orders

import (
	"context"
	"fmt"

	"github.com/exemple/monolithe/internal/catalog"
)

// catalogGetter est l'interface définie PAR le consommateur : orders
// ne dépend que de la poignée de méthodes qu'il utilise réellement.
type catalogGetter interface {
	Get(ctx context.Context, id string) (catalog.Product, error)
}

type Service struct {
	catalog catalogGetter
}

// New reçoit la dépendance par INJECTION EXPLICITE — de simples
// constructeurs, aucun framework d'injection de dépendances.
func New(cat catalogGetter) *Service {
	return &Service{catalog: cat}
}

// Describe illustre un appel inter-modules : un appel de fonction
// EN MÉMOIRE à travers l'interface — jamais du HTTP vers localhost.
func (s *Service) Describe(ctx context.Context, productID string) (string, error) {
	p, err := s.catalog.Get(ctx, productID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("commande possible : %s (%d ct)", p.Name, p.Price), nil
}
