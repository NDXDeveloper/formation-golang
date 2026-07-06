🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 7.4 Migrations de schéma (golang-migrate, goose)

Un schéma évolue : nouvelles tables, colonnes, index, contraintes. Ces changements doivent être **versionnés, ordonnés, reproductibles** d'un environnement à l'autre (dev, staging, production) et appliqués **incrémentalement** — pas rejoués à la main, pas régénérés à chaque démarrage par un `AutoMigrate` ([§ 7.3](03-sqlc-vs-orm.md)). C'est le rôle des migrations : du SQL (ou du Go) sous contrôle de version, relu comme du code. Deux outils dominent l'écosystème Go — **golang-migrate** et **goose** —, tous deux disponibles en **CLI et en bibliothèque**. Point de liaison utile : ces fichiers de migration constituent aussi le schéma que lit sqlc ([§ 7.3](03-sqlc-vs-orm.md)), et on peut les **embarquer dans le binaire unique** ([§ 6.4](../06-cli-outillage/04-distribution.md)).

## Le principe : le schéma sous contrôle de version

Trois idées communes à tous les outils. Chaque changement est une **migration** numérotée, avec un sens *avant* (appliquer) et souvent un sens *arrière* (annuler). L'outil tient dans la base une **table de suivi** enregistrant les migrations déjà appliquées, ce qui lui permet de n'exécuter que les manquantes. Et l'on ne **modifie jamais une migration déjà appliquée** : on en ajoute une nouvelle.

Deux conventions pratiques : préférer une **numérotation par horodatage** en équipe (elle évite les collisions de versions lorsque plusieurs personnes créent des migrations en parallèle), et rendre le SQL **idempotent** quand c'est possible (`IF NOT EXISTS`, `IF EXISTS`).

## golang-migrate

`github.com/golang-migrate/migrate/v4` est une bibliothèque et une CLI de migrations supportant de nombreux moteurs (PostgreSQL, MySQL, SQLite, MongoDB…) et sources (fichiers locaux, `embed`, S3…).

### Fichiers et nommage

Chaque migration se compose de **deux fichiers**, un pour chaque sens :

```
db/migrations/
  000001_create_authors.up.sql
  000001_create_authors.down.sql
```

```sql
-- 000001_create_authors.up.sql
CREATE TABLE IF NOT EXISTS authors (
    id   BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

-- 000001_create_authors.down.sql
DROP TABLE IF EXISTS authors;
```

La version peut être un horodatage Unix ou un numéro séquentiel ; l'horodatage est recommandé car il évite les conflits de fusion quand plusieurs développeurs créent des migrations simultanément.

### En ligne de commande

```console
# CLI compilée avec le support d'un moteur (sélectionné par build tags)
$ go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

$ migrate create -ext sql -dir db/migrations -seq create_authors  # -seq : séquentiel (sinon horodatage)
$ migrate -path db/migrations -database "$DATABASE_URL" up         # applique les migrations manquantes
$ migrate -path db/migrations -database "$DATABASE_URL" down 1     # annule la dernière
$ migrate -path db/migrations -database "$DATABASE_URL" version    # version courante
$ migrate -path db/migrations -database "$DATABASE_URL" force 2    # sort d'un état « dirty »
```

Particularité importante : une migration échouée est marquée « dirty » afin que l'on répare la base avant de continuer. On corrige alors manuellement, puis on fixe la version avec `force`.

### En bibliothèque, migrations embarquées

Combiné à `//go:embed`, golang-migrate permet un **artefact autonome** : les migrations voyagent dans le binaire ([§ 6.4](../06-cli-outillage/04-distribution.md)), idéal pour un déploiement conteneurisé.

```go
package db

import (
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func Migrate(databaseURL string) error {
	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("source des migrations : %w", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", src, databaseURL)
	if err != nil {
		return fmt.Errorf("initialisation : %w", err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("application : %w", err)
	}
	return nil
}
```

`migrate.ErrNoChange` (base déjà à jour) n'est pas une erreur : on la filtre avec `errors.Is` ([§ 2.9](../02-fondamentaux-langage/09-gestion-erreurs.md)).

## goose

`github.com/pressly/goose/v3` propose une CLI conviviale, supporte les migrations SQL **et** en Go, et affiche joliment le statut. Sa différence marquante : les **migrations en Go**, indispensables pour une migration de **données** qui nécessite du code.

### SQL et migrations en Go

Contrairement à golang-migrate, goose met les deux sens dans **un seul fichier**, délimités par des annotations :

```sql
-- 20260101120000_create_authors.sql
-- +goose Up
CREATE TABLE authors (
    id   BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

-- +goose Down
DROP TABLE authors;
```

Pour un corps contenant des points-virgules (fonction, trigger), on encadre l'instruction par `-- +goose StatementBegin` / `-- +goose StatementEnd`. Et pour une migration de données pilotée par du code :

```go
package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upBackfill, downBackfill)
}

func upBackfill(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `UPDATE authors SET name = 'inconnu' WHERE name = ''`)
	return err
}

func downBackfill(ctx context.Context, tx *sql.Tx) error { return nil }
```

### En ligne de commande et embarqué

```console
$ go install github.com/pressly/goose/v3/cmd/goose@latest
$ goose -dir db/migrations postgres "$DATABASE_URL" status  # état de chaque migration
$ goose -dir db/migrations postgres "$DATABASE_URL" up
$ goose -dir db/migrations postgres "$DATABASE_URL" down
$ goose create add_bio sql                                   # nouveau fichier horodaté
```

L'embarquement suit la même logique qu'avec golang-migrate :

```go
//go:embed migrations/*.sql
var migrationsFS embed.FS

goose.SetBaseFS(migrationsFS)
if err := goose.SetDialect("postgres"); err != nil {
	return err
}
if err := goose.Up(db, "migrations"); err != nil {
	return err
}
```

## golang-migrate ou goose ?

| Critère | golang-migrate | goose |
|---------|----------------|-------|
| Fichiers | deux (`.up.sql` / `.down.sql`) | un seul (annotations `+goose`) |
| Migrations en Go | non (SQL) | **oui** (migrations de données) |
| Moteurs / sources | très nombreux (fichier, `embed`, S3, GitHub…) | moteurs principaux |
| Gestion d'échec | marqueur « dirty » + `force` | — |
| Table de suivi | `schema_migrations` | `goose_db_version` |
| CLI + bibliothèque | oui | oui |
| Embarquable (`embed`) | oui (`iofs`) | oui (`SetBaseFS`) |

En résumé : **golang-migrate** pour un flux tout-SQL avec un large choix de moteurs et de sources et un modèle up/down strict ; **goose** lorsqu'on a besoin de **migrations en Go** (rétro-remplissages, transformations de données) ou qu'on préfère le fichier unique et le `status` lisible. L'essentiel est d'en **choisir un et de s'y tenir**. Deux alternatives à connaître : **Atlas** (`atlasgo.io`), plus déclaratif (schéma-comme-code, diff et lint, utilisé par Ent au [§ 7.3](03-sqlc-vs-orm.md)), et **tern**, des migrations natives pgx.

## Bonnes pratiques

- **Versionnées et relues** : les migrations vivent dans le dépôt et passent en revue comme du code ; ne jamais réécrire une migration appliquée, en ajouter une nouvelle.
- **Petites, ciblées, idempotentes** : une intention par migration, avec `IF [NOT] EXISTS` quand c'est pertinent ; tester aussi le sens *down*.
- **Compatibles avec un déploiement progressif** : pour un déploiement sans interruption (mises à jour continues Kubernetes, [§ 9.2](../09-conteneurs-cloud/02-kubernetes.md)), privilégier des changements rétro-compatibles selon le schéma *expand/contract* (ajouter une colonne nullable → remplir → déployer le code → supprimer/renommer plus tard).
- **Où les exécuter** : embarquer les migrations et les appliquer au démarrage donne un artefact autonome, pratique en conteneur. Mais avec **plusieurs répliques**, on évite que chaque instance ne se précipite : on exécute les migrations **une seule fois** via une étape dédiée — un *Job*/init container Kubernetes ([§ 9.2](../09-conteneurs-cloud/02-kubernetes.md)) ou une étape de CI/CD ([§ 15.2](../15-deploiement-devops/02-cicd.md)).
- **Une seule source de vérité** : ces fichiers servent aussi de schéma à sqlc ([§ 7.3](03-sqlc-vs-orm.md)).

## Côté IDE : GoLand et VS Code

Les fichiers de migration étant du **SQL ordinaire**, tout l'outillage base de données du [§ 7.1](01-database-sql.md) s'y applique : coloration, autocomplétion et inspections via la fenêtre *Database* de **GoLand** ou l'extension **SQLTools** de **VS Code**.

**GoLand** ajoute deux atouts pour ce sujet : on peut exécuter un fichier de migration contre une source de données de développement, et la **comparaison de schémas** (diff de DDL) permet de vérifier l'état obtenu après application. On lance `migrate` ou `goose` via un *External Tool* ou une configuration d'exécution, et l'on débogue le code Go d'application des migrations avec Delve.

**VS Code** exécute la CLI (`migrate`/`goose`) via une tâche `tasks.json`, prévisualise et lance le SQL avec SQLTools, et débogue le runner Go avec `dlv`.

## En résumé

- Migrations = changements de schéma **versionnés, ordonnés, reproductibles**, suivis par une table dans la base ; jamais d'`AutoMigrate` en production ([§ 7.3](03-sqlc-vs-orm.md)).
- **golang-migrate** : deux fichiers `.up.sql`/`.down.sql`, table `schema_migrations`, état « dirty » + `force`, très nombreux moteurs/sources, embarquable via `iofs` + `//go:embed`.
- **goose** : fichier unique avec annotations `-- +goose Up/Down`, **migrations en Go** pour les données, `status` lisible, embarquable via `SetBaseFS`.
- Bonnes pratiques : horodatage en équipe, SQL idempotent, tester le *down*, changements rétro-compatibles (expand/contract), et **exécution unique** au déploiement (Job/init container [§ 9.2](../09-conteneurs-cloud/02-kubernetes.md), ou CI [§ 15.2](../15-deploiement-devops/02-cicd.md)) plutôt qu'une course entre répliques.
- Ces fichiers sont aussi le schéma de sqlc ([§ 7.3](03-sqlc-vs-orm.md)), et peuvent voyager dans le binaire unique ([§ 6.4](../06-cli-outillage/04-distribution.md)).

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [7.5 — NoSQL et cache : MongoDB, Redis](05-nosql-redis.md)

⏭ [NoSQL et cache : MongoDB, Redis](/07-acces-donnees/05-nosql-redis.md)
