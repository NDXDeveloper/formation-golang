/* ============================================================================
   Section 5.5 : API REST complète (structure, versioning, Problem Details)
   Description : L'assemblage de la section — struct server à dépendances
                 injectées (zéro globale), routes() versionnées /v1, handlers
                 qui RENVOIENT une error, adaptateur handle(), writeError
                 centralisé au format Problem Details (RFC 9457), statuts REST
                 justes (200, 201+Location, 204, 404, 422)
   Fichier source : 05-api-rest-complete.md
   ============================================================================ */

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
)

// ---- Erreurs de domaine : traduites en statuts par writeError ----

var ErrNotFound = errors.New("introuvable")

type ValidationError struct{ msg string }

func (e *ValidationError) Error() string { return e.msg }

// ---- Le dépôt (interface : injectable, remplaçable par un mock) ----

type Item struct{ ID, Name string }

type Store interface {
	Item(ctx context.Context, id string) (Item, error)
}

type memStore map[string]Item

func (m memStore) Item(_ context.Context, id string) (Item, error) {
	it, ok := m[id]
	if !ok {
		// L'erreur est ENVELOPPÉE : writeError la retrouvera via errors.Is.
		return Item{}, fmt.Errorf("item %s : %w", id, ErrNotFound)
	}
	return it, nil
}

// ---- Le serveur : dépendances injectées, aucune variable globale ----

type server struct {
	store  Store
	logger *slog.Logger
}

// Les handlers RENVOIENT une error : plus de réponse d'erreur dupliquée.
type apiHandler func(w http.ResponseWriter, r *http.Request) error

// L'adaptateur central : un apiHandler devient un http.Handler standard.
func (s *server) handle(h apiHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			s.writeError(w, r, err)
		}
	})
}

// routes() centralise l'enregistrement — préfixe de VERSION /v1 dans les motifs.
func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /v1/items/{id}", s.handle(s.getItem))
	mux.Handle("POST /v1/items", s.handle(s.createItem))
	mux.Handle("DELETE /v1/items/{id}", s.handle(s.deleteItem))
	return mux // en production : enveloppé par recovery, logging… (§ 5.2)
}

func (s *server) getItem(w http.ResponseWriter, r *http.Request) error {
	item, err := s.store.Item(r.Context(), r.PathValue("id")) // ctx propagé
	if err != nil {
		return err // writeError décidera du statut
	}
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(item)
}

func (s *server) createItem(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Name == "" {
		return &ValidationError{msg: "name requis"} // → 422 via errors.As
	}
	w.Header().Set("Location", "/v1/items/7") // 201 pointe la ressource créée
	w.WriteHeader(http.StatusCreated)
	return nil
}

func (s *server) deleteItem(w http.ResponseWriter, r *http.Request) error {
	w.WriteHeader(http.StatusNoContent) // 204 : succès sans corps
	return nil
}

// ---- Problem Details (RFC 9457) : le format d'erreur UNIQUE de l'API ----

type Problem struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail,omitempty"`
}

func (s *server) writeError(w http.ResponseWriter, r *http.Request, err error) {
	status, detail := http.StatusInternalServerError, "erreur interne"

	var ve *ValidationError
	switch {
	case errors.Is(err, ErrNotFound): // sentinelle, même enveloppée
		status, detail = http.StatusNotFound, "ressource introuvable"
	case errors.As(err, &ve): // erreur typée : on extrait son message
		status, detail = http.StatusUnprocessableEntity, ve.Error()
	}

	if status >= 500 {
		// Le détail COMPLET reste côté serveur ; le client n'en voit rien.
		s.logger.Error("requête échouée", "err", err, "path", r.URL.Path)
	}

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Problem{
		Type: "about:blank", Title: http.StatusText(status), Status: status, Detail: detail,
	})
}

func main() {
	s := &server{
		store:  memStore{"42": {ID: "42", Name: "Café"}},
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	api := s.routes()

	do := func(method, path, body string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		api.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
		return rec
	}

	fmt.Println("=== Chemin nominal ===")
	r1 := do("GET", "/v1/items/42", "")
	fmt.Println("GET 42     →", r1.Code, strings.TrimSpace(r1.Body.String()))

	fmt.Println("=== Erreurs au format Problem Details (RFC 9457) ===")
	r2 := do("GET", "/v1/items/99", "")
	fmt.Println("GET 99     →", r2.Code, "·", r2.Header().Get("Content-Type"))
	fmt.Println("            ", strings.TrimSpace(r2.Body.String()))
	r3 := do("POST", "/v1/items", `{}`)
	fmt.Println("POST vide  →", r3.Code, strings.TrimSpace(r3.Body.String()))

	fmt.Println("=== Statuts REST justes ===")
	r4 := do("POST", "/v1/items", `{"name":"Thé"}`)
	fmt.Println("POST ok    →", r4.Code, "· Location :", r4.Header().Get("Location"))
	r5 := do("DELETE", "/v1/items/42", "")
	fmt.Println("DELETE     →", r5.Code, "· corps vide :", r5.Body.Len() == 0)
}
