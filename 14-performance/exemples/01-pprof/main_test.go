/* ============================================================================
   Section 14.1 : Profilage avec pprof (CPU, heap, goroutines)
   Description : La voie la plus simple pour produire un profil (§14.1) : un
                 benchmark. « go test -bench=. -cpuprofile=cpu.out
                 -memprofile=mem.out » écrit les profils d'un chemin chaud
                 isolé, que l'on ouvre ensuite avec go tool pprof.
   Fichier source : 01-pprof.md
   Lancer : go test -run=^$ -bench=. -benchmem -cpuprofile=cpu.out -memprofile=mem.out
   ============================================================================ */

package main

import "testing"

func BenchmarkTravail(b *testing.B) {
	for b.Loop() {
		travail(20_000)
	}
}
