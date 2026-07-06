# Exemples du chapitre 04 — Concurrence : le point fort de Go

Un exemple **complet et exécutable par section** : les extraits des `.md` y sont assemblés en programmes auto-démonstratifs, conçus pour des **sorties déterministes** (compteurs et tris plutôt qu'ordres d'ordonnancement). L'exemple de la section 4.6 est, par nature, un **package de tests**. Chaque fichier porte un en-tête **Section / Description / Fichier source**. Tous ont été **compilés, vérifiés (`go vet`) et exécutés** avec la toolchain **go1.26.0** — sorties ci-dessous telles que constatées.

## Prérequis communs

**Installation** : **Go 1.25 minimum** (`wg.Go`, `testing/synctest`) — la formation cible Go 1.26 ; cf. [section 1.4](../../01-introduction-go/04-installation-outils.md).  
**Configuration** : aucune (`GOTOOLCHAIN=auto` sélectionne la toolchain d'après le `go.mod`). **Réseau** requis une fois pour `03-synchronisation` (module `golang.org/x/sync` ; `go.sum` fourni).  
**Lancer** : `cd <dossier> && go run .` — sauf `06-tester` : `go test ./...` puis `go test -race ./...`.

## Vue d'ensemble

| Dossier | Section | Fichier source | Ce que ça démontre |
|---|---|---|---|
| `01-goroutines/` | 4.1 | `01-goroutines.md` | arguments figés, loopvar 1.22, attente, panique récupérée, runtime |
| `02-channels/` | 4.2 | `02-channels.md` | rendez-vous, buffer, fermeture/diffusion, `select`, générateur |
| `03-synchronisation/` | 4.3 | `03-synchronisation.md` | `WaitGroup`/`wg.Go`, `Mutex`, atomic, `OnceValue`, `errgroup` |
| `04-context/` | 4.4 | `04-context.md` | annulation, timeout, motif `select`, `WithValue`, `Cause`/`AfterFunc`/`WithoutCancel` |
| `05-patterns/` | 4.5 | `05-patterns-concurrence.md` | pipeline ordonné, fan-in/out, **preuve anti-fuite** |
| `06-tester/` | 4.6 | `06-tester-code-concurrent.md` | `-race` sur code synchronisé, `synctest.Test`/`Wait` |

---

## 01-goroutines — section 4.1 (`01-goroutines.md`)

**Description** : le cycle de vie complet — arguments d'un `go` **figés à l'instruction**, variable de boucle **par itération** (Go 1.22), attente par `WaitGroup` (jamais `time.Sleep`), **panique récupérée dans la goroutine** (le programme survit — le filet de la section), `GOMAXPROCS`/`NumGoroutine`.  
**Sortie attendue** (extraits ; GOMAXPROCS dépend de la machine) :

```text
valeur reçue : 1  (et non 2 )
valeurs capturées : [0 1 2]
goroutines terminées : 5
panique récupérée : boom
le programme a survécu ✔
```

## 02-channels — section 4.2 (`02-channels.md`)

**Description** : rendez-vous synchronisant, tampon (`len`/`cap`), fermeture (`range` draine puis s'arrête, virgule-ok `false`), **la fermeture est une diffusion** (3 goroutines réveillées par un seul `close`), `select` (cas d'**envoi** choisi, `default` non bloquant, canal `nil` jamais choisi), générateur idiomatique.  
**Sortie attendue** (extraits) :

```text
reçu : prêt  (envoi et réception synchronisés)
len = 2 · cap = 2
après fermeture : v = 0 · ok = false
goroutines réveillées d'un coup : 3
envoyé sur ch2 → 42
default : rien de prêt (le canal nil est ignoré)
```

## 03-synchronisation — section 4.3 (`03-synchronisation.md`)

**Description** : `WaitGroup` classique (`Add` avant / `defer Done`) **et** `wg.Go` 🆕 1.25, le `Counter` de la section (map + `Mutex`) sous 100 goroutines, `atomic`, `sync.OnceValue` (une seule exécution), `errgroup` — première erreur propagée, annulation du groupe observée, `SetLimit(2)` respecté.  
**Prérequis spécifiques** : réseau au premier build (`golang.org/x/sync`).  
**Sortie attendue** (extraits) :

```text
count["clé"] = 100  (exact : le verrou protège la map)
valeur = 42 · exécutions de f : 1
   l'autre tâche voit l'annulation : true
g.Wait() → échec rapide
concurrence maximale observée : 2 (limite : 2)
```

## 04-context — section 4.4 (`04-context.md`)

**Description** : `WithCancel` (sortie exacte « arrêt : context canceled »), `WithTimeout`/`DeadlineExceeded`, le **motif `worker`** (`select` jobs/annulation), `WithValue` avec clé non exportée et accesseurs typés, et les raffinements : `WithCancelCause`/`Cause`, `AfterFunc`, `WithoutCancel`.  
**Sortie attendue** (extraits) :

```text
arrêt : context canceled
DeadlineExceeded : true
RequestID : req-42 · présent : true
context.Cause → source indisponible
WithoutCancel : survit à l'annulation du parent ✔
```

## 05-patterns — section 4.5 (`05-patterns-concurrence.md`)

**Description** : pipeline annulable (**ordre préservé** : 4, 9, 16), fan-out de deux étages `sq` + fan-in `merge` (l'ordre se perd, les valeurs non), et **la preuve anti-fuite** : après un `break` précoce, `cancel()` dénoue tout le pipeline — `NumGoroutine` identique avant/après.  
**Sortie attendue** (extraits) :

```text
sq → 4
sq → 9
sq → 16
résultats triés : [4 9 16 25] (l'arrivée était désordonnée)
goroutines : 1 avant → 1 après (aucune fuite) ✔
```

## 06-tester — section 4.6 (`06-tester-code-concurrent.md`) — package de tests

**Description** : trois tests — le compteur **synchronisé** (`atomic` : la version corrigée du `TestCounter` de la section, propre sous `-race`), le **timeout en horloge virtuelle** (`synctest.Test` : 1 s virtuelle, instantané en réel), et `synctest.Wait` (stabiliser une goroutine d'arrière-plan : 3 minutes virtuelles en quelques millisecondes). Pour **reproduire la course** de la section : remplacez l'`atomic` par `n++` sur un entier partagé et relancez `go test -race ./...` — le rapport `WARNING: DATA RACE` apparaît.  
**Lancer et sortie attendue** :

```console
$ go test ./...
ok  	github.com/exemple/tester	0.003s
$ go test -race ./...
ok  	github.com/exemple/tester	1.015s   # plus lent : l'instrumentation -race
```

---

## Nettoyage des binaires

`go run .` et `go test` ne laissent aucun binaire. Après un `go build` manuel : `go clean` dans le dossier concerné. Le `go.sum` de `03-synchronisation` fait partie de l'exemple (empreintes de dépendance, cf. section 1.3).

---

*Tous les exemples testés le 2026-07-04 (toolchain go1.26.0, Linux amd64) : sorties conformes au chapitre.*
