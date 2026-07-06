/* ============================================================================
   Section 14.4 : Benchmarking rigoureux (benchstat)
   Description : Deux implémentations de Reverse — une lente (concaténation de
                 chaînes, quadratique) et une rapide (permutation de runes en
                 place) — pour illustrer le workflow benchstat : mesurer l'une,
                 puis l'autre, avec -count=10, et prouver STATISTIQUEMENT le
                 gain. Une exécution unique ne prouve rien ; seule la
                 comparaison répétée sur la même machine fait foi.
   Fichier source : 04-benchmarking.md
   ============================================================================ */

package strutil

// ReverseSlow : concaténation — réalloue la chaîne à chaque rune (quadratique).
func ReverseSlow(s string) string {
	out := ""
	for _, r := range s {
		out = string(r) + out
	}
	return out
}

// ReverseFast : permutation de runes en place — linéaire, deux allocations.
func ReverseFast(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
