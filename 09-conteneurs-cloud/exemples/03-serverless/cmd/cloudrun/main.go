/* ============================================================================
   Section 9.3 : Serverless (AWS Lambda, Cloud Run)
   Description : Le service Cloud Run de la section — AUCUN SDK : un serveur
                 net/http standard (module 5). L'unique contrainte de la
                 plateforme : écouter sur le port fourni par la variable
                 d'environnement PORT (8080 par défaut). À l'arrêt, Cloud Run
                 envoie SIGTERM — même arrêt propre qu'au § 9.2. Exécutable
                 en local : PORT=8493 go run ./cmd/cloudrun
   Fichier source : 03-serverless.md
   ============================================================================ */

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "ok cloud-run")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	port := os.Getenv("PORT") // Cloud Run fournit le port
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{Addr: ":" + port, Handler: mux}
	log.Println("écoute sur", srv.Addr)
	// À l'arrêt, Cloud Run envoie SIGTERM : même arrêt propre qu'au § 9.2.
	log.Fatal(srv.ListenAndServe())
}
