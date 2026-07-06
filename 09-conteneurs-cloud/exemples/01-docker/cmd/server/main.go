/* ============================================================================
   Section 9.1 : Dockerfile multi-stage, images distroless / scratch
   Description : Le service HTTP minimal destiné à l'image DISTROLESS du
                 Dockerfile ci-contre — tout ce que l'image doit prouver :
                 la version injectée au build (ARG VERSION → -ldflags -X),
                 l'utilisateur non-root (uid 65532), les fuseaux horaires
                 (embarqués par distroless), et le point /healthz que les
                 sondes de Kubernetes interrogeront (§ 9.2) puisque
                 HEALTHCHECK est impossible sans shell.
   Fichier source : 01-docker.md
   Lancer : voir le README (docker build puis docker run)
   ============================================================================ */

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// version est injectée à la COMPILATION par le Dockerfile :
// ARG VERSION=dev → -ldflags "-X main.version=${VERSION}".
var version = "dev"

func main() {
	mux := http.NewServeMux()

	// Le point de santé : HEALTHCHECK n'existe pas en distroless/scratch
	// (pas de shell, pas de curl) — c'est le BINAIRE qui l'expose, et
	// Kubernetes qui l'interroge (§ 9.2).
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// La page d'accueil affiche ce que l'image doit prouver.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		loc, err := time.LoadLocation("Europe/Paris") // distroless embarque tzdata
		fmt.Fprintf(w, "version=%s uid=%d tz-paris-ok=%t\n",
			version, os.Getuid(), err == nil && loc != nil)
	})

	log.Println("écoute sur :8080 — version", version)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
