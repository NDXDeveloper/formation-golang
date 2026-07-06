🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 7.3 sqlc (SQL typé généré) vs ORM (GORM, Ent) — grille de choix

Au-dessus du `database/sql` brut ([§ 7.1](01-database-sql.md)) se pose la question centrale du module : **quel niveau d'abstraction ?** Trois familles d'outils y répondent différemment — **sqlc** génère du Go typé à partir de votre SQL, **GORM** est un ORM par réflexion, **Ent** définit le schéma comme du code Go et génère un client typé. Aucune n'est « la bonne » dans l'absolu ; cette section les situe et propose une grille de décision. Conformément à l'esprit de la formation (du SQL explicite, pas d'ORM par réflexe), sqlc est présenté comme le **choix par défaut idiomatique**, les ORM l'étant honnêtement, avec leurs forces et leurs coûts.

## Le spectre, des helpers aux frameworks

Il ne s'agit pas d'un duel à deux, mais d'un curseur :

- **`database/sql` brut** ([§ 7.1](01-database-sql.md)) — contrôle total, balayage manuel ;
- **helpers légers** — `sqlx` (scan de structs, requêtes nommées) ou les `pgx.CollectRows`/`RowToStructByName` du [§ 7.2](02-drivers.md), qui suppriment le boilerplate sans rien cacher du SQL ;
- **sqlc** — on écrit le SQL, l'outil génère les types ;
- **ORM** — soit par **réflexion** (GORM), soit par **schéma-comme-code** (Ent).

Deux axes résument les compromis : *SQL d'abord ou Go d'abord*, et *types générés à la compilation ou réflexion à l'exécution*.

## sqlc : le SQL comme source de vérité

sqlc (`github.com/sqlc-dev/sqlc`) lit votre **schéma** et vos **requêtes** en SQL, puis génère des fonctions et des structures Go typées. Pas de réflexion : les requêtes sont analysées et vérifiées contre le schéma **au moment de la génération**, et vous gardez la maîtrise de chaque requête.

### `sqlc.yaml` et `go generate`

La configuration tient dans un fichier `sqlc.yaml` en version « 2 », où le champ `sql_package` vaut `pgx/v5` (ou `database/sql`) :

```yaml
version: "2"
sql:
  - engine: "postgresql"
    schema: "db/schema.sql"   # ou le dossier de migrations
    queries: "db/queries.sql"
    gen:
      go:
        package: "db"
        out: "db/gen"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_interface: true  # génère une interface Querier — utile pour les mocks (§ 13.2)
```

La génération se déclenche par `sqlc generate`, que l'on **rattache au code** grâce à une directive `go:generate` — c'est l'usage canonique de ce mécanisme dans la formation :

```go
//go:generate sqlc generate
package db
```

`go generate ./...` parcourt les fichiers à la recherche de ces directives (de simples commentaires) et exécute les commandes associées. Point important : `go generate` **n'est pas** lancé par `go build` ; on l'exécute explicitement, généralement en local puis en CI, et l'on versionne le code généré. sqlc va plus loin que la simple génération : `sqlc vet` applique des règles de contrôle (par exemple `sqlc/db-prepare`), et `sqlc verify` analyse les requêtes existantes face à un changement de schéma et échoue si celui-ci les casserait — un filet de sécurité précieux.

### Ce que ça produit

À partir d'une requête annotée par un commentaire `-- name: … :one` (ou `:many`, `:exec`, `:execrows`, `:copyfrom`)…

```sql
-- name: GetAuthor :one
SELECT id, name, bio FROM authors WHERE id = $1;

-- name: CreateAuthor :one
INSERT INTO authors (name, bio) VALUES ($1, $2)
RETURNING id, name, bio;
```

…sqlc génère une structure `Queries`, ses méthodes typées et les structures de paramètres/résultats. On l'utilise avec le pool pgx du [§ 7.2](02-drivers.md) :

```go
q := db.New(pool)                  // db.New accepte *pgx.Conn ou *pgxpool.Pool
author, err := q.GetAuthor(ctx, 1) // → db.Author, error
created, err := q.CreateAuthor(ctx, db.CreateAuthorParams{
	Name: "Kernighan",
	Bio:  pgtype.Text{String: "…", Valid: true},
})
```

### Forces et limites

**Pour** : transparence totale (vous voyez et possédez le SQL), sécurité de type **à la compilation**, coût d'exécution minimal, compétences SQL directement réutilisables, et vérification des requêtes en amont. **Contre** : les **requêtes dynamiques** (clauses `WHERE` conditionnelles, tris variables) sont malaisées — sqlc vise des requêtes essentiellement **statiques** — et toute évolution de schéma ou de requête impose une régénération.

## GORM : l'ORM par réflexion

GORM (`gorm.io/gorm`) est un ORM complet et convivial pour Go, de très loin le plus répandu. Les modèles sont des structures annotées, les relations se déclarent par des champs, et l'on manipule la base par une API fluide plutôt qu'en SQL.

```go
import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Author struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"not null"`
	Bio   string
	Books []Book `gorm:"foreignKey:AuthorID"` // relation has-many
}

db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})

var a Author
db.WithContext(ctx).Preload("Books").First(&a, id) // charge l'auteur ET ses livres
```

GORM apporte associations, `Preload` (chargement anticipé), hooks, *soft deletes*, `AutoMigrate`, et une gestion d'erreurs à la [§ 7.1](01-database-sql.md) (`errors.Is(err, gorm.ErrRecordNotFound)`). C'est un accélérateur de développement. **Le revers** : il repose sur la **réflexion** (coût d'exécution), le SQL réellement émis est masqué (risque de requêtes inefficaces, problème classique du **N+1**), et la « magie » complique le débogage. `AutoMigrate` est commode mais à réserver au développement — en production, on préfère des migrations explicites ([§ 7.4](04-migrations.md)). À noter : **GORM Gen** (`gorm.io/gen`) ajoute par-dessus une couche de génération de code fournissant une API 100 % typée, sans `interface{}`, qui atténue le typage faible des requêtes GORM.

## Ent : le schéma comme code

Ent (`entgo.io/ent`) prend le parti inverse de sqlc : le schéma n'est pas du SQL mais **du code Go**, à partir duquel Ent génère un client fortement typé et une API de traversée de graphe.

```go
// ent/schema/author.go — le schéma est du Go
func (Author) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.String("bio").Optional(),
	}
}
func (Author) Edges() []ent.Edge {
	return []ent.Edge{edge.To("books", Book.Type)} // relation
}
```

Après `go generate ./ent`, on interroge la base via des *builders* typés :

```go
a, err := client.Author.Query().
	Where(author.ID(1)).
	WithBooks(). // charge la relation
	Only(ctx)
```

Ent brille sur les **domaines riches** au graphe de relations typées : forte sécurité de type, API expressive, migrations intégrées (via Atlas). En contrepartie : outil **opiniâtre**, schéma dans un DSL Go propre à Ent, courbe d'apprentissage plus raide et volume important de code généré.

## Grille de choix

| Critère | sqlc | GORM | Ent |
|---------|------|------|-----|
| Approche | SQL d'abord | Go d'abord (structs + tags) | Schéma en Go |
| Génération de code | oui | non — réflexion (GORM Gen en option) | oui |
| Sécurité de type | compilation | exécution (partielle) | compilation |
| Requêtes dynamiques | difficile | facile | facile (builder) |
| Transparence du SQL | totale | faible | faible |
| Coût à l'exécution | minimal | réflexion | minimal |
| Migrations | non incluses ([§ 7.4](04-migrations.md)) | `AutoMigrate` (+ [§ 7.4](04-migrations.md)) | incluses (Atlas) |
| Apprentissage | SQL requis | doux | plus raide |

En pratique :

- **sqlc** — le défaut de cette formation pour la plupart des backends : SQL explicite, types générés, aucune magie, performances de pgx. Idéal si l'équipe est à l'aise avec SQL et que les requêtes sont majoritairement statiques.
- **GORM** — prototypage rapide, applications CRUD, équipes préférant ne pas écrire de SQL, associations simples et nombreuses ; on accepte le coût de la réflexion et la moindre transparence (GORM Gen aide sur le typage).
- **Ent** — domaine complexe et fortement relationnel, envie d'un schéma-comme-code typé et d'une API de traversée générée ; projets d'envergure.
- **`database/sql` brut ou `sqlx`** — pour un petit outil ou quelques requêtes, aucun des trois n'est nécessaire : pas de sur-ingénierie.

Tendance de fond à garder en tête : la **génération de code** gagne du terrain (Ent, GORM Gen, ou encore SQLBoiler qui génère depuis un schéma existant). La vraie question moderne est souvent *quelle approche générée* — SQL d'abord (sqlc) ou schéma-comme-code (Ent) — l'ORM par réflexion pur restant surtout pertinent pour le développement rapide. Enfin, **aucun de ces outils ne dispense d'une stratégie de migrations** : c'est l'objet du [§ 7.4](04-migrations.md).

## Côté IDE : GoLand et VS Code

Le choix a des conséquences directes sur l'outillage.

Avec **sqlc**, vos requêtes vivent dans des fichiers `.sql` ordinaires : tout l'outillage base de données du [§ 7.1](01-database-sql.md) s'y applique pleinement — coloration, autocomplétion et inspections via la fenêtre *Database* de **GoLand** ou l'extension **SQLTools** de **VS Code**. Le Go généré porte un bandeau `// Code generated … DO NOT EDIT.` : on ne le modifie pas à la main, on régénère. Pour lancer la génération sans quitter l'IDE, GoLand propose *Go Generate* (menu contextuel ou configuration d'exécution) et VS Code une tâche exécutant `go generate ./...`.

Avec un **ORM**, requêtes et modèles sont du Go : l'outillage SQL aide moins, et l'on observe plutôt le SQL réellement émis — le *logger* de GORM ou le mode `Debug()` d'Ent — pour repérer les requêtes inefficaces ou un N+1. Le débogage passe par Delve, comme partout.

## En résumé

- Trois approches au-dessus de `database/sql` : **sqlc** (SQL d'abord, types générés), **GORM** (ORM par réflexion), **Ent** (schéma-comme-code généré).
- **sqlc** : `sqlc.yaml` (v2, `sql_package: pgx/v5`), requêtes annotées `-- name: … :one/:many/:exec`, génération pilotée par `//go:generate sqlc generate` ; `vet`/`verify` attrapent les régressions. Transparence et sécurité de type maximales ; requêtes dynamiques malaisées.
- **GORM** : rapide et riche (associations, `Preload`, `AutoMigrate`) mais réflexion, SQL masqué et risque de N+1 ; GORM Gen ajoute du typage généré.
- **Ent** : schéma en Go, client typé et traversée de graphe générés ; puissant sur les domaines complexes, plus exigeant.
- `go generate` exécute les directives `//go:generate` (hors `go build`), à lancer explicitement et en CI ; le code généré se versionne.
- Défaut idiomatique : **sqlc** ; ORM selon l'équipe et le domaine ; rien de tout cela pour un petit outil. Et dans tous les cas, prévoir des **migrations** ([§ 7.4](04-migrations.md)).

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [7.4 — Migrations de schéma (golang-migrate, goose)](04-migrations.md)

⏭ [Migrations de schéma (golang-migrate, goose)](/07-acces-donnees/04-migrations.md)
