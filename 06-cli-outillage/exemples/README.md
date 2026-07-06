# Exemples du chapitre 06 — Applications CLI et outillage

Un exemple **complet par section**. Un CLI se démontre en l'**invoquant** : chaque fiche ci-dessous liste des invocations et leurs **sorties exactes, telles que constatées** (toolchain **go1.26.0**, Linux amd64). Les exemples 02 et 03 embarquent en plus leurs **tests** (`go test ./...`) : exécution in-memory d'une commande Cobra, et modèle Bubble Tea testé **sans terminal**. Chaque fichier porte un en-tête **Section / Description / Fichier source** et des commentaires d'intention.

## Prérequis communs

**Installation** : **Go 1.25 minimum** (la formation cible Go 1.26) ; cf. [section 1.4](../../01-introduction-go/04-installation-outils.md). **GoReleaser** et **Docker** uniquement pour les volets optionnels de `04-distribution`.  
**Configuration** : aucune (`GOTOOLCHAIN=auto`). **Réseau** au premier build pour `02-cobra-viper` (spf13) et `03-tui-bubbletea` (charmbracelet) — `go.sum` fournis.  
**Lancer** : `cd <dossier> && go run . <arguments>` (voir chaque fiche).

## Vue d'ensemble

| Dossier | Section | Fichier source | Ce que ça démontre |
|---|---|---|---|
| `01-flag-args-env/` | 6.1 | `01-flag-args-env.md` | squelette `run(ctx)` complet : drapeaux typés, `Func`/`TextVar`, précédence env, pièges |
| `02-cobra-viper/` | 6.2 | `02-cobra-viper.md` | arbre Cobra par constructeurs + Viper : **précédence à 4 niveaux**, tests in-memory |
| `03-tui-bubbletea/` | 6.3 | `03-tui-bubbletea.md` | le compteur TEA (v1) + le modèle **testé sans TTY** |
| `04-distribution/` | 6.4 | `04-distribution.md` | estampillage 2 voies, builds statiques, cross-compilation, `.goreleaser.yaml` |

---

## 01-flag-args-env — section 6.1 (`01-flag-args-env.md`)

**Description** : le squelette idiomatique d'un outil stdlib — `FlagSet` en `ContinueOnError` dans `run(ctx, args) error` (les `defer` s'exécutent, `Ctrl-C` honoré via `signal.NotifyContext`), drapeaux typés (`Int`/`Bool`/`Duration`), **répétable** (`flag.Func`), **auto-validé** (`flag.TextVar` + `netip`), précédence **drapeau > `$PORT` > défaut**, résultats sur stdout / diagnostics sur stderr.  
**Invocations et sorties constatées** :

```console
$ go run . -port 9000 -verbose -timeout 2s fichier.txt
port=9000 verbose=true timeout=2s addr=127.0.0.1
arguments positionnels : [fichier.txt]
travail terminé

$ go run .                            # défaut
port=8080 …
$ PORT=9999 go run .                  # l'environnement l'emporte sur le défaut
port=9999 …
$ PORT=9999 go run . -port 7777       # le drapeau l'emporte sur tout
port=7777 …

$ go run . -H "Accept: application/json" -H "Authorization: Bearer x" -addr ::1
addr=::1 · en-têtes : [Accept: application/json Authorization: Bearer x]

# Les DEUX pièges de la section, reproduits :
$ go run . fichier.txt -port 9000     # positionnel d'abord → -port NON interprété
arguments positionnels : [fichier.txt -port 9000]
$ go run . -verbose=false             # un booléen se nie avec « = » (jamais « -verbose false »)
```

## 02-cobra-viper — section 6.2 (`02-cobra-viper.md`)

**Description** : l'assemblage complet de la section — commandes par **constructeurs** (`newRootCmd`/`newServeCmd`, ni globales ni `init()`), `RunE` + `SilenceUsage`, `cobra.NoArgs`, `--config` persistant, câblage Viper dans `PersistentPreRunE` (`monoutil.yaml` fourni + env `MONOUTIL_*` + `BindPFlag`).  
**Prérequis spécifiques** : réseau au premier build (`spf13/cobra`, `spf13/viper`).  
**Invocations et sorties constatées** (le `monoutil.yaml` du dossier porte `port: 6000`) :

```console
$ go run . serve                                  # fichier > défaut
écoute sur :6000
$ MONOUTIL_PORT=7000 go run . serve               # env > fichier
écoute sur :7000
$ MONOUTIL_PORT=7000 go run . serve --port 9000   # drapeau > env > fichier
écoute sur :9000
$ go run . --version
monoutil version 1.4.0
$ go run . sevre                                  # suggestion automatique
Did you mean this?
	serve
```

**Tests** (exécution in-memory, sans binaire — `SetArgs`/`SetOut`/`Execute`) :

```console
$ go test ./...
ok  	github.com/exemple/monoutil
```

## 03-tui-bubbletea — section 6.3 (`03-tui-bubbletea.md`)

**Description** : le compteur de la section — l'architecture Elm (Model / `Update` / `View`, récepteur par valeur, `tea.Quit`), API **v1**. La TUI elle-même exige un **vrai terminal** ; la logique, elle, se teste **sans TTY** — c'est la vertu de TEA démontrée par `model_test.go` (+1 +1 −1 → 1, « q » → `tea.Quit`).  
**Prérequis spécifiques** : réseau au premier build (`charmbracelet/bubbletea`) ; un TTY pour `go run .`.  
**Lancer / tester** :

```console
$ go run .          # dans un vrai terminal : ↑/↓ (ou k/j) ajustent, q quitte
Compteur : 0
↑/↓ pour ajuster · q pour quitter

$ go test ./...     # AUCUN terminal nécessaire
ok  	github.com/exemple/compteur
```

## 04-distribution — section 6.4 (`04-distribution.md`)

**Description** : l'outil à distribuer (layout `./cmd/monoutil`) avec les **deux voies d'estampillage** — variables `-ldflags "-X …"` et `runtime/debug.ReadBuildInfo` — plus le **`.goreleaser.yaml` de la section** (v2, matrice 3 OS × 2 arch, statique, archives, checksums).  
**Prérequis spécifiques** : GoReleaser pour le volet release (`go install github.com/goreleaser/goreleaser/v2@latest`) ; Docker pour la variante container.  
**Builds et sorties constatées** :

```console
$ go build -o monoutil ./cmd/monoutil && ./monoutil
monoutil dev (commit none, date unknown)
version du module (ReadBuildInfo) : (devel)

$ CGO_ENABLED=0 go build -trimpath \
    -ldflags="-s -w -X main.version=1.2.3 -X main.commit=abc123 -X main.date=2026-07-05" \
    -o monoutil ./cmd/monoutil
$ file monoutil            # → statically linked
$ ./monoutil
monoutil 1.2.3 (commit abc123, date 2026-07-05)

$ GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o monoutil.exe ./cmd/monoutil
$ file monoutil.exe        # → PE32+ executable (console) x86-64

$ goreleaser check                        # valide le .goreleaser.yaml
$ goreleaser release --snapshot --clean   # 6 binaires + archives + checksums dans dist/
```

**Variante Docker** — le « binaire unique » dans un container nu (le pitch de la section) :

```console
$ CGO_ENABLED=0 go build -o monoutil ./cmd/monoutil       # statique
$ docker run --rm -v "$PWD/monoutil:/monoutil:ro" busybox /monoutil   # lancer (–-rm : auto-supprimé)
monoutil dev (commit none, date unknown)
$ docker ps -a                                            # (vide : --rm a nettoyé)
$ docker rmi busybox                                      # supprimer l'image téléchargée
$ docker system df                                        # vérifier : 0 B, aucun volume créé
$ rm -f monoutil                                          # supprimer le binaire
```

---

## Nettoyage des binaires

`go run .` et `go test` ne laissent rien. Après un `go build` manuel : `go clean` (et `rm -f monoutil monoutil.exe` dans `04-distribution`, `rm -rf dist/` après un snapshot GoReleaser). Les `go.sum` de 02 et 03 font partie des exemples (empreintes de dépendances, cf. section 1.3).

---

*Tous les exemples testés le 2026-07-05 (toolchain go1.26.0, Linux amd64) : sorties conformes au chapitre. La variante container (binaire statique dans busybox) et le snapshot GoReleaser (6 binaires) ont été validés lors des tests exhaustifs du chapitre.*
