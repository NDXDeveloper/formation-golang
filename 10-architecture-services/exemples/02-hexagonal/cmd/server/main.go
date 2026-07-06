/* ============================================================================
   Section 10.2 : Clean architecture / hexagonale en Go (sans sur-ingénierie)
   Description : Le composition root — lui-même un adaptateur : il relie le
                 monde extérieur au domaine par de simples constructeurs,
                 AUCUN framework d'injection. La démo prouve : (1) la
                 satisfaction IMPLICITE du port par l'adaptateur PostgreSQL
                 (la ligne var _ compile = c'est prouvé) ; (2) le VRAI
                 bénéfice du port : tester le domaine SANS base — le faux en
                 mémoire memRepo (écrit à la main, quelques lignes) ; (3) la
                 chaîne complète driving → use case → driven : POST /orders
                 sur l'adaptateur HTTP, DTO total_cents, et la règle métier
                 qui remonte en 422.
   Fichier source : 02-clean-architecture.md
   Lancer : go run ./cmd/server        (aucune base requise)
   ============================================================================ */

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/exemple/hexagonal/internal/orders"
	"github.com/exemple/hexagonal/internal/orders/httpapi"
	"github.com/exemple/hexagonal/internal/orders/postgres"
)

// La satisfaction implicite, PROUVÉE à la compilation : si OrderRepo cessait
// de satisfaire le port Repository, cette ligne ne compilerait pas.
var _ orders.Repository = (*postgres.OrderRepo)(nil)

// memRepo : le faux en mémoire du cours — quelques lignes écrites à la main,
// il satisfait orders.Repository. C'est LE bénéfice concret du port.
type memRepo map[string]orders.Order

func (m memRepo) ByID(_ context.Context, id string) (orders.Order, error) {
	o, ok := m[id]
	if !ok {
		return orders.Order{}, orders.ErrNotFound
	}
	return o, nil
}

func (m memRepo) Save(_ context.Context, o orders.Order) error {
	m[o.ID] = o
	return nil
}

func main() {
	ctx := context.Background()

	fmt.Println("=== Le domaine testé SANS base : le faux en mémoire ===")
	repo := memRepo{}
	svc := orders.NewService(repo) // le port accepte n'importe quel adaptateur

	o, err := svc.Place(ctx, "Alice", 8900)
	fmt.Println("Place(Alice, 8900) →", o.ID, "· err =", err)

	_, err = svc.Place(ctx, "Bob", 0)
	fmt.Println("Place(Bob, 0)      → règle métier :", err)

	_, err = repo.ByID(ctx, "absent")
	fmt.Println("ByID(absent)       → ErrNotFound :", errors.Is(err, orders.ErrNotFound))

	fmt.Println()
	fmt.Println("=== La chaîne complète : adaptateur HTTP → use case → dépôt ===")
	api := httpapi.New(svc)
	mux := http.NewServeMux()
	api.Register(mux)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, _ := http.Post(srv.URL+"/orders", "application/json",
		strings.NewReader(`{"customer":"Chloé","total":12900}`))
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Println("POST /orders (12900) →", resp.StatusCode, strings.TrimSpace(string(body)))
	fmt.Println("(total_cents : le tag JSON vit dans le DTO de l'adaptateur, pas dans l'entité ✔)")

	resp2, _ := http.Post(srv.URL+"/orders", "application/json",
		strings.NewReader(`{"customer":"Dan","total":0}`))
	resp2.Body.Close()
	fmt.Println("POST /orders (0)     →", resp2.StatusCode, "(la règle métier remonte en 422 ✔)")
}
