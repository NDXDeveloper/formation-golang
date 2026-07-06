/* ============================================================================
   Section 13.5 : Linters — go vet, staticcheck, golangci-lint, go fix
   Description : Du code VOLONTAIREMENT écrit avec des idiomes datés (le « Go
                 de 2020 » que produit souvent l'IA) — interface{} au lieu de
                 any, un if/else réimplémentant max, une boucle manuelle de
                 copie de map. Il COMPILE, mais « go fix ./... » (Go 1.26) le
                 modernise mécaniquement. Lancer « go fix -diff ./... » pour
                 voir les réécritures avant de les appliquer.
   Fichier source : 05-linters.md
   Essayer : go fix -diff ./...   puis   go fix ./...
   ============================================================================ */

package linters

import "fmt"

func plusGrand(a, b int) int { // → go fix : remplacé par le builtin max
	if a > b {
		return a
	}
	return b
}

func copierMap(src map[string]int) map[string]int {
	dst := make(map[string]int)
	for k, v := range src { // → go fix : maps.Copy(dst, src)
		dst[k] = v
	}
	return dst
}

func Demo() {
	var v interface{} = 42 // → go fix : any
	fmt.Println(v, plusGrand(3, 7), copierMap(map[string]int{"a": 1}))
}
