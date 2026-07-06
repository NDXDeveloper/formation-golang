/* ============================================================================
   Section 5.4 : Frameworks — stdlib vs Gin / Echo / Chi
   Description : Le bloc Chi de la section, en action — groupes de routes,
                 suite de middlewares fournie, et LA preuve du positionnement :
                 un chi.Router EST un http.Handler standard (on le sert tel
                 quel), ses handlers sont 100 % net/http. (Gin et Echo, à
                 contexte propre, restent illustrés dans le .md.)
   Fichier source : 04-frameworks.md
   Prérequis : réseau au premier build (module go-chi/chi — zéro dépendance)
   ============================================================================ */

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type User struct {
	ID string `json:"id"`
}

func main() {
	r := chi.NewRouter()

	// Chi fournit une suite de middlewares prêts (ici : RequestID) — de
	// simples func(http.Handler) http.Handler, comme ceux du § 5.2.
	r.Use(middleware.RequestID)

	// Groupe de routes : préfixe partagé /users pour toutes les sous-routes.
	r.Route("/users", func(r chi.Router) {
		r.Get("/{id}", func(w http.ResponseWriter, req *http.Request) {
			id := chi.URLParam(req, "id") // l'équivalent chi de r.PathValue
			// L'ID posé par middleware.RequestID se lit dans le CONTEXTE —
			// même mécanique qu'au § 5.2 ; on le renvoie ici en en-tête.
			w.Header().Set("X-Request-ID", middleware.GetReqID(req.Context()))
			// Le handler manipule les types STANDARD : w et req, rien d'autre.
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(User{ID: id})
		})
	})

	// LA démonstration du « pas de verrouillage » : chi.Router satisfait
	// http.Handler — httptest.NewServer (ou http.Server) le sert TEL QUEL.
	srv := httptest.NewServer(r)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/users/42")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Print("GET /users/42 → ", resp.StatusCode, " ", string(body))
	fmt.Println("request-ID du middleware chi (lu du contexte) :", resp.Header.Get("X-Request-ID") != "")
	fmt.Println("chi.Router est bien un http.Handler standard ✔")
}
