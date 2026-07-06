/* ============================================================================
   Section 4.3 : Synchronisation (sync.WaitGroup, Mutex, Once, errgroup)
   Description : WaitGroup (Add avant / defer Done, puis wg.Go Go 1.25), Mutex
                 protégeant une map (le Counter de la section, ×100 goroutines),
                 atomic, sync.OnceValue (une seule exécution), errgroup :
                 première erreur propagée, annulation du groupe, SetLimit
   Fichier source : 03-synchronisation.md
   ============================================================================ */

package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"
)

// Le Counter de la section : le verrou à côté des données qu'il garde.
type Counter struct {
	mu    sync.Mutex
	count map[string]int
}

func (c *Counter) Inc(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count[key]++
}

func main() {
	fmt.Println("=== WaitGroup : Add avant, defer Done — et wg.Go (Go 1.25) ===")
	var wg sync.WaitGroup
	var n atomic.Int32
	wg.Add(1) // forme classique : AVANT de lancer
	go func() {
		defer wg.Done()
		n.Add(1)
	}()
	for range 4 {
		wg.Go(func() { n.Add(1) }) // forme moderne : Add+go+Done condensés
	}
	wg.Wait()
	fmt.Println("tâches terminées :", n.Load())

	fmt.Println("=== Mutex : le Counter sous 100 goroutines ===")
	c := &Counter{count: map[string]int{}}
	var wg2 sync.WaitGroup
	for range 100 {
		wg2.Go(func() { c.Inc("clé") })
	}
	wg2.Wait()
	fmt.Println("count[\"clé\"] =", c.count["clé"], " (exact : le verrou protège la map)")

	fmt.Println("=== atomic : plus léger qu'un verrou pour un compteur ===")
	var a atomic.Int64
	a.Add(1)
	a.Add(1)
	fmt.Println("atomic =", a.Load())

	fmt.Println("=== sync.OnceValue : calculé une seule fois ===")
	var appels atomic.Int32
	valeur := sync.OnceValue(func() int {
		appels.Add(1)
		return 42
	})
	_, _, _ = valeur(), valeur(), valeur()
	fmt.Println("valeur =", valeur(), "· exécutions de f :", appels.Load())

	fmt.Println("=== errgroup : première erreur + annulation du groupe ===")
	g, ctx := errgroup.WithContext(context.Background())
	g.Go(func() error { return errors.New("échec rapide") })
	g.Go(func() error {
		select {
		case <-ctx.Done():
			fmt.Println("   l'autre tâche voit l'annulation :", context.Cause(ctx) != nil)
		case <-time.After(2 * time.Second):
		}
		return nil
	})
	fmt.Println("g.Wait() →", g.Wait())

	fmt.Println("=== errgroup.SetLimit : concurrence bornée ===")
	var cur, max atomic.Int32
	g2 := new(errgroup.Group)
	g2.SetLimit(2)
	for range 6 {
		g2.Go(func() error {
			if c := cur.Add(1); c > max.Load() {
				max.Store(c)
			}
			time.Sleep(5 * time.Millisecond)
			cur.Add(-1)
			return nil
		})
	}
	_ = g2.Wait()
	fmt.Println("concurrence maximale observée :", max.Load(), "(limite : 2)")
}
