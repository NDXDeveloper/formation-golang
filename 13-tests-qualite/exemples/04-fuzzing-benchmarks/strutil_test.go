/* ============================================================================
   Section 13.4 : Fuzzing natif et benchmarks
   Description : Le FUZZING par propriétés (f.Add pour les graines, f.Fuzz pour
                 la cible) — on affirme deux invariants : la sortie reste de
                 l'UTF-8 valide, et l'aller-retour Reverse(Reverse(s)) == s.
                 Et le BENCHMARK au style moderne « for b.Loop() » (Go 1.24+),
                 qui gère le chronomètre et empêche l'élimination de code mort.
                 Aucune dépendance : le fuzzing et les benchmarks sont natifs.
   Fichier source : 04-fuzzing-benchmarks.md
   Lancer : go test ./...                            (graines seules, rapide)
            go test -run=^$ -fuzz=FuzzReverse -fuzztime=20s   (campagne active)
            go test -bench=. -benchmem -run=^$        (benchmark)
   ============================================================================ */

package strutil

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// FuzzReverse : le moteur mute les entrées et cherche à violer les invariants.
func FuzzReverse(f *testing.F) {
	f.Add("Golang") // graines : exemples + cas de non-régression
	f.Add("café")
	f.Add("")

	f.Fuzz(func(t *testing.T, s string) {
		if !utf8.ValidString(s) {
			t.Skip() // Reverse suppose une entrée UTF-8 valide
		}
		rev := Reverse(s)
		if !utf8.ValidString(rev) {
			t.Errorf("Reverse(%q) a produit un UTF-8 invalide : %q", s, rev)
		}
		if got := Reverse(rev); got != s {
			t.Errorf("Reverse(Reverse(%q)) = %q, want %q", s, got, s)
		}
	})
}

// BenchmarkReverse : style moderne b.Loop (Go 1.24+) — la préparation est
// hors chrono, les opérandes restent vivantes (pas d'élimination de code mort).
func BenchmarkReverse(b *testing.B) {
	s := strings.Repeat("café ", 1000) // préparé UNE fois, hors de la boucle
	for b.Loop() {
		Reverse(s)
	}
}

// Sous-benchmarks par taille (comme une table de tests).
func BenchmarkReverseTailles(b *testing.B) {
	for _, taille := range []int{10, 1000} {
		b.Run("n="+itoa(taille), func(b *testing.B) {
			s := strings.Repeat("é", taille)
			for b.Loop() {
				Reverse(s)
			}
		})
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var d []byte
	for n > 0 {
		d = append([]byte{byte('0' + n%10)}, d...)
		n /= 10
	}
	return string(d)
}
