🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 3.1 Structs, méthodes, receveurs valeur vs pointeur

Deux briques suffisent à modéliser presque tout en Go : la **struct**, qui regroupe des données, et la **méthode**, qui attache un comportement à un type. Leur assemblage soulève une question qui revient à chaque type que vous écrirez — faut-il un **receveur valeur ou pointeur ?** C'est le choix de conception le plus fréquent du quotidien Go, et la source de confusion n°1 chez les débutants. Cette section pose les bases ; l'embedding ([§ 3.2](02-composition-embedding.md)) et les interfaces ([§ 3.3](03-interfaces.md)) viendront ensuite s'appuyer dessus.

## Déclarer et instancier une struct

Une struct est un type composite qui agrège des champs nommés :

```go
type User struct {
	ID    int
	Name  string
	email string // non exporté : minuscule
}
```

Comme partout en Go, la casse gouverne la visibilité : `ID` et `Name` sont exportés (accessibles depuis un autre package), `email` reste privé au package (cf. [§ 2.1](../02-fondamentaux-langage/01-structure-packages.md)).

**Zéro-value.** Une struct déclarée sans initialisation n'est jamais `nil` : chacun de ses champs prend sa [zéro-value](../02-fondamentaux-langage/02-types-variables.md) (`0`, `""`, `nil`, `false`…). Concevoir des types dont la zéro-value est directement utilisable est un idiome Go fort — `bytes.Buffer` ou `sync.Mutex` s'emploient sans initialisation préalable.

```go
var u User // {0 "" ""} — prêt à l'emploi, pas nil
```

**Littéraux.** On instancie via un littéral, de préférence **avec champs nommés** (robuste face à l'ajout ou au réordonnancement de champs) plutôt que positionnel :

```go
u1 := User{ID: 1, Name: "Ada"}     // idiomatique
u2 := User{2, "Alan", "alan@x.io"} // positionnel : fragile, à éviter
p := &User{Name: "Grace"}          // p est un *User ; les champs omis restent à zéro
```

La forme `&User{...}` — plus courante que `new(User)` — alloue et renvoie un pointeur en une seule expression.

**Comparabilité.** Deux structs se comparent avec `==` si tous leurs champs sont comparables ; une telle struct peut alors servir de **clé de map**. À l'inverse, une struct contenant un slice, une map ou une fonction n'est pas comparable.

**Struct anonyme.** Pour un besoin ponctuel et local (configuration jetable, regroupement temporaire), on peut se passer de type nommé :

```go
opts := struct {
	Verbose bool
	Retries int
}{Verbose: true, Retries: 3}
```

**Tags de champs.** Un champ peut porter un **tag** — une chaîne de métadonnées écrite entre *backticks* — que le langage lui-même ignore, mais que des bibliothèques lisent par réflexion. Reprenons notre `User` en annotant ses champs :

```go
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	email string // non exporté : invisible pour le JSON — tag inutile
}
```

Un tag ne change rien au comportement du langage : c'est une annotation à destination d'outils (`encoding/json`, ORM, validation…). Ici, `json.Marshal` produira `{"id":1,"name":"Ada"}` — les clés viennent des tags, et le champ non exporté est ignoré. Syntaxe détaillée et options (`omitempty`…) au moment du JSON, en [section 5.3](../05-backend-http/03-json.md).

## Méthodes : attacher un comportement à un type

Une méthode est une fonction dotée d'un **receveur**, déclaré entre `func` et le nom :

```go
func (u User) DisplayName() string {
	return u.Name
}
```

Point souvent surprenant : une méthode peut être attachée à **n'importe quel type nommé du package**, pas seulement à une struct — y compris à un type dérivé d'un type de base :

```go
type Celsius float64

func (c Celsius) Fahrenheit() float64 {
	return float64(c)*9/5 + 32
}
```

Seule contrainte : le type receveur doit être **défini dans le même package**. Impossible d'ajouter une méthode à `int` ou à un type d'un autre package ; il faut d'abord le redéfinir localement (`type MyInt int`).

**Convention de nommage.** Le receveur porte un nom **court et cohérent** — une ou deux lettres, en général l'initiale du type (`u` pour `User`, `srv` pour `Server`). Go n'utilise ni `this` ni `self`, et un même type garde le même nom de receveur sur toutes ses méthodes.

**Pas de constructeur, mais une convention.** Le langage n'a aucun mot-clé de construction. L'usage est une fonction `NewXxx` renvoyant le type prêt à l'emploi — indispensable dès que des champs non exportés doivent être initialisés :

```go
func NewUser(name string) *User {
	return &User{Name: name, email: strings.ToLower(name) + "@example.com"}
}
```

Enfin, une méthode n'est qu'une fonction : `u.DisplayName` en isole une (méthode *valeur*), et `User.DisplayName` en donne l'*expression* avec le receveur passé en premier paramètre — pratique pour fournir une méthode en argument.

## Receveur valeur ou receveur pointeur

Voilà le cœur de la section.

### Ce que change le receveur

Un **receveur valeur** reçoit une **copie** : les mutations qu'il opère sont perdues. Un **receveur pointeur** reçoit l'adresse de l'original : ses mutations persistent.

```go
type Counter struct{ n int }

func (c Counter) IncValue()    { c.n++ } // agit sur une copie
func (c *Counter) IncPointer() { c.n++ } // agit sur l'original
```

```go
c := Counter{}
c.IncValue()   // c.n == 0 : la copie a été jetée
c.IncPointer() // c.n == 1 : Go prend l'adresse pour vous → (&c).IncPointer()
```

Ce dernier commentaire souligne une commodité du compilateur : sur une **valeur adressable**, l'appel d'une méthode à receveur pointeur prend automatiquement l'adresse ; réciproquement, l'appel d'une méthode à receveur valeur sur un pointeur le déréférence. Vous écrivez `c.IncPointer()`, jamais `(&c).IncPointer()`.

### Un détail qui compte : les *method sets*

Cette symétrie a une limite qui prépare le terrain des interfaces ([§ 3.3](03-interfaces.md)). Le *method set* (ensemble de méthodes) d'un type détermine les interfaces qu'il satisfait :

- le method set de `T` ne contient **que** les méthodes à receveur **valeur** ;
- celui de `*T` contient les méthodes à receveur **valeur *et* pointeur**.

Conséquence pratique : un type dont une méthode a un receveur pointeur ne satisfait une interface que **via son pointeur**.

```go
type Stringer interface{ String() string }

type T struct{}

func (t *T) String() string { return "T" } // receveur POINTEUR

var _ Stringer = &T{} // OK : *T possède String()
// var _ Stringer = T{} // NE COMPILE PAS : la valeur T ne l'a pas
```

Le piège corollaire concerne les valeurs **non adressables**, comme un élément de map : la prise d'adresse automatique n'y joue pas.

```go
m := map[string]Counter{"a": {}}
// m["a"].IncPointer() // NE COMPILE PAS : m["a"] n'est pas adressable
v := m["a"]
v.IncPointer()
m["a"] = v // détour obligé : copier, modifier, réaffecter
```

### Règle de décision

Préférez un **receveur pointeur** lorsque :

1. la méthode doit **modifier** le receveur ;
2. la struct est **volumineuse** et vous voulez éviter le coût de la copie (nuances de perf en [§ 14.2](../14-performance/02-gc-allocations.md)) ;
3. le type contient un champ **qui ne doit pas être copié** — `sync.Mutex`, `sync.WaitGroup`, etc. (cf. [§ 4.3](../04-concurrence/03-synchronisation.md)) ;
4. par **cohérence** : dès qu'*une* méthode du type exige un receveur pointeur, donnez-le à **toutes** ses méthodes.

Un **receveur valeur** convient pour de petits types sans mutation (une paire de coordonnées, une durée, un identifiant). En cas de doute, le pointeur est le choix par défaut le plus sûr. Ce qu'il faut surtout éviter, c'est de **mélanger** receveurs valeur et pointeur sur un même type : cela sème la confusion et fragilise la satisfaction d'interfaces.

### Pièges classiques

**Copier un verrou.** Un receveur valeur sur un type qui embarque un `sync.Mutex` copie le verrou : la protection devient inopérante. `go vet` le détecte (analyseur `copylocks`, cf. [§ 13.5](../13-tests-qualite/05-linters.md)).

```go
type Registry struct {
	mu    sync.Mutex
	items map[string]int
}

func (r *Registry) Add(k string) { // pointeur : obligatoire ici
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[k]++
}
```

**Champ slice et receveur valeur.** Copier une struct copie l'**en-tête** de ses slices, mais pas le tableau sous-jacent (cf. [§ 2.7](../02-fondamentaux-langage/07-slices-maps.md)). Modifier un élément existant via un receveur valeur reste donc visible par l'appelant… tandis qu'un `append` qui réalloue passera inaperçu. Une raison de plus de préférer le pointeur dès qu'un état interne évolue.

**Receveur pointeur nil.** Un receveur pointeur peut valoir `nil` : c'est légal, et parfois voulu, si la méthode le gère explicitement.

```go
type List struct {
	val  int
	next *List
}

func (l *List) Len() int {
	if l == nil {
		return 0 // cas de base propre pour une liste chaînée
	}
	return 1 + l.next.Len()
}
```

## Côté IDE : GoLand et VS Code

La plupart de ces gestes s'automatisent des deux côtés.

- **GoLand** : `Alt+Inser` (Win/Linux) ou `⌘N` (macOS) — *Generate* — crée constructeur, *getters/setters* ou squelette de méthode ; le complètement remplit un littéral de struct (*Fill fields*) ; des inspections signalent l'incohérence entre receveurs valeur et pointeur et proposent de convertir. Le refactoring *Change Signature* met à jour une méthode et tous ses appels.
- **VS Code + extension Go (gopls)** : l'action rapide *Fill struct* (ampoule 💡) complète un littéral avec les champs à zéro ; *Go: Add Tags to Struct Fields* gère les tags via `gomodifytags` ; gopls génère automatiquement les méthodes manquantes pour satisfaire une interface (utile au [§ 3.3](03-interfaces.md)).

Dans les deux, `gofmt` aligne les champs et normalise la mise en forme à l'enregistrement — inutile de peaufiner l'alignement à la main.

## En résumé

- Une struct groupe des données ; sa **zéro-value** doit idéalement être utilisable telle quelle.
- Instanciez avec des **littéraux à champs nommés** ; construisez via `&T{}` ou une fonction `NewXxx`.
- Une **méthode** s'attache à tout type nommé du package ; receveur court et cohérent, jamais `this`/`self`.
- **Receveur pointeur** si : mutation, gros type, champ non-copiable, ou cohérence. **Receveur valeur** pour les petits types sans mutation. **Ne mélangez pas** les deux sur un même type.
- Les **method sets** décident de la satisfaction d'interface : `*T` en offre plus que `T` — à garder en tête pour le [§ 3.3](03-interfaces.md).

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [3.2 — Composition plutôt qu'héritage (embedding)](02-composition-embedding.md)

⏭ [Composition plutôt qu'héritage (embedding)](/03-types-interfaces/02-composition-embedding.md)
