# Exemples du chapitre 17 — Développer en Go avec l'IA

Ce module est **conceptuel** (comment bien utiliser l'IA en Go), donc ses exemples illustrent surtout le **code idiomatique que l'IA devrait produire** et **l'outillage qui rattrape ses écarts**. Trois projets sont exécutables (stdlib pure) ; le quatrième rassemble des **fichiers de configuration** (pas de Go). Chaque fichier porte un en-tête **Section / Description / Fichier source**. Tout le code a été **compilé, vérifié (`go vet`) et exécuté** (toolchain **go1.26.0**, Linux amd64).

## Prérequis communs

**Installation** : **Go 1.25 minimum** (la formation cible Go 1.26 ; `02-pieges/modernize` **exige Go 1.26** pour `go fix`). **Aucune dépendance externe**, **aucun accès réseau**, **aucun Docker** : tout est en stdlib.  
**Lancer** : `cd <dossier> && go run .` (ou `go test ./...`).

## Vue d'ensemble

| Dossier | Section | Fichier source | Ce que ça démontre |
|---|---|---|---|
| `01-prompting/` | 17.1 | `01-prompting-go.md` | le code idiomatique d'un bon prompt (`LoadConfig`) + `AGENTS.md` |
| `02-pieges/` | 17.2 | `02-pieges-ia.md` | pièges ❌/✅ + **`modernize/`** : les 6 réécritures `go fix` du §4.2 |
| `03-tests-migration/` | 17.3 | `03-tests-migration-ia.md` | test **table-driven**, migration exception→erreur |
| `04-assistants/` | 17.4 | `04-assistants-ide.md` | fichiers de config IA (`.aiignore`, `settings.json`) |

---

## 01-prompting — section 17.1 (`01-prompting-go.md`)

**Description** : le Go qu'un **prompt cadré** produit (signature imposée, stdlib, erreurs wrappées `%w`, jamais de `panic`) — le `LoadConfig` qu'on n'a pas à réécrire. Le `AGENTS.md` voisin est l'exemple d'**instructions persistantes** (règles Go *non mécanisables*, l'outillage se chargeant du reste).  
**Lancer** : `go run .`  
**Sortie attendue** :

```text
config chargée : Addr=:8080 Timeout=30
fichier absent → erreur wrappée : lecture config "absent.json" : open absent.json: no such file or directory
```

## 02-pieges — section 17.2 (`02-pieges-ia.md`)

**Description** : les pièges récurrents de l'IA, en versions ✅ idiomatiques compilables (champ exporté vs getter, `defer Close` **avec capture d'erreur**, `error` plutôt que `panic`, assertion à deux valeurs), plus un banc qui les vérifie. Le sous-dossier **`modernize/`** contient un fichier « Go de 2020 » que **`go fix` redresse**.  
**Lancer** : `go test ./...` · **Modernize** : `cd modernize && go fix -diff ./...`  
**Sortie attendue** (`go fix -diff` — les 6 réécritures du tableau §4.2) :

```diff
+func vague(v any) any { return v }        // interface{} → any
+	return os.ReadFile(path)               // ioutil.ReadFile → os.ReadFile
+	x := max(b, a)                          // min/max manuel → max()
+	maps.Copy(dst, src)                     // boucle map → maps.Copy
+	var s strings.Builder                   // s += x → strings.Builder
+	for i := range n {                      // for i := 0; i < n → range n
```

> `go fix ./...` applique ces réécritures — **sûres, préservant la sémantique, conscientes de la ligne `go` du `go.mod`**.

## 03-tests-migration — section 17.3 (`03-tests-migration-ia.md`)

**Description** : le test **table-driven** idiomatique (`discount_test.go`) — dont chaque cas *affirme* un comportement attendu, pas le comportement observé — et la **migration** exception→erreur (`LoadUser` : `raise NotFoundError` du Python devient `(nil, ErrNotFound)`, testable par `errors.Is`).  
**Lancer** : `go run .` · **Tester** : `go test -v ./...`  
**Sortie attendue** :

```text
Discount(100,10)=90 · Discount(100,150) rejeté=true
LoadUser(1) = {ID:1 Name:alice}
LoadUser(99) → errors.Is(ErrNotFound)=true : utilisateur 99 : utilisateur introuvable
```

`go test ./...` → `ok` (5 sous-tests table-driven).

## 04-assistants — section 17.4 (`04-assistants-ide.md`) — configuration (pas de Go)

**Description** : cette section est conceptuelle (assistants IDE). Ses artefacts concrets sont des **fichiers de configuration** : `.aiignore` (tenir secrets et `vendor/` hors du contexte de l'IA, côté GoLand/Junie) et `settings.json` (associer Copilot à l'extension Go dans VS Code — l'IA écrit, gopls et golangci-lint vérifient).  
**Utilisation** : copier `.aiignore` à la racine du dépôt ; fusionner `settings.json` dans `.vscode/settings.json`. Rien à exécuter.

---

## Nettoyage des binaires et résidus

`go run` / `go test` ne laissent aucun binaire. `01-prompting` crée puis supprime `config.json` lui-même ; après un `go build` manuel : `go clean ./...`.

**Aucun conteneur** n'est utilisé par ce module (il ne met en jeu aucun backend) : rien à arrêter, supprimer, ni aucune image ou volume à purger côté Docker.

---

*Tous les exemples testés le 2026-07-06 (toolchain go1.26.0, Linux amd64). Sorties conformes au chapitre.*
