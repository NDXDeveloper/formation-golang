/* ============================================================================
   Section 14.3 : Optimisations — sync.Pool, préallocation, PGO
   Description : Le banc qui PROUVE chaque optimisation via -benchmem : on
                 compare la colonne allocs/op de la version lente et de la
                 version rapide. Résultats constatés : préallocation 11 → 1,
                 strings.Builder 999 → 1, sync.Pool 1 → 0 allocation par
                 opération. « Vérifier, toujours » : une optimisation ne se
                 déclare victorieuse qu'avec la mesure.
   Fichier source : 03-optimisations-pgo.md
   Lancer : go test -run=^$ -bench=. -benchmem
   ============================================================================ */

package opt

import (
	"strings"
	"testing"
)

var (
	users = make([]User, 1000)
	parts = make([]string, 1000)
	msgs  = make([]string, 1000)
)

func init() {
	for i := range parts {
		parts[i] = "abc"
		msgs[i] = "message"
		users[i] = User{Name: strings.Repeat("x", 8)}
	}
}

func BenchmarkNamesSlow(b *testing.B) {
	for b.Loop() {
		NamesSlow(users)
	}
}
func BenchmarkNamesFast(b *testing.B) {
	for b.Loop() {
		NamesFast(users)
	}
}
func BenchmarkJoinSlow(b *testing.B) {
	for b.Loop() {
		JoinSlow(parts)
	}
}
func BenchmarkJoinFast(b *testing.B) {
	for b.Loop() {
		JoinFast(parts)
	}
}
func BenchmarkFormatSansPool(b *testing.B) {
	for b.Loop() {
		FormatSansPool(msgs)
	}
}
func BenchmarkFormatAvecPool(b *testing.B) {
	for b.Loop() {
		FormatAvecPool(msgs)
	}
}
