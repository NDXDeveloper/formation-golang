/* ============================================================================
   Section 8.1 : Consommer des API REST (http.Client, timeouts, retries,
                 résilience)
   Description : Le client robuste de la section, auto-démonstratif — les
                 « API » sont des serveurs httptest intégrés (aucun réseau
                 externe). Au menu : fetchUser (contexte, statut vérifié —
                 un 404 n'est PAS une erreur de Do —, décodage en flux),
                 doWithRetry (backoff exponentiel + gigue : succès à la 3e
                 tentative, temps mesuré ; un 404 n'est jamais ré-essayé),
                 circuit breaker gobreaker (s'OUVRE après 6 échecs : échec
                 rapide, le serveur mort n'est plus appelé), et un
                 http.RoundTripper personnalisé (le middleware côté client).
   Fichier source : 01-consommer-api.md
   Lancer : go run .        (sorties déterministes, temps de backoff affiché)
   ============================================================================ */

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"

	"github.com/sony/gobreaker/v2"
)

type User struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// fetchUser — la requête « correcte » de la section : contexte attaché,
// statut vérifié (Do n'échoue que sur le TRANSPORT), corps fermé, décodage
// en flux. (URL de base paramétrée pour viser nos serveurs de démo.)
func fetchUser(ctx context.Context, client *http.Client, base string, id int64) (User, error) {
	url := fmt.Sprintf("%s/users/%d", base, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil) // le contexte porte délai/annulation
	if err != nil {
		return User{}, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return User{}, fmt.Errorf("appel API : %w", err) // erreur de transport (réseau, timeout…)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK { // Do ne renvoie PAS d'erreur sur 4xx/5xx
		return User{}, fmt.Errorf("statut inattendu : %s", resp.Status)
	}

	var u User
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil { // décodage en flux
		return User{}, fmt.Errorf("décodage : %w", err)
	}
	return u, nil
}

// doWithRetry — le ré-essai de la section : uniquement le transitoire
// (transport, 5xx, 429), jamais les 4xx ; backoff exponentiel + gigue ;
// le contexte peut interrompre l'attente.
func doWithRetry(ctx context.Context, client *http.Client, req *http.Request, maxAttempts int) (*http.Response, error) {
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		resp, err := client.Do(req)
		switch {
		case err != nil:
			lastErr = err // transitoire (transport)
		case resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests:
			resp.Body.Close() // fermer avant de ré-essayer, sinon fuite
			lastErr = fmt.Errorf("statut %d", resp.StatusCode)
		default:
			return resp, nil // succès, ou 4xx : on ne ré-essaie pas
		}
		if attempt == maxAttempts {
			break
		}
		backoff := time.Duration(1<<uint(attempt-1)) * 100 * time.Millisecond
		jitter := time.Duration(rand.Int63n(int64(backoff)/2 + 1))
		select {
		case <-time.After(backoff + jitter):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return nil, lastErr
}

// countingTripper — un RoundTripper qui enveloppe le transport : le point
// d'extension idiomatique côté client (pendant des middlewares serveur).
type countingTripper struct {
	base  http.RoundTripper
	calls atomic.Int32
}

func (c *countingTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	c.calls.Add(1) // ici : compter ; en vrai : tracer, journaliser, ré-essayer…
	return c.base.RoundTrip(req)
}

func main() {
	ctx := context.Background()

	// L'« API » : un serveur httptest maîtrisé.
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/users/42":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"id":42,"name":"Alice"}`)
		default:
			http.Error(w, "introuvable", http.StatusNotFound)
		}
	}))
	defer api.Close()

	// UN client, partagé (jamais un par requête) — avec timeout.
	client := &http.Client{Timeout: 10 * time.Second}

	fmt.Println("=== fetchUser : nominal, puis 404 (une réponse, pas une erreur de Do) ===")
	u, err := fetchUser(ctx, client, api.URL, 42)
	fmt.Println("42  →", u.Name, "· err =", err)
	_, err = fetchUser(ctx, client, api.URL, 99)
	fmt.Println("99  →", err)

	fmt.Println("=== doWithRetry : 2 × 500 puis 200 — backoff mesuré ===")
	var hits atomic.Int32
	flaky := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits.Add(1) <= 2 {
			http.Error(w, "boom", http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, "ok")
	}))
	defer flaky.Close()
	req, _ := http.NewRequestWithContext(ctx, "GET", flaky.URL, nil)
	start := time.Now()
	resp, err := doWithRetry(ctx, client, req, 5)
	if err == nil {
		resp.Body.Close()
	}
	fmt.Printf("succès à la tentative %d · attente cumulée %s (backoff 100 ms puis 200 ms + gigue)\n",
		hits.Load(), time.Since(start).Round(10*time.Millisecond))

	fmt.Println("=== jamais de retry sur un 4xx ===")
	var h404 atomic.Int32
	srv404 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h404.Add(1)
		http.Error(w, "non", http.StatusNotFound)
	}))
	defer srv404.Close()
	req, _ = http.NewRequestWithContext(ctx, "GET", srv404.URL, nil)
	if r, _ := doWithRetry(ctx, client, req, 5); r != nil {
		r.Body.Close()
	}
	fmt.Println("tentatives sur 404 :", h404.Load(), "(une seule : l'erreur cliente ne se corrige pas d'elle-même)")

	fmt.Println("=== circuit breaker : échouer VITE quand l'aval est mort ===")
	var down atomic.Int32
	dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		down.Add(1)
		http.Error(w, "down", http.StatusInternalServerError)
	}))
	defer dead.Close()
	cb := gobreaker.NewCircuitBreaker[*http.Response](gobreaker.Settings{
		Name:        "users-api",
		MaxRequests: 3,                // requêtes autorisées en half-open
		Timeout:     10 * time.Second, // durée de l'état open avant de re-tester
		ReadyToTrip: func(c gobreaker.Counts) bool { return c.ConsecutiveFailures > 5 },
	})
	call := func() error {
		_, err := cb.Execute(func() (*http.Response, error) {
			r, err := client.Get(dead.URL)
			if err != nil {
				return nil, err
			}
			if r.StatusCode >= 500 {
				r.Body.Close()
				return nil, fmt.Errorf("statut %d", r.StatusCode)
			}
			return r, nil
		})
		return err
	}
	for range 6 {
		_ = call() // 6 échecs consécutifs → le circuit S'OUVRE
	}
	err = call() // 7e appel : refus IMMÉDIAT, sans toucher le serveur
	fmt.Println("7e appel  →", err, "· ErrOpenState :", errors.Is(err, gobreaker.ErrOpenState))
	fmt.Println("le serveur mort n'a reçu que", down.Load(), "appels (échec rapide ✔)")

	fmt.Println("=== RoundTripper : instrumenter sans changer les points d'appel ===")
	ct := &countingTripper{base: http.DefaultTransport}
	instrumented := &http.Client{Transport: ct, Timeout: 5 * time.Second}
	_, _ = fetchUser(ctx, instrumented, api.URL, 42)
	_, _ = fetchUser(ctx, instrumented, api.URL, 42)
	fmt.Println("requêtes vues par le RoundTripper :", ct.calls.Load())
}
