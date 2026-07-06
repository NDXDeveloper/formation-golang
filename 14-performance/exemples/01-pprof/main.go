/* ============================================================================
   Section 14.1 : Profilage avec pprof (CPU, heap, goroutines)
   Description : Un service volontairement gourmand, prêt à être profilé de
                 trois façons. (1) net/http/pprof expose /debug/pprof/ sur un
                 écouteur INTERNE dédié (localhost:6060) — jamais public.
                 (2) runtime/pprof écrit un profil CPU + un profil de tas au
                 démarrage. Le handler /travail mêle calcul (point chaud CPU)
                 et allocations (pression mémoire) pour que les profils aient
                 quelque chose à montrer. Voir aussi le benchmark
                 (main_test.go) pour la voie « -cpuprofile », et le
                 sous-programme flightrecorder/ pour la trace Go 1.25.
   Fichier source : 01-pprof.md
   Lancer : go run .   puis   go tool pprof http://localhost:6060/debug/pprof/profile?seconds=5
   ============================================================================ */

package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // effet de bord : enregistre /debug/pprof/ sur DefaultServeMux
	"os"
	"runtime/pprof"
)

// travail : un point chaud — calcul (CPU) + allocations qui S'ÉCHAPPENT (tas).
// Chaque buf est rangé dans le slice retourné, donc il ne peut pas rester sur
// la pile : le profil de tas a de quoi montrer.
func travail(n int) [][]byte {
	out := make([][]byte, 0, n)
	for i := 0; i < n; i++ {
		buf := make([]byte, 256)
		for j := range buf {
			buf[j] = byte((i + j) % 251)
		}
		out = append(out, buf) // buf s'échappe → allocation tas
	}
	return out
}

func main() {
	// (2) runtime/pprof : profiler une portion précise, dans le code.
	cpu, _ := os.Create("cpu.prof")
	_ = pprof.StartCPUProfile(cpu)
	for i := 0; i < 50; i++ {
		travail(20_000)
	}
	pprof.StopCPUProfile()
	_ = cpu.Close()

	heap, _ := os.Create("heap.prof")
	_ = pprof.WriteHeapProfile(heap) // instantané du tas
	_ = heap.Close()
	fmt.Println("profils écrits : cpu.prof, heap.prof (go tool pprof cpu.prof)")

	// (1) net/http/pprof sur un écouteur INTERNE — jamais exposé publiquement.
	http.HandleFunc("/travail", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%d tampons alloués\n", len(travail(50_000)))
	})
	fmt.Println("service sur http://localhost:6060 — /debug/pprof/ et /travail")
	fmt.Println("Ctrl-C pour arrêter.")
	log.Fatal(http.ListenAndServe("localhost:6060", nil))
}
