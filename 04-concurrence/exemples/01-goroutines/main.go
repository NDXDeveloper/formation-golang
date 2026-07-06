/* ============================================================================
   Section 4.1 : Goroutines et le scheduler
   Description : Cycle de vie complet — arguments figés à l'instruction go,
                 variable de boucle par itération (Go 1.22), attente par
                 WaitGroup (jamais time.Sleep), panique récupérée DANS la
                 goroutine (le programme survit), GOMAXPROCS et NumGoroutine
   Fichier source : 01-goroutines.md
   ============================================================================ */

package main

import (
	"fmt"
	"log"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
)

func task() { panic("boom") }

func main() {
	log.SetFlags(0)

	fmt.Println("=== Les arguments d'un go sont évalués à l'instruction ===")
	fige := make(chan int, 1)
	i := 1
	go func(v int) { fige <- v }(i) // v est figé ICI, comme pour go fmt.Println(i)
	i = 2
	fmt.Println("valeur reçue :", <-fige, " (et non", i, ")")

	fmt.Println("=== Variable de boucle par itération (Go 1.22) ===")
	vals := make(chan int, 3)
	var wg sync.WaitGroup
	for j := 0; j < 3; j++ {
		wg.Go(func() { vals <- j }) // chaque goroutine capture SON j
	}
	wg.Wait()
	close(vals)
	got := []int{}
	for v := range vals {
		got = append(got, v)
	}
	sort.Ints(got) // l'ordre d'exécution est indéterminé — pas les valeurs
	fmt.Println("valeurs capturées :", got)

	fmt.Println("=== Attendre ses goroutines (WaitGroup, jamais time.Sleep) ===")
	var compte atomic.Int32
	var wg2 sync.WaitGroup
	for range 5 {
		wg2.Go(func() { compte.Add(1) })
	}
	wg2.Wait() // sans cette ligne, main pourrait sortir avant les goroutines
	fmt.Println("goroutines terminées :", compte.Load())

	fmt.Println("=== Panique récupérée DANS la goroutine : le programme survit ===")
	var wg3 sync.WaitGroup
	wg3.Go(func() {
		defer func() {
			if p := recover(); p != nil {
				log.Printf("panique récupérée : %v", p)
			}
		}()
		task()
	})
	wg3.Wait()
	fmt.Println("le programme a survécu ✔")

	fmt.Println("=== Le runtime en deux fonctions ===")
	fmt.Println("GOMAXPROCS  :", runtime.GOMAXPROCS(0), "(parallélisme réel)")
	fmt.Println("NumGoroutine:", runtime.NumGoroutine(), "(goroutines vivantes)")
}
