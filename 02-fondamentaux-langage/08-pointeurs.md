🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 2.8 Pointeurs, sémantique valeur vs référence

Go a des pointeurs, mais domestiqués : ils portent une **adresse**, se prennent avec `&` et se déréférencent avec `*`, et — c'est capital — **ne supportent aucune arithmétique**. Les comprendre va de pair avec la règle la plus importante du langage : **tout est passé par valeur**. Cette section explique les deux, et éclaire au passage pourquoi slices et maps se comportent « comme des références ».

*(Le choix d'un receveur valeur ou pointeur pour les méthodes, décision voisine, est traité en [section 3.1](../03-types-interfaces/01-structs-methodes.md).)*

## Qu'est-ce qu'un pointeur

Un pointeur contient l'**adresse mémoire** d'une valeur. Son type s'écrit `*T` (pointeur vers un `T`). L'opérateur `&` prend l'adresse d'une variable ; l'opérateur `*` **déréférence** le pointeur pour accéder à la valeur.

```go
x := 42
p := &x         // p : *int, adresse de x
fmt.Println(*p) // 42 — déréférencement
*p = 7          // modifie x à travers p
fmt.Println(x)  // 7

var q *int      // zéro-value : nil
// fmt.Println(*q) // PANIQUE : déréférencement d'un pointeur nil
```

La zéro-value d'un pointeur est `nil`, et **déréférencer un pointeur `nil` provoque une panique**. Surtout, **Go n'autorise aucune arithmétique de pointeur** : pas de `p++`, pas de `p + 1`. C'est ce qui rend les pointeurs sûrs. (Le package `unsafe` permet des manipulations de bas niveau, mais c'est un anti-pattern à éviter — voir [annexe B](../annexes/go-idiomatique/README.md).)

## Tout est passé par valeur

En Go, **un argument est toujours copié** lors d'un appel. Pour un type valeur (entier, structure, tableau), la fonction reçoit une copie et **ne peut pas modifier la variable de l'appelant**. Pour lui permettre de la modifier, on passe un **pointeur**.

```go
func doubleValeur(n int)    { n *= 2 } // reçoit une COPIE : sans effet au-dehors
func doublePointeur(n *int) { *n *= 2 } // modifie la valeur pointée

x := 10
doubleValeur(x)
fmt.Println(x) // 10 — inchangé

doublePointeur(&x)
fmt.Println(x) // 20 — modifié
```

## Slices et maps : pourquoi elles semblent « par référence »

La règle « passage par valeur » vaut aussi pour les slices et les maps. Mais leur **valeur** est un petit **en-tête** contenant un **pointeur** vers des données partagées. Copier cet en-tête copie le pointeur — donc les données sous-jacentes restent partagées. Modifier un élément existant est visible pour l'appelant.

En revanche — subtilité fréquente — une opération qui **remplace l'en-tête** (comme `append` provoquant une réallocation, ou une réaffectation de la slice) reste invisible, car la fonction ne détient qu'une **copie de l'en-tête** :

```go
func remplir(s []int) { s[0] = 99 }        // modifie un élément : visible
func ajouter(s []int) { s = append(s, 1) } // réaffecte l'en-tête (copie) : INVISIBLE

nums := []int{0, 0, 0}
remplir(nums)
fmt.Println(nums) // [99 0 0] — la modification est visible

ajouter(nums)
fmt.Println(nums) // [99 0 0] — l'append n'a PAS d'effet sur l'appelant
```

Pour qu'un `append` soit visible, on **renvoie** la slice (`nums = ajouter(nums)`) ou l'on passe un **pointeur de slice** (`*[]int`). Ce comportement prolonge directement les pièges vus en [section 2.7](07-slices-maps.md).

## Créer un pointeur

Plusieurs façons d'obtenir un pointeur :

```go
p := &x                  // adresse d'une variable existante
pt := &Point{X: 1, Y: 2} // pointeur vers une structure littérale (structs : section 3.1)
q := new(int)            // *int pointant vers la zéro-value (0)
```

🆕 Rappel : **depuis Go 1.26, `new` accepte une expression d'initialisation** ([section 2.2](02-types-variables.md)), ce qui donne directement un pointeur vers une valeur — pratique pour un **champ optionnel** :

```go
r := new(42) // *int pointant vers 42
```

Un point rassurant propre à Go : **renvoyer l'adresse d'une variable locale est sûr**. Grâce à l'analyse d'échappement, la valeur est automatiquement placée sur le tas si nécessaire ; il n'y a pas de pointeur fou (*dangling pointer*) comme en C (voir [section 14.2](../14-performance/02-gc-allocations.md)).

## Quand utiliser des pointeurs

On recourt à un pointeur pour :

- **modifier** la valeur de l'appelant ;
- **éviter de copier** une grande structure (performance) ;
- représenter une valeur **optionnelle** (le `nil` signalant « absente »).

En contrepartie, un pointeur ajoute de l'indirection, peut provoquer une allocation sur le tas (pression sur le GC) et introduit le risque du `nil` et de l'aliasing. L'habitude idiomatique : **passer les petites valeurs par copie** (plus simple et sûr), et réserver les pointeurs à la mutation ou aux grandes structures. Cette même décision se repose pour le **receveur des méthodes**, en [section 3.1](../03-types-interfaces/01-structs-methodes.md).

## En résumé

Un pointeur porte une **adresse**, s'obtient avec `&`, se déréférence avec `*`, vaut `nil` par défaut, et **ne connaît pas l'arithmétique**. Comme **tout est passé par valeur**, on passe un **pointeur pour modifier** la valeur de l'appelant. Slices et maps semblent « par référence » parce que leur en-tête pointe vers des données partagées — d'où la subtilité de l'`append` invisible. Reste le dernier pilier des fondamentaux, et l'un des plus idiomatiques : la **gestion des erreurs**, en [section suivante](09-gestion-erreurs.md).

---

[🔝 Sommaire](../SOMMAIRE.md) · [⏭️ Section suivante : 2.9 Gestion des erreurs](09-gestion-erreurs.md)

⏭ [Gestion des erreurs — l'idiome Go](/02-fondamentaux-langage/09-gestion-erreurs.md)
