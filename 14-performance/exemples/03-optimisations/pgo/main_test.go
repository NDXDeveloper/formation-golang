/* ============================================================================
   Section 14.3 : PGO — l'optimisation guidée par profil
   Description : Le benchmark qui produit le profil CPU représentatif. On le
                 capture dans default.pgo, on le place ici, et go build
                 l'utilise. En production, on collecte plutôt le profil réel
                 via net/http/pprof (§14.1) ; ici, le benchmark en tient lieu.
   Fichier source : 03-optimisations-pgo.md
   Générer le profil : go test -run=^$ -bench=. -cpuprofile=default.pgo
   ============================================================================ */

package main

import "testing"

func BenchmarkChaud(b *testing.B) {
	for b.Loop() {
		chaud(mod7{}, 1_000_000)
	}
}
