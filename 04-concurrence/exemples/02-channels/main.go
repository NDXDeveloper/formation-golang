/* ============================================================================
   Section 4.2 : Channels (buffered/unbuffered, select, fermeture)
   Description : Rendez-vous, tampon (len/cap), fermeture (range + virgule-ok),
                 diffusion par close, select (cas d'envoi, default non bloquant,
                 canal nil jamais choisi), générateur idiomatique
   Fichier source : 02-channels.md
   ============================================================================ */

package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Le générateur idiomatique : produit puis ferme (l'émetteur possède le canal).
func gen(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

func main() {
	fmt.Println("=== Rendez-vous (non bufferisé) ===")
	ch := make(chan string)
	go func() { ch <- "prêt" }()
	fmt.Println("reçu :", <-ch, " (envoi et réception synchronisés)")

	fmt.Println("=== Bufferisé : découplé jusqu'à n ===")
	bc := make(chan int, 2)
	bc <- 1
	bc <- 2 // aucun des deux n'attend
	fmt.Println("len =", len(bc), "· cap =", cap(bc))

	fmt.Println("=== Fermer : range draine puis s'arrête, virgule-ok ensuite ===")
	cc := make(chan int, 3)
	cc <- 10
	cc <- 20
	close(cc)
	for v := range cc {
		fmt.Println("range →", v)
	}
	v, ok := <-cc
	fmt.Println("après fermeture : v =", v, "· ok =", ok)

	fmt.Println("=== La fermeture est une DIFFUSION ===")
	done := make(chan struct{})
	var reveils atomic.Int32
	var wg sync.WaitGroup
	for range 3 {
		wg.Go(func() {
			<-done // les trois attendent le même signal
			reveils.Add(1)
		})
	}
	close(done) // un seul close débloque tout le monde
	wg.Wait()
	fmt.Println("goroutines réveillées d'un coup :", reveils.Load())

	fmt.Println("=== select : cas d'envoi, délai, default, canal nil ===")
	ch1 := make(chan int)    // personne n'enverra
	ch2 := make(chan int, 1) // l'envoi est prêt (tampon libre)
	select {
	case v := <-ch1:
		fmt.Println("reçu :", v)
	case ch2 <- 42:
		fmt.Println("envoyé sur ch2 →", <-ch2)
	case <-time.After(time.Second):
		fmt.Println("délai dépassé")
	}
	var nilCh chan int
	select {
	case <-nilCh: // un canal nil n'est JAMAIS choisi
		fmt.Println("impossible")
	default:
		fmt.Println("default : rien de prêt (le canal nil est ignoré)")
	}

	fmt.Println("=== Générateur ===")
	for v := range gen(1, 2, 3) {
		fmt.Println("gen →", v)
	}
}
