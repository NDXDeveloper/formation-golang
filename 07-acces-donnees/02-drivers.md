🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 7.2 Drivers : PostgreSQL (pgx) ⭐, MySQL, SQLite

Le [§ 7.1](01-database-sql.md) a présenté l'interface `database/sql`, mais un `*sql.DB` ne sert à rien sans un **driver** qui parle réellement à un moteur. Cette section branche les trois cibles que rencontrent la plupart des équipes Go — PostgreSQL, MySQL, SQLite — et montre deux points qui vont au-delà de la simple configuration : PostgreSQL, via **pgx**, offre une voie *native* plus riche que `database/sql` ; et SQLite impose un choix **cgo ou pur Go** dont les conséquences remontent jusqu'à la distribution ([module 6](../06-cli-outillage/README.md)).

## Comment un driver s'enfiche dans `database/sql`

Un driver s'enregistre lui-même, dans sa fonction `init()`, via `sql.Register("nom", …)`. Il suffit donc de l'importer pour ses effets de bord (import « blanc ») ; le premier argument de `sql.Open` est ce **nom enregistré**, distinct du nom du paquet. Deux détails changent d'un moteur à l'autre : la **syntaxe des placeholders** (`$1` en PostgreSQL, `?` ailleurs) et le **DSN** (chaîne de connexion).

## PostgreSQL : pgx ⭐

**pgx** (`github.com/jackc/pgx`, v5 étant la version majeure stable) est le driver et la boîte à outils PostgreSQL de référence. Il est **écrit en pur Go** et s'utilise de deux façons.

### Via `database/sql` (adaptateur `stdlib`)

L'adaptateur permet de rester exactement sur l'API du [§ 7.1](01-database-sql.md) :

```go
import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib" // s'enregistre sous le nom "pgx"
)

db, err := sql.Open("pgx", os.Getenv("DATABASE_URL")) // puis QueryContext, etc.
```

C'est le choix de la **compatibilité maximale** : tout l'outillage bâti sur `database/sql` (migrations, [§ 7.4](04-migrations.md)) fonctionne sans adaptation.

### Via l'API native (`pgx` / `pgxpool`)

L'interface native est plus rapide et donne accès à des fonctionnalités propres à PostgreSQL — LISTEN / NOTIFY, COPY — qui ne sont pas exposées par `database/sql`. Un `*pgx.Conn` seul n'est pas sûr en concurrence : on utilise le pool `pgxpool`.

```go
import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL")) // pool sûr en concurrence
if err != nil {
	return err
}
defer pool.Close()

var name string
err = pool.QueryRow(ctx, `SELECT name FROM users WHERE id = $1`, id).Scan(&name)
```

Surtout, pgx fournit des **fonctions génériques de collecte** qui remplacent avantageusement le trio `Next` / `Scan` / `Err` du [§ 7.1](01-database-sql.md) : `CollectRows` et `ForEachRow` sont une manière plus simple et plus sûre de traiter les lignes que d'appeler manuellement `Rows.Close`, `Rows.Next`, `Rows.Scan` et `Rows.Err`.

```go
type User struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

rows, err := pool.Query(ctx, `SELECT id, name FROM users WHERE active = $1`, true)
if err != nil {
	return nil, err
}
users, err := pgx.CollectRows(rows, pgx.RowToStructByName[User]) // → []User, mappage par nom
```

### Laquelle choisir ?

Par défaut, pour un service dédié à PostgreSQL, l'**API native avec `pgxpool`** est le meilleur compromis : performances, fonctionnalités PostgreSQL, et ergonomie des helpers de lignes. On revient à l'**adaptateur `database/sql`** lorsqu'on veut rester sur l'interface standard pour un outillage générique. Ce n'est pas strictement l'un ou l'autre — le choix se recroise d'ailleurs avec sqlc et les ORM au [§ 7.3](03-sqlc-vs-orm.md). À noter : pgx utilise les paramètres positionnels `$1, $2` et ne gère pas les paramètres nommés ; côté compatibilité, il suit la politique de ses amonts — les **deux dernières versions majeures de Go**, et les versions de **PostgreSQL des cinq dernières années**.

## MySQL : go-sql-driver/mysql

Le driver de référence est `github.com/go-sql-driver/mysql`, lui aussi en **pur Go**. Il s'enregistre sous `"mysql"`, utilise les placeholders `?`, et son DSN réserve un piège classique : sans **`parseTime=true`**, les colonnes `DATE`/`DATETIME`/`TIMESTAMP` reviennent en octets bruts et refusent de se scanner dans un `time.Time`.

```go
import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql" // s'enregistre sous "mysql"
)

// parseTime=true est presque toujours nécessaire ; loc fixe le fuseau d'interprétation.
dsn := "user:pass@tcp(localhost:3306)/mydb?parseTime=true&loc=UTC"
db, err := sql.Open("mysql", dsn)
```

Contrairement à PostgreSQL, MySQL fournit `LastInsertId` (colonnes `AUTO_INCREMENT`), utilisable directement via le `sql.Result` du [§ 7.1](01-database-sql.md).

## SQLite : cgo ou pur Go, un choix qui compte

SQLite est une base **embarquée** (un simple fichier, sans serveur), idéale pour un outil CLI, un service léger ou des tests. En Go, deux drivers principaux s'affrontent, et le choix a des conséquences concrètes.

- **`github.com/mattn/go-sqlite3`** — enveloppe la bibliothèque C officielle. Parce que c'est un paquet cgo, il exige `CGO_ENABLED=1` et la présence d'un compilateur gcc. Fonctionnalités complètes, performances maximales, mais il **casse les binaires statiques** et complique la cross-compilation ([§ 6.4](../06-cli-outillage/04-distribution.md)). Il s'enregistre sous **`"sqlite3"`**. Piège sournois : compilé malgré tout avec `CGO_ENABLED=0`, **le build réussit** (un *stub* sans cgo prend la place) — et le binaire **panique au premier usage** : `go-sqlite3 requires cgo to work. This is a stub`. L'erreur ne se voit qu'à l'exécution, pas à la compilation.
- **`modernc.org/sqlite`** — un portage sans cgo de la bibliothèque C de SQLite, donc **pur Go** : compatible `CGO_ENABLED=0`, il se cross-compile aussi trivialement que n'importe quel code Go. Légèrement moins rapide dans certains cas, mais sans dépendance C. Il s'enregistre sous **`"sqlite"`**.

Le code fonctionne avec l'un ou l'autre en changeant simplement le chemin d'import et le nom de driver passé à `sql.Open` (mattn s'enregistre sous « sqlite3 », modernc sous « sqlite ») — attention à cette différence de nom.

```go
// Pur Go (aucun cgo, cross-compilation triviale) : souvent le bon défaut pour un binaire distribué.
import _ "modernc.org/sqlite"

db, err := sql.Open("sqlite", "./app.db") // SQLite crée le fichier s'il n'existe pas

// cgo (fonctionnalités et performances maximales, mais CGO_ENABLED=1 + compilateur C requis) :
// import _ "github.com/mattn/go-sqlite3"
// db, err := sql.Open("sqlite3", "./app.db")
```

Pour un outil que l'on distribue en **binaire statique unique** ([§ 6.4](../06-cli-outillage/04-distribution.md)), `modernc.org/sqlite` est souvent le choix pragmatique ; on réserve mattn aux cas exigeant la parité complète avec le C. Une autre option pure Go existe (`github.com/ncruces/go-sqlite3`, fondée sur WASM). Dernier point de vigilance : SQLite n'accepte **qu'un seul écrivain à la fois** ; en contexte concurrent, activez le mode WAL et/ou limitez le pool d'écriture (`db.SetMaxOpenConns(1)`), la syntaxe exacte du DSN variant selon le driver.

## Récapitulatif : noms de driver et placeholders

| Moteur | Nom (`sql.Open`) | Import | Placeholder | Pur Go ? |
|--------|------------------|--------|-------------|----------|
| PostgreSQL | `pgx` | `github.com/jackc/pgx/v5/stdlib` (ou l'API native `.../v5`) | `$1` | oui |
| MySQL | `mysql` | `github.com/go-sql-driver/mysql` | `?` | oui |
| SQLite (cgo) | `sqlite3` | `github.com/mattn/go-sqlite3` | `?` | non (`CGO_ENABLED=1`) |
| SQLite (pur Go) | `sqlite` | `modernc.org/sqlite` | `?` | oui |

## Côté IDE : GoLand et VS Code

L'enjeu est de connecter l'IDE au **même moteur** que celui visé par le code, pour que l'exploration du schéma et l'injection SQL ([§ 7.1](01-database-sql.md)) valident le bon dialecte.

**GoLand** gère nativement PostgreSQL, MySQL et SQLite dans la fenêtre *Database* : ajoutez une *Data Source* avec le même DSN que l'application (pour SQLite, pointez sur le fichier `.db`). Que vous utilisiez l'API native pgx ou l'adaptateur `database/sql`, GoLand dialogue avec le serveur indépendamment de votre code.

**VS Code** s'appuie sur **SQLTools** et ses extensions de pilote (PostgreSQL, MySQL, SQLite), ou sur des extensions propres à un moteur ; on configure une connexion par base.

Point concret lié à la distribution : un driver **pur Go** (pgx, mysql, `modernc.org/sqlite`) se cross-compile depuis l'IDE sans réglage ([§ 6.4](../06-cli-outillage/04-distribution.md)) ; avec `mattn/go-sqlite3`, l'environnement de build de l'IDE doit fournir `CGO_ENABLED=1` **et** un compilateur C, sous peine d'échec de compilation.

## En résumé

- Un driver s'enregistre par import « blanc » ; `sql.Open` prend son **nom enregistré** ; placeholders et DSN dépendent du moteur.
- **PostgreSQL / pgx** (pur Go, v5) : soit l'adaptateur `database/sql` (`stdlib`, compatibilité maximale), soit l'API native `pgx` + `pgxpool` (plus rapide, LISTEN/NOTIFY, COPY, et `CollectRows`/`RowToStructByName` qui suppriment le boilerplate de scan).
- **MySQL** / `go-sql-driver/mysql` (pur Go) : placeholders `?`, DSN avec `parseTime=true` presque obligatoire ; `LastInsertId` disponible.
- **SQLite** : arbitrage **cgo vs pur Go** — `mattn/go-sqlite3` (`"sqlite3"`, complet mais `CGO_ENABLED=1` + gcc ; sans cgo, le build passe mais le binaire **panique à l'exécution** — un stub) contre `modernc.org/sqlite` (`"sqlite"`, pur Go, cross-compilation triviale). Même code, on change import et nom de driver ; le choix se répercute sur la distribution ([§ 6.4](../06-cli-outillage/04-distribution.md)).

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [7.3 — sqlc (SQL typé généré) vs ORM (GORM, Ent) — grille de choix](03-sqlc-vs-orm.md)

⏭ [sqlc (SQL typé généré) vs ORM (GORM, Ent) — grille de choix](/07-acces-donnees/03-sqlc-vs-orm.md)
