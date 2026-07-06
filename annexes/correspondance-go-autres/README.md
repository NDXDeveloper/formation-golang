🔝 Retour au [Sommaire](/SOMMAIRE.md)

# Annexe A — Correspondance syntaxique Go ↔ autres langages (aide-mémoire)

Cet aide-mémoire met la syntaxe de Go en regard de quatre langages de référence — **Python**, **Java**, **C#** et **Rust** — pour vous permettre de projeter ce que vous savez déjà sur le vocabulaire de Go. Il prolonge, côté syntaxe, la grille de décision de la [section 1.6.1 (« Go vs Rust / Python / Java / C# »)](../../01-introduction-go/06.1-go-vs-autres-langages.md).

> ⚠️ **Une correspondance n'est pas une équivalence.** Les tableaux ci-dessous donnent la construction *la plus proche*, pas un calque exact. Traduire mot à mot la syntaxe d'un autre langage produit presque toujours du Go non idiomatique : un `try/catch` Java rendu en `panic`/`recover`, une hiérarchie de classes plaquée en interfaces… restent des faux amis. Pour écrire du Go *idiomatique* (et non seulement compilable), lisez cet aide-mémoire en parallèle de l'[annexe B — Go idiomatique](../go-idiomatique/README.md).

Les exemples ciblent **Go 1.26**. Les apports récents du langage sont signalés 🆕.

---

## Le choc des paradigmes (à lire en premier)

Avant toute table de correspondance, voici les partis pris de Go qui déroutent le plus les développeurs venant d'un autre langage. Les garder en tête évite 90 % des contresens.

| Ce à quoi on s'attend | Le choix de Go |
|---|---|
| Des exceptions pour signaler les erreurs | **Les erreurs sont des valeurs** renvoyées en dernier retour ; `panic`/`recover` sont réservés à l'irrécupérable |
| De l'héritage de classes | **Pas d'héritage** : composition par *embedding* |
| Des interfaces qu'on déclare implémenter (`implements`) | **Interfaces implicites** (structurelles) : avoir les bonnes méthodes suffit |
| `public` / `private` | **La casse décide** : `Majuscule` = exporté, `minuscule` = privé au package |
| Un opérateur ternaire `?:` | **Aucun** : un `if`/`else` classique |
| Plusieurs mots-clés de boucle | **Un seul** : `for` (qui joue aussi `while`) |
| Des conversions numériques implicites | **Aucune** : `int → int64` doit être explicite |
| Un `switch` qui « tombe » d'un cas à l'autre | **Pas de fall-through** par défaut, `break` inutile |
| Surcharge de fonctions, paramètres par défaut | **Ni l'un ni l'autre** |
| Des imports/variables inutilisés tolérés | **Erreur de compilation** |
| `null` distinct de la valeur par défaut | **`nil` est la zéro-value** ; une struct n'est jamais `nil`, elle est « à zéro » |

Chacun de ces points est détaillé plus bas.

---

## Déclaration : variables et constantes

```go
// Go
var x int = 5       // déclaration explicite
var y = 5           // type inféré
z := 5              // forme courte (uniquement dans une fonction)
const Pi = 3.14159  // constante (compile-time, typée à l'usage)

const (             // groupe + compteur iota
    Rouge = iota    // 0
    Vert            // 1
    Bleu            // 2
)
```

```python
# Python
x: int = 5      # annotation optionnelle, non contraignante
y = 5           # typage dynamique
PI = 3.14159    # « constante » par convention (rien ne l'impose)
```

```java
// Java
int x = 5;
var y = 5;              // inférence locale (Java 10+)
final double PI = 3.14; // constante d'exécution
```

```csharp
// C#
int x = 5;
var y = 5;
const double Pi = 3.14; // constante de compilation
```

```rust
// Rust
let x: i32 = 5;       // immuable par défaut
let mut y = 5;        // mutable
const PI: f64 = 3.14; // constante
```

**À retenir côté Go** : `:=` déclare **et** affecte, mais uniquement à l'intérieur d'une fonction ; au niveau package, on utilise `var`/`const`. Les constantes Go sont évaluées à la compilation et peuvent rester *non typées* jusqu'à leur utilisation. Il n'existe pas de type énuméré : on combine `const` + `iota` avec un type nommé.

🆕 Depuis **Go 1.26**, `new` accepte une expression d'initialisation :

```go
p := new(int)   // *int pointant vers 0 (zéro-value)
q := new(42)    // 🆕 Go 1.26 : *int pointant vers 42
```

---

## Types primitifs

| Catégorie | Go | Python | Java | C# | Rust |
|---|---|---|---|---|---|
| Entier « par défaut » | `int` (32/64 bits selon plateforme) | `int` (précision arbitraire) | `int` / `long` | `int` / `long` | `i32` / `i64` / `isize` |
| Entier 64 bits | `int64` | `int` | `long` | `long` | `i64` |
| Entier non signé | `uint`, `uint64` | *(pas de type dédié)* | *(pas de type natif)* | `uint`, `ulong` | `u32`, `u64`, `usize` |
| Flottant | `float64` | `float` | `double` | `double` | `f64` |
| Booléen | `bool` | `bool` | `boolean` | `bool` | `bool` |
| Caractère | `rune` (= `int32`, point de code Unicode) | `str` de longueur 1 | `char` (UTF-16) | `char` (UTF-16) | `char` (scalaire Unicode, 4 o) |
| Octet | `byte` (= `uint8`) | `int` / `bytes` | `byte` | `byte` | `u8` |
| Chaîne | `string` (UTF-8, immuable) | `str` | `String` | `string` | `String` / `&str` |
| Tableau d'octets | `[]byte` | `bytes` / `bytearray` | `byte[]` | `byte[]` | `Vec<u8>` / `&[u8]` |
| Absence de valeur | `nil` | `None` | `null` | `null` | `Option::None` |

**Pièges classiques** : `rune` (point de code, 32 bits) et `byte` (octet) recouvrent des réalités différentes — indexer une chaîne (`s[i]`) donne un **octet**, la parcourir avec `range` donne des **runes**. Contrairement à Java/C# dont le `char` est une unité UTF-16, un `rune` Go est un vrai point de code. Aucune conversion numérique n'est implicite en Go : additionner un `int` et un `int64` ne compile pas sans `int64(x)`.

---

## Opérateurs et mots-clés

| Concept | Go | Python | Java / C# | Rust |
|---|---|---|---|---|
| ET / OU logiques | `&&` `\|\|` | `and` `or` | `&&` `\|\|` | `&&` `\|\|` |
| NON logique | `!` | `not` | `!` | `!` |
| Égalité / inégalité | `==` `!=` | `==` `!=` | `==` `!=` (`.equals` pour objets en Java) | `==` `!=` |
| Import | `import` | `import` | `import` / `using` | `use` |
| Retour | `return` | `return` | `return` | `return` ou dernière expression |
| Ternaire | *absent* | `a if c else b` | `c ? a : b` | `if c { a } else { b }` |
| Déclaration de constante | `const` | *(convention)* | `final` / `const` | `const` |
| Portée exportée | *casse (Majuscule)* | *(convention `_`)* | `public` | `pub` |

L'absence de ternaire est volontaire : Go préfère un `if`/`else` explicite. Attention aussi à l'égalité : en Go `==` compare structurellement les valeurs comparables (y compris les structs sans champ non comparable), là où en Java `==` sur des objets compare les références.

---

## Structures de contrôle

### Conditions

```go
// Go — pas de parenthèses, accolades obligatoires
if x > 0 {
    // ...
} else if x == 0 {
    // ...
} else {
    // ...
}

// forme idiomatique avec instruction d'initialisation
if v, err := calcul(); err != nil {
    return err
} else {
    use(v) // v n'existe que dans le if/else
}
```

En Python `if:`/`elif:`/`else:` (indentation), en Java/C# `if (...) { }` (parenthèses obligatoires, accolades facultatives), en Rust `if cond { }` (une expression, sans parenthèses). L'instruction d'initialisation du `if` (`if x := …; cond`) est une signature de Go.

### Sélection (`switch` / `match`)

```go
// Go — pas de fall-through, break implicite
switch x {
case 1, 2:               // plusieurs valeurs par cas
    fmt.Println("un ou deux")
case 3:
    fmt.Println("trois")
    fallthrough          // à demander explicitement si voulu
default:
    fmt.Println("autre")
}

// switch « sans expression » = if/else if lisible
switch {
case x < 0:
    // ...
case x == 0:
    // ...
}
```

C'est l'**inverse** de C/Java/C#, où chaque cas « tombe » sur le suivant sauf `break`. En Go il faut au contraire écrire `fallthrough` pour l'obtenir. Le `switch` sans expression remplace élégamment une cascade de `if`/`else if`. Le [`type switch`](../../02-fondamentaux-langage/04-conditions.md) (`switch v := x.(type)`) n'a pas d'équivalent direct ailleurs ; le `match` de Rust et celui de Python 3.10+ relèvent d'un filtrage par motifs plus riche mais d'un esprit voisin.

### Boucles

Go n'a qu'**un seul** mot-clé de boucle : `for`.

```go
for i := 0; i < n; i++ { }      // forme classique
for cond { }                    // équivalent d'un « while »
for { }                         // boucle infinie
for i, v := range s { }         // parcours (slice, map, string, canal)
for i := range n { }            // 🆕 Go 1.22 : itère de 0 à n-1
for v := range monIterateur { } // 🆕 Go 1.23 : itérateurs (range-over-func)
```

| Besoin | Go | Python | Java | C# | Rust |
|---|---|---|---|---|---|
| Compteur | `for i := 0; …; i++` | `for i in range(n)` | `for (…; …; …)` | `for (…; …; …)` | `for i in 0..n` |
| Tant que | `for cond` | `while cond` | `while (cond)` | `while (cond)` | `while cond` |
| Parcours | `for _, v := range c` | `for v in c` | `for (var v : c)` | `foreach (var v in c)` | `for v in c` |
| Infinie | `for` | `while True` | `for (;;)` | `while (true)` | `loop` |

---

## Fonctions

```go
// Go
func add(a, b int) int { return a + b }

// retours multiples — l'idiome Go
func divmod(a, b int) (q, r int) { // valeurs de retour nommées
    return a / b, a % b
}

// variadique
func somme(nums ...int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}

// fonction de première classe + closure
func multiplicateur(f int) func(int) int {
    return func(x int) int { return x * f }
}
```

```python
# Python
def add(a: int, b: int) -> int:
    return a + b

def divmod_(a, b):
    return a // b, a % b        # tuple

def somme(*nums):               # variadique
    return sum(nums)
```

```java
// Java — pas de retour multiple natif : record, tableau ou objet
int add(int a, int b) { return a + b; }
record DivMod(int q, int r) {}
DivMod divmod(int a, int b) { return new DivMod(a / b, a % b); }
int somme(int... nums) { /* varargs */ }
```

```rust
// Rust
fn add(a: i32, b: i32) -> i32 { a + b }
fn divmod(a: i32, b: i32) -> (i32, i32) { (a / b, a % b) } // tuple
```

**Les retours multiples** sont au cœur de Go, notamment le duo `(résultat, error)`. Là où Python et Rust renvoient un tuple et C# des value tuples, Java doit passer par un `record` ou un objet. Go n'a **ni surcharge de fonction ni paramètres par défaut** : un nom = une fonction ; pour de la configuration optionnelle, on utilise des variadiques ou le motif des *functional options*.

---

## Collections

### Tableaux et slices / listes

La distinction **tableau (taille fixe) / slice (dynamique)** est spécifique à Go et source de confusion.

```go
// Go
var arr [3]int = [3]int{1, 2, 3} // tableau : taille fixe, TYPE VALEUR (copié à l'affectation)
s := []int{1, 2, 3}              // slice : dynamique, vue sur un tableau sous-jacent
s = append(s, 4)                 // ajout (peut réallouer)
sub := s[1:3]                    // sous-slice : PARTAGE le tableau sous-jacent
n := len(s)                      // longueur
c := cap(s)                      // capacité
```

| Structure | Go | Python | Java | C# | Rust |
|---|---|---|---|---|---|
| Taille fixe | `[N]T` | *(n/a)* | `T[]` | `T[]` | `[T; N]` |
| Dynamique | `[]T` + `append` | `list` | `ArrayList<T>` | `List<T>` | `Vec<T>` |
| Vue / tranche | `s[a:b]` | `s[a:b]` | *(n/a native)* | `Span<T>` / `s[a..b]` | `&s[a..b]` |

Le plus proche d'une `list` Python ou d'un `Vec` Rust est le **slice** Go, pas le tableau. Attention : copier un slice copie son *en-tête* (pointeur, longueur, capacité), pas les données — deux slices peuvent donc partager le même tableau et se modifier mutuellement.

### Maps / dictionnaires

```go
// Go
m := map[string]int{"a": 1}
v := m["a"]          // 0 (zéro-value) si la clé est absente
v, ok := m["a"]      // « comma-ok » : ok == false si absente
delete(m, "a")
clear(m)             // 🆕 Go 1.21 : vide la map
```

| Opération | Go | Python | Java | C# | Rust |
|---|---|---|---|---|---|
| Type | `map[K]V` | `dict` | `HashMap<K,V>` | `Dictionary<K,V>` | `HashMap<K,V>` |
| Accès sûr | `v, ok := m[k]` | `m.get(k)` | `m.getOrDefault(k, d)` | `m.TryGetValue(k, out v)` | `m.get(&k)` → `Option` |
| Suppression | `delete(m, k)` | `del m[k]` | `m.remove(k)` | `m.Remove(k)` | `m.remove(&k)` |

Le motif **comma-ok** (`v, ok := m[k]`) distingue « clé absente » de « valeur nulle » sans exception. Point crucial : **l'ordre de parcours d'une map Go est volontairement aléatoire** — n'écrivez jamais de code qui en dépend.

---

## Types structurés et méthodes

### Struct vs classe

```go
// Go — des structs, pas de classes
type Point struct {
    X, Y int
}

p := Point{X: 1, Y: 2}
q := Point{1, 2}          // positionnel (fragile, préférer les noms)
```

```python
# Python
from dataclasses import dataclass

@dataclass
class Point:
    x: int
    y: int
```

```java
// Java — record pour un simple porteur de données (Java 16+)
record Point(int x, int y) {}
```

```csharp
// C# — record (C# 9+)
record Point(int X, int Y);
```

```rust
// Rust
struct Point { x: i32, y: i32 }
```

### Méthodes et receveurs

En Go, les méthodes sont définies **hors** du type, via un *receveur*.

```go
// receveur VALEUR : reçoit une copie (ne mute pas l'original)
func (p Point) Norme() float64 {
    return math.Hypot(float64(p.X), float64(p.Y))
}

// receveur POINTEUR : peut modifier l'original
func (p *Point) Deplacer(dx, dy int) {
    p.X += dx
    p.Y += dy
}
```

Le choix receveur valeur / pointeur n'a pas d'équivalent direct : dans les langages à objets classiques, `this`/`self` est toujours une référence. En Go, un receveur valeur travaille sur une copie — utile pour de petits types immuables, à éviter dès qu'on veut muter l'état ou pour de gros structs (voir [§3.1](../../03-types-interfaces/01-structs-methodes.md)).

### Composition vs héritage

```go
// Go — composition par embedding (PAS d'héritage)
type Animal struct{ Nom string }

func (a Animal) Parler() string { return a.Nom + " fait un bruit" }

type Chien struct {
    Animal      // embedding : les champs et méthodes d'Animal sont « promus »
    Race string
}

c := Chien{Animal{"Rex"}, "Berger"}
c.Parler() // OK : méthode promue
c.Nom      // OK : champ promu
```

| Langage | Réutilisation de comportement |
|---|---|
| Go | Composition (*embedding*) — pas d'héritage du tout |
| Python | Héritage (multiple possible) |
| Java | Héritage simple (`extends`) + interfaces |
| C# | Héritage simple (`:`) + interfaces |
| Rust | Composition + traits — pas d'héritage |

Sur ce point, **Rust et Go se ressemblent** : ni l'un ni l'autre n'a d'héritage de classe. L'embedding promeut champs et méthodes sans créer de relation « est-un » ; il se combine avec les interfaces implicites pour obtenir du polymorphisme sans hiérarchie.

---

## Polymorphisme : interfaces

C'est **la** différence de design la plus marquante. En Go, une interface est satisfaite **implicitement** : aucun mot-clé, aucune déclaration. Si un type possède les bonnes méthodes, il satisfait l'interface — que l'auteur du type le sache ou non.

```go
// Go
type Stringer interface {
    String() string
}

type Couleur struct{ R, G, B uint8 }

// Aucun « implements Stringer » : la présence de la méthode suffit
func (c Couleur) String() string {
    return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}

var s Stringer = Couleur{255, 128, 0} // accepté automatiquement
```

| Langage | Mécanisme | Typage |
|---|---|---|
| **Go** | interface satisfaite **implicitement** | **structurel** |
| Python | duck typing / `typing.Protocol` (PEP 544) | **structurel** |
| Java | `implements Interface` | nominal |
| C# | `: IInterface` | nominal |
| Rust | `impl Trait for Type` | nominal |

Le plus proche des interfaces Go est le `Protocol` de Python (typage structurel), pas l'interface Java/C# ni le trait Rust, qui sont *nominaux* (il faut déclarer explicitement la relation). Cette implicité pousse à définir de **petites interfaces côté consommateur** (`io.Reader`, `io.Writer`…), au lieu d'imposer une hiérarchie côté producteur — voir [§3.3](../../03-types-interfaces/03-interfaces.md).

---

## Génériques

Les génériques existent en Go depuis la version 1.18.

```go
// Go — fonction générique avec contrainte
func Map[T, U any](s []T, f func(T) U) []U {
    r := make([]U, len(s))
    for i, v := range s {
        r[i] = f(v)
    }
    return r
}

// contrainte personnalisée (ensemble de types)
type Nombre interface {
    ~int | ~int64 | ~float64
}

func Somme[T Nombre](vals []T) T {
    var total T
    for _, v := range vals {
        total += v
    }
    return total
}
```

| Concept | Go | Java | C# | Rust | Python |
|---|---|---|---|---|---|
| Paramètre de type | `[T any]` | `<T>` | `<T>` | `<T>` | `[T]` (PEP 695) |
| Contrainte / borne | `[T Ordered]` | `<T extends …>` | `where T : …` | `<T: Trait>` | `TypeVar(bound=…)` |
| « N'importe quel type » | `any` | `<?>` / `Object` | `object` | *(défaut)* | `Any` |

Go fournit deux contraintes prédéfinies : `any` (tout type) et `comparable` (types utilisables avec `==`). Les contraintes personnalisées se déclarent comme des interfaces listant des ensembles de types (avec `~` pour inclure les types sous-jacents). 🆕 **Go 1.26** autorise les génériques **auto-référentiels**, du type `type Adder[A Adder[A]]`, pour exprimer des contraintes récursives. Comme dans les autres langages, le conseil reste : n'introduisez des génériques que lorsqu'ils réduisent réellement la duplication (voir [§3.4](../../03-types-interfaces/04-generiques.md)).

---

## Gestion des erreurs

Le changement de mentalité le plus important. **En Go, une erreur est une valeur ordinaire**, renvoyée en dernière position et vérifiée par le code appelant. Il n'y a pas d'exceptions pour le flux d'erreurs normal.

```go
// Go
f, err := os.Open("data.txt")
if err != nil {
    return fmt.Errorf("ouverture data.txt : %w", err) // wrapping avec %w
}
defer f.Close()

// inspection d'erreur
if errors.Is(err, os.ErrNotExist) { /* erreur sentinelle */ }

var perr *fs.PathError
if errors.As(err, &perr) { /* erreur typée */ }
```

```python
# Python — exceptions
try:
    f = open("data.txt")
except FileNotFoundError as e:
    raise RuntimeError("ouverture data.txt") from e
```

```java
// Java — exceptions vérifiées / non vérifiées
try (var f = new FileInputStream("data.txt")) {
    // ...
} catch (IOException e) {
    throw new RuntimeException("ouverture data.txt", e);
}
```

```rust
// Rust — Result<T, E> + opérateur ?
let f = File::open("data.txt")
    .map_err(|e| MonErreur::Ouverture(e))?; // ? propage l'erreur
```

| Approche | Langages | Idée |
|---|---|---|
| Erreurs comme **valeurs** | **Go**, Rust | Le type de retour dit que la fonction peut échouer ; on traite l'erreur sur place |
| **Exceptions** | Python, Java, C# | On lève et on rattrape ; le chemin d'erreur est séparé du flux principal |

Go se range du côté de **Rust** (erreurs-valeurs), mais sans type somme ni opérateur `?` : on écrit explicitement `if err != nil`. Outils clés : `%w` pour envelopper (*wrapping*) et remonter une chaîne d'erreurs, `errors.Is` pour comparer à une sentinelle, `errors.As` pour extraire un type concret, et `errors.Join` (Go 1.20) pour agréger plusieurs erreurs.

> `panic`/`recover` **ne sont pas** l'équivalent de `try/catch`. Un `panic` signale un bug ou une situation irrécupérable ; `recover` ne s'emploie que dans un `defer`, à une frontière bien choisie (par exemple un middleware HTTP, cf. [§5.2](../../05-backend-http/02-middleware.md)). Ne vous en servez jamais comme d'un mécanisme de contrôle de flux ordinaire (voir [§2.10](../../02-fondamentaux-langage/10-defer-panic-recover.md)).

---

## Absence de valeur : `nil`, `null`, `Option`

```go
// Go — nil est la zéro-value de plusieurs types
var p *int            // nil
var s []int           // nil, mais UTILISABLE : len(s)==0, append(s, 1) fonctionne
var m map[string]int  // nil : lecture OK, mais ÉCRITURE => panic
var f func()          // nil
var err error         // nil
```

| Langage | Représentation | Particularité |
|---|---|---|
| Go | `nil` (zéro-value des pointeurs, slices, maps, canaux, fonctions, interfaces) | Une struct **n'est jamais** `nil` : elle est « à zéro » |
| Python | `None` | Objet singleton |
| Java | `null` | La « faute à un milliard de dollars » |
| C# | `null` (+ types de référence *nullable*) | Vérifications activables par le compilateur |
| Rust | *(pas de null)* — `Option<T>` | L'absence est encodée dans le type |

Go n'a **pas** de type `Option` : il s'appuie sur `nil`. Deux pièges fréquents : écrire dans une map `nil` provoque un `panic` (il faut `make`), alors qu'ajouter à un slice `nil` fonctionne parfaitement. Et une **interface contenant un pointeur nil n'est pas elle-même nil** — une subtilité qui piège même les habitués.

---

## Concurrence

La concurrence est le point fort de Go, avec un modèle CSP : des **goroutines** légères qui communiquent par **canaux**.

```go
// Go
go traiter(x) // lance une goroutine (le mot-clé fait tout)

ch := make(chan int)
go func() { ch <- 42 }() // envoi
v := <-ch                // réception

select { // attend sur plusieurs canaux
case v := <-ch1:
    use(v)
case <-time.After(time.Second):
    // timeout
}
```

| Langage | Modèle | Primitives |
|---|---|---|
| **Go** | CSP — goroutines + canaux | `go`, `chan`, `select` |
| Python | Threads (GIL) / asyncio | `threading`, `async`/`await` |
| Java | Threads, threads virtuels (21+) | `Thread`, `CompletableFuture` |
| C# | Asynchrone à base de `Task` | `async`/`await`, `Task` |
| Rust | Threads + async (runtime externe) | `thread::spawn`, `async`/`await` |

Les goroutines sont bon marché (démarrage rapide, pile qui croît à la demande) et multiplexées sur un petit nombre de threads système. L'analogue le plus proche aujourd'hui sont les **threads virtuels de Java 21** (projet Loom). Les canaux Go s'apparentent aux canaux `mpsc` de Rust. Pour l'annulation et les délais, Go passe partout un [`context.Context`](../../04-concurrence/04-context.md), là où C#/Java utilisent un `CancellationToken`/`Future`. Le test du code concurrent bénéficie du détecteur de courses (`-race`) et de [`testing/synctest`](../../04-concurrence/06-tester-code-concurrent.md).

---

## Chaînes et formatage

Go n'a **pas** de syntaxe d'interpolation dans la chaîne : le formatage passe par le package `fmt` et ses *verbes*.

```go
// Go
nom := "monde"
s := fmt.Sprintf("bonjour %s (%d)", nom, 42)
fmt.Printf("%v %+v %#v %T %q\n", p, p, p, p, nom)
```

| Verbe Go | Rôle |
|---|---|
| `%v` | valeur, format par défaut |
| `%+v` | idem, avec les noms de champs |
| `%#v` | représentation en syntaxe Go |
| `%T` | type dynamique |
| `%d` / `%s` / `%q` | entier / chaîne / chaîne entre guillemets |
| `%f` / `%g` | flottant (fixe / compact) |
| `%x` / `%p` | hexadécimal / pointeur |
| `%w` | enveloppe une erreur (*uniquement* dans `fmt.Errorf`) |

| Langage | Interpolation |
|---|---|
| Go | *(aucune)* — verbes `fmt` |
| Python | f-strings : `f"bonjour {nom} ({n})"` |
| Java | `String.format(...)` / `"…".formatted(...)` |
| C# | interpolation : `$"bonjour {nom} ({n})"` |
| Rust | macro : `format!("bonjour {nom} ({n})")` |

Les chaînes Go sont des séquences d'octets **UTF-8 immuables**. Pour construire efficacement une chaîne, on utilise `strings.Builder` plutôt que la concaténation répétée (voir [§2.6](../../02-fondamentaux-langage/06-chaines.md)).

---

## Organisation : modules, packages, visibilité

```go
// Go — un dossier = un package ; la casse gouverne la visibilité
package payment

func Charge()  {} // exporté (Majuscule) : visible hors du package
func compute() {} // non exporté (minuscule) : privé au package

type Client struct { // exporté
    ID  string       // champ exporté
    seq int          // champ non exporté
}
```

| Langage | Unité | Visibilité |
|---|---|---|
| Go | package (= dossier) | **Casse** : `Majuscule` = exporté, `minuscule` = privé au package |
| Python | module (= fichier) | Convention `_prefixe` (non imposée) |
| Java | package | `public` / `protected` / `private` + *package-private* |
| C# | namespace / assembly | `public` / `internal` / `protected` / `private` |
| Rust | module (`mod`) | `pub` (tout est privé par défaut) |

La **visibilité par la casse** est propre à Go : il n'existe aucun mot-clé `public`/`private`. Le regroupement se fait par package (un dossier), et la dépendance entre packages se gère via les modules Go (`go.mod`, cf. [§1.3](../../01-introduction-go/03-ecosysteme-go.md) et [§3.5](../../03-types-interfaces/05-organisation-code.md)).

---

## Gestion de la mémoire

| Langage | Modèle mémoire | Pointeurs |
|---|---|---|
| **Go** | Ramasse-miettes concurrent | Oui, **sans arithmétique** de pointeurs |
| Python | GC (comptage de références + cycles) | Références implicites |
| Java | GC | Références |
| C# | GC | Références (+ pointeurs en bloc `unsafe`) |
| Rust | **Ownership / borrowing** — pas de GC | `&`, `&mut`, `Box<T>` |

Go possède des pointeurs mais **interdit l'arithmétique de pointeurs** : on prend une adresse avec `&`, on déréférence avec `*`, on alloue avec `new` ou `make`. Le compilateur décide seul, par *escape analysis*, si une valeur va sur la pile ou le tas — pas de `new`/`delete` manuel. C'est ici que **Rust diverge le plus** : son système de possession supprime le besoin d'un GC, au prix d'un modèle mental plus exigeant. Côté performances, le GC de Go continue d'évoluer (voir [§14.2](../../14-performance/02-gc-allocations.md)).

---

## Faux amis (pièges de traduction)

| Ça ressemble à… | Mais en Go… |
|---|---|
| `switch` de C/Java/C# | Pas de fall-through par défaut (l'inverse) ; `break` inutile |
| `=` d'affectation | `:=` déclare **et** affecte ; `=` réaffecte seulement |
| Tableau « dynamique » | Un `[N]T` est de **taille fixe** ; le dynamique, c'est le **slice** `[]T` |
| Copier un tableau | Copier un slice copie l'en-tête, pas les données (tableau partagé) |
| Indexer une chaîne | `s[i]` renvoie un **octet**, pas un caractère ; `range` donne des runes |
| `null` | `nil` est une zéro-value ; une struct n'est jamais nil, elle est « à zéro » |
| Conversion implicite | Aucune : `int` → `int64` doit être **explicite** |
| Import / variable inutilisé | **Erreur de compilation**, pas un simple avertissement |
| Surcharge de fonction | Inexistante : un nom = une fonction |
| Paramètres par défaut | Inexistants : variadiques ou *functional options* |
| `finally` / `using` / RAII | `defer` (exécuté à la sortie de la fonction) |
| Ordre d'itération d'une map | **Volontairement aléatoire** — ne pas en dépendre |
| Exceptions | Erreurs comme valeurs ; `panic`/`recover` ≠ `try/catch` |
| Énumérations (`enum`) | Absentes : `const` + `iota` (+ type nommé) |
| `this`/`self` toujours référence | Le receveur peut être **valeur** (copie) ou **pointeur** |

---

## Idiomes Go sans équivalent direct

Certaines tournures n'ont pas de traduction propre ailleurs — les reconnaître aide à lire du code Go :

- **Comma-ok** — `v, ok := m[k]`, `v, ok := x.(T)`, `v, ok := <-ch` : teste la présence, l'assertion de type ou la fermeture d'un canal, sans exception.
- **`defer`** — diffère un appel jusqu'à la sortie de la fonction, empilé en LIFO ; sert à libérer les ressources (`f.Close()`, `mu.Unlock()`) au plus près de leur acquisition.
- **Retours `(valeur, error)`** — la convention universelle de signalement d'erreur, sans tuple générique ni exception.
- **Zéro-value utile** — beaucoup de types sont conçus pour être directement utilisables « à zéro » (`sync.Mutex`, `bytes.Buffer`, `strings.Builder`) sans constructeur.
- **Embedding** — composition qui promeut champs et méthodes, en remplacement de l'héritage.
- **Interfaces implicites** — on définit l'interface côté consommateur ; les types la satisfont sans le déclarer.
- **`iota`** — génère des constantes incrémentales pour émuler des énumérations typées.

---

## Pour aller plus loin

- **Quand choisir Go** plutôt qu'un autre langage : [§1.6.1](../../01-introduction-go/06.1-go-vs-autres-langages.md).
- **GoLand vs VS Code** : [§1.6.2](../../01-introduction-go/06.2-goland-vs-vscode.md).
- **Écrire du Go idiomatique** (et non seulement compilable) : [annexe B](../go-idiomatique/README.md).
- **Glossaire et acronymes** : [annexe F](../glossaire/README.md).

---

🔝 [Retour au sommaire](../../SOMMAIRE.md) · ⏭️ [Annexe B — Go idiomatique : *Effective Go* condensé et anti-patterns](../go-idiomatique/README.md)


⏭ [Go idiomatique : *Effective Go* condensé et anti-patterns](/annexes/go-idiomatique/README.md)
