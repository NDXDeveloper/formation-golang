/* ============================================================================
   Section 14.3 : Optimisations — sync.Pool, préallocation, PGO
   Description : Les trois techniques « allouer moins » de la section, chacune
                 avec sa version lente et sa version optimisée, à comparer au
                 banc (opt_test.go). PRÉALLOCATION : réserver la capacité
                 connue supprime les réallocations d'append. STRINGS.BUILDER :
                 construire une chaîne sans la concaténation quadratique.
                 SYNC.POOL : recycler un tampon coûteux pour soulager le GC
                 (avec Reset impératif avant réutilisation). Chaque technique
                 ne vaut que sur un chemin chaud IDENTIFIÉ par le profil.
   Fichier source : 03-optimisations-pgo.md
   Comparer : go test -run=^$ -bench=. -benchmem
   ============================================================================ */

package opt

import (
	"bytes"
	"strings"
	"sync"
)

type User struct{ Name string }

// --- Préallocation ---

func NamesSlow(users []User) []string {
	var out []string // réalloué à mesure qu'append dépasse la capacité
	for _, u := range users {
		out = append(out, u.Name)
	}
	return out
}

func NamesFast(users []User) []string {
	out := make([]string, 0, len(users)) // capacité connue : une seule allocation
	for _, u := range users {
		out = append(out, u.Name)
	}
	return out
}

// --- strings.Builder ---

func JoinSlow(parts []string) string {
	s := "" // chaque += recopie toute la chaîne (quadratique)
	for _, p := range parts {
		s += p
	}
	return s
}

func JoinFast(parts []string) string {
	n := 0
	for _, p := range parts {
		n += len(p)
	}
	var b strings.Builder
	b.Grow(n) // préalloue le tampon exact
	for _, p := range parts {
		b.WriteString(p)
	}
	return b.String()
}

// --- sync.Pool ---

var bufPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}

func FormatSansPool(msgs []string) int {
	total := 0
	for _, m := range msgs {
		var b bytes.Buffer // alloue un nouveau buffer à chaque tour
		b.WriteString(m)
		total += b.Len()
	}
	return total
}

func FormatAvecPool(msgs []string) int {
	total := 0
	for _, m := range msgs {
		b := bufPool.Get().(*bytes.Buffer)
		b.Reset() // impératif : l'objet réutilisé porte un état ancien
		b.WriteString(m)
		total += b.Len()
		bufPool.Put(b)
	}
	return total
}
