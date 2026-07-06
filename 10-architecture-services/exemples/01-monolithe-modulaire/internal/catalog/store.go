/* ============================================================================
   Section 10.1 : Monolithe modulaire vs microservices
   Description : L'accès aux données PROPRES au module catalogue — type non
                 exporté : aucun autre module ne peut le manipuler ni lire
                 « ses tables » directement (la règle de propriété des
                 données du cours). Ici, un simple jeu de données en mémoire
                 tient lieu de base.
   Fichier source : 01-monolithe-vs-microservices.md
   ============================================================================ */

package catalog

import (
	"context"
	"errors"
)

type store struct {
	produits map[string]Product // les « tables » du module, à lui seul
}

func newStore() *store {
	return &store{produits: map[string]Product{
		"p-42": {ID: "p-42", Name: "Clavier mécanique (appel EN MÉMOIRE)", Price: 8900},
	}}
}

func (st *store) byID(_ context.Context, id string) (Product, error) {
	p, ok := st.produits[id]
	if !ok {
		return Product{}, errors.New("produit introuvable")
	}
	return p, nil
}
