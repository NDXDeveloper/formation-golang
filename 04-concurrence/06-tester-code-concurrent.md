🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 4.6 Tester le code concurrent : détecteur de races (`-race`) et `testing/synctest`

Le code concurrent résiste aux tests ordinaires : le non-déterminisme fait qu'un test peut passer **par chance**, laissant un bug se réveiller une fois sur mille en production. Deux outils complémentaires comblent ce fossé — le **détecteur de races** (`-race`), qui débusque les courses de données survenues pendant l'exécution, et **`testing/synctest`** (stable en Go 1.25), qui rend déterministes et instantanés les tests dépendant du temps. Les fondamentaux du package `testing` — tests, sous-tests, table-driven — sont traités au [§ 13.1](../13-tests-qualite/01-tests-unitaires.md) ; cette section se concentre sur ce qui est propre à la concurrence.

## Pourquoi les tests ordinaires ne suffisent pas

Une course de données ne se produit qu'à certains entrelacements d'exécution ; un test vert ne prouve donc **pas** son absence. Pire, l'attente d'une goroutine par un `time.Sleep` bricolé rend la suite à la fois **lente** (on attend pour rien) et **instable** (trop court, elle échoue sous charge ; trop long, elle traîne). Il faut un outillage dédié.

## Le détecteur de races (`-race`)

Activé par l'indicateur `-race`, il **instrumente** le programme (via ThreadSanitizer) pour repérer les **courses de données** — deux goroutines accédant à la même mémoire en concurrence, dont au moins une en écriture, sans synchronisation. Il s'utilise sur toutes les commandes :

```console
$ go test -race ./...
$ go run -race .
```

```go
func TestCounter(t *testing.T) {
	var count int
	var wg sync.WaitGroup
	for range 100 {
		wg.Go(func() { count++ }) // accès concurrent non synchronisé
	}
	wg.Wait()
	// Même si count atteint 100 par chance, `-race` signale la course sous-jacente.
}
```

Le remède, ici, est un `atomic.Int64` ou un `Mutex` ([§ 4.3](03-synchronisation.md)). Trois propriétés à retenir. C'est un détecteur **dynamique** : il ne voit que les courses qui **surviennent réellement** — il faut donc des tests qui exercent les chemins concurrents, d'où l'intérêt de l'exécuter sur toute la suite en intégration continue ([§ 15.2](../15-deploiement-devops/02-cicd.md)). Il a un **coût** (exécution 2 à 20× plus lente, mémoire 5 à 10× supérieure) : réservé aux tests et à la CI, jamais à la production. Et il cible **uniquement** les courses de données — pas les interblocages (le runtime détecte le blocage total, [§ 4.2](02-channels.md)), pas les conditions de course logiques, pas les fuites de goroutines.

## Le problème du temps

Tester des délais, des minuteurs (`Ticker`) ou des relances avec back-off en **temps réel** est un cauchemar : soit on attend vraiment (suite lente), soit on ne teste pas la temporisation du tout. C'est exactement ce que résout `synctest`.

## `testing/synctest` (stable en Go 1.25)

Introduit à titre expérimental en Go 1.24 puis stabilisé dans le paquet `testing/synctest` en Go 1.25, il exécute un groupe de goroutines dans une **bulle** isolée dotée d'une **horloge virtuelle**. Le temps n'y avance que lorsque **toutes** les goroutines de la bulle sont bloquées — un `time.Sleep(time.Hour)` se termine alors **instantanément** en temps réel, de façon parfaitement déterministe.

`synctest.Test(t, f)` lance la bulle ; à l'intérieur, `time.Sleep`, `time.After`, les `Ticker` et les échéances de `context` utilisent l'horloge feinte.

```go
import "testing/synctest"

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
```

Le test est instantané *et* exact : le délai d'une seconde s'écoule en temps virtuel, et `time.Since` renvoie précisément `1s`. La fonction compagne `synctest.Wait()` bloque jusqu'à ce que toutes les autres goroutines de la bulle soient durablement bloquées — pratique pour laisser des goroutines d'arrière-plan atteindre un état stable avant d'affirmer un résultat. Deux contraintes : les goroutines lancées dans la bulle doivent **toutes se terminer** avant la fin de `f` (sans quoi le test échoue), et elles ne doivent pas se bloquer sur le monde extérieur (E/S réelles).

## Détecter les fuites de goroutines

Ni `-race` ni les assertions classiques ne voient une goroutine qui **ne se termine jamais**. Deux approches : la bibliothèque `go.uber.org/goleak` (externe), dont `defer goleak.VerifyNone(t)` fait échouer un test s'il laisse des goroutines derrière lui ; et `synctest` lui-même, qui signale les goroutines non terminées d'une bulle. On complète au besoin par le profil goroutine de pprof ([§ 14.1](../14-performance/01-pprof.md)). Les détails de mise en place des tests relèvent du [§ 13.1](../13-tests-qualite/01-tests-unitaires.md).

## Côté IDE : GoLand et VS Code

- **Activer `-race`** : dans **GoLand**, activez le détecteur de races dans la configuration de test (option dédiée, ou `-race` dans les arguments de l'outil Go) ; dans **VS Code**, réglez `"go.testFlags": ["-race"]`. Le rapport de course s'affiche dans la sortie des tests des deux côtés.
- Le vrai réflexe reste de faire tourner **toute la suite avec `-race` en CI** ([§ 15.2](../15-deploiement-devops/02-cicd.md)) — c'est là que les entrelacements rares finissent par se produire. Pour explorer un test bloqué, le débogueur et sa vue des goroutines ([§ 4.1](01-goroutines.md)) prennent le relais.

## En résumé

- Un test vert ne prouve **pas** l'absence de course ; n'attendez jamais une goroutine par `time.Sleep`.
- **`-race`** détecte dynamiquement les **courses de données** (`go test -race ./...`) ; il ne voit que celles qui surviennent, coûte cher (2–20× en temps, 5–10× en mémoire) et ne couvre ni interblocages ni fuites — à lancer systématiquement en CI ([§ 15.2](../15-deploiement-devops/02-cicd.md)).
- **`testing/synctest`** (stable Go 1.25) exécute une **bulle** à **horloge virtuelle** : les tests de délais/minuteurs/`context` deviennent **instantanés et déterministes** via `synctest.Test(t, f)` et `synctest.Wait()`.
- Les **fuites** se traquent avec `goleak`, la vérification intégrée de `synctest`, ou le profil pprof ([§ 14.1](../14-performance/01-pprof.md)).
- Fondamentaux des tests au **[§ 13.1](../13-tests-qualite/01-tests-unitaires.md)** ; ici, l'outillage propre à la concurrence.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [5. Backend HTTP — le scénario phare](../05-backend-http/README.md)

⏭ [Backend HTTP — le scénario phare](/05-backend-http/README.md)
