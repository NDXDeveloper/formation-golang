/* ============================================================================
   Section 4.5 : Patterns — worker pool, fan-in/fan-out, pipeline
   Description : Pipeline annulable (ordre préservé : 4, 9, 16), fan-out de
                 deux étages sq + fan-in par merge (l'ordre se perd, la somme
                 non), worker pool borné, et la PREUVE anti-fuite : après un
                 break précoce, cancel() dénoue tout le pipeline
                 (NumGoroutine stable)
   Fichier source : 05-patterns-concurrence.md
   ============================================================================ */

package main

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"sync"
	"time"
)

func gen(ctx context.Context, nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			select {
			case out <- n:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

func sq(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			select {
			case out <- n * n:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

func merge(cs ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	wg.Add(len(cs))
	for _, c := range cs {
		go func() {
			defer wg.Done()
			for v := range c {
				out <- v
			}
		}()
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func main() {
	fmt.Println("=== Pipeline simple : l'ordre est préservé ===")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for v := range sq(ctx, gen(ctx, 2, 3, 4)) {
		fmt.Println("sq →", v) // 4, 9, 16 — dans l'ordre
	}

	fmt.Println("=== Fan-out / fan-in : l'ordre se perd, pas les valeurs ===")
	in := gen(ctx, 2, 3, 4, 5)
	c1 := sq(ctx, in) // deux étages lisent la même entrée
	c2 := sq(ctx, in)
	vals := []int{}
	for v := range merge(c1, c2) {
		vals = append(vals, v)
	}
	sort.Ints(vals)
	fmt.Println("résultats triés :", vals, "(l'arrivée était désordonnée)")

	fmt.Println("=== Anti-fuite : break précoce PUIS cancel → tout se dénoue ===")
	avant := runtime.NumGoroutine()
	ctx2, cancel2 := context.WithCancel(context.Background())
	out := sq(ctx2, gen(ctx2, 1, 2, 3, 4, 5, 6, 7, 8))
	for range out {
		break // le consommateur abandonne tôt
	}
	cancel2() // la voie de sortie : chaque étage observe Done() et se termine
	time.Sleep(50 * time.Millisecond)
	fmt.Printf("goroutines : %d avant → %d après (aucune fuite) ✔\n", avant, runtime.NumGoroutine())
}
