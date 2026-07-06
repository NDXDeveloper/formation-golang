/* ============================================================================
   Section 17.2 §4.2 : Du « Go de 2020 » — APIs et motifs périmés
   Description : Un fichier écrit avec les idiomes que l'IA reproduit depuis son
                 corpus d'entraînement (Go d'il y a 3-5 ans). Il COMPILE, mais
                 « go fix ./... » (Go 1.26) le modernise mécaniquement, de façon
                 SÛRE et consciente de la ligne `go` du go.mod. Lancez
                 « go fix -diff ./... » pour prévisualiser les six réécritures
                 du tableau §4.2, puis « go fix ./... » pour les appliquer.
   Fichier source : 17.2 (02-pieges-ia.md)
   Essayer : go fix -diff ./...
   ============================================================================ */

package modernize

import "io/ioutil"

func vague(v interface{}) any { return v } // → any (Go 1.18)

func lire(path string) ([]byte, error) { // ioutil → os.ReadFile (Go 1.16)
	return ioutil.ReadFile(path)
}

func plusGrand(a, b int) int { // → max intégré (Go 1.21)
	x := a
	if b > x {
		x = b
	}
	return x
}

func copierMap(src map[string]int) map[string]int { // → maps.Copy (Go 1.21)
	dst := make(map[string]int)
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func concat(parts []string) string { // → strings.Builder
	s := ""
	for _, p := range parts {
		s += p
	}
	return s
}

func somme(n int) int { // → for i := range n (Go 1.22)
	total := 0
	for i := 0; i < n; i++ {
		total += i
	}
	return total
}
