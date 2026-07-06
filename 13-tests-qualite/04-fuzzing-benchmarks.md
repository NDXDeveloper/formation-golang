🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 13.4 Fuzzing natif et benchmarks

Deux outils partagent le package `testing` et la commande `go test`, chacun prolongeant ce que les tables et les doublures ne couvrent pas. Le **fuzzing** génère automatiquement des entrées pour débusquer les bugs auxquels on n'a pas pensé. Les **benchmarks** mesurent la performance — temps et allocations par opération. Les deux sont natifs : aucun framework externe.

---

## Le fuzzing natif

### Le principe

Le fuzzing est intégré à Go depuis la 1.18. Il est **guidé par la couverture** : le moteur mute des entrées, conserve celles qui atteignent de nouveaux chemins de code, et traque celles qui font paniquer le programme, échouer une assertion ou violer un invariant.

Il complète les tests table-driven ([§ 13.1](01-tests-unitaires.md)) : la table vérifie les cas qu'on a imaginés, le fuzzing trouve ceux qu'on n'a pas vus — entrée vide, entrée géante, UTF-8 invalide, débordement d'entier. Il est particulièrement précieux pour tout ce qui analyse des données non maîtrisées : parseurs, décodeurs, entrées réseau ([§ 16.1](../16-securite/01-owasp-go.md)).

### Écrire un test de fuzz

Une fonction `func FuzzXxx(f *testing.F)` se compose de deux temps : un **corpus de graines** via `f.Add`, puis la **cible de fuzz** via `f.Fuzz`. Les arguments de la cible sont limités à un ensemble fixe de types : `[]byte`, `string`, les variantes de `int`/`uint`, `float32`/`float64`, `bool` et `rune` — pas de structs.

Que vérifier ? L'absence de panique est implicite ; au-delà, on affirme des **propriétés** : aller-retour (encoder puis décoder redonne l'original), différentiel (comparer à une implémentation de référence) ou cohérence (ne jamais produire un résultat mal formé). L'aller-retour est le plus courant :

```go
import (
	"testing"
	"unicode/utf8"
)

func FuzzReverse(f *testing.F) {
	f.Add("Golang") // graines : exemples de départ + cas de non-régression
	f.Add("café")
	f.Add("")

	f.Fuzz(func(t *testing.T, s string) {
		if !utf8.ValidString(s) {
			t.Skip() // Reverse suppose une entrée UTF-8 valide
		}
		rev := Reverse(s)
		if !utf8.ValidString(rev) {
			t.Errorf("Reverse(%q) a produit un UTF-8 invalide : %q", s, rev)
		}
		if got := Reverse(rev); got != s {
			t.Errorf("Reverse(Reverse(%q)) = %q, want %q", s, got, s)
		}
	})
}
```

Cette cible affirme deux invariants — la validité de la sortie et l'identité de l'aller-retour. C'est, à peu de choses près, l'exemple du tutoriel officiel, celui-là même qui a révélé un bug de gestion de l'UTF-8 dans une implémentation naïve de `Reverse`.

### Exécuter le fuzzing

Un test de fuzz a **deux modes d'exécution** :

- **Graines seules** (par défaut) — un simple `go test` exécute la cible comme un test normal, sur le seul corpus de graines. Rapide, déterministe : parfait en CI comme filet de non-régression.
- **Fuzzing actif** — `go test -fuzz=FuzzReverse` mute les entrées en continu jusqu'à une défaillance ou une interruption. C'est un processus **ouvert** : on le borne avec `-fuzztime=30s` (ou `-fuzztime=1000x`). Une seule cible et un seul package par exécution `-fuzz` ; on ajoute `-run=^$` pour ne pas rejouer les tests unitaires au passage.

```sh
go test -run=^$ -fuzz=FuzzReverse -fuzztime=30s
```

### Le corpus et la reproduction des bugs

Deux corpus coexistent : le **corpus de graines** (dans le code via `f.Add`, plus les fichiers versionnés sous `testdata/fuzz/FuzzXxx/`) et le **corpus généré** (mis en cache sous `$GOCACHE/fuzz`, local à la machine).

Quand le moteur trouve une défaillance, il **écrit l'entrée fautive** dans `testdata/fuzz/FuzzXxx/<hash>` et affiche une commande de reproduction. Ce fichier se commite et devient un **cas de non-régression permanent** : dès lors, un simple `go test` le rejoue de façon déterministe. C'est toute la boucle — *le fuzzing découvre → le cas fautif est enregistré comme graine → `go test` le reproduit*.

En CI, le corpus de graines tourne dans la suite normale (rapide) ; le fuzzing continu s'exécute à part, dans un job dédié et borné dans le temps (nocturne via `-fuzztime`, ou OSS-Fuzz pour l'open source). C'est un outil de sécurité de première ligne pour le code d'analyse d'entrées ([§ 16](../16-securite/README.md)) ; la mise en place en pipeline relève de [§ 15.2](../15-deploiement-devops/02-cicd.md).

---

## Les benchmarks (`go test -bench`)

### Anatomie d'un benchmark

Un benchmark est une fonction `func BenchmarkXxx(b *testing.B)`, lancée par `go test -bench=.`. Historiquement, on écrivait une boucle de `0` à `b.N`. **Depuis Go 1.24, on préfère `for b.Loop() { … }`**, plus robuste et plus sûr :

```go
func BenchmarkReverse(b *testing.B) {
	s := strings.Repeat("café ", 1000) // préparé une seule fois, hors chrono
	for b.Loop() {
		Reverse(s)
	}
}
```

`b.Loop` corrige d'un coup les trois pièges classiques du style `b.N` :

1. il **réinitialise le chronomètre** à son premier appel et **l'arrête** quand il renvoie `false` — la préparation avant la boucle et le nettoyage après sont automatiquement exclus de la mesure, sans `b.ResetTimer`/`b.StopTimer` manuels ;
2. il **maintient en vie les opérandes du corps de boucle** (arguments, résultats et variables assignées entre les accolades), empêchant le compilateur d'éliminer le travail comme code mort — le vieux truc du *sink* (`_ = f(x)`) devient inutile ;
3. sa **montée en charge « one-shot »** exécute la fonction une seule fois plutôt que de la relancer avec un `b.N` croissant : c'est plus rapide et plus prévisible. Dans la boucle, `b.N` vaut `0` ; après le retour de `Loop`, il contient le nombre total d'itérations, utile pour calculer des métriques dérivées.

Deux règles : on utilise `Loop` **ou** une boucle `b.N`, jamais les deux ; et l'on ne fait ni `break`/`return` hors d'une boucle `b.Loop()`, ni `b.StopTimer()` en son sein sans la relancer — ces usages produisent des résultats aberrants (Go en détecte une partie). À noter enfin la nuance 1.24 → 1.26 : en 1.24, le maintien en vie était obtenu en **interdisant l'inlining** dans la boucle, ce qui pouvait introduire une allocation tas parasite absente du code réel ; **Go 1.26 lève cette restriction**, si bien que tout benchmark en style `b.N` peut migrer vers `b.Loop` sans effet de bord.

Pour mémoire, la forme antérieure à 1.24 reste valide (par exemple sous un `go.mod` plus ancien) :

```go
func BenchmarkReverse(b *testing.B) {
	s := strings.Repeat("café ", 1000)
	b.ResetTimer()        // exclut la préparation du chrono
	for range b.N {
		_ = Reverse(s)    // le sink évite l'élimination de code mort
	}
}
```

### Lancer et lire les résultats

```sh
go test -bench=. -benchmem
go test -bench=Reverse -benchmem -count=10   # 10 mesures, pour l'analyse statistique
```

Une ligne de résultat se lit ainsi :

```
BenchmarkReverse-8        39114             30655 ns/op           26624 B/op          2 allocs/op
```

De gauche à droite : le nom suivi de `-8` (la valeur de `GOMAXPROCS` durant l'exécution), le nombre d'itérations retenues, puis **`ns/op`** (temps par opération), **`B/op`** et **`allocs/op`** (octets et allocations par opération, affichés par `-benchmem` ou `b.ReportAllocs()`). Sur ces trois métriques, plus bas vaut mieux.

Drapeaux utiles : `-benchtime=10s` ou `-benchtime=100x` (durée ou nombre fixe), `-count=N` (répétitions pour la stabilité), `-cpu=1,2,4` (faire varier le parallelisme), `-run=^$` (ignorer les tests). `-benchmem` ajoute les statistiques d'allocation.

### Mesurer avec justesse

La préparation coûteuse doit rester **hors de la boucle** (automatique avec `b.Loop` ; via `b.ResetTimer` en style `b.N`). Les **allocations** sont souvent le chiffre le plus actionnable en Go, car elles nourrissent la pression sur le GC ([§ 14.2](../14-performance/02-gc-allocations.md)) : on les expose avec `b.ReportAllocs()` ou `-benchmem`. Pour des mesures sur mesure, `b.ReportMetric(x, "unité/op")`. Pour mesurer le débit sous concurrence, `b.RunParallel(func(pb *testing.PB){ for pb.Next() { … } })`. Enfin, on balaie les tailles d'entrée avec des sous-benchmarks — `b.Run("taille=1k", func(b *testing.B){ … })` —, exactement comme une table de tests.

### D'une mesure à une conclusion

Une exécution isolée est du bruit. Pour affirmer que la variante A est plus rapide que B, on lance `-count=10` (ou plus) sur les deux et l'on compare avec **benchstat**, qui indique si l'écart est statistiquement significatif — la démarche rigoureuse est détaillée en [§ 14.4](../14-performance/04-benchmarking.md). Pour comprendre **pourquoi** un chemin chaud est lent, on le profile : `go test -bench=. -cpuprofile=cpu.out -memprofile=mem.out`, puis analyse avec pprof ([§ 14.1](../14-performance/01-pprof.md)). Réduction des allocations et *escape analysis* : [§ 14.2](../14-performance/02-gc-allocations.md) ; `sync.Pool`, préallocation et PGO : [§ 14.3](../14-performance/03-optimisations-pgo.md). Cette section montre **comment les écrire** ; le [module 14](../14-performance/README.md) montre **comment les exploiter**.

---

## Côté IDE : GoLand et VS Code

**GoLand.** Une icône ▶ dans la gouttière lance chaque fonction `Benchmark` ou `Fuzz`. **Run with Profiler** (CPU/mémoire) branche directement pprof sur un benchmark isolé, et les résultats s'affichent dans la fenêtre d'exécution. Une cible de fuzz se lance aussi depuis la gouttière — on pense à fixer un `-fuzztime` dans la configuration pour qu'elle se termine.

**VS Code** (extension Go officielle). Un *CodeLens* surmonte chaque fonction `Benchmark` et `Fuzz (« run benchmark » / « run fuzz test »)` ; `"go.testFlags"` et `"go.testTimeout"` transmettent `-benchmem`, `-fuzztime`, etc. Le profilage et la sortie des benchmarks passent par l'extension.

Dans les deux cas, le fuzzing étant **ouvert**, on le borne toujours (`-fuzztime`) faute de quoi il tourne jusqu'à interruption.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [13.5 — Linters : `go vet`, staticcheck, golangci-lint](05-linters.md)

⏭ [Linters : `go vet`, staticcheck, golangci-lint](/13-tests-qualite/05-linters.md)
