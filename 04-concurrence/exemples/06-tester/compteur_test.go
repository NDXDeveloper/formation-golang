/* ============================================================================
   Section 4.6 : Tester le code concurrent (-race et testing/synctest)
   Description : Trois tests — le compteur sous -race (version synchronisée),
                 le timeout de la section en horloge VIRTUELLE (synctest.Test :
                 1 s virtuelle, instantané en réel), et synctest.Wait pour
                 stabiliser une goroutine d'arrière-plan
   Fichier source : 06-tester-code-concurrent.md
   Lancer :  go test ./...        puis  go test -race ./...
   ============================================================================ */

package tester

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

func TestCompteurSynchronise(t *testing.T) {
	if got := Incremente(100); got != 100 {
		t.Fatalf("count = %d, attendu 100", got)
	}
	// Avec -race : aucun signalement — la synchronisation (atomic) est correcte.
}

func TestRequestTimeout(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		start := time.Now()
		<-ctx.Done() // attend l'échéance — instantané en temps virtuel

		if d := time.Since(start); d != time.Second {
			t.Fatalf("échéance après %v, attendu 1s", d)
		}
		if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
			t.Fatalf("err = %v", ctx.Err())
		}
	})
}

func TestArrierePlanStabilise(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var etat atomic.Int32
		go func() {
			time.Sleep(time.Minute) // une minute VIRTUELLE
			etat.Store(1)
		}()
		synctest.Wait() // attend que la goroutine soit durablement bloquée
		if etat.Load() != 0 {
			t.Fatal("état modifié trop tôt")
		}
		time.Sleep(2 * time.Minute) // l'horloge virtuelle dépasse l'échéance
		synctest.Wait()
		if etat.Load() != 1 {
			t.Fatal("état non modifié après l'échéance virtuelle")
		}
	})
}
