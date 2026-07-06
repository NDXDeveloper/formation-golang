🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 14.2 Le GC de Go, allocations, escape analysis

Une fois le point chaud localisé ([§ 14.1](01-pprof.md)), la cause est très souvent la **mémoire** : trop d'allocations, donc trop de travail pour le ramasse-miettes. Cette section explique où vivent les valeurs (pile ou tas), comment le compilateur en décide (*escape analysis*), comment fonctionne le GC de Go — et pourquoi le nouveau collecteur **Green Tea**, activé par défaut en Go 1.26, change la donne sans une ligne de code.

---

## Pile ou tas

Go gère la mémoire automatiquement. Une valeur vit soit sur la **pile** de sa goroutine — allouée et libérée au retour de la fonction, pour un coût quasi nul —, soit sur le **tas**, où c'est le GC qui la récupère plus tard. Une allocation sur le tas coûte deux fois : à la création, puis au ramassage.

Le programmeur ne choisit pas directement : c'est le **compilateur** qui tranche, par l'*escape analysis*. Mais la façon d'écrire le code oriente sa décision. (Les piles de goroutines sont petites — quelques kilo-octets — et croissent ou rétrécissent automatiquement.)

---

## L'escape analysis

À la compilation, le compilateur détermine si une valeur peut rester sur la pile ou doit « s'échapper » vers le tas. Elle s'échappe dès que sa durée de vie peut dépasser la fonction. On visualise ses décisions avec `-gcflags=-m` :

```go
type Point struct{ X, Y int }

// p s'échappe : on renvoie son adresse, sa durée de vie dépasse newPoint.
func newPoint() *Point {
	p := Point{X: 1, Y: 2}
	return &p
}

// p ne s'échappe pas : il naît et meurt dans la boucle, sur la pile.
func sumX(pts []Point) int {
	total := 0
	for _, p := range pts {
		total += p.X
	}
	return total
}
```

```sh
$ go build -gcflags=-m ./...
./geo.go:6:6: can inline newPoint
./geo.go:7:2: moved to heap: p
./geo.go:12:11: pts does not escape
```

La ligne à lire est `moved to heap: p` : la variable déclarée ligne 7 « déménage » sur le tas, parce que son adresse est renvoyée. (Le message frère `escapes to heap` apparaît, lui, quand une *valeur* s'échappe en passant par une interface — un argument passé à `fmt.Println`, par exemple.) À l'inverse, `pts does not escape` confirme que le paramètre de `sumX` reste sur la pile.

Les causes d'échappement les plus fréquentes : renvoyer un pointeur vers une variable locale ; ranger une valeur dans une **interface** (`any`) — le *boxing* force souvent l'allocation ; une *closure* qui capture une variable dont la vie dépasse la fonction ; une slice ou une map dont la taille n'est pas connue à la compilation ou qui croît ; passer des arguments à `fmt.Println` ou à une variadique `...any` (ils deviennent `any` et s'échappent). Go 1.26 améliore le tableau : le compilateur alloue désormais le tableau sous-jacent de davantage de slices **sur la pile**, réduisant la pression mémoire du code manipulant beaucoup de slices.

> **L'échappement n'est pas un défaut.** Il est souvent nécessaire et correct. L'objectif n'est jamais de le supprimer partout, mais d'éviter les allocations tas **inutiles sur le chemin chaud** identifié au profilage ([§ 14.1](01-pprof.md)). Contorsionner du code lisible pour économiser une allocation sans importance est une fausse bonne idée.

---

## Le ramasse-miettes

Le GC de Go est un collecteur **concurrent**, à **marquage tricolore et balayage** (*mark-and-sweep*), **non générationnel** et **non compactant**. Il s'exécute en même temps que le programme et vise avant tout la **faible latence** : les pauses « stop-the-world » sont typiquement bien inférieures à la milliseconde. La phase coûteuse — celle qui dispute le CPU à l'application — est le **marquage** : parcourir le graphe des objets vivants.

Deux boutons de réglage :

- **`GOGC`** (défaut `100`) — déclenche un cycle quand le tas a crû de `GOGC %` depuis le dernier ramassage (`100` = tas doublé). Plus haut : moins de cycles, plus de mémoire ; plus bas : cycles plus fréquents, moins de mémoire. `GOGC=off` désactive le GC.
- **`GOMEMLIMIT`** (Go 1.19+) — une **limite mémoire souple**. À mesure que le tas vivant s'en approche, le GC travaille plus fort pour éviter l'*OOM*. Essentiel en conteneur : on la fixe **sous** la limite du conteneur. Couplée au `GOMAXPROCS` conscient des conteneurs (Go 1.25), elle rend le comportement de Go bien plus sain sous les *limits* Kubernetes ([§ 9.2](../09-conteneurs-cloud/02-kubernetes.md)).

Le principe à retenir : **le coût du GC croît avec le taux d'allocation et la taille du tas vivant**. Le levier le plus puissant n'est donc pas de régler `GOGC`, mais d'**allouer moins** ([§ 14.3](03-optimisations-pgo.md)).

---

## Le collecteur Green Tea (Go 1.26) 🆕

Expérimental en Go 1.25, **Green Tea est le collecteur par défaut depuis Go 1.26**. C'est une refonte de la phase de marquage motivée par le matériel moderne. Le GC traditionnel traite le tas comme un graphe plat : il suit les pointeurs **objet par objet**, sautant d'une zone mémoire à une autre au gré des références — d'où de nombreux défauts de cache, au point qu'**au moins 35 %** du temps de marquage était perdu à attendre la mémoire, et le problème s'aggrave avec le nombre de cœurs.

Green Tea devient **conscient de la mémoire** : au lieu de scanner des objets épars, il scanne la mémoire par **spans** — des blocs contigus de 8 KiB regroupant des objets de même taille — enfilés dans une file de travail. Ce parcours séquentiel exploite bien mieux les caches CPU et passe à l'échelle avec les cœurs. Résultat : une **réduction de 10 à 40 %** de l'*overhead* du GC sur les charges qui l'utilisent beaucoup (le gain est le plus net sur les tas réguliers et fortement ramifiés — arbres, graphes, index — et croît avec le nombre de cœurs), **sans changement de code**. Sur les amd64 récents (Intel Ice Lake, AMD Zen 4 et au-delà), une accélération vectorielle (SIMD) ajoute encore ~10 %.

Le repli existe — `GOEXPERIMENT=nogreenteagc` à la compilation — mais il est **destiné à disparaître en Go 1.27** : c'est le sens de l'histoire. Une nuance éclaire toute la section : Green Tea **ne crée pas de localité à partir de rien**. Si des objets qui se référencent mutuellement sont dispersés en mémoire, il retombe sur un comportement proche de l'ancien GC (optimisation « un seul objet par span »). Autrement dit, une bonne disposition mémoire reste payante — une raison de plus de soigner ses allocations.

---

## Observer et diagnostiquer

Pour rendre le GC et les allocations visibles :

- **`GODEBUG=gctrace=1`** imprime une ligne par cycle de GC — tailles de tas, durées de pause, part CPU :

```sh
$ GODEBUG=gctrace=1 ./app
gc 14 @3.207s 0%: 0.019+0.62+0.004 ms clock, ... 12->13->7 MB, 14 MB goal, ...
#     ^cycle          ^pauses (STW+concurrent+STW)  ^tas avant->après marquage->vivant  ^cible
```

- le **profil d'allocations** ([§ 14.1](01-pprof.md)) : `-alloc_space` / `-alloc_objects` pour localiser *où* l'on alloue ;
- **`-benchmem`** ([§ 13.4](../13-tests-qualite/04-fuzzing-benchmarks.md)) : les colonnes `B/op` et `allocs/op`, la métrique d'allocation du quotidien ;
- `runtime.ReadMemStats` et le package `runtime/metrics` pour des mesures programmatiques ;
- la **trace d'exécution** ([§ 14.1](01-pprof.md)) matérialise les pauses GC sur la chronologie.

---

## Réduire la pression mémoire

Le détail des techniques revient à [§ 14.3](03-optimisations-pgo.md), mais l'esprit est ici : **préallouer** avec une capacité connue (`make([]T, 0, n)`), **réutiliser** les tampons (`sync.Pool`, `bytes.Buffer`), éviter le *boxing* en interface et les pointeurs superflus dans les boucles chaudes, préférer les types valeur quand cela évite un échappement, et tenir `fmt` hors des chemins chauds. On règle `GOGC`/`GOMEMLIMIT` seulement quand c'est justifié — et toujours après avoir **mesuré** ([§ 14.1](01-pprof.md)) : alléger les allocations ne compte que sur le chemin chaud ; ailleurs, la clarté prime.

---

## Côté IDE : GoLand et VS Code

**Escape analysis.** Sous GoLand, on ajoute `-gcflags=-m` aux **Go tool arguments** de la configuration de compilation ; la vue d'allocations du profileur montre par ailleurs les allocations tas. Sous VS Code (extension Go), on règle `"go.buildFlags": ["-gcflags=-m"]`, ou l'on lance `go build -gcflags=-m ./...` au terminal.

**Trace GC.** Dans les deux éditeurs, on renseigne `GODEBUG=gctrace=1` dans les variables d'environnement de la configuration d'exécution pour suivre les cycles en direct. Le profilage d'allocations passe, lui, par pprof ([§ 14.1](01-pprof.md)) dans l'interface de chaque profileur.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [14.3 — Optimisations : `sync.Pool`, préallocation, PGO](03-optimisations-pgo.md)

⏭ [Optimisations : `sync.Pool`, préallocation, PGO](/14-performance/03-optimisations-pgo.md)
