/* ============================================================================
   Section 14.4 : Benchmarking rigoureux (benchstat)
   Description : Le benchmark de référence. On le lance DEUX fois avec
                 -count=10 (une par implémentation) en redirigeant vers
                 old.txt puis new.txt, puis « benchstat old.txt new.txt »
                 rend un verdict statistique (delta, p-value, variation). On
                 s'appuie sur « for b.Loop() » (Go 1.24), qui gère le
                 chronomètre et l'élimination de code mort. Ne jamais mesurer
                 sous -race ni -cover (l'instrumentation fausse tout).
   Fichier source : 04-benchmarking.md
   Workflow :
     1. go test -run=^$ -bench=Reverse -benchmem -count=10 > old.txt   (sur ReverseSlow)
     2. basculer le benchmark sur ReverseFast
     3. go test -run=^$ -bench=Reverse -benchmem -count=10 > new.txt
     4. benchstat old.txt new.txt
   ============================================================================ */

package strutil

import (
	"strings"
	"testing"
)

// Entrée réaliste (runes multi-octets), préparée une fois hors du chrono.
var input = strings.Repeat("café ", 200)

// Basculer l'appel entre ReverseSlow et ReverseFast pour produire old/new.
func BenchmarkReverse(b *testing.B) {
	for b.Loop() {
		ReverseFast(input)
	}
}
