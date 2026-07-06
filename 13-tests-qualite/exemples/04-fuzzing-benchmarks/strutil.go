/* ============================================================================
   Section 13.4 : Fuzzing natif et benchmarks
   Description : Reverse correct (raisonne en runes) — cible du fuzzing par
                 PROPRIÉTÉS (aller-retour + sortie UTF-8 valide) et du
                 benchmark. Un Reverse naïf par octets (dans le _test) sert
                 à MONTRER le bug que le fuzzing débusque.
   Fichier source : 04-fuzzing-benchmarks.md
   ============================================================================ */

package strutil

// Reverse renvoie s avec ses runes en ordre inverse (UTF-8 préservé).
func Reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
