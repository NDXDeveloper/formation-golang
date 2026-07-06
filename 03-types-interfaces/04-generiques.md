🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 3.4 Génériques (contraintes, `any`, `comparable` — quand les utiliser ou pas)

Les génériques (paramètres de type) sont arrivés avec Go 1.18. Ils permettent d'écrire des fonctions et des types **paramétrés par le type**, avec une sécurité vérifiée à la compilation. Mais en Go, la vraie question n'est pas *comment* écrire un générique — c'est *quand*. La culture du langage les traite comme un outil que l'on dégaine **avec parcimonie**, après avoir envisagé le type concret puis l'interface. Cette section couvre la mécanique, puis consacre l'essentiel à cet arbitrage.

## Le problème qu'ils résolvent

Avant les génériques, écrire une pile ou un `Map` valable pour plusieurs types imposait deux mauvaises options : **dupliquer** le code pour chaque type, ou passer par `any` + assertions — au prix de la sécurité de type et de conversions à l'exécution ([§ 3.3](03-interfaces.md)). Les paramètres de type offrent une troisième voie : un **polymorphisme typé, résolu à la compilation**.

## Fonctions génériques et inférence

La syntaxe introduit les paramètres de type entre crochets, chacun borné par une **contrainte** :

```go
// Map applique f à chaque élément, quels que soient les types T et U.
func Map[T, U any](s []T, f func(T) U) []U {
	r := make([]U, len(s))
	for i, v := range s {
		r[i] = f(v)
	}
	return r
}

lengths := Map([]string{"go", "rust"}, func(s string) int { return len(s) })
// lengths == []int{2, 4} — T=string et U=int sont inférés, pas à écrire
```

L'**inférence de type** est la règle : dans la majorité des cas, le compilateur déduit les arguments de type depuis les paramètres, et l'on appelle un générique comme une fonction ordinaire.

## Les contraintes : `any`, `comparable`, interfaces, unions

Une contrainte est une **interface** utilisée comme borne. Les plus courantes :

**`any`** — la contrainte la plus large (alias de `interface{}`) : aucun opérateur ni méthode garanti, seulement stocker et passer la valeur, comme pour `T`/`U` ci-dessus.

**`comparable`** — contrainte **intégrée** pour les types utilisables avec `==` et `!=`, donc comme **clés de map** :

```go
type Set[T comparable] map[T]struct{}

func (s Set[T]) Add(v T)      { s[v] = struct{}{} }
func (s Set[T]) Has(v T) bool { _, ok := s[v]; return ok }

s := Set[string]{}
s.Add("go")
s.Has("go") // true
```

*(Au passage : `struct{}` est la **struct vide** — zéro champ, zéro octet. La valeur `struct{}{}` ne coûte rien : c'est l'idiome pour signaler la simple **présence** d'une clé, comme ici dans un ensemble.)*

**Interfaces à méthodes** — pour contraindre par un comportement, exactement comme au [§ 3.3](03-interfaces.md). **Unions de types et `~`** — une contrainte peut aussi lister des types concrets avec `|`, et le **tilde `~`** élargit à tout type dont le *sous-jacent* correspond :

```go
// ~int | ~float64 : accepte int, float64 et tout type dérivé d'eux.
type Number interface {
	~int | ~float64
}

func Sum[T Number](nums []T) T {
	var total T
	for _, n := range nums {
		total += n // autorisé : le jeu de types de Number supporte +
	}
	return total
}

type Celsius float64

Sum([]Celsius{1.5, 2.5}) // OK : Celsius a pour sous-jacent float64 → ~float64
```

Pour les types **ordonnés** (comparables avec `<`, `>`), la stdlib fournit `cmp.Ordered` depuis Go 1.21 — inutile désormais de recourir au vieux package `golang.org/x/exp/constraints` :

```go
func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

Max(3, 7)     // int
Max("a", "b") // string
```

## Types génériques

Un type peut être paramétré ; ses méthodes réutilisent le paramètre du receveur :

```go
type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }

func (s *Stack[T]) Pop() (T, bool) {
	var zero T // valeur zéro d'un paramètre de type
	if len(s.items) == 0 {
		return zero, false
	}
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v, true
}
```

Deux points à retenir : l'idiome `var zero T` pour obtenir la valeur zéro d'un paramètre de type, et une **limite du langage** — une méthode ne peut pas introduire ses *propres* paramètres de type au-delà de ceux du receveur (pas de `func (s *Stack[T]) MapTo[U any](...)`). Pour une telle transformation, on écrit une fonction générique de premier niveau.

## Quand les utiliser — et quand s'en abstenir

Voilà le cœur de la section. La règle d'or tient en un ordre de préférence : **type concret → interface → générique**. On choisit l'outil le **moins puissant** qui fait le travail.

**Utilisez un générique quand :**

- vous écrivez un **conteneur** ou une structure de données paramétrée par le type de ses éléments (pile, ensemble, arbre, cache typé) ;
- vous écrivez un **algorithme général** dont la logique est *identique* pour tous les types et ne dépend d'aucune méthode spécifique (`Map`, `Filter`, `Reduce`, `Keys`, `Max`) ;
- vous voulez **remplacer** un `any` + assertions par du code type-sûr, ou **préserver le type concret** en sortie (là où une interface l'effacerait).

**Préférez une interface quand :**

- l'opération est définie par un **comportement / une méthode**. Si votre fonction ne fait qu'appeler des méthodes sur son argument, prenez une interface, pas un paramètre de type : `func Print(w io.Writer)` et non `func Print[W io.Writer](w W)` — le paramètre de type n'apporterait rien et alourdirait la signature ;
- des types différents nécessitent des **implémentations différentes** de la même opération : c'est du polymorphisme par interfaces (dispatch dynamique), pas de la généricité.

**Abstenez-vous quand** il n'existe en pratique qu'un seul type concerné, ou quand le générique ajoute de la complexité sans réutilisation réelle. Dupliquer vingt lignes deux fois vaut souvent mieux qu'une abstraction générique alambiquée — la sur-abstraction est un anti-pattern classique ([annexe B](../annexes/go-idiomatique/README.md)).

Deux formules pour trancher : *si l'implémentation est la même pour tous les types → générique ; si elle diffère selon le type → interface*. Et surtout : **ne remplacez pas une bonne interface par un paramètre de type**.

## Auto-référence : les génériques F-bornés (Go 1.26)

Go 1.26 lève une restriction : un paramètre de type peut désormais **se référencer lui-même** dans sa propre contrainte. On peut donc déclarer `type Adder[A Adder[A]]` — le *polymorphisme F-borné*, qui exprime « un type capable d'opérer sur lui-même et de renvoyer un lui-même » :

```go
// Go 1.26 : la contrainte de A est l'interface elle-même, paramétrée par A.
type Adder[A Adder[A]] interface {
	Add(A) A
}

type Vec struct{ X, Y int }

func (v Vec) Add(o Vec) Vec { return Vec{v.X + o.X, v.Y + o.Y} }

// Vec satisfait Adder[Vec] ; l'algorithme renvoie un A concret, pas une interface.
func Total[A Adder[A]](items []A) A {
	var sum A
	for _, it := range items {
		sum = sum.Add(it)
	}
	return sum
}
```

C'est un outil **avancé**, dont la plupart du code n'aura jamais besoin. Son intérêt sur une interface classique est de **conserver le type concret** dans les signatures (`Add(A) A` plutôt que `Add(Adder) Adder`). À réserver aux bibliothèques génériques qui en tirent un réel bénéfice — fidèle, ici encore, au principe de parcimonie de la section.

## Les génériques que vous consommez d'abord

En pratique, on **consomme** les génériques bien plus qu'on n'en écrit. Depuis Go 1.21, la stdlib expose l'essentiel :

```go
nums := []int{3, 1, 2}
slices.Sort(nums)                        // [1 2 3]
slices.Contains(nums, 2)                 // true
i, found := slices.BinarySearch(nums, 2) // 1, true
```

Le package `slices` (tri, recherche, `Index`, `Max`…), le package `maps` (`Keys`, `Values`, qui renvoient des itérateurs `range`-over-func, [§ 2.5](../02-fondamentaux-langage/05-boucles.md)) et le package `cmp` (`cmp.Ordered`, `cmp.Compare`, `cmp.Or`) couvrent la majorité des besoins courants sans écrire une seule contrainte soi-même.

## Performance : ne pas surinvestir

Le compilateur implémente les génériques par un mélange de spécialisation et de *dictionnaires* (partage par « forme GC »). Ce n'est **pas** un gain de performance automatique : selon les cas, un générique peut être plus rapide, équivalent, ou plus lent qu'une interface. Ne choisissez pas les génériques *pour* la performance — mesurez (cf. [module 14](../14-performance/README.md)).

## Côté IDE : GoLand et VS Code

Les deux prennent en charge les génériques de bout en bout.

- **GoLand** : complètement conscient des contraintes, indices d'inlay affichant les arguments de type inférés, inspections signalant une contrainte non satisfaite, refactorings compatibles.
- **VS Code + extension Go (gopls)** : complètement et diagnostics sur les paramètres de type, survol (hover) et *inlay hints* révélant les types inférés à l'appel — utile pour vérifier ce que l'inférence a déduit.

## En résumé

- Les génériques (Go 1.18) offrent un **polymorphisme typé** à la compilation ; l'inférence évite d'écrire les arguments de type.
- Une **contrainte** est une interface : `any` (la plus large), `comparable` (pour `==` et les clés de map), interfaces à méthodes, unions de types avec `|` et `~` (sous-jacent), `cmp.Ordered` pour l'ordre.
- Ordre de préférence : **concret → interface → générique**. Générique si la logique est *identique* pour tous les types ; interface si le comportement *diffère* selon le type.
- **Ne remplacez pas une interface par un paramètre de type** ; évitez la sur-abstraction ([annexe B](../annexes/go-idiomatique/README.md)).
- Go 1.26 autorise l'**auto-référence** (`type Adder[A Adder[A]]`) — puissant mais rarement nécessaire.
- Vous **consommez** surtout les génériques via `slices`, `maps` et `cmp` ; ce n'est pas un levier de performance ([module 14](../14-performance/README.md)).

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [3.5 — Organisation du code : packages, `internal/`, layout, workspaces](05-organisation-code.md)

⏭ [Organisation du code : packages, `internal/`, layout standard, workspaces (`go work`)](/03-types-interfaces/05-organisation-code.md)
