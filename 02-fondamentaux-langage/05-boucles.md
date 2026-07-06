🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 2.5 Boucles et itérateurs

Go n'a qu'un seul mot-clé de boucle — **`for`** — qui couvre tous les besoins. Associé à **`range`**, il parcourt les collections ; et depuis Go 1.23, `range` sait aussi piloter des **itérateurs définis par l'utilisateur** (*range-over-func*). Deux évolutions récentes rendent cette section actuelle : la **variable de boucle par itération** (Go 1.22) et le **motif d'itérateur** (Go 1.23).

## Le `for`, unique boucle de Go

Il n'y a ni `while` ni `do-while` : `for` prend trois formes.

```go
// 1. Forme complète (init ; condition ; post)
for i := 0; i < n; i++ {
	// …
}

// 2. Condition seule (l'équivalent d'un « while »)
for cond {
	// …
}

// 3. Boucle infinie (avec break)
for {
	if done {
		break
	}
}
```

`break` et `continue` fonctionnent comme attendu. Pour les boucles imbriquées, une **étiquette** permet de rompre ou de continuer la boucle externe :

```go
Recherche:
	for _, ligne := range grille {
		for _, cellule := range ligne {
			if cellule == cible {
				break Recherche // sort des deux boucles
			}
		}
	}
```

*(Pour être complet : Go possède aussi un `goto` vers une étiquette de la même fonction. Son usage est rarissime — on le croise surtout dans du code généré — et les `break`/`continue` étiquetés couvrent l'essentiel des besoins légitimes.)*

## La variable de boucle (depuis Go 1.22) 🆕

**Depuis Go 1.22, chaque itération dispose de sa propre variable de boucle.** Auparavant, cette variable était partagée entre toutes les itérations — une source classique de bugs dès qu'une closure ou une goroutine la capturait (voir [section 2.3](03-fonctions.md)).

```go
funcs := make([]func(), 3)
for i := 0; i < 3; i++ {
	funcs[i] = func() { fmt.Println(i) }
}
for _, f := range funcs {
	f() // affiche 0, 1, 2 — et non 3, 3, 3 comme avant Go 1.22
}
```

Chaque closure capture désormais un `i` distinct, ce qui correspond à l'intuition.

## `range` : itérer sur les collections

`range` parcourt une collection en fournissant, selon le type, un indice/clé et une valeur.

```go
// Slice : indice et élément
for i, v := range nums {
	fmt.Println(i, v)
}

// Map : clé et valeur — l'ordre n'est PAS déterministe
for key, val := range m {
	fmt.Println(key, val)
}

// Chaîne : indice d'octet et rune (point de code Unicode ; voir section 2.6)
for i, r := range "héllo" {
	fmt.Printf("%d: %c\n", i, r)
}
```

Deux cas particuliers :

- Sur un **channel**, `range` reçoit les valeurs jusqu'à sa fermeture (voir [module 4](../04-concurrence/README.md)).
- 🆕 Sur un **entier** (depuis Go 1.22), `for i := range n` itère de `0` à `n-1` :

```go
for i := range 3 {
	fmt.Println(i) // 0, 1, 2
}
```

On ignore les valeurs superflues avec `_`, ou on les omet : `for range ch { … }` (aucune variable), `for i := range nums { … }` (indice seul).

## Les itérateurs `range-over-func` (depuis Go 1.23) 🆕

**Depuis Go 1.23, `range` peut aussi parcourir une fonction** ayant la signature d'un itérateur. Le package `iter` définit deux types : `iter.Seq[V]` (`func(yield func(V) bool)`) pour une séquence de valeurs, et `iter.Seq2[K, V]` pour des paires clé/valeur.

Le principe : l'itérateur **produit** ses éléments en appelant `yield`. La boucle `range` pilote l'ensemble ; si elle s'interrompt (`break`), `yield` renvoie `false`, ce qui signale à l'itérateur de s'arrêter.

```go
// Compte renvoie un itérateur sur 0, 1, … n-1.
func Compte(n int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := 0; i < n; i++ {
			if !yield(i) { // false = la boucle appelante s'est arrêtée
				return
			}
		}
	}
}

for v := range Compte(3) {
	fmt.Println(v) // 0, 1, 2
}
```

L'intérêt : parcourir **n'importe quelle collection personnalisée** avec la même syntaxe `range`, y compris des séquences paresseuses. La bibliothèque standard fournit désormais de tels itérateurs — par exemple `slices.Values` et `slices.All`, ou `maps.Keys` et `maps.Values` :

```go
for v := range slices.Values(nums) { // itère sur les éléments d'un slice
	fmt.Println(v)
}
```

## En résumé

Go se contente d'un **unique `for`** — forme complète, conditionnelle ou infinie —, complété par **`range`** pour parcourir slices, maps, chaînes, channels et même les entiers. Deux apports récents modernisent l'ensemble : la **variable de boucle par itération** (Go 1.22), qui élimine un piège classique des closures, et les **itérateurs `range-over-func`** (Go 1.23), qui unifient le parcours des collections personnalisées. Poursuivons avec un type que `range` traite déjà à part : les **chaînes**, en [section suivante](06-chaines.md).

---

[🔝 Sommaire](../SOMMAIRE.md) · [⏭️ Section suivante : 2.6 Chaînes, runes, UTF-8](06-chaines.md)

⏭ [Chaînes, runes, UTF-8 (`strings`, `strconv`, `fmt`)](/02-fondamentaux-langage/06-chaines.md)
