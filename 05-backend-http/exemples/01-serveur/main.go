/* ============================================================================
   Section 5.1 : net/http — serveur, handlers, ServeMux
   Description : Serveur de PRODUCTION complet — routing Go 1.22 (méthode,
                 jokers {id} et {path...}, racine exacte {$}), http.Server
                 avec les quatre timeouts, et arrêt propre (graceful shutdown)
                 sur SIGINT/SIGTERM via signal.NotifyContext
   Fichier source : 01-net-http.md
   Lancer : go run .   puis   curl http://localhost:8080/items/42
   Arrêter : Ctrl-C → les requêtes en cours se terminent avant la sortie
   ============================================================================ */

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Le ServeMux à soi (jamais le DefaultServeMux global en production).
	mux := http.NewServeMux()

	// Depuis Go 1.22, le motif porte la MÉTHODE et des JOKERS.
	// Le mux répond seul 405 (avec l'en-tête Allow) et 404.
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "accueil — {$} ne matche QUE la racine /")
	})
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok")) // premier Write ⇒ 200 implicite
	})
	mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		// L'ordre : en-têtes → statut → corps.
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		// {id} capture UN segment, lu par PathValue ; ?verbose=1 via Query.
		fmt.Fprintf(w, "item %s (verbose=%s)\n", r.PathValue("id"), r.URL.Query().Get("verbose"))
	})
	mux.HandleFunc("GET /files/{path...}", func(w http.ResponseWriter, r *http.Request) {
		// {path...} en fin de motif capture TOUT le reste du chemin.
		fmt.Fprintf(w, "chemin capturé : %s\n", r.PathValue("path"))
	})

	// Un serveur de production a des TIMEOUTS (sinon : connexions lentes
	// qui monopolisent les ressources — attaque slowloris).
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second, // parade anti-slowloris
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// ctx sera annulé à la réception de SIGINT (Ctrl-C) ou SIGTERM (docker stop).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Le serveur écoute dans une goroutine : main reste libre d'attendre le signal.
	go func() {
		log.Println("écoute sur http://localhost:8080 — Ctrl-C pour un arrêt propre")
		// À l'arrêt propre, ListenAndServe renvoie ErrServerClosed :
		// ce n'est PAS une erreur — d'où le filtre errors.Is.
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-ctx.Done() // bloque jusqu'au signal d'arrêt

	// Shutdown cesse d'accepter de nouvelles connexions et LAISSE FINIR les
	// requêtes en vol — dans la limite du contexte (10 s ici).
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("arrêt forcé : %v", err)
		return
	}
	log.Println("arrêt propre : toutes les requêtes en vol sont terminées")
}
