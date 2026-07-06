/* ============================================================================
   Section 10.1 : Monolithe modulaire vs microservices
   Description : Le COMPOSITION ROOT (le main) — le câblage se fait une seule
                 fois, par un simple enchaînement de constructeurs lisible de
                 haut en bas ; aucune magie, aucune annotation. La démo
                 déroule la trajectoire du cours : (1) le monolithe
                 modulaire, appels EN MÉMOIRE entre modules ; (2) le jour de
                 l'EXTRACTION, le catalogue devient un service distant
                 (httptest) et un client réseau satisfait la MÊME interface
                 — orders n'a pas changé d'une ligne.
   Fichier source : 01-monolithe-vs-microservices.md
   Lancer : go run ./cmd/server        (aucun service requis)
   ============================================================================ */

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/exemple/monolithe/internal/catalog"
	"github.com/exemple/monolithe/internal/catalogclient"
	"github.com/exemple/monolithe/internal/orders"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== 1. Le monolithe modulaire : appels en mémoire ===")
	// Le graphe de dépendances : un enchaînement de constructeurs.
	cat := catalog.New()   // module catalogue
	ord := orders.New(cat) // orders dépend de catalog → injection explicite
	desc, err := ord.Describe(ctx, "p-42")
	fmt.Println(desc, "· err =", err)

	fmt.Println()
	fmt.Println("=== 2. L'extraction : le catalogue devient un service DISTANT ===")
	// Le « catalogue distant » : en vrai, un microservice ; ici, httptest.
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(catalog.Product{
			ID: "p-42", Name: "Clavier mécanique (appel RÉSEAU)", Price: 8900,
		})
	}))
	defer remote.Close()

	// Un client réseau satisfait la MÊME interface catalogGetter :
	// orders.New n'a pas changé — c'est toute la thèse du chapitre.
	ord2 := orders.New(catalogclient.New(remote.URL))
	desc2, err := ord2.Describe(ctx, "p-42")
	fmt.Println(desc2, "· err =", err)

	fmt.Println()
	fmt.Println("orders est resté identique entre les deux mondes : extraction = refactor mécanique ✔")
}
