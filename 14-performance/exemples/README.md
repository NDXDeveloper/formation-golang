# Exemples du chapitre 14 — Performance et gestion de la mémoire

Un projet par section, illustrant le **cycle de la performance** : profiler (01) → comprendre les allocations (02) → optimiser (03) → prouver le gain (04). **Tous autonomes**, en bibliothèque standard — **aucun conteneur, aucun service externe**. Chaque fichier porte un en-tête **Section / Description / Fichier source** et des commentaires d'intention. Tout a été **compilé, vérifié (`go vet`) et exécuté** (toolchain **go1.26.0**, Linux amd64) — sorties telles que constatées.

## Prérequis communs

**Installation** : **Go 1.25 minimum** (la formation cible Go 1.26). Pour `04` (et la voie benchstat de `03`) : **benchstat** — `go install golang.org/x/perf/cmd/benchstat@latest`.  
**Configuration** : aucune (`GOTOOLCHAIN=auto`). **Aucune dépendance externe** ni accès réseau : tout est en stdlib.  
**Pas de Docker** : ce chapitre ne met en jeu aucun backend — il n'y a donc **aucune image à télécharger, aucun conteneur à lancer ou supprimer, aucun volume**.

## Vue d'ensemble

| Dossier | Section | Fichier source | Ce que ça démontre |
|---|---|---|---|
| `01-pprof/` | 14.1 | `01-pprof.md` | net/http/pprof, runtime/pprof, `go tool pprof`, **flight recorder** (1.25) |
| `02-gc-allocations/` | 14.2 | `02-gc-allocations.md` | **escape analysis** (`-gcflags=-m`), `gctrace`, `GOMEMLIMIT` |
| `03-optimisations/` | 14.3 | `03-optimisations-pgo.md` | préallocation, `strings.Builder`, **`sync.Pool`**, **PGO** |
| `04-benchmarking/` | 14.4 | `04-benchmarking.md` | workflow **benchstat** (significativité statistique) |

---

## 01-pprof — section 14.1 (`01-pprof.md`)

**Description** : un service volontairement gourmand, profilable de trois façons — `net/http/pprof` sur un écouteur **interne** (`localhost:6060`), `runtime/pprof` (profils CPU + tas écrits au démarrage), et un benchmark (`main_test.go`) pour la voie `-cpuprofile`. Le sous-programme `flightrecorder/` illustre l'enregistreur de vol (Go 1.25).  
**Lancer** :
- `go run .` — démarre le service ; puis `go tool pprof http://localhost:6060/debug/pprof/profile?seconds=5`
- `go test -run=^$ -bench=. -benchmem -cpuprofile=cpu.out -memprofile=mem.out` — profils via benchmark
- `go tool pprof -top cpu.out` · `go tool pprof -http=:8080 cpu.out` (flame graph)
- `go run ./flightrecorder` → `go tool trace trace.out`

**Sortie attendue** (extraits) :

```text
# benchmark
BenchmarkTravail-4   121   10081657 ns/op   5603396 B/op   20001 allocs/op
# profil de tas (go tool pprof -sample_index=alloc_space)
  280.51MB 97.96%   main.travail
# service
/debug/pprof/ → 200   ·   /travail → "50000 tampons alloués"
```

> ⚠️ `/debug/pprof` n'est **jamais** exposé publiquement — ici sur `localhost` uniquement.

## 02-gc-allocations — section 14.2 (`02-gc-allocations.md`)

**Description** : l'*escape analysis* rendue visible (`escape/geo.go`) et le GC observable (`main.go` alloue en boucle). `go build -gcflags=-m` montre les décisions du compilateur ; `GODEBUG=gctrace=1` imprime un cycle de GC par ligne ; `GOMEMLIMIT` force le GC même avec `GOGC=off`.  
**Lancer** :
- `go build -gcflags=-m ./escape/` — escape analysis
- `GODEBUG=gctrace=1 go run .` — trace du GC
- `GOMEMLIMIT=32MiB GOGC=off GODEBUG=gctrace=1 go run .` — la limite souple déclenche le GC
- `GOEXPERIMENT=nogreenteagc go build .` — repli sur l'ancien GC (Go 1.26)

**Sortie attendue** (escape analysis) :

```text
escape/geo.go:20:2: moved to heap: p        ← NewPoint renvoie l'adresse d'une locale
escape/geo.go:25:11: pts does not escape     ← SumX ne fait que lire : reste sur la pile
```

```text
# GODEBUG=gctrace=1 : une ligne par cycle
gc 1 @0.014s 2%: 0.21+0.94+0.22 ms clock, ... 3->3->0 MB, 4 MB goal, ... 4 P
```

## 03-optimisations — section 14.3 (`03-optimisations-pgo.md`)

**Description** : les trois techniques « allouer moins », chacune avec sa version lente et rapide, prouvées au banc (`-benchmem`) ; plus un sous-projet **`pgo/`** montrant l'auto-détection de `default.pgo`.  
**Lancer** :
- `go test -run=^$ -bench=. -benchmem` — compare les allocations lentes vs rapides
- PGO : dans `pgo/`, `go test -run=^$ -bench=. -cpuprofile=default.pgo .` puis `go build .` (détecte `default.pgo` ; visible avec `go build -x`)

**Sortie attendue** (allocations par opération) :

```text
BenchmarkNamesSlow-4          11 allocs/op     BenchmarkNamesFast-4        1 alloc/op
BenchmarkJoinSlow-4          999 allocs/op     BenchmarkJoinFast-4         1 alloc/op
BenchmarkFormatSansPool-4   1000 allocs/op     BenchmarkFormatAvecPool-4   0 allocs/op
```

Le `pgo/default.pgo` fourni (~900 o) est un profil CPU réel ; `go build -x` montre `-pgoprofile=…` transmis au compilateur — PGO **sans changement de code**.

## 04-benchmarking — section 14.4 (`04-benchmarking.md`) — benchstat requis

**Description** : deux implémentations de `Reverse` (lente/rapide) pour dérouler le **workflow benchstat** — mesurer chacune avec `-count=10`, puis comparer statistiquement.  
**Prérequis** : `go install golang.org/x/perf/cmd/benchstat@latest`.  
**Workflow** :

```sh
# 1. mesurer ReverseSlow (l'appel dans BenchmarkReverse)
go test -run=^$ -bench=Reverse -benchmem -count=10 > old.txt
# 2. basculer l'appel sur ReverseFast, puis :
go test -run=^$ -bench=Reverse -benchmem -count=10 > new.txt
# 3. comparer
benchstat old.txt new.txt
```

**Sortie attendue** (extrait) :

```text
          │    sec/op    │   sec/op     vs base                │
Reverse-4   233.16µ ± 3%   13.44µ ± 4%  -94.24% (p=0.000 n=10)
```

Un `p` < 0,05 et une variation faible : le gain est **réel**. Ne jamais mesurer sous `-race` ni `-cover` (l'instrumentation fausse tout : ×3,6 mesuré).

---

## Nettoyage des binaires et résidus

`go run` / `go test` ne laissent aucun binaire. Les profils produits (`cpu.out`, `mem.out`, `cpu.prof`, `heap.prof`, `trace.out`, `old.txt`, `new.txt`) sont des **artefacts de travail** — les supprimer après analyse :

```sh
find . -name '*.out' -o -name '*.prof' -o -name '*.txt' | grep -v go.sum | xargs rm -f
go clean ./...
```

Le seul fichier binaire **conservé volontairement** est `03-optimisations/pgo/default.pgo` : un profil PGO se **commite** avec le code (§14.3).

---

*Tous les exemples testés le 2026-07-06 (toolchain go1.26.0, Linux amd64). Les chiffres absolus dépendent de la machine ; seules comptent les comparaisons relatives.*
