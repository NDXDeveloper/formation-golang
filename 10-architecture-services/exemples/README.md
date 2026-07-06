# Exemples du chapitre 10 — Architecture de services

Trois projets, un par section — **tous autonomes** : aucun service externe, aucune base, aucun container (la démo « service distant » du 01 utilise `httptest`). Chaque fichier porte un en-tête **Section / Description / Fichier source** et des commentaires d'intention. Tous ont été **compilés, vérifiés (`go vet`) et exécutés** (toolchain **go1.26.0**, Linux amd64) — sorties telles que constatées.

## Prérequis communs

**Installation** : **Go 1.25 minimum** (la formation cible Go 1.26) ; **golangci-lint v2** uniquement pour jouer la règle `depguard` du 01 (`go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`) — facultatif pour exécuter les démos.  
**Configuration** : aucune (`GOTOOLCHAIN=auto`). **Réseau** au premier build du 03 (godotenv — `go.sum` fourni) ; 01 et 02 sont en **stdlib pure**.  
**Lancer** : la commande figure dans chaque section ci-dessous.

## Vue d'ensemble

| Dossier | Section | Fichier source | Ce que ça démontre |
|---|---|---|---|
| `01-monolithe-modulaire/` | 10.1 | `01-monolithe-vs-microservices.md` | packages par domaine, `internal/`, **interface consommateur**, composition root — et **l'extraction : un client réseau satisfait la même interface, `orders` inchangé** |
| `02-hexagonal/` | 10.2 | `02-clean-architecture.md` | domaine **pur**, port `Repository`, adaptateurs *driven* (postgres) et *driving* (HTTP+DTO), **faux en mémoire** — testé sans base |
| `03-config-flags/` | 10.3 | `03-configuration-12factor.md` | précédence **défauts < env < flags**, *fail fast*, godotenv `Load`/`Overload`, **feature flags derrière un port**, rechargement `SIGHUP` |

---

## 01-monolithe-modulaire — section 10.1 (`01-monolithe-vs-microservices.md`)

**Description** : le monolithe modulaire du cours, structuré comme l'arborescence de la section — `internal/catalog` (petite surface : `Service` exporté, `store` non exporté, propriétaire de « ses tables »), `internal/orders` (l'interface `catalogGetter` **définie par le consommateur**), `cmd/server` (le *composition root* : de simples constructeurs). La démo joue **la thèse du chapitre** : d'abord les appels **en mémoire**, puis l'extraction — le catalogue devient un service distant (`httptest`) et `internal/catalogclient` satisfait la **même** interface : `orders` ne change pas d'une ligne. Le `.golangci.yml` porte la règle `depguard` du cours (format **v2**).  
**Lancer** : `go run ./cmd/server` — puis, si golangci-lint v2 est installé : `golangci-lint run` (→ `0 issues.` ; ajoutez un import interdit dans `catalog/` pour voir la détection).  
**Sortie attendue** :

```text
=== 1. Le monolithe modulaire : appels en mémoire ===
commande possible : Clavier mécanique (appel EN MÉMOIRE) (8900 ct) · err = <nil>

=== 2. L'extraction : le catalogue devient un service DISTANT ===
commande possible : Clavier mécanique (appel RÉSEAU) (8900 ct) · err = <nil>

orders est resté identique entre les deux mondes : extraction = refactor mécanique ✔
```

## 02-hexagonal — section 10.2 (`02-clean-architecture.md`)

**Description** : ports & adaptateurs à la Go, avec **l'asymétrie volontaire** du cours — le domaine `internal/orders` est **pur** (entité sans tag JSON, port `Repository` déclaré côté consommateur, règle métier dans `Place`) ; l'adaptateur *driven* `postgres/` satisfait le port **implicitement** (prouvé à la compilation par `var _ orders.Repository = (*postgres.OrderRepo)(nil)`) et traduit `sql.ErrNoRows` → `orders.ErrNotFound` ; l'adaptateur *driving* `httpapi/` dépend du service **concret** (pas d'interface « au cas où ») et porte les **tags JSON dans son DTO** (`total` → `total_cents`). La démo teste le domaine **sans base** (le faux `memRepo`), puis la chaîne complète HTTP → use case → dépôt.  
**Lancer** : `go run ./cmd/server`  
**Sortie attendue** :

```text
=== Le domaine testé SANS base : le faux en mémoire ===
Place(Alice, 8900) → ord-1 · err = <nil>
Place(Bob, 0)      → règle métier : le montant doit être positif
ByID(absent)       → ErrNotFound : true

=== La chaîne complète : adaptateur HTTP → use case → dépôt ===
POST /orders (12900) → 200 {"id":"ord-2","customer":"Chloé","total_cents":12900}
POST /orders (0)     → 422 (la règle métier remonte en 422 ✔)
```

## 03-config-flags — section 10.3 (`03-configuration-12factor.md`)

**Description** : la configuration 12-factor de bout en bout — `config/config.go` est le chargeur du cours (structure typée, **précédence défauts < env < flags** sur un `FlagSet` local, **validation fail fast**) ; puis godotenv (**`Load` n'écrase jamais** une variable de la plateforme, **`Overload` force** — le `.env.exemple` n'existe que pour la démo) ; les **feature flags derrière un port** (`staticFlags` en dev, un fournisseur satisferait la même interface en prod — `checkout` bascule V1/V2) ; et le **rechargement à chaud** du niveau de log sur `SIGHUP`.  
**Lancer** : `go run .`  
**Sortie attendue** :

```text
addr=:9000 (env > défaut) · log=debug (flag > tout) · timeout=5s (défaut) · err=<nil>
sans DATABASE_URL → fail fast : configuration invalide : DATABASE_URL est requis
après Load     : défini-par-la-plateforme (la plateforme a gagné)
après Overload : défini-par-le-fichier-.env (le .env a forcé)
flag à false → checkout V1 (parcours actuel)
flag à true  → checkout V2 (nouveau parcours)
niveau de log : info
niveau de log : debug (changé à chaud, sans redémarrage)
```

---

## Nettoyage des binaires

`go run` ne laisse aucun binaire. Après un `go build` manuel : `go clean`. Aucun container ni volume à gérer dans ce chapitre — pas de service externe.

---

*Tous les exemples testés le 2026-07-05 (toolchain go1.26.0, Linux amd64) ; depguard vérifié avec golangci-lint v2.12. Sorties conformes au chapitre.*
