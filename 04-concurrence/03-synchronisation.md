🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 4.3 Synchronisation (`sync.WaitGroup`, `Mutex`, `Once`, `errgroup`)

Les channels ne sont pas toujours le bon outil. Pour **protéger un état partagé** ou **attendre un groupe de goroutines**, le paquet `sync` offre des primitives souvent plus simples et plus rapides. Rappel du principe ([§ 4.2](02-channels.md)) : on emploie un channel pour *transmettre* des données et *coordonner*, une primitive `sync` pour *garder* une donnée partagée — et, dans le doute, le plus lisible des deux. Cette section couvre les quatre chevaux de trait du titre, plus les atomics, plus légers encore.

## `sync.WaitGroup` : attendre un groupe de goroutines

Un `WaitGroup` attend qu'un ensemble de goroutines se terminent. Trois méthodes : `Add(n)` incrémente un compteur, `Done()` le décrémente, `Wait()` bloque jusqu'à zéro. La règle d'or : **`Add` avant de lancer** la goroutine (jamais à l'intérieur, sous peine de course avec `Wait`), et `defer wg.Done()` dedans.

```go
var wg sync.WaitGroup

for _, url := range urls {
	wg.Add(1) // AVANT de lancer
	go func() {
		defer wg.Done() // décrémente à la sortie, quel que soit le chemin
		fetch(url)
	}()
}
wg.Wait() // bloque jusqu'à ce que le compteur retombe à 0
```

La valeur zéro est prête à l'emploi (aucune initialisation) ; en revanche, un `WaitGroup` **ne doit pas être copié** après usage — passez-le par pointeur (`go vet` le détecte, [§ 13.5](../13-tests-qualite/05-linters.md)). Depuis **Go 1.25**, `wg.Go(f)` condense `Add(1)` + `go` + `Done` :

```go
var wg sync.WaitGroup
for _, url := range urls {
	wg.Go(func() { fetch(url) })
}
wg.Wait()
```

## `sync.Mutex` : protéger un état partagé

Un `Mutex` garantit qu'une seule goroutine à la fois entre dans une **section critique**. On verrouille avec `Lock()`, on libère avec `Unlock()` — et l'idiome est `defer mu.Unlock()` juste après le `Lock`, pour libérer sur tous les chemins de sortie. Placez le verrou **à côté** de ce qu'il protège, en champ **non exporté** (rappel du [§ 3.2](../03-types-interfaces/02-composition-embedding.md) : un champ nommé `mu`, pas un embedding qui exposerait `Lock`/`Unlock`) :

```go
type Counter struct {
	mu    sync.Mutex // le verrou, juste à côté des données qu'il garde
	count map[string]int
}

func (c *Counter) Inc(key string) { // receveur pointeur : le type contient un verrou (§ 3.1)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count[key]++
}
```

La `map` sera initialisée par un constructeur (`NewCounter`). Deux réflexes : garder les sections critiques **courtes** — jamais d'E/S bloquante sous verrou — et, là encore, ne pas copier le `Mutex`.

Pour un simple **compteur** ou un **drapeau**, `sync/atomic` est plus léger qu'un verrou :

```go
var n atomic.Int64
n.Add(1) // incrément atomique
n.Load() // lecture atomique
```

Quand les **lectures dominent** massivement les écritures, `sync.RWMutex` (`RLock`/`RUnlock`) autorise plusieurs lecteurs simultanés ; ce n'est pas un choix par défaut, son intérêt n'apparaît que sous forte concurrence en lecture. Enfin, `sync.Map` existe pour quelques profils d'accès particuliers, mais pour la plupart des cas une `map` + `Mutex` reste préférable ; et `sync.Pool`, orienté performance, est traité au [§ 14.3](../14-performance/03-optimisations-pgo.md).

## `sync.Once` : initialiser une seule fois

`once.Do(f)` garantit que `f` s'exécute **exactement une fois**, même appelée depuis plusieurs goroutines — idéal pour une initialisation paresseuse (singleton, chargement de configuration). Tous les appelants attendent la fin du premier `Do`, puis poursuivent.

```go
var (
	once   sync.Once
	client *http.Client
)

func Client() *http.Client {
	once.Do(func() {
		client = &http.Client{ /* configuration coûteuse */ }
	})
	return client // initialisé une seule fois, même en concurrence
}
```

Depuis **Go 1.21**, `sync.OnceValue` (et `OnceFunc`, `OnceValues`) encapsule ce motif de façon plus ergonomique :

```go
var Client = sync.OnceValue(func() *http.Client {
	return &http.Client{ /* configuration coûteuse */ }
})

// Client() calcule à la première invocation, puis renvoie la valeur mémorisée.
```

## `errgroup` : un `WaitGroup` qui propage les erreurs

Un `WaitGroup` attend des goroutines, mais ne remonte pas leurs **erreurs**. C'est le rôle d'`errgroup`, du module officiel (mais externe) `golang.org/x/sync`. Il combine attente, **propagation de la première erreur** et **annulation** des autres tâches via `context` ([§ 4.4](04-context.md)) :

```go
import "golang.org/x/sync/errgroup"

func fetchAll(ctx context.Context, urls []string) error {
	g, ctx := errgroup.WithContext(ctx) // ctx est annulé dès la première erreur
	results := make([]string, len(urls))

	for i, url := range urls {
		g.Go(func() error {
			body, err := fetch(ctx, url)
			if err != nil {
				return err // la première erreur non nil sera renvoyée par Wait
			}
			results[i] = body // chaque goroutine écrit à un indice distinct : pas de partage
			return nil
		})
	}
	return g.Wait() // attend tout le groupe ; renvoie la première erreur rencontrée
}
```

Dès qu'une goroutine renvoie une erreur, le `ctx` dérivé est annulé — les autres, si elles l'observent, s'arrêtent (*fail fast*). Après `g.Wait()`, la lecture de `results` est sûre : les écritures visent des indices distincts et `Wait` fournit la synchronisation. `g.SetLimit(n)` permet en outre de **borner** le nombre de goroutines simultanées — un pont vers les *worker pools* du [§ 4.5](05-patterns-concurrence.md). C'est aussi le point de rencontre de la concurrence et de l'idiome d'erreurs ([§ 2.9](../02-fondamentaux-langage/09-gestion-erreurs.md)).

## Côté IDE : GoLand et VS Code

- **Copie de verrou** : copier un `Mutex`, un `WaitGroup` ou un `Once` (qui ne doivent pas l'être) déclenche l'analyseur `copylocks` de `go vet` — signalé par les inspections de **GoLand** comme par **gopls** dans VS Code (vet à l'enregistrement, [§ 13.5](../13-tests-qualite/05-linters.md)).
- **Le reste échappe à l'analyse statique.** Un `Lock` oublié ou un accès concurrent non protégé sont invisibles au compilateur : c'est le **détecteur de races** ([§ 4.6](06-tester-code-concurrent.md)), en exécutant les tests avec `-race`, qui les révèle. GoLand ajoute quelques inspections (verrou copié, `defer` manquant), mais la garantie vient des tests, pas de l'éditeur.

## En résumé

- **`WaitGroup`** attend un groupe : `Add` **avant** de lancer, `defer Done`, `Wait` pour bloquer ; ne pas copier (passer par pointeur) ; `wg.Go(f)` depuis Go 1.25.
- **`Mutex`** protège un état partagé : `Lock` + `defer Unlock`, verrou en champ non exporté à côté des données ([§ 3.2](../03-types-interfaces/02-composition-embedding.md)), sections critiques courtes. Pour un compteur/drapeau, `sync/atomic` est plus léger ; `RWMutex` seulement si les lectures dominent.
- **`Once`** initialise une seule fois (`once.Do`) ; `sync.OnceValue`/`OnceFunc` (Go 1.21) en sont la forme moderne.
- **`errgroup`** (`x/sync`) = un `WaitGroup` qui **propage la première erreur** et **annule** le reste via `context` ([§ 4.4](04-context.md)) ; `SetLimit` borne la concurrence ([§ 4.5](05-patterns-concurrence.md)).
- Les bugs de synchronisation se traquent à l'exécution avec **`-race`** ([§ 4.6](06-tester-code-concurrent.md)), pas par l'analyse statique.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [4.4 — `context.Context` (annulation, timeout, propagation)](04-context.md)

⏭ [`context.Context` (annulation, timeout, propagation)](/04-concurrence/04-context.md)
