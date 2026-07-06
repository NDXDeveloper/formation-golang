🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 14.1 Profilage avec pprof (CPU, heap, goroutines)

Profiler, c'est répondre empiriquement à « où le programme passe-t-il réellement son temps et sa mémoire ? ». C'est le fondement du cycle de ce module : on ne devine pas le point chaud, on le mesure. Go embarque pour cela **pprof**, un profileur par échantillonnage livré avec le *runtime*, doublé d'un outil d'analyse mûr (`go tool pprof`). Cette section couvre les types de profils, trois façons de les produire, la lecture d'un profil (`top`, `list`, *flame graph*, *flat* contre *cum*), leur comparaison, et — nouveauté de Go 1.25 — les **traces d'exécution** et le **flight recorder** pour les questions temporelles que pprof ne sait pas trancher.

---

## Les types de profils

pprof produit plusieurs profils, chacun répondant à une question différente :

| Profil | Ce qu'il révèle |
|--------|-----------------|
| **CPU** | où passe le temps de calcul (échantillonnage des piles d'appels, ~100 Hz) |
| **Tas** (heap) | les allocations : mémoire **vivante** (`inuse_space`/`inuse_objects`) et allocations **cumulées** (`alloc_space`/`alloc_objects`) |
| **Goroutines** | les piles de toutes les goroutines vivantes — pour traquer fuites et blocages |
| **Blocage** (block) | où les goroutines attendent sur la synchronisation (activer `runtime.SetBlockProfileRate`) |
| **Mutex** | la contention sur les verrous (activer `runtime.SetMutexProfileFraction`) |

---

## Produire un profil

### Depuis un benchmark — le plus simple

C'est le prolongement direct du [§ 13.4](../13-tests-qualite/04-fuzzing-benchmarks.md) : `go test` écrit les profils d'un chemin chaud isolé.

```sh
go test -bench=. -benchmem \
  -cpuprofile=cpu.out -memprofile=mem.out -blockprofile=block.out
```

### Depuis un service en cours — `net/http/pprof`

Un import à effet de bord enregistre les points d'accès de profilage sur le *mux* HTTP par défaut :

```go
import _ "net/http/pprof" // expose /debug/pprof/ sur http.DefaultServeMux

// … puis, sur un écouteur interne dédié :
go func() { log.Println(http.ListenAndServe("localhost:6060", nil)) }()
```

On récupère alors les profils à distance :

```sh
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30  # CPU sur 30 s
go tool pprof http://localhost:6060/debug/pprof/heap                # tas
go tool pprof http://localhost:6060/debug/pprof/goroutine           # goroutines
```

> ⚠️ **`/debug/pprof` ne doit jamais être exposé publiquement.** Il révèle le fonctionnement interne du service et laisse quiconque déclencher des profils coûteux. On le sert sur un écouteur interne (ci-dessus, `localhost`) ou derrière une authentification — jamais sur le *mux* public. À rapprocher du durcissement HTTP en [§ 16.3](../16-securite/03-durcissement-http.md).

### Dans le code — `runtime/pprof`

Pour un outil en ligne de commande ou une portion précise, on pilote le profilage à la main :

```go
f, _ := os.Create("cpu.prof")
pprof.StartCPUProfile(f)
defer pprof.StopCPUProfile()
// … code à profiler
```

Le profil de tas s'écrit avec `pprof.WriteHeapProfile(f)`.

---

## Analyser avec `go tool pprof`

L'outil s'ouvre sur un profil (local ou distant) et propose une session interactive. La distinction cardinale est **flat contre cum** : le temps *flat* est celui passé dans la fonction **elle-même**, le temps *cum* (cumulé) inclut **tout ce qu'elle appelle**. Une fonction au *flat* élevé est un vrai point chaud ; un *cum* élevé désigne un chemin coûteux à descendre.

```text
$ go tool pprof cpu.out
(pprof) top
      flat  flat%   sum%        cum   cum%
     1.31s 43.7%  43.7%      1.98s 66.0%  encoding/json.(*decodeState).object
     0.42s 14.0%  57.7%      0.42s 14.0%  runtime.mapassign_faststr
     0.28s  9.3%  67.0%      0.61s 20.3%  encoding/json.(*decodeState).value
(pprof) list decodeState.*object   # source annotée, temps ligne à ligne
```

`top -cum` trie par temps cumulé ; `list <regexp>` annote le code source ligne à ligne ; `peek` et `traces` explorent les appelants/appelés. Mais l'entrée recommandée est **l'interface web** :

```sh
go tool pprof -http=:8080 cpu.out
```

Elle ouvre dans le navigateur un *flame graph*, un graphe d'appels, la vue `top` et le source annoté. Sur un *flame graph*, la **largeur** d'un cadre représente le temps (ou les allocations) et la **profondeur** la pile d'appels : on cherche les cadres les plus larges.

### Profil de tas : vivant ou cumulé

Un profil de tas s'interprète selon deux axes. Pour une **fuite** ou l'empreinte mémoire, on regarde la mémoire vivante (`-inuse_space`, `-inuse_objects`). Pour **réduire les allocations** — et donc la pression sur le GC ([§ 14.2](02-gc-allocations.md)) —, on regarde les allocations cumulées (`-alloc_space`, `-alloc_objects`). On bascule entre axes avec `-sample_index`.

### Comparer deux profils

Pour valider une optimisation, on compare l'avant et l'après :

```sh
go tool pprof -base=avant.out apres.out   # diff : ce qui a changé
```

C'est le complément *qualitatif* de la comparaison statistique par benchstat ([§ 14.4](04-benchmarking.md)).

---

## Traces d'exécution et flight recorder (Go 1.25) 🆕

pprof répond « **où** passe le temps ? » (un agrégat échantillonné). La **trace d'exécution** répond « **que** s'est-il passé, **quand**, et pourquoi cette goroutine a-t-elle calé ? » — une chronologie fine de l'ordonnancement des goroutines, des pauses GC et des blocages (syscall, réseau, synchronisation). C'est l'outil des problèmes de **latence**.

On produit une trace avec `go test -trace=trace.out`, via `/debug/pprof/trace?seconds=5`, ou avec `runtime/trace.Start`/`Stop`, puis on l'ouvre :

```sh
go tool trace trace.out   # UI web : timeline (« view trace by proc »), analyse par goroutine, latence, GC
```

### Le flight recorder

Tracer en continu est coûteux, et l'on ignore *à l'avance* quand surviendra l'anomalie rare que l'on cherche — une trace complète est donc impraticable pour un service de longue durée. Le **flight recorder** (Go 1.25) résout ce dilemme : il conserve en mémoire une **fenêtre glissante** des derniers instants de trace (léger, toujours actif), que l'on **capture à la demande**, précisément au moment où le programme détecte un problème (requête hors budget de latence, erreur). On saisit ainsi ce qui a **précédé** l'événement.

```go
import "runtime/trace"

fr := trace.NewFlightRecorder(trace.FlightRecorderConfig{
	MinAge:   5 * time.Second, // ~2× la fenêtre de l'événement surveillé
	MaxBytes: 1 << 20,          // borne mémoire du tampon (1 MiB) — prioritaire sur MinAge
})
if err := fr.Start(); err != nil {
	log.Fatal(err)
}
defer fr.Stop()

// … plus loin, à la détection d'une anomalie :
f, _ := os.Create("trace.out")
if _, err := fr.WriteTo(f); err != nil { // instantané de la fenêtre glissante
	log.Print(err)
}
_ = f.Close()
```

L'instantané s'ouvre ensuite dans `go tool trace` comme n'importe quelle trace. Un seul flight recorder peut être actif à la fois ; `MaxBytes` borne la mémoire consommée et l'emporte sur `MinAge`.

---

## Du profilage à l'optimisation

Le profilage n'est pas une fin : il alimente la suite du module.

- Le **profil CPU** collecté ici est exactement ce que consomme la **PGO** ([§ 14.3](03-optimisations-pgo.md)) : un fichier `default.pgo` déposé dans le package `main` est un profil pprof que le compilateur exploite pour guider *inlining* et dévirtualisation.
- Les **profils de tas/allocations** guident la réduction des allocations et l'*escape analysis* ([§ 14.2](02-gc-allocations.md)).
- Toute optimisation se **valide** en reprofilant (`-base`) et en comparant les benchmarks avec benchstat ([§ 14.4](04-benchmarking.md)).

En production, des outils de **profilage continu** (Grafana Pyroscope, Google Cloud Profiler, Datadog…) collectent en permanence des profils pprof sur l'ensemble du parc — sujet relevant de l'observabilité ([§ 12.4](../12-erreurs-debogage/04-observabilite.md)).

---

## Côté IDE : GoLand et VS Code

**GoLand** propose **Run with Profiler** (CPU, allocations…) sur une application, un test ou un benchmark : les résultats s'affichent en *flame graph* et arbre d'appels intégrés. Il ouvre aussi un fichier `.out`/`.pprof` existant, et les traces via son outil dédié.

**VS Code** (extension Go officielle) profile un benchmark via *CodeLens* et ouvre l'interface web de pprof ; les traces se consultent avec `go tool trace`.

Dans les deux cas, `go tool pprof -http=:8080 profil.out` et `go tool trace trace.out` restent le dénominateur commun, indépendant de l'éditeur.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [14.2 — Le GC de Go, allocations, escape analysis](02-gc-allocations.md)

⏭ [Le GC de Go, allocations, escape analysis](/14-performance/02-gc-allocations.md)
