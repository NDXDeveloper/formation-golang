🔝 Retour au [Sommaire](/SOMMAIRE.md)

# Annexe B — Go idiomatique : *Effective Go* condensé et anti-patterns ⭐

Cette annexe condense ce qui fait qu'un code est *idiomatique* en Go, par opposition à un code seulement compilable. Elle distille quatre sources de référence — [*Effective Go*](https://go.dev/doc/effective_go), les *Go Code Review Comments*, le [*Google Go Style Guide*](https://google.github.io/styleguide/go/) et les *Go Proverbs* de Rob Pike — et met chaque idiome en regard de l'**anti-pattern** qu'il corrige.

> **Pourquoi l'idiome pèse autant en Go.** Le langage est délibérément petit : il offre peu de façons de faire une chose. La communauté a donc convergé, plus qu'ailleurs, sur un style commun — largement **outillé** (`gofmt`, `go vet`, `staticcheck`). En Go, « idiomatique » n'est pas une affaire de goût personnel : c'est une norme partagée, souvent vérifiable automatiquement. Voir la [§13.5 (linters)](../../13-tests-qualite/05-linters.md) pour l'outillage détaillé.

Les exemples ciblent **Go 1.26** et privilégient la bibliothèque standard avant les frameworks. Les apports récents sont signalés 🆕.

---

## La philosophie en une page

Quelques *Go Proverbs* résument l'état d'esprit ; ils reviennent en filigrane dans toute l'annexe.

- **Clear is better than clever.** La lisibilité prime sur l'astuce.
- **Errors are values.** On traite les erreurs comme des données, pas comme des exceptions.
- **Don't panic.** `panic` est réservé à l'irrécupérable.
- **The bigger the interface, the weaker the abstraction.** Les petites interfaces valent mieux.
- **Make the zero value useful.** Un type doit être exploitable « à zéro ».
- **`interface{}` says nothing.** (aujourd'hui `any`) — un type trop vague n'apporte rien.
- **Don't communicate by sharing memory; share memory by communicating.** Les canaux orchestrent.
- **A little copying is better than a little dependency.** Un peu de duplication vaut mieux qu'un couplage inutile.
- **Gofmt's style is no one's favorite, yet gofmt is everyone's favorite.** Le formatage n'est pas négociable.

L'ossature idiomatique tient en trois piliers, martelés par cette formation : **erreurs explicites**, **composition** (plutôt qu'héritage), **petites interfaces** (satisfaites implicitement) — et toujours la **stdlib avant les frameworks**.

---

## Formatage : non négociable

Le formatage n'est pas une opinion. `gofmt` (ou `goimports`, qui gère en plus les imports) impose un style unique : indentation par tabulations, alignement, espaces. On ne discute pas la sortie de `gofmt` ; on la subit, et c'est précisément ce qui met fin aux débats de style.

L'**intégration IDE** est traitée en fin d'annexe ([Outillage](#outillage-qui-fait-respecter-lidiome)). Règle de base : formatage **à l'enregistrement**, dans GoLand comme dans VS Code.

---

## Nommage

| ❌ À éviter | ✅ Idiomatique | Pourquoi |
|---|---|---|
| `GetName()` | `Name()` | Pas de préfixe `Get` sur un accesseur ; le mutateur est `SetName()` |
| `userId`, `htmlUrl`, `parseJson` | `userID`, `htmlURL`, `parseJSON` | Acronymes en casse **constante** : `ID`, `URL`, `HTTP`, `JSON` |
| `user_count` | `userCount` | `MixedCaps`/`mixedCaps`, **jamais** snake_case |
| package `utils`, `common`, `helpers`, `base` | package au nom **métier** (`payment`, `auth`, `store`) | Un package nomme une responsabilité, pas un fourre-tout |
| `http.HTTPServer`, `strings.StringReader` | `http.Server`, `strings.Reader` | Éviter le **bégaiement** (nom du package + nom du type) |
| `IUserRepo`, `UserRepoImpl` | `UserRepo` | Ni notation hongroise, ni suffixe `Impl` |
| receveur `this` / `self` | `u`, `s`, `r` (court et cohérent) | Convention Go des receveurs |

Règles complémentaires :

- **Portée courte, nom court.** `i`, `r`, `buf` dans une petite fonction ; un nom descriptif pour une variable de portée large.
- **Interfaces à une méthode** : nom = méthode + suffixe `-er` (`Reader`, `Writer`, `Formatter`, `Stringer`).
- **Erreurs** : sentinelles nommées `ErrXxx` (`io.EOF`, `sql.ErrNoRows`) ; types d'erreur nommés `XxxError`.
- **Cohérence du receveur** : le même nom court pour toutes les méthodes d'un type.

### Documentation

Les commentaires de documentation commencent par le nom de l'élément documenté et forment des phrases complètes ; `go doc` et pkg.go.dev les rendent tels quels.

```go
// Server traite les requêtes HTTP entrantes.
//
// La valeur zéro n'est pas utilisable : passer par NewServer.
type Server struct { /* ... */ }
```

La dépréciation suit une convention reconnue par l'outillage :

```go
// Deprecated: utiliser NewServer à la place.
func New() *Server { /* ... */ }
```

---

## Gestion des erreurs — l'idiome central

En Go, **une erreur est une valeur** renvoyée en dernier retour, vérifiée sur place. C'est le cœur du style ; le maîtriser, c'est écrire du Go idiomatique. (Détail complet en [§2.9](../../02-fondamentaux-langage/09-gestion-erreurs.md).)

**Ne jamais ignorer une erreur silencieusement :**

```go
// ❌ l'erreur disparaît
data, _ := os.ReadFile(path)

// ✅ on l'enrichit de contexte et on la remonte
data, err := os.ReadFile(path)
if err != nil {
    return fmt.Errorf("lecture de %q : %w", path, err) // %w enveloppe (wrapping)
}
```

**Inspecter par comportement, pas par comparaison de chaîne :**

```go
if errors.Is(err, os.ErrNotExist) { /* sentinelle */ }

var perr *fs.PathError
if errors.As(err, &perr) { /* type concret */ }
```

**Traiter une erreur une seule fois.** Loguer *et* renvoyer, c'est signaler deux fois le même problème :

```go
// ❌ double signalement
if err != nil {
    log.Printf("échec : %v", err)
    return err
}

// ✅ on enrichit et on remonte ; le log se fait à la frontière (main, handler)
if err != nil {
    return fmt.Errorf("traitement commande %d : %w", id, err)
}
```

**Style des messages d'erreur** : minuscule à l'initiale, **pas de ponctuation finale** — parce qu'ils sont presque toujours concaténés à d'autres contextes. Éviter aussi les mots redondants (« erreur : … », « échec de … »).

| ❌ | ✅ |
|---|---|
| `errors.New("Connexion refusée.")` | `errors.New("connexion refusée")` |
| `fmt.Errorf("Error: échec de lecture: %w", err)` | `fmt.Errorf("lecture config : %w", err)` |

**En bibliothèque, pas de `log.Fatal` ni de `os.Exit`** : on renvoie l'erreur et on laisse l'appelant décider. `panic` est réservé aux bugs de programmation et à l'irrécupérable ; `recover` ne s'emploie que dans un `defer`, à une frontière choisie (par ex. un middleware HTTP, cf. [§5.2](../../05-backend-http/02-middleware.md)). Détail en [§2.10](../../02-fondamentaux-langage/10-defer-panic-recover.md).

---

## Structures de contrôle : viser la ligne de mire

Idiome majeur : **le chemin heureux reste aligné à gauche**, les erreurs sortent tôt. On évite l'`else` inutile qui grimpe l'indentation.

```go
// ❌ else superflu, indentation qui gonfle
func lire(f string) ([]byte, error) {
    if data, err := os.ReadFile(f); err != nil {
        return nil, err
    } else {
        return traiter(data), nil
    }
}

// ✅ retour anticipé, lecture linéaire
func lire(f string) ([]byte, error) {
    data, err := os.ReadFile(f)
    if err != nil {
        return nil, err
    }
    return traiter(data), nil
}
```

Autres réflexes idiomatiques :

- **`if` avec initialisation** pour limiter la portée d'une variable : `if v, ok := m[k]; ok { … }`.
- **`switch` sans expression** pour remplacer une cascade de `if`/`else if`.
- Pas de `fallthrough` sauf intention explicite (Go ne « tombe » pas d'un cas à l'autre par défaut).

---

## Fonctions et méthodes

**Résultats nommés : avec parcimonie.** Utiles pour documenter, ou pour qu'un `defer` modifie la valeur de retour (enveloppe d'erreur) — mais les *naked returns* dans une longue fonction nuisent à la lisibilité.

```go
// usage légitime : enrichir l'erreur en un seul point via defer
func charger(path string) (cfg Config, err error) {
    defer func() {
        if err != nil {
            err = fmt.Errorf("chargement %q : %w", path, err)
        }
    }()
    // ... plusieurs points de sortie, tous enrichis
    return cfg, nil
}
```

**`defer` pour libérer les ressources**, au plus près de leur acquisition :

```go
f, err := os.Open(path)
if err != nil {
    return err
}
defer f.Close() // exécuté à la sortie de la fonction (LIFO)
```

> Piège : un `defer` **dans une boucle** s'accumule jusqu'à la fin de la fonction. Pour libérer à chaque tour, extraire le corps dans une fonction, ou fermer explicitement.

**Receveur valeur vs pointeur : rester cohérent.** Si une seule méthode a besoin d'un receveur pointeur (mutation, gros struct), **toutes** les méthodes du type devraient être à receveur pointeur.

```go
// ❌ mélange des deux → confusion, jeux de méthodes incohérents
func (c Compteur) Valeur() int { return c.n }
func (c *Compteur) Incr()      { c.n++ }

// ✅ un seul style pour tout le type
func (c *Compteur) Valeur() int { return c.n }
func (c *Compteur) Incr()       { c.n++ }
```

Rappels : **pas de surcharge** de fonctions, **pas de paramètres par défaut** — pour de la configuration optionnelle, utiliser des variadiques ou le motif des *functional options*.

---

## Interfaces : petites et côté consommateur

C'est le design signature de Go. Deux règles portent l'essentiel :

**1. Définir l'interface là où on la consomme, pas là où on l'implémente.** L'interface décrit un *besoin*, exprimé par l'appelant ; grâce à la satisfaction implicite, le type concret n'a rien à déclarer.

```go
// ✅ le consommateur ne demande QUE ce dont il a besoin
type userReader interface {
    User(ctx context.Context, id string) (*User, error)
}

func Bienvenue(ctx context.Context, r userReader, id string) (string, error) {
    u, err := r.User(ctx, id)
    if err != nil {
        return "", err
    }
    return "Bonjour " + u.Nom, nil
}
```

**2. Accepter des interfaces, renvoyer des types concrets.** Accepter une interface garde l'appelant flexible ; renvoyer une struct lui laisse toutes les capacités du type. (Guide, non dogme : renvoyer `error` — une interface — est parfaitement idiomatique, tout comme renvoyer parfois un `io.Reader`.)

**Anti-pattern : la pollution d'interfaces.** Créer une interface pour *une seule* implémentation « au cas où » ajoute de l'indirection sans bénéfice.

```go
// ❌ interface énorme, définie côté producteur, à une seule implémentation
type Store interface {
    Create(...) error
    Read(...) (X, error)
    Update(...) error
    Delete(...) error
    List(...) ([]X, error)
    // ... et 15 autres méthodes
}
```

> Introduisez l'abstraction quand une **deuxième implémentation réelle** ou un **double de test** la justifie — pas avant. Voir [§3.3](../../03-types-interfaces/03-interfaces.md).

**`any` (interface{}) ne dit rien.** Quand un type concret ou un générique suffit, préférez-les au filtrage de type à répétition.

```go
// ❌ any + type switch partout
func Traiter(v any) { /* switch v.(type) { ... } */ }

// ✅ générique typé (Go 1.18+)
func Traiter[T Message](v T) { /* ... */ }
```

---

## Composition plutôt qu'héritage

Go n'a pas d'héritage : on **compose** par *embedding*, qui promeut champs et méthodes.

```go
type baseHandler struct{ log *slog.Logger }

func (h baseHandler) logErr(err error) { h.log.Error("échec", "err", err) }

type UserHandler struct {
    baseHandler // embedding : logErr est promue
    repo UserRepo
}
```

Le proverbe « *a little copying is better than a little dependency* » complète l'idée : ne créez pas d'abstraction partagée pour économiser trois lignes. La sur-abstraction est un anti-pattern récurrent (et l'un des pièges de l'IA, cf. [§17.2](../../17-developpement-ia/02-pieges-ia.md)). Voir [§3.2](../../03-types-interfaces/02-composition-embedding.md).

---

## Données : slices, maps, zéro-value

**`new` vs `make`** : `new(T)` renvoie `*T` (à zéro) ; `make` initialise slices, maps et canaux (et renvoie le type, pas un pointeur).

```go
p := new(int)            // *int → 0
s := make([]int, 0, 10)  // slice initialisé, capacité 10
m := make(map[string]int)
ch := make(chan int)
```

**Écrire dans une map `nil` panique** — l'initialiser d'abord :

```go
// ❌ panic: assignment to entry in nil map
var m map[string]int
m["x"] = 1

// ✅
m := make(map[string]int)
m["x"] = 1
```

**Préallouer quand la taille est connue** évite les réallocations : `make([]T, 0, n)`.

**Piège d'aliasing d'`append`** : ajouter à un sous-slice peut écraser le tableau sous-jacent partagé.

```go
// ❌ b partage le tableau sous-jacent de a
a := []int{1, 2, 3, 4}
b := a[:2]
b = append(b, 99) // écrit dans a[2] ! → a == [1 2 99 4]
```

```go
// ✅ « full slice expression » a[:2:2] borne la capacité → nouvelle allocation
a := []int{1, 2, 3, 4}
b := a[:2:2]
b = append(b, 99) // a reste [1 2 3 4]
```

**`nil` slice vs slice vide** : les deux se parcourent et acceptent `append`. Nuance à connaître pour JSON : un slice `nil` s'encode en `null`, un slice vide en `[]` (cf. [§5.3](../../05-backend-http/03-json.md)).

**Faire de la valeur zéro une valeur utile** est un principe de conception : `sync.Mutex`, `bytes.Buffer`, `strings.Builder` s'emploient directement, sans constructeur.

```go
var mu sync.Mutex   // prêt à l'emploi
var buf bytes.Buffer
buf.WriteString("ok") // aucun make/New nécessaire
```

Détails en [§2.7](../../02-fondamentaux-langage/07-slices-maps.md).

---

## Concurrence idiomatique

« *Don't communicate by sharing memory; share memory by communicating.* » Les canaux orchestrent, les mutex sérialisent. (Chapitre complet : [§4](../../04-concurrence/README.md).)

**Ne jamais lancer une goroutine sans savoir comment elle s'arrête** — sinon, fuite de goroutine.

```go
// ❌ boucle sans condition d'arrêt : fuite
go func() {
    for { traiter(<-jobs) }
}()

// ✅ annulation via context
go func() {
    for {
        select {
        case <-ctx.Done():
            return
        case job := <-jobs:
            traiter(job)
        }
    }
}()
```

**Le `context.Context` porte l'annulation et les délais** : premier paramètre, nommé `ctx`, jamais stocké dans une struct.

```go
func Fetch(ctx context.Context, url string) (*Resp, error) { /* ... */ }
```

Voir [§4.4](../../04-concurrence/04-context.md).

**Fermeture des canaux : c'est l'émetteur qui ferme, jamais le récepteur.** Envoyer sur un canal fermé — ou le fermer deux fois — provoque un `panic`.

**Attendre proprement** : `sync.WaitGroup` pour la synchronisation, `golang.org/x/sync/errgroup` quand on veut aussi propager la première erreur et annuler le reste.

🆕 **Variable de boucle (Go 1.22).** Le piège historique — toutes les goroutines capturant le *même* `i` — a disparu : depuis Go 1.22 (et si votre `go.mod` déclare `go 1.22` ou plus), chaque itération possède sa propre variable. Le `x := x` défensif n'est plus nécessaire dans le code moderne ; restez vigilant en lisant du code dont le `go.mod` cible une version antérieure.

**Toujours tester le code concurrent avec le détecteur de courses** (`go test -race`), et `testing/synctest` pour un temps virtuel déterministe (cf. [§4.6](../../04-concurrence/06-tester-code-concurrent.md)).

---

## Packages et organisation

- **Pas de dépendances circulaires** : le compilateur les interdit ; c'est un signal de découpage à revoir.
- **`internal/`** cache les packages d'implémentation : ils ne sont importables que par le sous-arbre parent.
- **Un package = une responsabilité nommée.** Éviter les packages fourre-tout (`utils`, `helpers`, `common`) et le découpage purement technique par « couches » (`models`, `controllers`) quand un découpage par domaine est plus clair.
- **API de package minimale** : n'exportez que le nécessaire ; tout ce qui est en minuscule reste un détail d'implémentation libre d'évoluer.

Voir [§3.5](../../03-types-interfaces/05-organisation-code.md) et l'[annexe E — layout de projet](../layout-projet/README.md).

---

## Anti-patterns notables

Au-delà de ceux déjà cités, quelques pièges reviennent souvent en revue de code.

### `reflect` : à éviter dans le code applicatif

La réflexion est lente, contourne le typage statique et rend le code difficile à lire et à faire évoluer. Elle a sa place **à l'intérieur de bibliothèques** génériques — c'est d'ailleurs ce qui fait fonctionner `encoding/json`, qui inspecte les champs et lit les *struct tags* par réflexion :

```go
type User struct {
    ID   string `json:"id"`
    Nom  string `json:"nom"`
    seq  int    // non exporté : ignoré par encoding/json
}
```

Ce coût est acceptable *dans* la bibliothèque ; dans du code métier, préférez presque toujours un **type concret**, une **petite interface** ou un **générique**. La règle : si vous atteignez `reflect` pour votre propre logique, cherchez d'abord une conception sans réflexion.

### `unsafe` : hors périmètre, sauf cas extrême

`unsafe` brise les garanties de sûreté et de portabilité du langage et peut casser d'une version de Go à l'autre. Réservez-le à des cas très pointus (interopérabilité bas niveau, optimisation critique prouvée par le profilage) et documentez chaque usage. Dans l'immense majorité des projets : à proscrire.

### Autres pièges fréquents

| Anti-pattern | À la place |
|---|---|
| **Fuite de goroutine** (aucune condition d'arrêt) | `context` + `select { case <-ctx.Done() }` |
| **Ignorer les erreurs** (`_ =`) | Enrichir (`%w`) et remonter |
| **`panic` pour le contrôle de flux** | Erreurs comme valeurs ; `panic` = irrécupérable |
| **État global mutable partagé** | Injecter les dépendances (struct, constructeur) |
| **Sur-utilisation d'`init()`** (logique lourde, ordre implicite) | Constructeur explicite (`NewXxx`) |
| **Interface prématurée** (une seule implémentation) | Type concret ; abstraire au 2ᵉ besoin réel |
| **Optimisation prématurée** | Mesurer d'abord (pprof, benchmarks, cf. [§14.1](../../14-performance/01-pprof.md)) |
| **Getters `GetX`** | `X()` / `SetX()` |
| **Packages `utils` / `common`** | Package au nom métier |

---

## Outillage qui fait respecter l'idiome

Une grande partie de l'idiome Go est **vérifiable automatiquement**. À intégrer dès le départ, **des deux côtés** (GoLand et VS Code).

- **`gofmt` / `goimports`** — formatage et imports, à l'enregistrement.
  - *GoLand* : *Reformat Code* (⌘⌥L / Ctrl+Alt+L) ; activer **Settings → Tools → Actions on Save** (« Reformat code » + « Optimize imports »). `gofmt` est intégré.
  - *VS Code* : extension Go + `gopls` ; `"editor.formatOnSave": true` et `"editor.codeActionsOnSave": { "source.organizeImports": true }`.
- **`go vet`** — intégré à `go test` ; repère les constructions suspectes (verbes `Printf` erronés, copie de `Mutex`, *struct tags* mal formés…).
- **`staticcheck` / `golangci-lint`** — analyse statique avancée ; `golangci-lint` agrège `staticcheck`, `go vet` et bien d'autres, et fait office de standard de fait.
  - *GoLand* : inspections intégrées très riches (complémentaires), avec *quick-fixes* ; intégration `golangci-lint` configurable.
  - *VS Code* : `"go.lintTool": "golangci-lint"`, `"go.lintOnSave": "package"`.
- 🆕 **`go fix` (Go 1.26)** — réoutillé avec des « modernizers » qui réécrivent automatiquement d'anciens motifs en idiomes récents ; directive `//go:fix inline`. Détails en [§13.5](../../13-tests-qualite/05-linters.md).

L'objectif : que le style et de nombreux anti-patterns soient signalés (voire corrigés) **avant** la revue humaine, qui peut alors se concentrer sur la conception.

---

## Pour aller plus loin

- **Gestion des erreurs** en profondeur : [§2.9](../../02-fondamentaux-langage/09-gestion-erreurs.md) et [§2.10](../../02-fondamentaux-langage/10-defer-panic-recover.md).
- **Interfaces** et **composition** : [§3.2](../../03-types-interfaces/02-composition-embedding.md), [§3.3](../../03-types-interfaces/03-interfaces.md).
- **Concurrence** idiomatique : [§4](../../04-concurrence/README.md), context en [§4.4](../../04-concurrence/04-context.md).
- **Linters et outillage qualité** : [§13.5](../../13-tests-qualite/05-linters.md).
- **Correspondance depuis un autre langage** : [annexe A](../correspondance-go-autres/README.md).
- **Ressources et veille** (dont *Effective Go*, *Go Code Review Comments*, *Go Proverbs*) : [§18.3](../../18-strategie-roadmap/03-communaute-ressources.md).

---

🔝 [Retour au sommaire](../../SOMMAIRE.md) · ⏭️ [Annexe C — Bonnes pratiques de codage Go (+ avec l'IA 🤖)](../bonnes-pratiques/README.md)


⏭ [Bonnes pratiques de codage Go (+ avec l'IA 🤖)](/annexes/bonnes-pratiques/README.md)
