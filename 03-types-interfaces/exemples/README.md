# Exemples du chapitre 03 — Types, méthodes et interfaces

Un exemple **complet et exécutable par section** : les extraits des `.md` y sont assemblés en programmes auto-démonstratifs ; la section 3.5 (organisation) est illustrée par **deux mini-projets réels** (layout `cmd/`+`internal/`, et un workspace `go work`). Chaque fichier porte un en-tête **Section / Description / Fichier source**. Tous ont été **compilés, vérifiés (`go vet`) et exécutés** avec la toolchain **go1.26.0** — les sorties ci-dessous sont celles réellement constatées.

## Prérequis communs

**Installation** : **Go** 1.25 minimum ; **1.26 requis** pour `04-generiques` (auto-référence F-bornée). Cf. [section 1.4](../../01-introduction-go/04-installation-outils.md).  
**Configuration** : aucune — `GOTOOLCHAIN=auto` (défaut) sélectionne la bonne toolchain d'après chaque `go.mod`/`go.work`. Aucun réseau requis (aucune dépendance externe).  
**Lancer** : `cd <dossier> && go run .` (cas particuliers de 3.5 indiqués plus bas).

## Vue d'ensemble

| Dossier | Section | Fichier source | Ce que ça démontre |
|---|---|---|---|
| `01-structs-methodes/` | 3.1 | `01-structs-methodes.md` | littéraux, tags JSON, comparabilité, receveurs, method sets |
| `02-composition-embedding/` | 3.2 | `02-composition-embedding.md` | promotion, **pas** de dispatch virtuel, décorateur |
| `03-interfaces/` | 3.3 | `03-interfaces.md` | satisfaction implicite, **nil typé**, assertions, capacité optionnelle |
| `04-generiques/` | 3.4 | `04-generiques.md` | contraintes, `~`, `Stack`, auto-référence 🆕 1.26 |
| `05-organisation-code/` | 3.5 | `05-organisation-code.md` | `cmd/`+`internal/`, workspace `go work` |

---

## 01-structs-methodes — section 3.1 (`01-structs-methodes.md`)

**Description** : tout le socle des types — littéraux (nommé/positionnel/`&T{}`), zéro-value, struct anonyme, **tags de champs** (JSON réel), **struct comme clé de map**, méthodes (type dérivé `Celsius`, méthode-valeur vs méthode-expression), **receveur valeur vs pointeur** (`Counter`), method sets (affectation muette ; la ligne `var _ Stringer = T{}` en commentaire ne compile pas si on la réactive), élément de map non adressable (le détour copier-modifier-réaffecter), `Registry` sous verrou, **receveur pointeur `nil`** (`List`).  
**Sortie attendue** (extraits clés) :

```text
json.Marshal(u1) → {"id":1,"name":"Ada"}
p1 == p2 → true
m[Point{1,2}] → présent
Celsius(100).Fahrenheit() → 212
après IncValue  : {0}
après IncPointer: {1}
mc["a"] = {1}
(*List)(nil).Len() → 0
```

## 02-composition-embedding — section 3.2 (`02-composition-embedding.md`)

**Description** : promotion (`Dog`/`Animal`), pointeur embarqué (`*log.Logger`), la démonstration centrale — **pas de dispatch virtuel** (`Greeting()` → « Bonjour, Base ») et sa solution idiomatique (interface injectée), `Cache` (Mutex embarqué : `Lock` promue) vs `SafeCache` (champ nommé encapsulé), **décorateur** `LoggingStore` (seule `Get` est décorée), interface composée (`ReadWriter` satisfaite par `*bytes.Buffer`).  
**Sortie attendue** (extraits) :

```text
d.Speak() → Rex émet un son
dv.Hello()    → Derived
dv.Greeting() → Bonjour, Base  ← et NON « Bonjour, Derived »
via interface → Bonjour, Derived
GET "clé"
Get → valeur
ReadWriter via *bytes.Buffer → ok
```

## 03-interfaces — section 3.3 (`03-interfaces.md`)

**Description** : satisfaction **implicite** (`EmailSender` ne nomme jamais `Notifier`), affectation muette, petites interfaces universelles (`firstLine` sur un `io.Reader`), « accepter des interfaces, retourner des structs », **le piège du nil typé** (un `*MyError` nil rangé dans `error` n'est **pas** nil), assertions virgule-ok, type switch, **capacité optionnelle** (`io.Closer` fermé seulement s'il existe).  
**Sortie attendue** (extraits) :

```text
typedNil()  == nil → false  ← LE PIÈGE : pointeur nil typé ≠ interface nil
literalNil() == nil → true
i.(int)    → 0 false  (pas de panique grâce à ok)
describe → erreur : échec
Close appelé sur le fermable : true
```

## 04-generiques — section 3.4 (`04-generiques.md`) — Go 1.26

**Description** : `Map` avec inférence, `Set[T comparable]` et la struct vide, union `~int | ~float64` (`Sum` accepte `Celsius`), `Max[cmp.Ordered]`, type générique `Stack` (`var zero T`), **auto-référence F-bornée 🆕 Go 1.26** (`Adder[A Adder[A]]` — `Total` conserve le type concret `Vec`), génériques de la stdlib (`slices`).  
**Sortie attendue** (extraits) :

```text
Map(longueurs) → [2 4]
Sum([]int) → 6 · Sum([]Celsius) → 4
Pop sur pile vide → false  (var zero T)
Total([]Vec) → {4 6}  — le type concret Vec est conservé
Sort → [1 2 3] · Contains(2) → true · BinarySearch(2) → 1 true
```

## 05-organisation-code — section 3.5 (`05-organisation-code.md`) — deux mini-projets

### `projet/` — layout `cmd/` + `internal/`

Un module réel `github.com/acme/app` : le binaire `cmd/api` importe le package privé `internal/store`. **Lancer** :

```bash
cd 05-organisation-code/projet
go run ./cmd/api
```

**Sortie attendue** : `store interne : accessible depuis cmd/api (même module)`
Depuis **tout autre module**, l'import de `github.com/acme/app/internal/store` serait refusé : `use of internal package … not allowed` (garanti par le compilateur).

### `workspace/` — développer deux modules ensemble avec `go work`

`app` consomme `example.com/lib`, **jamais publiée** : sans le `go.work`, `app` ne compile pas ; avec lui, la version **locale** est utilisée. **Lancer** :

```bash
cd 05-organisation-code/workspace/app
go run .
```

**Sortie attendue** : `app → version LOCALE de lib, vue via go.work`  
**À essayer** : modifiez `lib/lib.go` (`Message()`) et relancez — la modification est prise **immédiatement**, sans publication ni `replace`.
*Note : un `go.work` est normalement exclu du versionnement ; il est fourni ici parce qu'il **est** l'objet de l'exemple.*

---

## Nettoyage des binaires

`go run` ne laisse aucun binaire. Après un `go build` manuel : `go clean` dans le dossier concerné (ou supprimez l'exécutable produit).

---

*Tous les exemples testés le 2026-07-04 (toolchain go1.26.0, Linux amd64) : sorties conformes au chapitre.*
