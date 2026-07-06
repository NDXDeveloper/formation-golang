/* ============================================================================
   Section 11.3 : Migrer un service Python / Java / Node vers Go : stratégies
   Description : Les deux piliers de la section, auto-démonstratifs.
                 (1) La façade STRANGLER FIG du cours — un ServeMux et un
                 NewSingleHostReverseProxy de la stdlib : la tranche migrée
                 (/api/orders/) est servie par Go, tout le reste part vers
                 « l'existant » (le service Python/Java/Node hérité, simulé
                 ici par httptest) ; on retire l'ancien chemin tranche par
                 tranche, jamais d'un bloc. (2) Le piège n°1 — la
                 TRANSLITTÉRATION : émuler les exceptions avec panic
                 (« Java/Python-en-Go », le service meurt) face au Go
                 idiomatique (l'erreur est une valeur retournée).
   Fichier source : 03-migrer-vers-go.md
   Lancer : go run .        (aucun service requis)
   ============================================================================ */

package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
)

// newOrdersHandler : la tranche MIGRÉE — désormais servie par Go.
func newOrdersHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "commandes servies par GO (tranche migrée)")
	})
}

// ---- Le piège n°1 : translittération vs Go idiomatique ----

type User struct{ Name string }

type storeT struct{}

func (storeT) find(id string) (User, error) {
	if id == "" {
		return User{}, errors.New("utilisateur introuvable")
	}
	return User{Name: "Alice"}, nil
}

var store = storeT{}

// « Java/Python-en-Go » : émuler les exceptions avec panic/recover — à fuir.
func getUserBad(id string) User {
	u, err := store.find(id)
	if err != nil {
		panic(err) // ce n'est pas ainsi qu'on gère une erreur en Go
	}
	return u
}

// Go idiomatique : l'erreur est une valeur, retournée explicitement.
func getUser(id string) (User, error) {
	return store.find(id)
}

func main() {
	fmt.Println("=== 1. La façade strangler fig ===")
	// « L'existant » : le système hérité (Python/Java/Node) — simulé.
	legacySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "servi par L'EXISTANT :", r.URL.Path)
	}))
	defer legacySrv.Close()

	// La façade du cours : route la tranche migrée vers le service Go,
	// le reste vers l'existant — avec la seule bibliothèque standard.
	legacy, _ := url.Parse(legacySrv.URL)
	legacyProxy := httputil.NewSingleHostReverseProxy(legacy)

	mux := http.NewServeMux()
	mux.Handle("/api/orders/", newOrdersHandler()) // migré : servi par Go
	mux.Handle("/", legacyProxy)                   // le reste : encore l'existant

	facade := httptest.NewServer(mux)
	defer facade.Close()

	get := func(p string) string {
		resp, err := http.Get(facade.URL + p)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return string(b)
	}
	fmt.Print("/api/orders/42 → ", get("/api/orders/42"))
	fmt.Print("/api/users/7   → ", get("/api/users/7"))
	fmt.Println("(on monte tranche par tranche, l'ancien chemin se retire à la fin)")

	fmt.Println()
	fmt.Println("=== 2. Le piège n°1 : translittération vs idiomatique ===")
	u, err := getUser("42")
	fmt.Println("idiomatique      :", u.Name, "· err =", err)

	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("translittération : panic →", r, "(le service meurt — à fuir)")
			}
		}()
		_ = getUserBad("") // le « Java-en-Go » : l'erreur devient un crash
	}()
}
