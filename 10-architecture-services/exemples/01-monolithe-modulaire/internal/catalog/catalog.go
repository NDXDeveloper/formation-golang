/* ============================================================================
   Section 10.1 : Monolithe modulaire vs microservices
   Description : Le module « catalogue » — un bounded context sous internal/,
                 qui expose une PETITE SURFACE : le type Product, le point
                 d'entrée Service et son constructeur New ; le store, lui,
                 est non exporté (invisible hors du package). Les autres
                 modules passent obligatoirement par Service : c'est la
                 version Go de « exposer une API, cacher l'implémentation »,
                 tenue par la simple règle majuscule/minuscule.
   Fichier source : 01-monolithe-vs-microservices.md
   ============================================================================ */

package catalog

import "context"

// Product est le modèle que le module expose à l'extérieur.
type Product struct {
	ID    string
	Name  string
	Price int64 // en centimes
}

// Service est le point d'entrée public du module catalogue.
type Service struct {
	store *store // non exporté : les internes restent invisibles hors du package
}

// New construit le module à partir de ses dépendances (ici : aucune —
// dans un vrai service, on y injecterait un *sql.DB, cf. § 7.1).
func New() *Service {
	return &Service{store: newStore()}
}

func (s *Service) Get(ctx context.Context, id string) (Product, error) {
	return s.store.byID(ctx, id) // … logique métier …
}
