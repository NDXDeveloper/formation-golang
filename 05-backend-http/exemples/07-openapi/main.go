/* ============================================================================
   Section 5.7 : Documentation OpenAPI / Swagger (spec-first)
   Description : La chaîne spec-first du chapitre, complète — openapi.yaml est
                 la source de vérité ; oapi-codegen (directive go:generate
                 ci-dessous) produit api.gen.go : types + ServerInterface +
                 HandlerFromMux sur le ServeMux de Go 1.22 (cible std-http).
                 On implémente l'interface : tout écart avec la spec devient
                 une ERREUR DE COMPILATION. La spec est aussi servie à
                 GET /openapi.yaml, et une 404 sort en Problem Details.
   Fichier source : 07-openapi.md
   Régénérer : go generate ./...   (api.gen.go est commité : build sans outil)
   ============================================================================ */

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest -generate types,std-http -package main -o api.gen.go openapi.yaml

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
)

// api implémente la ServerInterface GÉNÉRÉE depuis la spec : si la spec
// change (paramètre, route), ce type ne compile plus — zéro dérive possible.
type api struct{}

// GetItem répond à « GET /v1/items/{id} » : la signature (id string) vient
// de la spec — c'est oapi-codegen qui l'a écrite, pas nous.
func (api) GetItem(w http.ResponseWriter, r *http.Request, id string) {
	if id != "42" {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusNotFound)
		title, status := "Not Found", 404
		_ = json.NewEncoder(w).Encode(Problem{Title: &title, Status: &status})
		return
	}
	name := "Café"
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(Item{Id: &id, Name: &name})
}

func main() {
	mux := http.NewServeMux()

	// Servir la spec elle-même : le contrat est consultable à une URL
	// (on y pointerait Swagger UI, Redoc ou Scalar).
	mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "openapi.yaml")
	})

	// HandlerFromMux (généré) enregistre les routes de la spec sur NOTRE mux.
	h := HandlerFromMux(api{}, mux)

	srv := httptest.NewServer(h)
	defer srv.Close()

	get := func(path string) (int, string, string) {
		resp, err := http.Get(srv.URL + path)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, resp.Header.Get("Content-Type"), strings.TrimSpace(string(b))
	}

	fmt.Println("=== Le serveur généré depuis la spec ===")
	c1, ct1, b1 := get("/v1/items/42")
	fmt.Println("GET /v1/items/42 →", c1, ct1, b1)
	c2, ct2, b2 := get("/v1/items/99")
	fmt.Println("GET /v1/items/99 →", c2, ct2, b2)

	fmt.Println("=== La spec servie (contrat consultable) ===")
	c3, _, b3 := get("/openapi.yaml")
	fmt.Println("GET /openapi.yaml →", c3, "· source de vérité :", strings.Contains(b3, "openapi: 3.0.3"))
}
