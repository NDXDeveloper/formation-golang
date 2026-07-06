/* ============================================================================
   Section 4.4 : context.Context (annulation, timeout, propagation)
   Description : WithCancel (« arrêt : context canceled »), WithTimeout et
                 DeadlineExceeded, le motif select sur ctx.Done(), WithValue
                 avec clé non exportée et accesseurs typés, et les raffinements
                 récents : WithCancelCause/Cause, AfterFunc, WithoutCancel
   Fichier source : 04-context.md
   ============================================================================ */

package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// WithValue : clé sentinelle non exportée + accesseurs typés (jamais une string)
type ctxKey struct{}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

func RequestID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(ctxKey{}).(string)
	return id, ok
}

// Le motif canonique : l'annulation en concurrence du travail
func worker(ctx context.Context, jobs <-chan int) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case j, ok := <-jobs:
			if !ok {
				return nil
			}
			fmt.Println("   traité :", j)
		}
	}
}

func main() {
	fmt.Println("=== WithCancel : l'annulation ferme Done() ===")
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Go(func() {
		<-ctx.Done()
		fmt.Println("arrêt :", ctx.Err())
	})
	cancel()
	wg.Wait()

	fmt.Println("=== WithTimeout : l'échéance l'emporte ===")
	ctx2, cancel2 := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel2() // TOUJOURS, même si le timeout s'annule seul (lostcancel)
	<-ctx2.Done()
	fmt.Println("DeadlineExceeded :", errors.Is(ctx2.Err(), context.DeadlineExceeded))

	fmt.Println("=== Le motif worker : jobs OU annulation ===")
	jobs := make(chan int, 2)
	jobs <- 1
	jobs <- 2
	close(jobs)
	fmt.Println("worker →", worker(context.Background(), jobs))

	fmt.Println("=== WithValue : accesseurs typés ===")
	rctx := WithRequestID(context.Background(), "req-42")
	id, ok := RequestID(rctx)
	fmt.Println("RequestID :", id, "· présent :", ok)

	fmt.Println("=== Raffinements : Cause, AfterFunc, WithoutCancel ===")
	ctx3, cancel3 := context.WithCancelCause(context.Background())
	cancel3(fmt.Errorf("source indisponible"))
	<-ctx3.Done()
	fmt.Println("context.Cause →", context.Cause(ctx3))

	nettoye := make(chan struct{})
	ctx4, cancel4 := context.WithCancel(context.Background())
	context.AfterFunc(ctx4, func() { close(nettoye) })
	cancel4()
	<-nettoye
	fmt.Println("AfterFunc : nettoyage exécuté à l'annulation ✔")

	parent, pcancel := context.WithCancel(context.Background())
	detache := context.WithoutCancel(parent)
	pcancel()
	select {
	case <-detache.Done():
		fmt.Println("WithoutCancel : annulé (inattendu)")
	default:
		fmt.Println("WithoutCancel : survit à l'annulation du parent ✔")
	}
}
