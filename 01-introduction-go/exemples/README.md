# Exemples du chapitre 01 — Introduction : Go et son écosystème

Chaque sous-dossier reconstitue un code source (ou concrétise un extrait) du chapitre, prêt à compiler. Tous les fichiers portent un en-tête indiquant la **section concernée**, la **description** et le **fichier source** `.md` d'origine. Tous ont été **compilés et testés** avec la toolchain **go1.26.0** ; les « sorties attendues » ci-dessous sont les sorties réellement constatées.

## Prérequis communs

**Installation :**
- **Go** : version 1.22 minimum (routage `"GET /"` du serveur) ; la formation cible **Go 1.26**. Installation : voir [section 1.4](../04-installation-outils.md) (`go.dev/dl`).
- **curl** (ou un navigateur) pour sonder le serveur HTTP.

**Configuration :**
- Aucune variable requise. Les `go.mod` déclarent `go 1.25.0`/`go 1.26.0` : si votre Go local est plus ancien, la gestion automatique de toolchain (`GOTOOLCHAIN=auto`, comportement par défaut — cf. [section 1.3](../03-ecosysteme-go.md)) télécharge la bonne version toute seule. Pour épingler explicitement : `GOTOOLCHAIN=go1.26.0 go build`.
- **Réseau** requis une seule fois pour `03-mon-service-uuid` (téléchargement de la dépendance via `proxy.golang.org`).

## Vue d'ensemble

| Dossier | Section | Fichier source | Ce que ça démontre |
|---|---|---|---|
| `01-serveur-http/` | 1.1 | `01-quest-ce-que-go.md` | Serveur web stdlib, routage par méthode |
| `03-mon-service-uuid/` | 1.3 | `03-ecosysteme-go.md` | `go.mod`, dépendance épinglée, `go.sum` |
| `04-vet-demo/` | 1.4 | `04-installation-outils.md` | `go vet` attrape ce que le compilateur accepte |
| `04-gofmt-demo/` | 1.4 | `04-installation-outils.md` | `gofmt -l` / `gofmt -w` |
| `05-hello/` | 1.5 | `05-premier-projet.md` | Premier programme, build, cross-compilation |

---

## 01-serveur-http — le serveur web de la section 1.1

- **Section** : 1.1 · **Fichier source** : `01-quest-ce-que-go.md`
- **Description** : serveur HTTP complet avec la seule bibliothèque standard — `ServeMux` moderne, routage par méthode (`"GET /"`), port 8080.
- **Lancer** :
  ```bash
  cd 01-serveur-http
  go run .            # ou : go build && ./serveur-http
  ```
- **Tester (autre terminal) et sortie attendue** :
  ```bash
  curl http://localhost:8080/            # → Bonjour depuis Go 👋
  curl -X POST http://localhost:8080/ -o /dev/null -w "%{http_code}\n" -s   # → 405
  ```
  Le `405 Method Not Allowed` en POST démontre le routage **par méthode** (`GET` seul est enregistré).
- **Arrêter** : `Ctrl+C` dans le terminal du serveur (ou `kill <pid>`).
- **Prérequis spécifiques** : port 8080 libre.

## 03-mon-service-uuid — le go.mod de la section 1.3, devenu projet

- **Section** : 1.3 · **Fichier source** : `03-ecosysteme-go.md`
- **Description** : le `go.mod` d'exemple de la section, **tel quel** (chemin de module, directive `go`, `require github.com/google/uuid v1.6.0`), complété d'un `main.go` qui consomme la dépendance. Le `go.sum` (empreintes cryptographiques, cf. 1.3) est fourni ; il a été généré par `go mod tidy`.
- **Compiler et exécuter** :
  ```bash
  cd 03-mon-service-uuid
  go run .
  ```
- **Sortie attendue** : un UUID v4, différent à chaque exécution, p. ex. :
  ```text
  5c795cc7-c5fb-4b31-a413-536f417aab52
  ```
- **Prérequis spécifiques** : réseau lors du premier build (message `go: downloading github.com/google/uuid v1.6.0`, puis cache local).

## 04-vet-demo — « go vet attrape un mauvais format Printf » (section 1.4)

- **Section** : 1.4 · **Fichier source** : `04-installation-outils.md`
- **Description** : programme qui **compile sans erreur** mais contient un `fmt.Printf("%d", "oops")` fautif. Les deux commandes font partie de la démo :
  ```bash
  cd 04-vet-demo
  go build ./...   # succès : le compilateur ne dit rien
  go vet ./...     # échec attendu — c'est le but
  ```
- **Sortie attendue de `go vet`** (code de sortie ≠ 0) :
  ```text
  main.go:14:14: fmt.Printf format %d has arg "oops" of wrong type string
  ```

## 04-gofmt-demo — reformater avec gofmt (section 1.4)

- **Section** : 1.4 · **Fichier source** : `04-installation-outils.md`
- **Description** : `bad_format.go.txt` est volontairement mal formaté (extension `.txt` pour ne pas être compilé ni reformaté par mégarde). La démo :
  ```bash
  cd 04-gofmt-demo
  cp bad_format.go.txt bad.go
  gofmt -l .    # liste bad.go (mal formaté)
  gofmt -w .    # réécrit au format canonique
  gofmt -l .    # plus rien : silence
  cat bad.go    # indentation en tabs, espacements normalisés
  rm bad.go     # on ne garde que le .txt de départ
  ```
- **Sortie attendue** : `gofmt -l` affiche `bad.go` avant, plus rien après ; le corps reformaté devient :
  ```go
  func main() {
  	x := 42
  	fmt.Println(x)
  }
  ```

## 05-hello — le premier projet de la section 1.5

- **Section** : 1.5 · **Fichier source** : `05-premier-projet.md`
- **Description** : le « hello world » du chapitre, avec le `go.mod` que `go mod init` écrit sous Go 1.26 (directive `go 1.25.0` — version N-1, cf. 1.5).
- **Compiler et exécuter** :
  ```bash
  cd 05-hello
  go run .          # → Bonjour, le monde !
  go build          # produit le binaire ./hello (hello.exe sous Windows)
  ./hello           # → Bonjour, le monde !
  ```
- **Cross-compilation** (section 1.5) :
  ```bash
  GOOS=windows GOARCH=amd64 go build   # hello.exe (PE32+ x86-64)
  GOOS=linux   GOARCH=arm64 go build   # binaire ELF ARM64, statiquement lié
  ```
- **Sortie attendue** : `Bonjour, le monde !`

---

## Nettoyage des binaires

Les binaires ne sont pas conservés dans le dépôt. Pour nettoyer après essais :

```bash
# depuis le dossier exemples/
rm -f 05-hello/hello 05-hello/hello.exe 01-serveur-http/serveur-http \
      03-mon-service-uuid/mon-service 04-vet-demo/vet-demo 04-gofmt-demo/bad.go
# ou, module par module (supprime le binaire du nom du module) :
go clean
```

---

*Tous les exemples testés le 2026-07-04 (toolchain go1.26.0, Linux amd64) : sorties conformes au chapitre.*
