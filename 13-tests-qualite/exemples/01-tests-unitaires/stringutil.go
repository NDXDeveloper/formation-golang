/* ============================================================================
   Section 13.1 : Package testing, table-driven tests et sous-tests
   Description : Le code testé — Reverse raisonne en RUNES (pas en octets),
                 d'où le cas « caractère accentué » de la table. Somme et
                 Render servent de support aux autres démonstrations (helper
                 générique, golden file).
   Fichier source : 01-tests-unitaires.md
   ============================================================================ */

package stringutil

// Reverse renvoie s avec ses runes en ordre inverse.
func Reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

// Somme additionne deux entiers (support du helper assertEqual).
func Somme(a, b int) int { return a + b }

// Render produit une sortie comparée à un golden file.
func Render(in string) []byte { return []byte("<<" + in + ">>") }
