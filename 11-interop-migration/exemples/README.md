# Exemples du chapitre 11 — Interopérabilité et migration

Trois projets, un par lisière. **Aucun service externe ni container** — mais deux chaînes particulières : le 01 touche la **frontière C** (gcc requis pour deux de ses trois démos), le 02 se construit **vers WebAssembly** (chaîne Go officielle seule). Chaque fichier porte un en-tête **Section / Description / Fichier source** et des commentaires d'intention. Tous ont été **compilés, vérifiés et exécutés** (toolchain **go1.26.0**, Linux amd64) — sorties telles que constatées.

## Prérequis communs

**Installation** : **Go 1.25 minimum** (la formation cible Go 1.26) ; **gcc** pour `01-cgo-ffi` (démos cgo et export — `purego`, lui, n'en a pas besoin : c'est son intérêt) ; rien d'autre (le moteur wasm du 02 est `wazero`, en Go pur).  
**Configuration** : aucune (`GOTOOLCHAIN=auto`). **Réseau** au premier build (purego, wazero — `go.sum` fournis).  
**Lancer** : les commandes exactes figurent dans chaque section.

## Vue d'ensemble

| Dossier | Section | Fichier source | Prérequis | Ce que ça démontre |
|---|---|---|---|---|
| `01-cgo-ffi/` | 11.1 | `01-cgo-ffi.md` | gcc (sauf purego) | cgo verbatim (mémoire C, `CGO_ENABLED`), **purego sans cgo**, **exposer du Go à C** (`//export`, c-shared, `main.c` réel) |
| `02-webassembly/` | 11.2 | `02-webassembly.md` | — | **la boucle Go→wasm→Go** : invité `go:wasmexport` (réacteur) hébergé par **wazero** (Go pur, capacités WASI) |
| `03-migration/` | 11.3 | `03-migrer-vers-go.md` | — | la façade **strangler fig** (stdlib) et le piège de la **translittération** (panic vs erreur-valeur) |

---

## 01-cgo-ffi — section 11.1 (`01-cgo-ffi.md`) — gcc requis (sauf purego)

**Description** : la frontière C dans les deux sens. `cmd/strlen-cgo` : le bloc cgo du cours (préambule C, `C.CString` **à libérer soi-même**, appel `C.c_strlen`) — refusé au build si `CGO_ENABLED=0`, c'est le point de bascule. `cmd/purego-libc` : appeler la libc **sans cgo** (chargement à l'exécution, build 100 % Go). `export/` : **exposer du Go à C** — `//export Additionner` + `-buildmode=c-shared` produisent `libgo.so` **et l'en-tête `libgo.h` généré**, consommés par le `main.c` de `consommateur/` (qui vit hors du package Go : sinon `go build` tenterait de le compiler avant que `libgo.h` n'existe).

```console
$ CGO_ENABLED=1 go run ./cmd/strlen-cgo          # → 6
$ CGO_ENABLED=0 go run ./cmd/purego-libc         # → Appel de C depuis Go, sans cgo   (dans un terminal)
$ cd export && go build -buildmode=c-shared -o libgo.so .
$ cd consommateur && gcc main.c -o prog_c -I.. -L.. -lgo -Wl,-rpath,'$ORIGIN/..'
$ ./prog_c
C appelle Go : Additionner(19, 23) = 42
bonjour depuis la bibliothèque Go
$ cd .. && rm -f libgo.so libgo.h consommateur/prog_c   # nettoyage des artefacts
```

*(L'ordre des deux lignes de sortie peut s'inverser : le `printf` C et le `fmt.Println` Go ont chacun leur tampon.)*

## 02-webassembly — section 11.2 (`02-webassembly.md`) — autonome

**Description** : la boucle complète. `guest/` : du Go compilé **vers** `wasip1` qui exporte `add` via **`go:wasmexport`** (Go 1.24+), en mode **réacteur** (`-buildmode=c-shared` — l'hôte appelle `_initialize`, pas `_start`). `host/` : le bloc **wazero** du cours — moteur **en Go pur, sans cgo** — qui accorde les imports WASI (la sécurité par capacités), instancie l'invité et appelle `add(2, 3)`. Le `.md` embarque le wasm par `//go:embed` ; ici on le lit sur disque pour ne pas committer de binaire.

```console
$ cd guest && GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -o ../host/add.wasm .
$ cd ../host && go run .
5
$ rm -f add.wasm                                  # nettoyage de l'artefact wasm
```

**Sortie attendue** : `5` — `add(2, 3)` exécutée dans le bac à sable, l'hôte restant `CGO_ENABLED=0`. Pour éprouver les autres cibles du cours : `GOOS=js GOARCH=wasm go build` (≈ 2,5 Mo, exécutable via `node $(go env GOROOT)/lib/wasm/wasm_exec_node.js`) et la CLI `wazero run` (`go install github.com/tetratelabs/wazero/cmd/wazero@latest`).

## 03-migration — section 11.3 (`03-migrer-vers-go.md`) — autonome

**Description** : les deux piliers. **La façade strangler fig** du cours — `ServeMux` + `httputil.NewSingleHostReverseProxy`, stdlib seule : la tranche migrée `/api/orders/` est servie **par Go**, tout le reste part vers **l'existant** (le service hérité, simulé par `httptest`) ; on retire l'ancien chemin tranche par tranche. Puis **le piège n°1** : la translittération « Java/Python-en-Go » (émuler les exceptions avec `panic` — le service meurt) face au Go idiomatique (l'erreur est une **valeur**).  
**Lancer** : `go run .`  
**Sortie attendue** :

```text
/api/orders/42 → commandes servies par GO (tranche migrée)
/api/users/7   → servi par L'EXISTANT : /api/users/7
idiomatique      : Alice · err = <nil>
translittération : panic → utilisateur introuvable (le service meurt — à fuir)
```

---

## Nettoyage des binaires

`go run` ne laisse rien. Les artefacts des cycles C et wasm (`libgo.so`, `libgo.h`, `prog_c`, `add.wasm`) se suppriment avec les commandes `rm` incluses ci-dessus — aucun n'est versionné. Pas de container ni volume dans ce chapitre.

---

*Tous les exemples testés le 2026-07-05 (toolchain go1.26.0, gcc 13, Linux amd64). Sorties conformes au chapitre.*
