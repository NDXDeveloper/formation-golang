🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 13. Tests et qualité du code

En Go, tester n'est pas une activité annexe confiée à un framework tiers : c'est une capacité de première classe intégrée à la chaîne d'outils. La commande `go test` découvre seule les fichiers `*_test.go` et les fonctions `TestXxx`, sans *runner* externe ni fichier de configuration. Cette intégration façonne toute une culture : le *table-driven test* comme idiome dominant, les petites interfaces qui rendent le *mocking* trivial sans bibliothèque dédiée, le fuzzing et le *benchmarking* livrés dans le même package `testing`, et une couverture de code obtenue d'un simple `-cover`.

Ce module couvre les tests au sens large — unitaires, d'intégration, fuzzing, benchmarks — puis la qualité statique (linters) et la mesure (couverture). Le tout dans l'esprit **« stdlib d'abord »** : on part de ce que la bibliothèque standard offre avant d'ajouter `testify` ou `golangci-lint` là où ils apportent réellement quelque chose. On visera des tests **lisibles et déterministes** plutôt qu'une couverture maximale affichée comme un trophée.

---

## 🎯 Objectifs du module

À l'issue de ce module, vous saurez :

- écrire des tests unitaires idiomatiques (table-driven, sous-tests via `t.Run`) avec le seul package `testing` ;
- isoler les dépendances par de petites interfaces et des *fakes*, et savoir quand `testify` ou `httptest` valent le détour ;
- monter des tests d'intégration réalistes avec Testcontainers (bases de données réelles et jetables) ;
- débusquer des cas limites par le fuzzing natif et mesurer les performances avec des benchmarks ;
- automatiser l'analyse statique (`go vet`, staticcheck, golangci-lint) et la modernisation du code (`go fix`) ;
- lire une couverture de code sans en faire une religion, et cadrer l'usage de l'IA pour générer et relire des tests.

---

## 🗺️ Plan du module

| # | Section | En bref |
|---|---------|---------|
| **13.1** ⭐ | [Package `testing`, table-driven tests, sous-tests](01-tests-unitaires.md) | Anatomie d'un test, `t.Error`/`t.Fatal`, `t.Helper()`, l'idiome table-driven, `t.Run` et `t.Parallel`. |
| **13.2** | [Mocks par interfaces, testify, `httptest`](02-mocks-testify.md) | Isoler par interfaces et *fakes* écrits à la main ; quand passer à `testify` ; tester du HTTP avec `httptest` ; génération de mocks pilotée par `go generate`. |
| **13.3** | [Tests d'intégration (Testcontainers, bases de données)](03-tests-integration.md) | Aller au-delà du *fake* : lancer une vraie base (PostgreSQL, Redis) en conteneur jetable, gérer *setup*/*teardown* et parallélisme. |
| **13.4** | [Fuzzing natif et benchmarks (`go test -bench`)](04-fuzzing-benchmarks.md) | Trouver les cas limites automatiquement (`FuzzXxx`, corpus, *seed*) ; mesurer avec `BenchmarkXxx` (`b.Loop`, `-benchmem`). |
| **13.5** | [Linters : `go vet`, staticcheck, golangci-lint](05-linters.md) | `go vet` intégré, staticcheck, l'agrégateur golangci-lint (config, CI), et le nouveau `go fix` (modernizers, `//go:fix inline`). |
| **13.6** 🤖 | [Couverture de code ; génération de tests par IA](06-couverture-tests-ia.md) | `-cover`/`-coverprofile`, `go tool cover` et ses pièges ; générer et relire des tests avec l'IA, avec garde-fous. |

---

## 🧰 La boîte à outils en une commande

Tout le module gravite autour d'un même point d'entrée, `go test`, épaulé par `go vet` et `go fix` :

```sh
go test ./...                 # lancer tous les tests du module
go test -run TestFoo ./...    # cibler un test (regex)
go test -v ./...              # sortie détaillée (un test = une ligne)
go test -race ./...           # activer le détecteur de data races
go test -cover ./...          # taux de couverture
go test -bench=. -benchmem    # benchmarks + statistiques d'allocation
go test -fuzz=FuzzFoo         # campagne de fuzzing (par package)
go vet ./...                  # analyse statique intégrée
go fix ./...                  # modernisation automatique du code (Go 1.26)
```

Aucune de ces commandes ne requiert d'installation supplémentaire : elles font partie de la distribution Go. Les outils tiers (staticcheck, golangci-lint, Testcontainers, `testify`) viennent en complément, pas en remplacement.

---

## 🆕 Nouveautés Go 1.25 → 1.26

Deux apports récents changent concrètement la façon de tester et d'entretenir un projet Go.

**`testing/synctest` — tester la concurrence sans *flakiness* (stable en Go 1.25).** Expérimental en 1.24, ce package est stable depuis Go 1.25. Il exécute un test dans une « bulle » isolée où l'horloge est virtualisée : le temps n'avance que lorsque *toutes* les goroutines sont durablement bloquées. Résultat, les tests qui dépendent de `time.Sleep`, `context.WithTimeout` ou de *timers* deviennent déterministes et quasi instantanés — fini le test qui passe en local et échoue une fois sur cent en CI. L'API tient en deux fonctions, `synctest.Test` et `synctest.Wait` (l'ancien `synctest.Run` de 1.24 est déprécié). Ce sujet est traité côté concurrence en [§ 4.6](../04-concurrence/06-tester-code-concurrent.md) et réutilisé ici pour les tests applicatifs.

**`go fix` réoutillé — garder un code idiomatique (Go 1.26).** La commande `go fix` a été entièrement réécrite sur le framework d'analyse de `go vet`. Elle embarque désormais une vingtaine de *modernizers* qui réécrivent sans risque les motifs datés vers les idiomes récents (`interface{}` → `any`, boucles manuelles → fonctions du package `maps`, helpers maison → `min`/`max` intégrés…), plus une directive `//go:fix inline` qui permet à un auteur de bibliothèque d'automatiser la migration de son API chez ses utilisateurs. Point notable à l'ère de l'IA : le code produit par les assistants tend à être « ancien » (idiomes d'avant les génériques, d'avant le package `maps`), et `go fix` le remet mécaniquement au goût du jour. Détails, configuration et usage en CI en [§ 13.5](05-linters.md).

---

## 🔗 Prérequis et suites

Ce module suppose acquis :

- la [gestion des erreurs](../02-fondamentaux-langage/09-gestion-erreurs.md) (§2.9) — on teste énormément de chemins d'erreur ;
- les [interfaces implicites](../03-types-interfaces/03-interfaces.md) (§3.3) — socle du *mocking* par interfaces ;
- la [concurrence](../04-concurrence/README.md), en particulier le [test du code concurrent](../04-concurrence/06-tester-code-concurrent.md) (§4.6).

Il se prolonge naturellement vers :

- la [performance](../14-performance/README.md)  : les benchmarks écrits ici s'exploitent avec `benchstat` en §14.4 ;
- le [CI/CD](../15-deploiement-devops/02-cicd.md) (§15.2) : tests, `-race` et linters s'y exécutent à chaque *push*.

---

## Côté IDE : GoLand et VS Code

Les deux environnements exécutent `go test` sous le capot ; seule l'ergonomie diffère.

- **GoLand** : icônes ▶ dans la gouttière, à gauche de chaque `TestXxx` et de chaque `t.Run`, pour lancer un test ou un sous-test isolément ; fenêtre *Run* présentant l'arbre des sous-tests ; couverture colorée dans l'éditeur via **Run with Coverage**.
- **VS Code** (extension Go officielle) : *CodeLens* **run test | debug test** au-dessus de chaque fonction de test ; **Test Explorer** latéral ; surbrillance de couverture via **Go: Toggle Test Coverage**.

Dans les deux cas, le débogage d'un test passe par Delve (voir [§ 12.2](../12-erreurs-debogage/02-debogage-delve.md)).

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [13.1 — Package `testing`, table-driven tests](01-tests-unitaires.md)

⏭ [Package `testing`, table-driven tests ⭐, sous-tests](/13-tests-qualite/01-tests-unitaires.md)
