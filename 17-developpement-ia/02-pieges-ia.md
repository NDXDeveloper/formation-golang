🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 17.2 Pièges de l'IA en Go (code non idiomatique, erreurs ignorées, sur-abstraction)

La [section précédente](01-prompting-go.md) cherchait à *prévenir* le Go non idiomatique par le prompt. Même bien prompté, un modèle garde des angles morts systématiques en Go — cette section les catalogue pour les *reconnaître et les corriger*.

Une bonne nouvelle traverse tout le module : la plupart de ces pièges sont **mécaniquement détectables** (le compilateur, `go vet`, `staticcheck`/`errcheck`, le détecteur de races, et désormais les modernizers de `go fix`). Les vrais dangers sont les autres : le code qui *compile* et *passe une relecture rapide* tout en étant faux ou bancal. D'où le réflexe mental à garder :

> **« Ça compile » n'est pas « c'est correct et idiomatique ».**

## 1. Code non idiomatique

L'IA importe des réflexes venus de Java ou de Python. Les symptômes sont reconnaissables.

### 1.1 Getters/setters et « classes » déguisées

```go
// ❌ Réflexe Java : accesseurs sur des champs triviaux.
type User struct{ name string }

func (u *User) GetName() string  { return u.name }
func (u *User) SetName(n string) { u.name = n }

// ✅ Go : champ exporté — sauf s'il y a une vraie logique à encapsuler.
type User struct{ Name string }
```

Même chose pour les types en `-Manager`, `-Helper`, `-Util` : en Go, ce sont le plus souvent des fonctions ou un simple paquet, pas des objets fourre-tout.

### 1.2 Le « stutter » et le nommage

Le paquet fait déjà partie du nom qualifié ; le répéter bégaie à l'usage.

```go
// ❌  paquet "config" → l'appelant écrit config.ConfigService
type ConfigService struct{}
func NewConfigService() *ConfigService { return &ConfigService{} }

// ✅  config.Service, config.New() — se lisent bien
type Service struct{}
func New() *Service { return &Service{} }
```

Autres conventions que l'IA néglige : receveurs courts (`u`, pas `user`), `error` en dernière valeur de retour, `MixedCaps` (jamais de `snake_case`).

### 1.3 Des types trop vagues

L'IA atteint vite `any` (ou `interface{}`) là où un type concret — ou un générique — exprimerait l'intention. Un `map[string]any` pour une structure connue, une fonction `func(any) any` : autant de renoncements au typage qui font perdre à Go son principal filet. Demandez le type le plus précis que la situation permet. (Le cas mécanique `interface{}` vs `any` relève de la §4.2.)

## 2. Erreurs ignorées — le piège le plus dangereux

C'est la catégorie qui casse la correction *en silence*, sans que rien ne se voie à la lecture.

### 2.1 L'erreur avalée

```go
_ = doSomething()                       // ❌ ignorée explicitement
if err != nil { log.Println(err) }      // ❌ journalisée puis... on continue comme si de rien n'était
```

Le linter `errcheck` (inclus dans golangci-lint) attrape la première forme ; la seconde exige de la vigilance.

### 2.2 `defer f.Close()` sur un fichier en écriture

Piège subtil et fréquent : `Close` peut échouer (c'est souvent lui qui *flushe* réellement), et l'ignorer perd une écriture tronquée.

```go
// ❌ L'erreur de Close est perdue → écriture potentiellement incomplète, sans alerte.
func writeFile(path string, data []byte) error {
    f, err := os.Create(path)
    if err != nil {
        return err
    }
    defer f.Close()
    _, err = f.Write(data)
    return err
}

// ✅ On capture l'erreur de Close via un retour nommé.
func writeFile(path string, data []byte) (err error) {
    f, err := os.Create(path)
    if err != nil {
        return err
    }
    defer func() {
        if cerr := f.Close(); cerr != nil && err == nil {
            err = cerr
        }
    }()
    _, err = f.Write(data)
    return err
}
```

(Pour un fichier ouvert en *lecture seule*, `defer f.Close()` sans capture est acceptable — l'erreur n'y porte pas de donnée perdue.)

### 2.3 `panic` au lieu de `return`

```go
// ❌ panic dans du code de bibliothèque : fait tomber l'appelant.
func mustParse(s string) int {
    n, err := strconv.Atoi(s)
    if err != nil {
        panic(err)
    }
    return n
}

// ✅ Retourner l'erreur ; laisser l'appelant décider (cf. 2.10).
func parse(s string) (int, error) {
    return strconv.Atoi(s)
}
```

### 2.4 Wrapping absent

```go
// ❌ Perd le contexte : d'où vient l'erreur ?
if err != nil {
    return err
}
// ✅ Contexte + chaîne préservée pour errors.Is / errors.As (cf. 2.9).
if err != nil {
    return fmt.Errorf("chargement utilisateur %d : %w", id, err)
}
```

### 2.5 Assertions de type non vérifiées

```go
s := x.(string)          // ❌ panique si x n'est pas un string
s, ok := x.(string)      // ✅ forme à deux valeurs
if !ok {
    return fmt.Errorf("type inattendu : %T", x)
}
```

## 3. Sur-abstraction

La catégorie la plus *séduisante*, parce qu'elle a l'air « professionnelle ». En Go, l'abstraction se paie et se mérite.

### 3.1 Interfaces prématurées

```go
// ❌ Interface à une seule implémentation, définie côté producteur.
type UserStore interface {
    Get(ctx context.Context, id int) (*User, error)
}
type pgUserStore struct{ db *sql.DB }
func NewUserStore(db *sql.DB) UserStore { return &pgUserStore{db} }

// ✅ Retourner la struct concrète. L'interface se définit côté CONSOMMATEUR, si un besoin réel apparaît.
type UserStore struct{ db *sql.DB }
func NewUserStore(db *sql.DB) *UserStore { return &UserStore{db} }
```

L'idiome Go : *accept interfaces, return structs* ([§ 3.3](../03-types-interfaces/03-interfaces.md)). Une interface anticipée « au cas où » est presque toujours de trop.

### 3.2 Génériques et abstractions inutiles

L'IA sur-emploie les génériques ([§ 3.4](../03-types-interfaces/04-generiques.md)) : un `Map[T, U]` là où une boucle est plus claire, un conteneur générique qui réinvente `slices`/`maps`. De même, les options fonctionnelles (`WithX(...)`) pour une struct à deux champs sont une cérémonie sans contrepartie — un littéral de struct suffit.

### 3.3 Cérémonie d'architecture

Demandez « un service utilisateur » et vous risquez d'obtenir des couches *clean architecture*, un conteneur d'injection de dépendances et trois *factories* pour un programme de 200 lignes. Le Go idiomatique préfère la simplicité jusqu'à ce que la complexité l'exige ([§ 10.2](../10-architecture-services/02-clean-architecture.md)).

## 4. Les pièges que la relecture rapide ne voit pas

Ceux qui *compilent* et qu'un survol laisse passer — souvent spécifiques à Go.

### 4.1 Concurrence

C'est le terrain où l'IA se trompe le plus, avec assurance : goroutines qui **fuient** (lancées sans moyen de les arrêter, sans `context`), **races** sur de la mémoire partagée sans synchronisation, interblocages, `sync.WaitGroup` mal séquencé. La parade décisive est le détecteur de races — `go test -race` transforme un bug invisible en échec net ([§ 4.6](../04-concurrence/06-tester-code-concurrent.md)). Passez toujours un `context.Context` et ne lancez pas de goroutine dont personne n'attend la fin ([§ 4.1](../04-concurrence/01-goroutines.md), [§ 4.4](../04-concurrence/04-context.md)).

Cas particulier, à connaître pour ne pas se laisser désorienter : **la portée des variables de boucle a changé en Go 1.22.**

```go
// Go < 1.22 : « u := u » était nécessaire pour capturer la bonne valeur.
// Go 1.22+ : chaque itération a sa propre variable — la copie est superflue.
for _, u := range urls {
    go handle(u) // capture correcte, sans « u := u »
}
```

Entraînée sur du code plus ancien, l'IA ajoute souvent ce `u := u` devenu inutile, ou vous met en garde contre un bug qui n'existe plus. À recouper avec votre version de Go (le modernizer `forvar` supprime d'ailleurs cette copie superflue).

### 4.2 Du « Go de 2020 » : APIs périmées

Les modèles sont entraînés sur l'immense corpus de Go *déjà publié* — qui penche vers ce qu'on écrivait il y a trois, quatre ou cinq ans. Résultat : du code d'avant les génériques, d'avant les paquets `slices`/`maps`, d'avant `min`/`max` intégrés, d'avant le `range` sur entier. Ça marche, mais c'est daté.

| Motif « de 2020 » produit par l'IA | Idiome moderne | Modernizer `go fix` |
|---|---|---|
| `interface{}` | `any` (Go 1.18) | `any` |
| `ioutil.ReadFile` | `os.ReadFile` (Go 1.16) | `//go:fix inline` |
| `s += x` dans une boucle | `strings.Builder` | `stringsbuilder` |
| boucle manuelle copiant une map | `maps.Copy` / `maps.Keys` (Go 1.21+) | `mapsloop` |
| `min`/`max` réécrits à la main | `min` / `max` intégrés (Go 1.21) | `minmax` |
| `for i := 0; i < n; i++` | `for i := range n` (Go 1.22) | `rangeint` |

L'antidote est arrivé avec **Go 1.26** : `go fix` a été réécrit sur le cadre d'analyse de `go vet` et embarque une vingtaine de **modernizers** — des réécritures **sûres, préservant la sémantique et conscientes de la version `go.mod`** (elles ne proposent rien de plus récent que ce que déclare votre module). Lancez `go fix ./...` (ou `go fix -diff ./...` pour prévisualiser) après une génération, et gopls fait remonter les mêmes suggestions *en direct* dans l'éditeur. La directive `//go:fix inline` va plus loin : une bibliothèque peut annoter une fonction dépréciée pour que `go fix` réécrive automatiquement les appels chez ses utilisateurs.

Distinct des APIs *périmées* : les APIs **hallucinées** (fonctions, méthodes ou paquets qui n'existent pas). Celles-là, au moins, ne compilent pas — le compilateur est votre premier relecteur.

### 4.3 Sécurité écrite avec assurance

L'IA produit sans broncher une requête SQL concaténée, un `InsecureSkipVerify: true`, ou un `math/rand` pour un jeton. Ces motifs sont traités au [module 16](../16-securite/README.md) ([§ 16.1](../16-securite/01-owasp-go.md), [§ 16.2](../16-securite/02-cryptographie-tls.md)) ; `gosec` les signale. Le fait qu'ils *compilent* est précisément ce qui les rend dangereux.

## 5. Le filet : quel outil attrape quel piège

La thèse du module, rendue opérationnelle. À chaque famille de pièges correspond un garde-fou mécanique :

- **erreurs ignorées** → `errcheck` / `staticcheck` (via golangci-lint) ;
- **style et non-idiomatique** → `staticcheck`, gopls ;
- **APIs périmées** → modernizers de `go fix` + gopls (suggestions inline) ;
- **races de concurrence** → `go test -race` ;
- **failles de sécurité** → `gosec` ;
- **APIs hallucinées** → le compilateur (le code ne construit pas).

Côté IDE, ces garde-fous s'activent des deux côtés du double environnement de la formation : **VS Code** (extension Go) exécute gopls — les modernizers s'affichent à la frappe — et golangci-lint à la sauvegarde, `go fix` se lançant depuis le terminal ou une *task* ; **GoLand** remonte les mêmes diagnostics gopls et ses propres inspections, et lance `go fix` / golangci-lint via une configuration d'exécution. Le détail d'installation est en [§ 13.5](../13-tests-qualite/05-linters.md) et [§ 17.4](04-assistants-ide.md).

La limite est nette, et c'est là que « l'outillage dispose » s'arrête : la chaîne attrape les pièges *mécaniques* (erreurs, races, APIs périmées, une partie du style). Elle ne voit **pas** la sur-abstraction ni une mauvaise conception — ce résidu-là reste du ressort de la **relecture humaine**.

## En résumé

- « Ça compile » ≠ « c'est correct et idiomatique » : l'IA a des angles morts systématiques en Go.
- **Non idiomatique** : getters/setters, « classes » déguisées, *stutter*, types trop vagues → champs exportés, fonctions, noms sans répétition, types précis.
- **Erreurs ignorées** (le plus dangereux) : `_ =`, log-et-continue, `defer Close` sur un writer, `panic` au lieu de `return`, wrapping absent, assertions non vérifiées.
- **Sur-abstraction** : interfaces prématurées, génériques inutiles, cérémonie d'architecture → simplicité d'abord, interface côté consommateur.
- **Pièges invisibles à la relecture** : concurrence (races, fuites — `-race`), « Go de 2020 » périmé (corrigé par `go fix` modernizers en Go 1.26), sécurité qui compile.
- **Le filet** : chaque piège mécanique a son outil ; la sur-abstraction et le design, eux, restent affaire de relecture humaine.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [17.3 Génération de tests, migration assistée, revue de code par IA](03-tests-migration-ia.md)

⏭ [Génération de tests, migration assistée, revue de code par IA](/17-developpement-ia/03-tests-migration-ia.md)
