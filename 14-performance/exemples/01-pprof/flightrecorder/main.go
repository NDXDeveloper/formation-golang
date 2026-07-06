/* ============================================================================
   Section 14.1 : Traces d'exécution et flight recorder (Go 1.25)
   Description : Le flight recorder (§14.1) — un enregistreur de vol léger,
                 toujours actif, qui conserve en mémoire une FENÊTRE GLISSANTE
                 des derniers instants de trace, capturée À LA DEMANDE au
                 moment où une anomalie est détectée. On saisit ainsi ce qui a
                 PRÉCÉDÉ l'événement, sans le coût d'une trace continue.
                 L'instantané s'ouvre ensuite dans « go tool trace ».
   Fichier source : 01-pprof.md
   Lancer : go run ./flightrecorder   →   go tool trace trace.out
   ============================================================================ */

package main

import (
	"fmt"
	"log"
	"os"
	"runtime/trace"
	"time"
)

func main() {
	fr := trace.NewFlightRecorder(trace.FlightRecorderConfig{
		MinAge:   5 * time.Second, // ~2× la fenêtre de l'événement surveillé
		MaxBytes: 1 << 20,         // borne mémoire (1 MiB) — prioritaire sur MinAge
	})
	if err := fr.Start(); err != nil {
		log.Fatal(err)
	}
	defer fr.Stop()

	// … travail applicatif normal, tracé en continu dans le tampon circulaire.
	acc := 0
	for i := 0; i < 5_000_000; i++ {
		acc += i % 7
	}
	_ = acc

	// … à la détection d'une anomalie (latence hors budget, erreur) : capturer.
	f, _ := os.Create("trace.out")
	if _, err := fr.WriteTo(f); err != nil { // instantané de la fenêtre glissante
		log.Print(err)
	}
	_ = f.Close()
	fmt.Println("instantané écrit dans trace.out — ouvrir avec : go tool trace trace.out")
}
