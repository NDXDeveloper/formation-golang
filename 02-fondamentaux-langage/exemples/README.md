# Exemples du chapitre 02 — Fondamentaux du langage

Un exemple **complet et exécutable par section** du chapitre : les extraits des `.md` (fragments pédagogiques) y sont assemblés en programmes auto-démonstratifs, qui affichent ce que le chapitre annonce. Chaque fichier `.go` porte un en-tête indiquant la **section concernée**, la **description** et le **fichier source** d'origine. Tous ont été **compilés, vérifiés (`go vet`) et exécutés** avec la toolchain **go1.26.0** — les sorties ci-dessous sont les sorties réellement constatées.

## Prérequis communs

**Installation :**
- **Go** : 1.23 minimum (itérateurs de la section 2.5) ; **1.26 requis** pour `02-types-variables` et `08-pointeurs` (`new(expr)`). La formation cible Go 1.26 — cf. [section 1.4](../../01-introduction-go/04-installation-outils.md).

**Configuration :**
- Aucune variable requise : avec `GOTOOLCHAIN=auto` (défaut), la bonne toolchain est sélectionnée/téléchargée d'après le `go.mod` de chaque exemple. Pour épingler : `GOTOOLCHAIN=go1.26.0 go run .`
- **Réseau** requis une seule fois pour `01-packages-greeting` (dépendance `github.com/google/uuid` du bloc d'imports de 2.1 ; `go.sum` fourni).

**Lancer un exemple** (tous se lancent pareil) :

```bash
cd <dossier-de-l-exemple>
go run .
```

## Vue d'ensemble

| Dossier | Section | Fichier source | Ce que ça démontre |
|---|---|---|---|
| `01-packages-greeting/` | 2.1 | `01-structure-packages.md` | package multi-fichiers, visibilité, imports, `init` |
| `02-types-variables/` | 2.2 | `02-types-variables.md` | conversions, déclarations, shadowing, `iota`, zéro-values, `new(expr)` 🆕 |
| `03-fonctions/` | 2.3 | `03-fonctions.md` | retours multiples, variadiques, closures |
| `04-conditions/` | 2.4 | `04-conditions.md` | `if` à initialisation, `switch`, `type switch` |
| `05-boucles-iterateurs/` | 2.5 | `05-boucles.md` | `for` unique, `break` étiqueté, loopvar 1.22, `range`, itérateurs 1.23 🆕 |
| `06-chaines-runes/` | 2.6 | `06-chaines.md` | octets vs runes, `strings`/`strconv`/`fmt` |
| `07-slices-maps/` | 2.7 | `07-slices-maps.md` | capacité, `append`, pièges de partage, `clear`, maps |
| `08-pointeurs/` | 2.8 | `08-pointeurs.md` | `&`/`*`, passage par valeur, `append` invisible, `new(expr)` 🆕 |
| `09-gestion-erreurs/` | 2.9 | `09-gestion-erreurs.md` | `%w`, `errors.Is`/`As`, sentinelles, erreurs typées |
| `10-defer-panic-recover/` | 2.10 | `10-defer-panic-recover.md` | LIFO, évaluation immédiate, frontière `recover`, `MustXxx` |

---

## 01-packages-greeting — section 2.1 (`01-structure-packages.md`)

**Description** : module réel multi-packages — le package `greeting` du chapitre (exporté vs non exporté), le programme qui le consomme via son chemin d'import, le bloc d'imports stdlib/externe **tel quel**, et la preuve de l'ordre d'initialisation (`startedAt`, `init`, puis `main`). La ligne commentée `greeting.defaultName` ne compile pas si on la décommente — c'est la démonstration de la visibilité.  
**Prérequis spécifiques** : réseau au premier build (uuid).  
**Sortie attendue** :

```text
Bonjour, Ada !
Bonjour, le monde !
init exécuté avant main : true
startedAt initialisée   : true
```

## 02-types-variables — section 2.2 (`02-types-variables.md`) — Go 1.26

**Description** : conversions toujours explicites, les formes de déclaration, le **piège du shadowing** (l'`err` externe reste `<nil>`), les constantes non typées (`pi` s'adapte), `iota` (énumération + drapeaux), les zéro-values utiles, et `new(expr)` 🆕.  
**Sortie attendue** (extraits clés) :

```text
i = 42 · f = 42 · u = 42
a = 0 · b = 7 · c = 7 · d = 7
e = 1 · f2 = deux · g = false · h = 3.14
err externe après le bloc : <nil>
circonférence = 12.56636 · maxUsers = 100 · greeting = salut
Sunday = 0 · Saturday = 6
Read = 1 · Write = 2 · Execute = 4
buf contient : ok
*p = 42 · *q = 42 · *r = ok
```

## 03-fonctions — section 2.3 (`03-fonctions.md`)

**Description** : `add`, le couple `(résultat, error)` avec `divide`, le motif `(valeur, ok)`, `split` (retours nommés + `return` nu), `sum` variadique et l'éclatement `xs...`, `applyTwice` (fonction valeur) et la closure `counter`.  
**Sortie attendue** (extraits) :

```text
add(2, 3) = 5
divide(10, 2) = 5
divide(1, 0) → erreur : division par zéro
split(17) → 7 10
sum(1, 2, 3) = 6
sum(xs...) = 15
applyTwice(double, 3) = 12
next() = 1 puis 2
```

## 04-conditions — section 2.4 (`04-conditions.md`)

**Description** : `if`/`else if`/`else`, `if` à initialisation (portée confinée, message « trop long : 29 caractères »), `switch` à cas multiples, à initialisation (`runtime.GOOS`), sans expression, et `type switch` sur `any`.  
**Sortie attendue** (extraits ; la ligne « système : » dépend de l'OS — `linux` ici) :

```text
score 85 → note B
checkLen(long)    → trop long : 29 caractères
samedi → week-end
vendredi → presque !
lundi → au travail
système : linux
entier : 42
chaîne de 2 caractères
booléen : true
type inconnu : float64
```

## 05-boucles-iterateurs — section 2.5 (`05-boucles.md`)

**Description** : les trois formes du `for`, `break` étiqueté (3 cellules visitées), la variable de boucle **par itération** (Go 1.22 : `0 1 2`, plus jamais `3 3 3`), `range` sur chaîne (indices d'octet !), sur entier, l'itérateur `Compte` (`iter.Seq`, Go 1.23) et `slices.Values`.  
**Sortie attendue** (extraits) :

```text
cible 3 trouvée après 3 cellules visitées
0 1 2 
0:h 1:é 3:l 4:l 5:o 
0 1 2 
0 1 2 
total = 30
```

## 06-chaines-runes — section 2.6 (`06-chaines.md`)

**Description** : `s[0]`=octet 104, `len`=6 octets mais 5 runes, chaîne brute, `[]rune`, `strings` (+ `Builder`), `strconv`, verbes `fmt`.  
**Sortie attendue** (extraits) :

```text
s[0] = 104 · len(s) = 6 · chemin = C:\temp
octets : 6 · runes : 5
0:h 1:é 3:l 4:l 5:o 
runes[1] = 233 · string(rune(65)) = A
true GO [a b c] a-b "salut"
Atoi = 42 · ParseFloat = 3.14 · ParseBool = true · Itoa = 42
42 go true
"salut"
float64
```

## 07-slices-maps — section 2.7 (`07-slices-maps.md`)

**Description** : copie complète des tableaux, longueur/capacité, croissance d'`append` (0 2 → 2 2 → 3 4), **les pièges** (partage du tableau, `append` qui écrase `a[2]`) et **les parades** (découpage à trois indices, `slices.Clone`), le package `slices`, les maps (comma-ok, `delete`, `clear`, lecture d'une map `nil`).  
**Sortie attendue** (extraits) :

```text
b = [1 2 3] · d = [9 2 3]
départ      : 0 2
append(1,2) : 2 2
append(3)   : 3 4
base après vue[0]=99 : [10 99 30 40]
a après append sur b2 : [1 2 99 4 5]
a2 préservé : [1 2 3 4 5] · c2 : [1 2 99]
après clear : len = 0
lecture map nil : 0
```

## 08-pointeurs — section 2.8 (`08-pointeurs.md`) — Go 1.26

**Description** : `&`/`*`, zéro-value `nil`, **tout est passé par valeur** (10 vs 20), le cas des slices (élément modifié visible, `append` invisible), `&Point{}`, `new(int)` et `new(42)` 🆕.  
**Sortie attendue** (extraits) :

```text
après doubleValeur  : 10
après doublePointeur : 20
après remplir : [99 0 0]
après ajouter : [99 0 0]
&Point{1,2} → {1 2} · *new(int) = 0
*new(42) = 42
```

## 09-gestion-erreurs — section 2.9 (`09-gestion-erreurs.md`)

**Description** : `errors.New`/`fmt.Errorf`, **wrapping `%w` sur une vraie erreur d'E/S** (fichier inexistant — `errors.Is(err, fs.ErrNotExist)` la retrouve à travers la chaîne), sentinelle (`errors.Is` ✔, `==` ✘ après wrap), erreur personnalisée extraite par `errors.As`.  
**Sortie attendue** :

```text
withdraw(100, 30) → 70 <nil>
withdraw(10, 20)  → erreur : solde insuffisant
withdraw(5, -1)   → erreur : montant invalide : -1
erreur enrichie : lecture de /chemin/inexistant.conf : open /chemin/inexistant.conf: no such file or directory
errors.Is(err, fs.ErrNotExist) → true
errors.Is(err, ErrNotFound) → true
err == ErrNotFound          → false (le wrap casse ==)
champ en cause : email
message : champ "email" : requis
```

## 10-defer-panic-recover — section 2.10 (`10-defer-panic-recover.md`)

**Description** : `defer` à côté de l'acquisition (`os.Open` sur `exemple.txt` fourni), **ordre LIFO**, **arguments évalués immédiatement** (le `defer` affiche `i = 0` bien que `i` vaille 10 ensuite), `defer` + retour nommé (`doWork : échec`), **frontière `recover`** (le programme survit à la panique) et le motif `MustXxx`. Les `defer` se déroulent après la dernière ligne de `main` — observez la fin de la sortie.  
**Sortie attendue (fin)** :

```text
panique récupérée : valeur négative interdite : -3
le programme a survécu à la panique ✔
re.MatchString("123") → true
re.MatchString("12a") → false
--- fin de main : les defer se déroulent maintenant (LIFO) ---
defer affiche i = 0
defer n° 3
defer n° 2
defer n° 1
```

---

## Nettoyage des binaires

`go run .` ne laisse aucun binaire. Après un `go build`, supprimez l'exécutable produit (nom du module, p. ex. `hello`, ou le nom passé à `-o`) :

```bash
go clean        # dans le dossier de l'exemple — supprime le binaire du module
```

Le `go.sum` de `01-packages-greeting` n'est **pas** un résidu : il fait partie de l'exemple (empreintes des dépendances, cf. section 1.3).

---

*Tous les exemples testés le 2026-07-04 (toolchain go1.26.0, Linux amd64) : sorties conformes au chapitre.*
