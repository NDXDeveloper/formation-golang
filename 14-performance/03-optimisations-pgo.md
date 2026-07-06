🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 14.3 Optimisations : `sync.Pool`, préallocation, PGO

C'est l'étape « agir » du cycle du module : après avoir **profilé** ([§ 14.1](01-pprof.md)) et **compris** ce qui alloue ([§ 14.2](02-gc-allocations.md)), on optimise — puis on **vérifie** ([§ 14.4](04-benchmarking.md)). Chaque technique ci-dessous ne s'applique qu'à un chemin chaud **identifié par la mesure**, jamais à l'aveugle : contorsionner du code lisible pour un gain qui n'existe pas est un mauvais échange. Deux familles se dégagent : **allouer moins** (préallocation, réutilisation, `sync.Pool`) et **laisser le compilateur en faire plus** (PGO).

---

## Préallocation

La cause d'allocations la plus fréquente et la plus facile à corriger est un tableau sous-jacent que `append` réalloue plusieurs fois faute de capacité initiale. Quand la taille finale est connue — ou estimable —, on la réserve d'emblée :

```go
// Sans : out est réalloué à mesure qu'append dépasse la capacité.
func namesSlow(users []User) []string {
	var out []string
	for _, u := range users {
		out = append(out, u.Name)
	}
	return out
}

// Avec : une seule allocation, dimensionnée dès le départ.
func namesFast(users []User) []string {
	out := make([]string, 0, len(users)) // capacité connue
	for _, u := range users {
		out = append(out, u.Name)
	}
	return out
}
```

`go test -bench=. -benchmem` fait apparaître la différence sur `allocs/op`. Le même réflexe vaut pour les **maps** (`make(map[K]V, n)` évite les redimensionnements) et pour la construction de chaînes avec `strings.Builder` :

```go
var b strings.Builder
b.Grow(tailleEstimée)  // préalloue le tampon
for _, s := range parts {
	b.WriteString(s)    // pas de réallocation à chaque ajout
}
result := b.String()
```

Enfin, on **réutilise** un tampon entre itérations en le remettant à longueur nulle tout en conservant sa capacité :

```go
buf := make([]byte, 0, 4096)
for sc.Scan() {
	buf = buf[:0]                 // longueur 0, capacité gardée : zéro allocation par tour
	buf = append(buf, transform(sc.Bytes())...)
	// … utiliser buf
}
```

---

## `sync.Pool`

Quand un même objet coûteux à allouer (un gros tampon, une grande structure) est créé puis jeté en boucle sur un chemin chaud, `sync.Pool` en recycle les instances pour soulager le GC :

```go
var bufPool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

func handle(w http.ResponseWriter, r *http.Request) {
	b := bufPool.Get().(*bytes.Buffer)
	b.Reset()             // impératif : un objet réutilisé porte un état ancien
	defer bufPool.Put(b)

	// … écrire dans b, puis b.WriteTo(w)
}
```

Trois règles encadrent son usage. D'abord, **les objets du pool peuvent être récupérés par le GC à tout moment** (le pool est vidé au fil des cycles, avec un cache « victime » d'un cycle depuis Go 1.13) : c'est un réservoir d'objets **transitoires** sous charge, jamais un cache à durée de vie garantie, ni un pool de connexions. Ensuite, on **réinitialise** systématiquement avant de réutiliser. Enfin, `sync.Pool` **se mérite** : pour des objets bon marché, le coût du *boxing* en `any` et de la synchronisation dépasse le gain — on ne l'emploie que là où le profil montre des allocations lourdes et répétées, et l'on mesure l'effet.

---

## Autres réductions d'allocations

Quelques causes récurrentes, à traiter uniquement sur le chemin chaud : le *boxing* en interface (`any`) dans une boucle serrée peut allouer à chaque valeur ; `fmt.Sprintf` transforme ses arguments en `any` et alloue — hors des chemins chauds, on lui préfère `strconv.AppendInt` ou `strings.Builder` ; les conversions `[]byte` ↔ `string` copient (beaucoup d'API de la stdlib acceptent les deux formes pour l'éviter ; les conversions `unsafe` relèvent de l'annexe B, pas de l'idiome par défaut).

Bonne nouvelle : **le runtime en absorbe de plus en plus tout seul**. Go 1.26 génère des routines d'allocation spécialisées par taille pour les **petits objets (< 512 octets)**, réduisant leur coût d'allocation jusqu'à ~30 % ; combiné à l'allocation de tableaux de slices sur la pile ([§ 14.2](02-gc-allocations.md)), cela supprime automatiquement nombre d'allocations qu'on aurait autrefois pourchassées à la main — une raison de plus de **mesurer avant** d'optimiser.

---

## PGO — l'optimisation guidée par profil

La PGO (*Profile-Guided Optimization*) donne au compilateur un **vrai profil CPU** ([§ 14.1](01-pprof.md)) pour qu'il prenne de meilleures décisions sur les chemins réellement chauds. Introduite en aperçu en Go 1.20, activée par défaut depuis la 1.21 et stabilisée avec une dévirtualisation et un *inlining* plus agressifs en 1.24, elle ne demande **aucun changement de code** :

```sh
# 1. Collecter un profil CPU représentatif (production de préférence, cf. §14.1)
curl -o default.pgo 'http://prod:6060/debug/pprof/profile?seconds=30'

# 2. Le placer dans le répertoire du package main, sous le nom exact default.pgo
mv default.pgo ./cmd/app/

# 3. Construire : go build le détecte et l'utilise (défaut : -pgo=auto)
go build ./cmd/app
```

Le fichier attendu est un profil pprof CPU — celui-là même que produisent `runtime/pprof` ou `net/http/pprof` ; on le **commite** avec le code. Deux mécanismes portent l'essentiel du gain : l'**inlining** des fonctions chaudes, et la **dévirtualisation** — convertir un appel de méthode d'interface, quand le type concret est prévisible sur ce chemin, en appel direct (souvent *inlinable* à son tour), avantage notable pour du Go riche en interfaces.

Le gain typique est de **2 à 14 %**, très dépendant de la charge (Datadog rapporte jusqu'à 14 % de CPU, Uber ~4 %) — donc à **mesurer** sur son propre programme. Point capital : le profil doit être **représentatif** du comportement de production ; un profil non représentatif ne rapporte quasiment rien. C'est pourquoi on le collecte en production, éventuellement en continu : des outils comme Datadog PGO ou Grafana Pyroscope agrègent les profils du parc en un `default.pgo` mis à jour automatiquement — le prolongement direct du profilage continu ([§ 12.4](../12-erreurs-debogage/04-observabilite.md)).

---

## Vérifier, toujours

Aucune de ces optimisations ne se déclare victorieuse sans preuve. On la confirme avec **benchstat** ([§ 14.4](04-benchmarking.md)) sur des séries répétées (`-count=N`), et l'on **reprofile** (`-base`, [§ 14.1](01-pprof.md)) pour vérifier que le point chaud s'est bien déplacé. Une « optimisation » qui ne montre pas d'amélioration statistiquement significative n'en est pas une : on la retire, au profit de la lisibilité.

---

## Côté IDE : GoLand et VS Code

**PGO.** Sous GoLand, on ajoute `-pgo=default.pgo` (ou l'on s'appuie sur l'auto-détection) dans les **Go tool arguments** de la configuration de compilation, et le profileur fournit le profil CPU à lui donner. Sous VS Code (extension Go), on règle `"go.buildFlags": ["-pgo=auto"]` et l'on collecte le profil via le profilage de l'extension.

**Vérification des allocations.** Le profilage d'allocations et `-benchmem` s'exécutent dans l'interface de profilage de chaque éditeur. Dans les deux cas, la boucle reste la même : profiler ([§ 14.1](01-pprof.md)) → appliquer → comparer avec benchstat ([§ 14.4](04-benchmarking.md)).

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [14.4 — Benchmarking rigoureux (benchstat)](04-benchmarking.md)

⏭ [Benchmarking rigoureux (benchstat)](/14-performance/04-benchmarking.md)
