🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 3.2 Composition plutôt qu'héritage (embedding)

Go ne propose **aucun héritage** — ni classes, ni hiérarchie, ni `extends`. Ce n'est pas un oubli, mais un parti pris : le langage privilégie la **composition**, tenue pour plus simple et plus robuste que l'héritage classique. L'outil dédié s'appelle l'**embedding**, décliné en deux formes : embarquer une struct, ou embarquer une interface. La règle à graver dès maintenant, car elle sous-tend toute la section : **l'embedding n'est pas de l'héritage** — il n'y a pas de dispatch virtuel.

## L'embedding de struct : la promotion

On embarque un type en le déclarant **sans nom de champ**, juste par son type :

```go
type Animal struct {
	Name string
}

func (a Animal) Speak() string { return a.Name + " émet un son" }

type Dog struct {
	Animal // champ embarqué : pas de nom, juste le type
	Breed  string
}
```

Les champs et méthodes du type embarqué sont **promus** : ils s'utilisent directement sur le type englobant, comme s'ils lui appartenaient.

```go
d := Dog{Animal: Animal{Name: "Rex"}, Breed: "Berger"}

d.Name        // "Rex"                — champ promu
d.Speak()     // "Rex émet un son"    — méthode promue
d.Animal.Name // accès explicite au champ embarqué (nom = nom du type)
```

Le nom implicite du champ est le **nom (non qualifié) du type** : `Animal` ici, `Mutex` pour un `sync.Mutex` embarqué, `Logger` pour un `*log.Logger`. On peut aussi embarquer un **pointeur**, à condition de l'initialiser (sinon un appel promu déréférence `nil` et panique) :

```go
type Service struct {
	*log.Logger // pointeur embarqué : à fournir à la construction
}

s := Service{Logger: log.Default()}
s.Println("démarrage") // méthode promue depuis *log.Logger
```

## Embedding n'est pas héritage : pas de dispatch virtuel

C'est la nuance décisive, et le piège classique de quiconque vient de Java, C# ou Python. On peut « masquer » une méthode promue en en définissant une de même nom sur le type englobant — mais **il n'y a pas de dispatch virtuel** : une méthode promue qui en appelle une autre appelle **la sienne**, pas votre redéfinition.

```go
type Base struct{}

func (b Base) Hello() string    { return "Base" }
func (b Base) Greeting() string { return "Bonjour, " + b.Hello() }

type Derived struct {
	Base
}

func (d Derived) Hello() string { return "Derived" } // « surcharge » apparente
```

```go
d := Derived{}
d.Hello()    // "Derived"        — la méthode externe masque la promue
d.Greeting() // "Bonjour, Base"  — et NON "Bonjour, Derived"
```

`Greeting` est promue depuis `Base` ; son receveur est le `Base` embarqué, donc son `b.Hello()` résout vers `Base.Hello`. Aucun retour vers `Derived.Hello`. En Java, `Greeting()` renverrait « Bonjour, Derived » ; en Go, non. **Ne cherchez pas à simuler l'héritage par l'embedding** — c'est un anti-pattern recensé en [annexe B](../annexes/go-idiomatique/README.md).

La bonne réponse idiomatique consiste à exprimer la variation par une **interface**, puis à l'injecter ([§ 3.3](03-interfaces.md)) :

```go
type Greeter interface{ Hello() string }

func Greeting(g Greeter) string { return "Bonjour, " + g.Hello() }
```

Ici `Greeting(Derived{})` appelle bien `Derived.Hello` : le polymorphisme passe par l'interface, pas par la parenté de types.

## Composer des comportements

L'intérêt réel de l'embedding : une méthode promue **entre dans le method set** du type englobant (cf. [§ 3.1](01-structs-methodes.md)). Embarquer un type qui satisfait une interface fait donc satisfaire cette interface au type englobant — sans une ligne de code de délégation.

Un cas fréquent est l'embarquement de `sync.Mutex` pour obtenir `Lock()`/`Unlock()`. Mais attention à un effet de bord idiomatiquement discutable : l'embedding **promeut ces méthodes dans l'API publique** de votre type.

```go
// Embarquer sync.Mutex promeut Lock()/Unlock() : appelables de l'extérieur…
type Cache struct {
	sync.Mutex
	data map[string]string
}

// … c'est pourquoi on préfère souvent un champ nommé, non exporté :
type SafeCache struct {
	mu   sync.Mutex // le verrou reste un détail d'implémentation (cf. § 4.3)
	data map[string]string
}
```

C'est le même arbitrage qu'au [§ 3.1](01-structs-methodes.md) : on **embarque quand on veut délibérément exposer** l'API du type interne ; sinon on utilise un **champ nommé** pour la garder encapsulée.

## Embarquer une interface : décoration et délégation

Embarquer une **interface** dans une struct est le socle du pattern décorateur, très idiomatique. Toutes les méthodes du contrat sont promues ; on n'en redéfinit qu'une poignée, en déléguant explicitement au type embarqué :

```go
type Store interface {
	Get(key string) (string, error)
	Set(key, val string) error
}

// On embarque l'interface : Get et Set sont promues…
type LoggingStore struct {
	Store
}

// … on ne redéfinit que Get, en déléguant au type embarqué.
func (s LoggingStore) Get(key string) (string, error) {
	log.Printf("GET %q", key)
	return s.Store.Get(key)
}
```

`LoggingStore` satisfait toujours `Store` (`Set` promue, `Get` redéfinie) : c'est exactement l'esprit d'un middleware, décliné ici au niveau des types (cf. [§ 5.2](../05-backend-http/02-middleware.md) pour son équivalent HTTP). La valeur embarquée doit évidemment être fournie — `LoggingStore{Store: vraiStore}` — faute de quoi tout appel non redéfini paniquerait sur une interface `nil`. Cette même technique sert à écrire des **stubs de test** minimalistes ([§ 13.2](../13-tests-qualite/02-mocks-testify.md)) : on embarque l'interface et l'on n'implémente que les méthodes attendues.

Enfin, une **interface peut en embarquer d'autres** pour composer des contrats — la bibliothèque standard en regorge (`io.ReadWriter` = `io.Reader` + `io.Writer`). Le mécanisme est détaillé au [§ 3.3](03-interfaces.md) :

```go
type ReadWriter interface {
	io.Reader
	io.Writer
}
```

## Quand embarquer, quand utiliser un champ nommé

- **Embarquer** lorsque le type englobant doit réellement s'utiliser *comme* le type interne : décoration, réutilisation d'une API existante, « mixin » de comportement.
- **Champ nommé** lorsque vous voulez seulement *utiliser* le type interne sans exposer ses méthodes — l'encapsulation par défaut, qui garde votre surface d'API sous contrôle.

Trois précautions accompagnent l'embedding : il **greffe toute l'API publique** du type embarqué sur la vôtre (vous exposez peut-être un `Close()` non voulu) ; en cas de **collision de noms à même profondeur**, l'appel devient ambigu et ne compile pas tant qu'on ne qualifie pas (`x.A.Foo()`) ; et les **chaînes d'embedding profondes** rendent la promotion difficile à suivre. Dans le doute, un champ nommé est plus explicite.

## Côté IDE : GoLand et VS Code

- **GoLand** : le complètement liste les membres promus et indique leur type d'origine ; l'inspection signale les promotions ambiguës. *Generate* (`Alt+Inser` / `⌘N`) crée constructeurs et squelettes, et *Implement Methods* (`Ctrl+I`) aide à satisfaire le contrat qu'on décore.
- **VS Code + extension Go (gopls)** : le complètement propose les membres promus ; *Go to Definition* sur une méthode promue saute vers le type embarqué, et le survol (hover) affiche sa provenance.

Dans les deux IDE, la **délégation explicite** vers un champ nommé (l'alternative encapsulée à l'embedding) s'écrit à la main — aucun générateur dédié côté Go.

## En résumé

- Go remplace l'héritage par la **composition** ; l'embedding promeut champs et méthodes du type interne.
- **Embedding ≠ héritage** : pas de dispatch virtuel. Pour du polymorphisme, passez par une **interface** ([§ 3.3](03-interfaces.md)), pas par une fausse « surcharge ».
- Une méthode promue entre dans le **method set** : embarquer un type qui satisfait une interface la fait satisfaire au type englobant.
- **Embarquez** pour exposer délibérément une API (décorateur, mixin) ; utilisez un **champ nommé** pour encapsuler (ex. `mu sync.Mutex`).
- Méfiez-vous des fuites d'API, des collisions de noms et des chaînes d'embedding profondes.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [3.3 — Interfaces implicites, le cœur du design Go](03-interfaces.md)

⏭ [Interfaces implicites — le cœur du design Go](/03-types-interfaces/03-interfaces.md)
