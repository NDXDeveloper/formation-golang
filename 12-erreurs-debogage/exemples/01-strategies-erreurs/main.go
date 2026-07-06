/* ============================================================================
   Section 12.1 : Stratégies d'erreurs à l'échelle d'une application
   Description : Les idiomes de la section, exécutés de bout en bout :
                 (1) errors.Join agrège plusieurs erreurs de validation, et
                 errors.As TRAVERSE le résultat joint ; (2) l'enrobage %w
                 ajoute du contexte tout en préservant l'erreur sous-jacente
                 (errors.Is traverse la chaîne) ; (3) la DISCIPLINE CARDINALE
                 à la frontière : httpError inspecte l'erreur, la journalise
                 UNE SEULE FOIS via slog, et la convertit en réponse propre
                 (Problem Details, RFC 9457) — sans JAMAIS fuiter les
                 détails internes au client.
   Fichier source : 01-strategies-erreurs.md
   Lancer : go run .        (aucun service requis)
   ============================================================================ */

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/exemple/strategieserreurs/store"
)

// problem : le corps d'une réponse Problem Details (RFC 9457).
type problem struct {
	Title  string `json:"title"`
	Status int    `json:"status"`
}

type User struct {
	Name string
	Age  int
}

// Validate rassemble PLUSIEURS erreurs avec errors.Join (nil si tout va bien).
func (u User) Validate() error {
	var errs []error
	if u.Name == "" {
		errs = append(errs, &store.ValidationError{Field: "name", Msg: "requis"})
	}
	if u.Age < 0 {
		errs = append(errs, &store.ValidationError{Field: "age", Msg: "négatif"})
	}
	return errors.Join(errs...) // errors.Is/As traversent le résultat
}

// loadConfig illustre l'enrobage %w en remontant (contexte + erreur préservée).
func loadConfig(path string) error {
	if _, err := os.Open(path); err != nil {
		return fmt.Errorf("ouverture de la config %q : %w", path, err)
	}
	return nil
}

// httpError : la frontière. Inspecte, journalise UNE fois, écrit une réponse
// propre — le client ne voit jamais le détail interne.
func httpError(w http.ResponseWriter, r *http.Request, err error, log *slog.Logger) {
	status, title := http.StatusInternalServerError, "erreur interne"

	var verr *store.ValidationError
	switch {
	case errors.Is(err, store.ErrNotFound): // identité → 404
		status, title = http.StatusNotFound, "introuvable"
	case errors.As(err, &verr): // extraction de données → 400
		status, title = http.StatusBadRequest, "requête invalide"
	}

	// détail journalisé EN INTERNE, une seule fois, avec tout le contexte…
	log.ErrorContext(r.Context(), "requête échouée",
		slog.String("path", r.URL.Path), slog.Any("err", err))

	// …et au client, seulement une réponse propre.
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(problem{Title: title, Status: status})
}

func main() {
	fmt.Println("=== Taxonomie : Join agrégé, As/Is qui traversent ===")
	err := User{Name: "", Age: -3}.Validate()
	fmt.Println("Join     :", strings.ReplaceAll(err.Error(), "\n", " | "))
	var verr *store.ValidationError
	fmt.Println("As(Join) :", errors.As(err, &verr), "→ premier champ fautif :", verr.Field)

	err2 := loadConfig("absente.yaml")
	fmt.Println("enrobé %w        :", err2)
	fmt.Println("Is(os.ErrNotExist):", errors.Is(err2, os.ErrNotExist), "(la chaîne est préservée)")

	fmt.Println()
	fmt.Println("=== La frontière : httpError journalise UNE fois + Problem Details ===")
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{} // sortie déterministe pour l'exemple
			}
			return a
		}}))

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/absent":
			httpError(w, r, fmt.Errorf("lecture : %w", store.ErrNotFound), log)
		case "/invalide":
			httpError(w, r, User{Age: -1}.Validate(), log)
		default:
			httpError(w, r, errors.New("panne interne secrète"), log)
		}
	})
	srv := httptest.NewServer(h)
	defer srv.Close()

	for _, p := range []string{"/absent", "/invalide", "/panne"} {
		resp, _ := http.Get(srv.URL + p)
		buf := make([]byte, 128)
		n, _ := resp.Body.Read(buf)
		resp.Body.Close()
		fmt.Printf("%-10s → %d  %s\n", p, resp.StatusCode, strings.TrimSpace(string(buf[:n])))
	}
	fmt.Println("(le client ne voit JAMAIS « panne interne secrète » — aucune fuite d'interne ✔)")
}
