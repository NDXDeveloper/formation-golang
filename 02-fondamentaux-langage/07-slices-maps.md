🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 2.7 Tableaux, slices et maps

Les **slices** et les **maps** sont les deux collections qu'on manipule sans cesse en Go — et les slices, en particulier, concentrent les pièges les plus notoires du langage. Tous découlent d'un même fait : **une slice est une vue sur un tableau sous-jacent partagé**. Maîtriser la distinction longueur / capacité, le comportement d'`append` et les pièges de partage est donc essentiel. Les tableaux (taille fixe) existent aussi, mais on les emploie rarement en direct.

*(La sémantique valeur vs référence, esquissée ici, est complétée en [section 2.8](08-pointeurs.md).)*

## Les tableaux : taille fixe

En Go, la **taille fait partie du type** : `[3]int` et `[4]int` sont deux types distincts. Un tableau se copie **intégralement** à l'affectation ou au passage en argument (sémantique de valeur).

```go
var a [3]int          // [0 0 0]
b := [3]int{1, 2, 3}
c := [...]int{1, 2, 3} // le compilateur compte : c'est un [3]int

d := b   // COPIE complète
d[0] = 9 // b[0] vaut toujours 1
```

En pratique, on utilise surtout les slices ; les tableaux servent de support (par exemple pour une empreinte de taille fixe).

## Les slices : une vue sur un tableau

Une **slice** est un descripteur léger : un **pointeur** vers un tableau sous-jacent, une **longueur** et une **capacité**. C'est la collection de travail par défaut.

```go
s := []int{1, 2, 3}     // littéral
t := make([]int, 2)     // longueur 2, capacité 2 → [0 0]
u := make([]int, 0, 10) // longueur 0, capacité 10 (préallouée)

fmt.Println(len(s), cap(s)) // 3 3
```

- La **longueur** (`len`) est le nombre d'éléments accessibles.
- La **capacité** (`cap`) est le nombre d'éléments depuis le début de la slice jusqu'à la fin du tableau sous-jacent.
- La zéro-value d'une slice est `nil` : une slice `nil` a une longueur nulle mais reste utilisable avec `append` et `range`. Une slice `nil` (`var s []int`) et une slice vide (`[]int{}`) sont toutes deux exploitables ; `nil` est l'idiome pour « aucun élément ».

## Capacité et `append`

`append` ajoute des éléments et **renvoie une slice (éventuellement nouvelle)** : il faut donc **réaffecter le résultat**.

```go
s := make([]int, 0, 2)
fmt.Println(len(s), cap(s)) // 0 2
s = append(s, 1, 2)         // tient dans la capacité
fmt.Println(len(s), cap(s)) // 2 2
s = append(s, 3)            // dépasse la capacité → réallocation
fmt.Println(len(s), cap(s)) // 3 4
```

Le mécanisme est le suivant : si la **capacité suffit**, `append` écrit dans le tableau existant (qui peut être partagé, voir plus bas) ; sinon, il **alloue un nouveau tableau plus grand** et y **recopie** les éléments. La façon dont la capacité croît est un détail d'implémentation sur lequel il ne faut pas compter. Quand la taille finale est connue, **préallouer** avec `make([]T, 0, n)` évite ces réallocations.

## Les pièges des slices ⚠️

**Partage du tableau sous-jacent.** Découper une slice ne copie rien : la sous-slice partage le tableau. Modifier l'une affecte l'autre.

```go
base := []int{10, 20, 30, 40}
vue := base[1:3] // partage le tableau : [20 30]
vue[0] = 99
fmt.Println(base) // [10 99 30 40] — base est modifié !
```

**Surprise d'`append`.** Ajouter à une slice qui dispose de capacité libre écrit dans le tableau partagé — et peut écraser les données d'une autre slice.

```go
a := []int{1, 2, 3, 4, 5}
b := a[:2]        // longueur 2, capacité 5 (partage le tableau de a)
b = append(b, 99) // capacité suffisante → écrit dans le tableau de a
fmt.Println(a)    // [1 2 99 4 5] — a[2] a été écrasé !
```

**Découpage à trois indices.** Pour éviter ce partage, on peut **borner la capacité** avec `s[bas:haut:max]` : le prochain `append` forcera alors une copie.

```go
b := a[0:2:2]     // longueur 2, capacité 2
b = append(b, 99) // réalloue → a est préservé
```

**Rétention mémoire.** Conserver une petite sous-slice d'un grand tableau **maintient tout le tableau en mémoire**. Pour détacher une copie indépendante, on utilise `slices.Clone` (ou `copy`).

```go
petit := slices.Clone(gros[:3]) // copie détachée ; le grand tableau peut être libéré
```

À retenir : **une slice ne « possède » pas ses données** — elle les regarde. Dès qu'un partage inattendu est possible, on copie explicitement.

## Le package `slices`

Le package `slices` fournit des utilitaires génériques pour les opérations courantes :

```go
slices.Contains([]int{1, 2, 3}, 2) // true
slices.Sort(nums)                  // tri en place
clone := slices.Clone(nums)        // copie indépendante
```

S'y ajoutent `Index`, `Equal`, `Insert`, `Delete`, `Reverse`, `Max`, `Min`, etc. (Les itérateurs `slices.Values`/`slices.All` ont été vus en [section 2.5](05-boucles.md).)

## Les maps

Une **map** associe des clés à des valeurs, **sans ordre garanti** : `map[K]V`.

```go
m := map[string]int{"a": 1, "b": 2}
n := make(map[string]int)

m["c"] = 3
absent := m["absent"] // 0 (zéro-value de la valeur) — pas de panique en lecture
v, ok := m["a"]       // v=1, ok=true ; le motif « comma-ok » distingue absent et zéro
delete(m, "b")
fmt.Println(len(m))
```

Pour **vider** une map de toutes ses entrées, la fonction intégrée `clear(m)` (Go 1.21) est plus directe qu'une boucle de `delete`. Elle vaut aussi pour les slices : `clear(s)` y réinitialise chaque élément à sa zéro-value, sans changer la longueur.

Trois points essentiels :

- Lire une **clé absente** renvoie la zéro-value du type de valeur, sans erreur. Le motif `v, ok := m[k]` permet de distinguer « absent » de « présent avec la valeur nulle ».
- **Piège de la map `nil`** : la zéro-value d'une map est `nil`. On peut la **lire** (tout renvoie la zéro-value), mais **écrire dedans provoque une panique**. Une map doit être initialisée avec `make` (ou un littéral) avant toute écriture.

```go
var p map[string]int // nil
_ = p["x"]           // OK : lecture → 0
// p["x"] = 1        // PANIQUE : écriture dans une map nil
```

- Les **clés doivent être comparables** (utilisables avec `==`) : pas de slice ni de map comme clé (voir `comparable` en [section 3.4](../03-types-interfaces/04-generiques.md)). L'ordre de parcours est **volontairement aléatoire** ([section 2.5](05-boucles.md)).

Le package `maps` complète l'ensemble (`maps.Clone`, `maps.Equal`, `maps.DeleteFunc`, et les itérateurs `maps.Keys`/`maps.Values`).

## Sémantique valeur vs référence

En résumé de ce qui précède : un **tableau** se copie (valeur), tandis qu'une **slice** ou une **map** se comporte comme une **référence** — copier ou passer l'en-tête partage les données sous-jacentes. C'est pourquoi une fonction peut modifier le contenu d'une slice ou d'une map qu'on lui passe. Ce modèle est détaillé, aux côtés des pointeurs, en [section 2.8](08-pointeurs.md).

## En résumé

Une **slice** est une vue (pointeur + longueur + capacité) sur un tableau partagé : d'où l'importance de comprendre `append` (réaffectation, réallocation) et les **pièges de partage**, que l'on désamorce en copiant explicitement (`slices.Clone`, découpage à trois indices). Une **map** associe des clés comparables à des valeurs, sans ordre, avec le motif *comma-ok* et le piège de la map `nil` en écriture. Slices et maps ayant une sémantique de référence, la suite logique est l'étude des **pointeurs** : [section suivante](08-pointeurs.md).

---

[🔝 Sommaire](../SOMMAIRE.md) · [⏭️ Section suivante : 2.8 Pointeurs](08-pointeurs.md)

⏭ [Pointeurs (sans arithmétique), sémantique valeur vs référence](/02-fondamentaux-langage/08-pointeurs.md)
