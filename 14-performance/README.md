🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 14. Performance et gestion de la mémoire

Go vise une performance proche des langages compilés tout en offrant un ramasse-miettes et l'absence de gestion mémoire manuelle. La discipline de performance y est donc **empirique** : on mesure d'abord, on optimise ensuite — jamais l'inverse. Le goulot d'étranglement se trouve presque toujours ailleurs qu'on ne le croit, et une micro-optimisation à l'aveugle sacrifie la lisibilité sans gain réel.

Ce module prolonge directement le [§ 13.4](../13-tests-qualite/04-fuzzing-benchmarks.md) : après avoir appris à **écrire** des benchmarks, on apprend ici à les **exploiter**. Le fil est un cycle — profiler pour localiser le point chaud (pprof), comprendre ce qui alloue et comment le GC réagit, appliquer des optimisations ciblées, puis vérifier le gain de façon rigoureuse (benchstat). L'ordre compte : optimiser sans profil, c'est deviner ; conclure sans benchstat, c'est mesurer du bruit. Et l'on n'optimise que le chemin chaud identifié — le reste du code reste simple et idiomatique.

---

## 🎯 Objectifs du module

À l'issue de ce module, vous saurez :

- profiler un programme Go (CPU, tas, goroutines, blocage) avec pprof et lire les résultats ;
- comprendre comment le GC, les allocations et l'*escape analysis* façonnent la performance, et réduire les allocations là où c'est utile ;
- appliquer des optimisations ciblées (`sync.Pool`, préallocation, PGO) sans sacrifier la clarté ;
- comparer des résultats de benchmark de façon rigoureuse avec benchstat (significativité statistique).

---

## 🗺️ Plan du module

| # | Section | En bref |
|---|---------|---------|
| **14.1** ⭐ | [Profilage avec pprof (CPU, heap, goroutines)](01-pprof.md) | Profiler un programme en cours, lire un profil (`top`, `list`, *flame graph*) ; traces d'exécution et *flight recorder* (Go 1.25). |
| **14.2** | [Le GC de Go, allocations, escape analysis](02-gc-allocations.md) | Comment le GC, les allocations et l'*escape analysis* pèsent sur la perf ; réduire la pression mémoire ; GC **Green Tea** (Go 1.26). |
| **14.3** 🆕 | [Optimisations : `sync.Pool`, préallocation, PGO](03-optimisations-pgo.md) | Réutilisation d'objets (`sync.Pool`), préallocation de slices/maps, et optimisation guidée par profil (PGO). |
| **14.4** | [Benchmarking rigoureux (benchstat)](04-benchmarking.md) | Conclure sérieusement : répétitions, variance et significativité statistique avec benchstat. |

---

## 🧰 La boîte à outils de la performance

Tout le module gravite autour de quelques commandes de la chaîne standard, complétées par benchstat :

```sh
go test -bench=. -benchmem                          # mesurer : temps + allocations
go test -bench=. -cpuprofile=cpu.out -memprofile=mem.out  # produire des profils
go tool pprof -http=:8080 cpu.out                   # analyser (flame graph interactif)
go build -gcflags=-m ./...                          # escape analysis : où et pourquoi ça alloue
go build -pgo=default.pgo                           # optimisation guidée par profil
benchstat old.txt new.txt                           # comparer deux séries, avec p-value
```

Pour un service en cours d'exécution, l'import `_ "net/http/pprof"` expose les profils sur `/debug/pprof/` — de quoi profiler en production, sujet repris côté observabilité en [§ 12.4](../12-erreurs-debogage/04-observabilite.md).

---

## 🆕 Nouveautés Go 1.25 → 1.26

Trois évolutions récentes traversent ce module :

- **Traces d'exécution et *flight recorder* (Go 1.25)** — un enregistreur léger, toujours actif, qui conserve les derniers instants d'exécution dans un tampon circulaire et permet d'en capturer un instantané au moment où une anomalie survient. Complément de pprof, traité en [§ 14.1](01-pprof.md).
- **GC Green Tea activé par défaut (Go 1.26)** — le nouveau ramasse-miettes réduit l'*overhead* du GC de l'ordre de 10 à 40 % selon les charges, sans changement de code ; repli possible via `GOEXPERIMENT=nogreenteagc`. Détaillé en [§ 14.2](02-gc-allocations.md).
- **PGO — optimisation guidée par profil** — stabilisée et de plus en plus employée : un profil CPU de production guide l'*inlining* et la dévirtualisation du compilateur pour quelques pour-cent de gain « gratuit ». En [§ 14.3](03-optimisations-pgo.md).

---

## 🔗 Prérequis et suites

Ce module suppose acquis :

- l'écriture des [benchmarks](../13-tests-qualite/04-fuzzing-benchmarks.md) (§13.4), point de départ direct ;
- les [slices et maps](../02-fondamentaux-langage/07-slices-maps.md) (§2.7) et les [pointeurs](../02-fondamentaux-langage/08-pointeurs.md) (§2.8) — l'intuition des allocations en dépend ;
- les [goroutines et le *scheduler*](../04-concurrence/01-goroutines.md), pour lire un profil de goroutines et raisonner la perf sous concurrence.

Il se prolonge vers :

- l'[observabilité](../12-erreurs-debogage/04-observabilite.md) (§12.4) : exposer profils et métriques en production ;
- le [déploiement](../15-deploiement-devops/README.md), d'où proviennent les profils de production qui alimentent la PGO.

---

## Côté IDE : GoLand et VS Code

Les deux environnements s'appuient sur pprof ; l'ergonomie diffère.

- **GoLand** propose **Run with Profiler** (CPU, allocations…) sur un test, un benchmark ou une application : les résultats s'affichent en *flame graph* et arbre d'appels intégrés, et l'on peut aussi ouvrir un fichier de profil `.out` existant.
- **VS Code** (extension Go officielle) lance le profilage d'un benchmark via *CodeLens* et ouvre l'interface web de pprof dans le navigateur.

Dans les deux cas, l'interface web interactive reste accessible indépendamment de l'éditeur : `go tool pprof -http=:8080 profile.out` ouvre *flame graphs*, vue `top` et annotation ligne à ligne.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [14.1 — Profilage avec pprof (CPU, heap, goroutines)](01-pprof.md)

⏭ [Profilage avec pprof (CPU, heap, goroutines)](/14-performance/01-pprof.md)
