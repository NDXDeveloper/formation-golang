/* ============================================================================
   Section 5.2 : Middleware (chaînage, logging, recovery, CORS)
   Description : La chaîne complète du chapitre — Chain(recovery, requestID,
                 logging, cors) autour d'un mux — exercée SANS réseau via
                 httptest : logging slog structuré avec statut capturé,
                 panic transformé en 500 propre, préflight CORS (origine
                 admise vs refusée), X-Request-ID posé et lu du contexte
   Fichier source : 02-middleware.md
   ============================================================================ */

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime/debug"
	"time"
)

// La signature canonique : prend le handler suivant, en renvoie un nouveau.
type Middleware func(http.Handler) http.Handler

// Chain applique les middlewares : le PREMIER de la liste est le plus EXTERNE
// (donc exécuté en premier) — d'où la boucle à rebours.
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// L'interface ResponseWriter n'expose pas le statut envoyé : on l'ENVELOPPE
// pour l'intercepter au passage. (Attention : l'enveloppe perd les interfaces
// optionnelles comme http.Flusher — cf. la section.)
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK} // 200 par défaut
		start := time.Now()
		next.ServeHTTP(rec, r) // on passe l'ENVELOPPE au handler
		slog.Info("requête",
			"method", r.Method, "path", r.URL.Path,
			"status", rec.status, "sous_la_ms", time.Since(start) < time.Millisecond)
	})
}

// recovery se place LE PLUS À L'EXTÉRIEUR : une panique traverse les
// middlewares internes (leur code post-appel ne s'exécute pas) et c'est
// lui qui la transforme en 500 propre.
func recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic récupéré", "err", err, "stack_presente", len(debug.Stack()) > 0)
				http.Error(w, "erreur interne", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// CORS : liste blanche stricte (jamais « * » reflété), réponse au préflight
// OPTIONS sans appeler le handler.
func cors(allowed string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if origin := r.Header.Get("Origin"); origin == allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			}
			if r.Method == http.MethodOptions { // préflight : on s'arrête là
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// Le middleware est l'endroit où poser les VALEURS DE PORTÉE REQUÊTE :
// clé non exportée + valeur dans le contexte, transmis via r.WithContext.
type ctxKey struct{}

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := "req-0042" // en vrai : identifiant aléatoire/ULID
		ctx := context.WithValue(r.Context(), ctxKey{}, id)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func main() {
	// slog en JSON, sans horodatage (sortie reproductible pour la démo).
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{} // supprime "time"
			}
			return a
		},
	})))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /whoami", func(w http.ResponseWriter, r *http.Request) {
		// L'identifiant posé par requestID se lit depuis le contexte.
		fmt.Fprint(w, "vu comme ", r.Context().Value(ctxKey{}).(string))
	})
	mux.HandleFunc("GET /boom", func(w http.ResponseWriter, r *http.Request) {
		panic("boom") // recovery le transformera en 500
	})

	// L'ordre type de la section : recovery → request-ID → logging → CORS.
	handler := Chain(mux, recovery, requestID, logging, cors("https://app.example.com"))

	// Exercice sans réseau : httptest fabrique requêtes et enregistreurs.
	do := func(method, path, origin string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(method, path, nil)
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		handler.ServeHTTP(rec, req)
		return rec
	}

	fmt.Println("=== Requête normale : X-Request-ID + contexte + log structuré ===")
	r1 := do("GET", "/whoami", "")
	fmt.Println("corps :", r1.Body.String(), "· X-Request-ID :", r1.Header().Get("X-Request-ID"))

	fmt.Println("=== Panique → 500 propre (recovery est le plus externe) ===")
	r2 := do("GET", "/boom", "")
	fmt.Println("statut :", r2.Code, "· corps :", r2.Body.String())

	fmt.Println("=== Préflight CORS : origine admise vs refusée ===")
	r3 := do("OPTIONS", "/whoami", "https://app.example.com")
	fmt.Println("admise  :", r3.Code, "· ACAO :", r3.Header().Get("Access-Control-Allow-Origin"))
	r4 := do("OPTIONS", "/whoami", "https://evil.example.com")
	fmt.Println("refusée :", r4.Code, "· ACAO : «", r4.Header().Get("Access-Control-Allow-Origin"), "» (vide : pas d'autorisation)")
}
