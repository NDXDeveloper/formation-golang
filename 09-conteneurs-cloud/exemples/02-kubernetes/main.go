/* ============================================================================
   Section 9.2 : Kubernetes : probes, configuration, graceful shutdown
   Description : Un service « prêt pour Kubernetes », auto-démonstratif SANS
                 cluster. Il prouve, dans l'ordre : (1) les SONDES — /healthz
                 léger et sans dépendances, /readyz qui Ping une vraie base
                 (SQLite) ; quand la base meurt, /readyz tombe à 503 mais
                 /healthz reste à 200 — retirer du trafic n'est pas
                 redémarrer ; (2) la CONFIGURATION par l'environnement
                 (APP_MESSAGE, comme l'injecterait un ConfigMap) ; (3) la
                 CONSCIENCE DES RESSOURCES (GOMAXPROCS Go 1.25, GOMEMLIMIT) ;
                 (4) l'ARRÊT PROPRE — le programme s'envoie SIGTERM (comme
                 le ferait Kubernetes) pendant qu'une requête longue est en
                 vol : Shutdown la laisse finir avant de rendre la main.
                 Le manifeste deploy.yaml ci-contre branche le tout.
   Fichier source : 02-kubernetes.md
   Lancer : go run .        (aucun cluster requis)
   ============================================================================ */

package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"syscall"
	"time"

	_ "modernc.org/sqlite" // driver pur Go (§ 7.2) — la « dépendance » du readyz
)

var db *sql.DB

func buildMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Liveness : le processus répond-il ? Léger, SANS dépendances —
	// si elle testait la base, une panne passagère ferait redémarrer
	// des pods parfaitement sains, en boucle.
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Readiness : peut-on servir ? C'est ICI qu'on teste les dépendances ;
	// un échec retire le pod du Service (plus de trafic), sans redémarrage.
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := db.PingContext(r.Context()); err != nil {
			http.Error(w, "base indisponible", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	// La configuration vient de l'ENVIRONNEMENT (12-factor) — en cluster,
	// un ConfigMap/Secret l'injecte (envFrom, cf. deploy.yaml).
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		msg, ok := os.LookupEnv("APP_MESSAGE")
		if !ok {
			msg = "(APP_MESSAGE non défini)"
		}
		fmt.Fprintln(w, msg)
	})

	// Une requête volontairement lente, pour prouver le drainage à l'arrêt.
	mux.HandleFunc("/lent", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1500 * time.Millisecond)
		fmt.Fprintln(w, "requête lente terminée proprement")
	})

	return mux
}

func main() {
	// ----- Conscience des ressources (l'apport Go 1.25 du chapitre) -----
	// En conteneur limité (docker run --cpus=2), GOMAXPROCS suit la LIMITE,
	// plus le nombre de cœurs de l'hôte ; GOMEMLIMIT se règle via l'env.
	fmt.Printf("NumCPU=%d GOMAXPROCS=%d GOMEMLIMIT=%d\n",
		runtime.NumCPU(), runtime.GOMAXPROCS(0), debug.SetMemoryLimit(-1))

	// ----- La « dépendance » du readyz : une vraie base -----
	var err error
	db, err = sql.Open("sqlite", "./ready.db")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove("./ready.db")

	// ----- Le contexte se ferme sur SIGINT/SIGTERM (Kubernetes envoie SIGTERM) -----
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{Addr: "127.0.0.1:8080", Handler: buildMux()}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("serveur : %v", err)
		}
	}()
	time.Sleep(200 * time.Millisecond) // le temps que le serveur écoute

	// ----- Démo 1 : les sondes, avant/après la mort de la dépendance -----
	get := func(p string) int {
		r, err := http.Get("http://127.0.0.1:8080" + p)
		if err != nil {
			return 0
		}
		r.Body.Close()
		return r.StatusCode
	}
	fmt.Println("== Sondes ==")
	fmt.Println("/healthz →", get("/healthz"), " /readyz →", get("/readyz"), "(tout va bien)")
	db.Close() // la dépendance « meurt »
	fmt.Println("base fermée : /readyz →", get("/readyz"), "(503 : retiré du trafic)",
		" /healthz →", get("/healthz"), "(200 : PAS de redémarrage)")

	// ----- Démo 2 : la configuration par l'environnement -----
	fmt.Println("== Config (env) ==")
	os.Setenv("APP_MESSAGE", "bonjour depuis l'environnement") // le ConfigMap, en local
	r, _ := http.Get("http://127.0.0.1:8080/")
	buf := make([]byte, 64)
	n, _ := r.Body.Read(buf)
	r.Body.Close()
	fmt.Print("GET / → ", string(buf[:n]))

	// ----- Démo 3 : l'arrêt propre, une requête longue EN VOL -----
	fmt.Println("== Arrêt propre ==")
	slow := make(chan string, 1)
	go func() { // la requête lente part…
		resp, err := http.Get("http://127.0.0.1:8080/lent")
		if err != nil {
			slow <- "COUPÉE : " + err.Error()
			return
		}
		b := make([]byte, 64)
		n, _ := resp.Body.Read(b)
		resp.Body.Close()
		slow <- string(b[:n])
	}()
	time.Sleep(100 * time.Millisecond)             // …puis l'orchestrateur décide d'arrêter :
	_ = syscall.Kill(os.Getpid(), syscall.SIGTERM) // Kubernetes enverrait ce SIGTERM
	<-ctx.Done()                                   // le signal est capté
	stop()
	fmt.Println("SIGTERM reçu — Shutdown draine les requêtes en vol…")

	// Délai borné, plus court que terminationGracePeriodSeconds (30 s).
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	start := time.Now()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("arrêt forcé : %v", err)
	}
	fmt.Printf("Shutdown rendu en %s · la requête lente : %s",
		time.Since(start).Round(100*time.Millisecond), <-slow)
	fmt.Println("(elle a été SERVIE, pas coupée — le drainage du .md)")
}
