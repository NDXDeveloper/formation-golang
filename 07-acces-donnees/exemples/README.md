# Exemples du chapitre 07 — Accès aux données

Un exemple **complet par section**. Deux sont **totalement autonomes** (`02` SQLite pur Go, `06` E/S stdlib — aucun service) ; les autres s'exécutent contre un **moteur en container Docker** (cycles complets ci-dessous : lancer, arrêter, supprimer, image, volumes). Chaque fichier porte un en-tête **Section / Description / Fichier source** et des commentaires d'intention. Tous ont été **compilés, vérifiés (`go vet`) et exécutés** (toolchain **go1.26.0**, Linux amd64) — sorties telles que constatées. GORM et Ent (les deux autres familles du § 7.3) restent illustrés dans le `.md` : validés lors des tests du chapitre, ils alourdiraient le dépôt sans bénéfice pédagogique supplémentaire.

## Prérequis communs

**Installation** : **Go 1.25 minimum** (la formation cible Go 1.26) ; **Docker** pour `01`, `03`, `04`, `05` ; **sqlc** uniquement pour *régénérer* `03` (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest` — le code généré est commité, le build n'en a pas besoin).  
**Configuration** : aucune (`GOTOOLCHAIN=auto`). **Réseau** au premier build (modules pgx, modernc/sqlite, migrate, mongo-driver v2, go-redis — `go.sum` fournis).

## Les containers (à lancer selon l'exemple)

```console
# PostgreSQL (exemples 01, 03, 04) — port hôte 5433
$ docker run -d --name pg07 -e POSTGRES_PASSWORD=secret -e POSTGRES_DB=formation \
    -p 127.0.0.1:5433:5432 postgres:16-alpine
$ export DATABASE_URL="postgres://postgres:secret@127.0.0.1:5433/formation?sslmode=disable"

# MongoDB + Valkey (exemple 05) — ports hôte 27018 et 6380
$ docker run -d --name mg07 -p 127.0.0.1:27018:27017 mongo:7
$ docker run -d --name vk07 -p 127.0.0.1:6380:6379 valkey/valkey:8-alpine
```

**Arrêt / suppression / nettoyage complet** (containers, images téléchargées, volumes) :

```console
$ docker stop pg07 mg07 vk07                 # arrêter
$ docker rm   pg07 mg07 vk07                 # supprimer les containers
$ docker rmi postgres:16-alpine mongo:7 valkey/valkey:8-alpine   # supprimer les images
$ docker volume prune -f                     # purger les volumes anonymes créés par les moteurs
$ docker system df                           # vérifier : 0 B partout
```

## Vue d'ensemble

| Dossier | Section | Fichier source | Service requis | Ce que ça démontre |
|---|---|---|---|---|
| `01-database-sql/` | 7.1 | `01-database-sql.md` | PostgreSQL | pool, 3 verbes, `ErrNoRows`, `RETURNING`, NULL, Tx + rollback réel |
| `02-drivers-sqlite/` | 7.2 | `02-drivers.md` | — | driver pur Go `"sqlite"`, `?`, `LastInsertId`, WAL ; piège mattn commenté |
| `03-sqlc/` | 7.3 | `03-sqlc-vs-orm.md` | PostgreSQL | SQL source de vérité → code **généré** (`Querier`), utilisé via `pgxpool` |
| `04-migrations/` | 7.4 | `04-migrations.md` | PostgreSQL | migrations **embarquées** (`embed`+`iofs`), `Up` idempotent (`ErrNoChange`) |
| `05-nosql-redis/` | 7.5 | `05-nosql-redis.md` | MongoDB + Valkey | driver mongo **v2**, `ErrNoDocuments`, `cur.All` ; **cache-aside** sur Valkey |
| `06-fichiers-io/` | 7.6 | `06-fichiers-io.md` | — | `io`/`bufio`/`os`/`io/fs`/`embed`, négatifs (`os.Root`, 64 Kio, `Flush`) |

---

## 01-database-sql — section 7.1 (`01-database-sql.md`) — PostgreSQL requis

**Description** : la section entière — pool borné + `PingContext`, les trois verbes `...Context`, `Next`/`Scan`/`Err`/`Close`, `sql.ErrNoRows` → sentinelle, `INSERT … RETURNING`, la **preuve** que `LastInsertId` n'existe pas en PostgreSQL, `NullString` + `sql.Null[T]` (1.22), requête préparée, transaction avec `defer tx.Rollback()` — **rollback réel** sur contrainte `CHECK` violée.  
**Lancer** : `DATABASE_URL=… go run .` — **Sortie attendue** (extraits) :

```text
LastInsertId → LastInsertId is not supported by this driver
actifs : 3 · premier : Ada
id inconnu → ErrNotFound : true
soldes : [70 80] (attendu [70 80])
Transfer(1→99, 1000) échoue : true
soldes après rollback : [70 80] (intacts)
```

## 02-drivers-sqlite — section 7.2 (`02-drivers.md`) — autonome

**Description** : le mécanisme des drivers via le plus autonome — `modernc.org/sqlite` (pur Go, nom **`"sqlite"`**, zéro serveur, `CGO_ENABLED=0` compatible) : import « blanc », placeholders `?`, `LastInsertId` (supporté ici — le contraste avec PostgreSQL), WAL + `SetMaxOpenConns(1)`. La variante cgo (mattn, `"sqlite3"`) et **son piège du stub** (build OK sans cgo, panique au premier usage) sont documentés en commentaire ; le DSN MySQL et son piège `parseTime=true` restent au `.md` (validés lors des tests du chapitre).  
**Sortie attendue** :

```text
fichier app.db créé · LastInsertId = 1 · err = <nil>
relu : bonjour
driver « sqlite » (modernc, pur Go) : compatible CGO_ENABLED=0 ✔
```

## 03-sqlc — section 7.3 (`03-sqlc-vs-orm.md`) — PostgreSQL requis

**Description** : le flux sqlc complet — `sqlc.yaml` (v2, `sql_package: pgx/v5`, `emit_interface`), `db/schema.sql` + `db/queries.sql` annotées (`-- name: … :one`) = **source de vérité** ; `db/gen/` = **généré et commité** (`db.go`, `models.go`, `querier.go`, `queries.sql.go`). Le `main` utilise l'API générée (`db.New(pool)`, `CreateAuthor`/`GetAuthor`, `pgtype.Text`) : zéro réflexion, types à la compilation.  
**Lancer** : `DATABASE_URL=… go run .` · **Régénérer** : `go generate ./...` — **Sortie attendue** :

```text
créé : id=1 name=Kernighan
relu : Kernighan · bio valide : true
sqlc : le SQL est la source de vérité, le Go est généré et typé ✔
```

## 04-migrations — section 7.4 (`04-migrations.md`) — PostgreSQL requis

**Description** : le bloc « bibliothèque + embarquées » — `//go:embed migrations/*.sql`, source `iofs`, `m.Up()` avec `migrate.ErrNoChange` filtré : l'appel est **idempotent** (relançable au démarrage). Les deux fichiers `.up.sql`/`.down.sql` de la section sont dans `migrations/`. Les **CLI** `migrate` (dont l'état *dirty* + `force`) et `goose` (migrations en Go) ont été validées lors des tests du chapitre et restent documentées au `.md`.  
**Lancer** : `DATABASE_URL=… go run .` — **Sortie attendue** :

```text
1er passage  : OK (migrations appliquées)
2e  passage  : OK (ErrNoChange filtré : déjà à jour)
```

## 05-nosql-redis — section 7.5 (`05-nosql-redis.md`) — MongoDB + Valkey requis

**Description** : les deux volets — **MongoDB driver v2** (`Connect` **sans** contexte, le contexte va sur `Ping` ; `bson.ObjectID`, `FindOne().Decode` + `ErrNoDocuments`, curseur `cur.All`, `UpdateOne $set`) puis **go-redis v9 pointé sur Valkey** (le code est identique — la compatibilité affirmée par la section, démontrée) : `Set` avec TTL, `redis.Nil` = cache miss, et le **cache-aside** complet (miss → `load` → hit, invalidation `Del`).  
**Sortie attendue** (extraits) :

```text
absent → ErrNoDocuments : true
prix < 10 : 2 produits
clé absente → redis.Nil : true
cache-aside : Chargé puis Chargé · load appelé 1 fois (le cache a servi le 2e)
après invalidation (Del) : rechargé, load = 2
```

## 06-fichiers-io — section 7.6 (`06-fichiers-io.md`) — autonome

**Description** : la section entière, **négatifs compris** — `io.Copy` et les assistants, le pattern « erreur de `Close` en écriture », `os.ReadDir`/`Stat`+`fs.ErrNotExist`, **`os.Root` (1.24) qui refuse `../`** (« path escapes from parent »), `bufio.Scanner` (64 Kio → `ErrTooLong` → parade `sc.Buffer`), **le `Flush` oublié = 0 octet écrit**, le même `WalkDir` sur `os.DirFS` et sur `embed.FS`, et `//go:embed` (string, FS, préfixe `all:` pour `templates/.cache`).  
**Sortie attendue** :

```text
   message Root : openat ../data.txt: path escapes from parent
✔ 7.6 : io/bufio/os/fs/embed — tous les comportements (et négatifs) conformes
```

---

## Nettoyage des binaires

`go run .` ne laisse aucun binaire (les exemples suppriment eux-mêmes leurs fichiers de travail : `app.db`, répertoires temporaires). Après un `go build` manuel : `go clean`. Les `go.sum` font partie des exemples ; `03-sqlc/db/gen/` est **généré mais commité** (le principe sqlc : on versionne le code généré).

---

*Tous les exemples testés le 2026-07-05 (toolchain go1.26.0, Linux amd64) contre postgres:16-alpine, mongo:7 et valkey/valkey:8-alpine : sorties conformes au chapitre.*
