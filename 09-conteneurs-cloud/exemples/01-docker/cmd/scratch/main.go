/* ============================================================================
   Section 9.1 : Dockerfile multi-stage, images distroless / scratch
   Description : Le binaire destiné à l'image SCRATCH (Dockerfile.scratch) —
                 il vérifie lui-même les trois éléments que « Ce qu'il faut
                 penser à embarquer » impose de fournir explicitement :
                 1. les certificats CA (un appel HTTPS sortant réel) ;
                 2. les fuseaux horaires (incorporés au binaire par l'import
                    blanc `_ "time/tzdata"`, faute de quoi seul UTC existe) ;
                 3. l'utilisateur non-root (USER 65532:65532 du Dockerfile).
   Fichier source : 01-docker.md
   Lancer : voir le README (docker build -f Dockerfile.scratch puis run)
   ============================================================================ */

package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
	_ "time/tzdata" // l'idiome scratch : la base des fuseaux DANS le binaire
)

func main() {
	// 1. HTTPS sortant : ne fonctionne QUE si les certificats CA sont
	//    présents dans l'image (COPY --from=builder ...ca-certificates.crt).
	status := "échec"
	resp, err := http.Get("https://proxy.golang.org/")
	if err == nil {
		status = resp.Status
		resp.Body.Close()
	}

	// 2. Fuseaux : sans time/tzdata (ni tzdata système), LoadLocation échoue.
	loc, errTz := time.LoadLocation("Europe/Paris")

	// 3. Non-root : le USER du Dockerfile.
	fmt.Printf("uid=%d https-sortant=%s (err=%v) tz-paris=%t\n",
		os.Getuid(), status, err, errTz == nil && loc != nil)
}
