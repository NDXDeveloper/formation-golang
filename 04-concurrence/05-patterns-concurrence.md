🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 4.5 Patterns : worker pool, fan-in/fan-out, pipeline

Ces patrons **composent** les primitives des sections précédentes — goroutines ([§ 4.1](01-goroutines.md)), channels ([§ 4.2](02-channels.md)), `WaitGroup`/`errgroup` ([§ 4.3](03-synchronisation.md)), `context` ([§ 4.4](04-context.md)) — en structures réutilisables pour traiter des données en concurrence. Le chemin nominal est le plus simple ; la vraie difficulté, celle qui traverse toute cette section, tient en trois soucis constants : **l'achèvement** (fermer les channels au bon moment), **l'annulation** (`context`), et **l'absence de fuite**.

## Le pipeline : des étages reliés par des channels

Un pipeline est une suite d'**étages** connectés par des channels. Chaque étage est une fonction qui reçoit d'un channel d'entrée, transforme, envoie sur un channel de sortie, et **ferme sa sortie** quand son entrée est épuisée ([§ 4.2](02-channels.md)).

```go
// Étage source : produit des entiers puis ferme.
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

// Étage : envoie les carrés, ferme out quand in est drainé.
func sq(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			out <- n * n
		}
	}()
	return out
}

for v := range sq(gen(2, 3, 4)) { // 4, 9, 16
	fmt.Println(v)
}
```

La composition `sq(gen(...))` se lit de gauche à droite. Mais un danger guette : si le consommateur **s'arrête tôt** (un `break` dans le `range`), l'étage `sq` reste bloqué sur `out <- …` faute de receveur, et `gen` derrière lui aussi — deux goroutines fuient ([§ 4.1](01-goroutines.md)).

## Annuler proprement : le contexte dans le pipeline

La parade : chaque étage doit pouvoir **abandonner**. On garde l'envoi par un `select` sur `ctx.Done()`.

```go
func sq(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			select {
			case out <- n * n: // envoi normal
			case <-ctx.Done(): // annulé : l'étage se termine sans bloquer
				return
			}
		}
	}()
	return out
}
```

`gen` suit le même schéma pour son propre envoi, et l'appelant crée le contexte avec `defer cancel()` : dès qu'il cesse de consommer, il annule, et **tout le pipeline se dénoue**. C'est la règle à retenir — un étage sans voie de sortie est une fuite en puissance.

## Fan-out / fan-in : paralléliser un étage

Quand un étage est lent, on le **duplique** : plusieurs goroutines lisent le **même** channel d'entrée (fan-out), et l'on **recombine** leurs sorties en un seul channel (fan-in). Le fan-in est le morceau délicat — il faut un `WaitGroup` pour savoir quand toutes les entrées sont drainées avant de fermer la sortie :

```go
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
		wg.Wait()  // quand toutes les entrées sont épuisées…
		close(out) // … on ferme la sortie
	}()

	return out
}
```

On assemble alors fan-out et fan-in autour d'un étage :

```go
in := gen(2, 3, 4, 5)
c1 := sq(ctx, in)              // deux étages sq lisent la même entrée :
c2 := sq(ctx, in)              // le travail se répartit entre eux
for v := range merge(c1, c2) { // résultats recombinés, ordre non garanti
	fmt.Println(v)
}
```

Attention : la parallélisation fait perdre l'**ordre** — les résultats sortent dans un ordre indéterminé.

## Worker pool : borner le parallélisme

Le fan-out crée autant de goroutines qu'on veut ; le *worker pool*, lui, en fixe le **nombre**. C'est le bon réflexe pour un volume de travail important : un nombre borné de *workers* consomme un channel de tâches partagé ([§ 4.1](01-goroutines.md) — ne pas lancer une goroutine par unité à l'infini).

```go
func workerPool(ctx context.Context, jobs <-chan Job, n int) <-chan Result {
	results := make(chan Result)
	var wg sync.WaitGroup

	wg.Add(n)
	for range n {
		go func() {
			defer wg.Done()
			for job := range jobs { // les n workers se partagent le même channel
				select {
				case results <- process(job):
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}
```

L'appelant **ferme `jobs`** quand il a fini de soumettre (pour que le `range` des workers se termine) et peut tout arrêter via `ctx`. Pour des tâches qui **renvoient une erreur**, une alternative bien plus concise existe déjà — `errgroup` avec `SetLimit` ([§ 4.3](03-synchronisation.md)) :

```go
g, ctx := errgroup.WithContext(ctx)
g.SetLimit(n) // au plus n goroutines simultanées
for _, job := range jobsList {
	g.Go(func() error { return process(ctx, job) })
}
err := g.Wait() // première erreur, ou nil
```

## Ne pas sur-concevoir

Ces patrons impressionnent, mais ne sont pas toujours la bonne réponse. Pour lancer quelques tâches concurrentes, un simple `WaitGroup` ou un `errgroup` ([§ 4.3](03-synchronisation.md)) suffit — inutile de dérouler un pipeline. Les pipelines et le fan-in/out brillent pour le **flux continu** et les **gros volumes** ; ailleurs, ajouter channels et goroutines là où une boucle séquentielle ferait l'affaire relève de la sur-ingénierie (anti-pattern, [annexe B](../annexes/go-idiomatique/README.md)). Deux repères : pour un **parcours séquentiel** de données, les itérateurs `range`-over-func ([§ 2.5](../02-fondamentaux-langage/05-boucles.md)) sont souvent plus clairs que des channels, réservés aux **étages concurrents** ; et la concurrence a un coût (ordonnancement, opérations de channel) — plus de goroutines n'est pas plus rapide, **mesurez** ([module 14](../14-performance/README.md)).

## Côté IDE : GoLand et VS Code

- Ces montages sont précisément ce que le **détecteur de races** vérifie : exécutez-les avec `-race` ([§ 4.6](06-tester-code-concurrent.md)) pour garantir l'absence d'accès concurrent non protégé.
- Une **fuite** — un étage qui ne se termine jamais — se voit dans le profil goroutine de pprof ([§ 14.1](../14-performance/01-pprof.md)) et se teste ; le débogueur (onglet *Goroutines* de GoLand, vue *Call Stack* de VS Code via Delve, [§ 4.1](01-goroutines.md)) aide à localiser un pipeline bloqué.

## En résumé

- Un **pipeline** enchaîne des étages reliés par des channels ; chaque étage **ferme sa sortie** en fin d'entrée.
- Tout étage doit pouvoir **abandonner** via `select { case <-ctx.Done(): return }`, sinon un consommateur qui s'arrête tôt fait **fuir** l'amont.
- **Fan-out** = plusieurs goroutines sur la même entrée ; **fan-in** (`merge`) = un `WaitGroup` ferme la sortie une fois toutes les entrées drainées ; l'**ordre** se perd.
- **Worker pool** = un nombre **borné** de workers sur un channel de tâches partagé ; `errgroup.SetLimit` ([§ 4.3](03-synchronisation.md)) en est la variante concise pour des tâches faillibles.
- **Ne sur-concevez pas** : `WaitGroup`/`errgroup` pour le simple, pipelines pour le flux ; itérateurs ([§ 2.5](../02-fondamentaux-langage/05-boucles.md)) pour le séquentiel, channels pour le concurrent ; **mesurez** ([module 14](../14-performance/README.md)).

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [4.6 — Tester le code concurrent : détecteur de races (`-race`) et `testing/synctest`](06-tester-code-concurrent.md)

⏭ [Tester le code concurrent : détecteur de races (`-race`) et `testing/synctest`](/04-concurrence/06-tester-code-concurrent.md)
