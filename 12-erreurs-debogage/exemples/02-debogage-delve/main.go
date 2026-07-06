/* ============================================================================
   Section 12.2 : Débogage avec Delve (dans GoLand et VS Code)
   Description : Un petit programme conçu pour être EXPLORÉ avec Delve. Il se
                 lance normalement (go run .), mais son intérêt est de servir
                 de terrain d'entraînement au débogueur : une fonction avec
                 une variable locale à inspecter (somme), et deux goroutines
                 concurrentes pour exercer la vue des goroutines et les
                 watchpoints. Le README détaille les commandes dlv (break,
                 continue, print, locals, goroutines, watch) et le débogage
                 distant headless.
   Fichier source : 02-debogage-delve.md
   Lancer : go run .
   Déboguer : dlv debug . (voir le README pour les commandes)
   ============================================================================ */

package main

import (
	"fmt"
	"sync"
)

// somme : une cible de point d'arrêt. Posez un break sur la ligne « return »
// et inspectez `total` avec `print total` / `locals`.
func somme(n int) int {
	total := 0
	for i := 1; i <= n; i++ {
		total += i
	}
	return total // break ici : total vaut n(n+1)/2
}

// compteur partagé : cible d'un watchpoint (`watch -w main.partage`).
var partage int

func main() {
	fmt.Println("somme(10) =", somme(10)) // attendu : 55

	// Deux goroutines : de quoi exercer `goroutines` et `goroutine <id>`.
	var wg sync.WaitGroup
	var mu sync.Mutex
	for id := 1; id <= 2; id++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 3; j++ {
				mu.Lock()
				partage += id // watchpoint : « qui a modifié partage ? »
				mu.Unlock()
			}
		}(id)
	}
	wg.Wait()
	fmt.Println("partage =", partage) // 1×3 + 2×3 = 9
}
