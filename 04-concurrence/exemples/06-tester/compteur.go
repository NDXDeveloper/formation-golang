/* ============================================================================
   Section 4.6 : Tester le code concurrent (-race et testing/synctest)
   Description : Le compteur SYNCHRONISÉ (atomic) — la version corrigée du
                 TestCounter de la section : `go test -race ./...` passe.
                 Pour reproduire la course détectée par -race, remplacez
                 l'atomic par « n++ » sur un int partagé et relancez avec -race.
   Fichier source : 06-tester-code-concurrent.md
   ============================================================================ */

package tester

import (
	"sync"
	"sync/atomic"
)

// Incremente lance n goroutines qui incrémentent un compteur partagé —
// correctement synchronisé (atomic), donc -race ne signale rien.
func Incremente(n int) int64 {
	var count atomic.Int64
	var wg sync.WaitGroup
	for range n {
		wg.Go(func() { count.Add(1) })
	}
	wg.Wait()
	return count.Load()
}
