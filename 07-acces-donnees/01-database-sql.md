🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 7.1 `database/sql` (pool de connexions, requêtes préparées, transactions)

`database/sql` est l'interface de la bibliothèque standard pour les bases relationnelles : **indépendante du moteur**, elle gère le pool de connexions, les requêtes paramétrées, les requêtes préparées et les transactions, pendant qu'un *driver* ([§ 7.2](02-drivers.md)) fournit l'implémentation concrète. Vous écrivez le SQL, le paquet s'occupe de la plomberie. C'est la fondation sur laquelle reposent aussi bien sqlc que les ORM ([§ 7.3](03-sqlc-vs-orm.md)) — la maîtriser, c'est comprendre tout le reste.

Deux conventions valables dans toute la section : on utilise systématiquement les variantes **`...Context`** des méthodes (annulation et délais, [module 4](../04-concurrence/README.md)), et les **placeholders** dépendent du driver — `$1, $2…` en PostgreSQL (le défaut de la formation, via pgx), `?` en MySQL/SQLite (détaillé en [§ 7.2](02-drivers.md)).

## `sql.DB` est un pool, pas une connexion

`sql.Open` ne se connecte pas : il valide les arguments et prépare un **pool** de connexions, ouvertes paresseusement à la première requête. Le driver est importé pour ses seuls effets de bord (import « blanc »), ce qui l'enregistre auprès de `database/sql`.

```go
package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // driver PostgreSQL (pgx) — voir § 7.2
)

func Open(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn) // ne se connecte pas encore
	if err != nil {
		return nil, fmt.Errorf("ouverture : %w", err)
	}

	db.SetMaxOpenConns(25)                 // borne le nombre total de connexions
	db.SetMaxIdleConns(25)                 // connexions inactives conservées (idéalement = max)
	db.SetConnMaxLifetime(5 * time.Minute) // recyclage périodique (load balancers, timeouts serveur)
	db.SetConnMaxIdleTime(1 * time.Minute) // ferme les connexions inactives trop longtemps

	if err := db.PingContext(ctx); err != nil { // vérifie *réellement* la connexion
		db.Close()
		return nil, fmt.Errorf("connexion : %w", err)
	}
	return db, nil
}
```

Trois points idiomatiques. `*sql.DB` est **sûr en concurrence** et prévu pour vivre **toute la durée de l'application** : on l'ouvre une fois et on le partage — l'ouvrir à chaque requête est un anti-patron classique. `Open` n'établissant aucune connexion, `PingContext` est le moyen de valider le paramétrage au démarrage. Enfin, le pool doit être **borné** (`SetMaxOpenConns`) pour ne pas saturer le serveur ; garder autant de connexions inactives que de connexions maximales évite un cycle ouverture/fermeture coûteux. `db.Close()` ne sert qu'à l'arrêt du programme.

## Exécuter des requêtes : trois verbes

- **`ExecContext`** pour ce qui ne renvoie pas de lignes (INSERT, UPDATE, DELETE, DDL) → un `sql.Result` ;
- **`QueryContext`** pour un SELECT de plusieurs lignes → un `*sql.Rows` à parcourir ;
- **`QueryRowContext`** pour au plus une ligne → un `*sql.Row` dont l'erreur est reportée sur `Scan`.

### Lire plusieurs lignes : `Next` / `Scan` / `Err` / `Close`

Le parcours d'un `*sql.Rows` suit toujours le même squelette, et l'oubli le plus fréquent — vérifier `rows.Err()` — est aussi le plus insidieux : `Next()` renvoie `false` **aussi bien** à la fin normale qu'en cas d'erreur d'itération, et seul `rows.Err()` les distingue.

```go
type User struct {
	ID   int64
	Name string
}

func ListActiveUsers(ctx context.Context, db *sql.DB) ([]User, error) {
	const q = `SELECT id, name FROM users WHERE active = $1 ORDER BY id`
	rows, err := db.QueryContext(ctx, q, true)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // impératif : libère la connexion vers le pool

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err() // distingue fin d'itération et erreur
}
```

`defer rows.Close()` est indispensable : tant que les `Rows` ne sont pas fermées, la connexion sous-jacente reste retenue hors du pool.

### Lire une seule ligne : `QueryRowContext` et `sql.ErrNoRows`

Quand aucune ligne ne correspond, `Scan` renvoie `sql.ErrNoRows`. Ce n'est pas une panne mais un **cas métier**, que l'on traduit en erreur sentinelle ([§ 2.9](../02-fondamentaux-langage/09-gestion-erreurs.md)) plutôt que de le laisser remonter brut.

```go
var ErrNotFound = errors.New("utilisateur introuvable")

func GetUser(ctx context.Context, db *sql.DB, id int64) (User, error) {
	const q = `SELECT id, name FROM users WHERE id = $1`
	var u User
	err := db.QueryRowContext(ctx, q, id).Scan(&u.ID, &u.Name)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return User{}, ErrNotFound
	case err != nil:
		return User{}, err
	}
	return u, nil
}
```

### Écrire : `ExecContext` et `sql.Result`

`sql.Result` expose `RowsAffected()` (largement supporté) et `LastInsertId()` — mais **ce dernier n'est pas fourni par tous les moteurs** : PostgreSQL, notamment, ne le propose pas ; on récupère alors l'identifiant généré via `RETURNING`.

```go
func RenameUser(ctx context.Context, db *sql.DB, id int64, name string) error {
	res, err := db.ExecContext(ctx, `UPDATE users SET name = $1 WHERE id = $2`, name, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound // aucune ligne mise à jour
	}
	return nil
}

// PostgreSQL : pas de LastInsertId, on utilise RETURNING.
func InsertUser(ctx context.Context, db *sql.DB, name string) (int64, error) {
	const q = `INSERT INTO users (name, active) VALUES ($1, $2) RETURNING id`
	var id int64
	err := db.QueryRowContext(ctx, q, name, true).Scan(&id)
	return id, err
}
```

## Valeurs `NULL` et types

Une colonne `NULL` ne peut pas être scannée dans un `string` ou un `int`. Trois approches, par ordre de préférence habituel : les types **`sql.Null…`**, la forme générique **`sql.Null[T]`** (depuis Go 1.22), ou un **pointeur** (`nil` pour `NULL`).

```go
var nick sql.NullString // { String string; Valid bool }
err := db.QueryRowContext(ctx, `SELECT nickname FROM users WHERE id = $1`, id).Scan(&nick)
// nick.Valid == false si NULL, sinon nick.String

var nick2 sql.Null[string] // Go 1.22+ : { V string; Valid bool }, uniforme pour tout type
```

## Requêtes préparées

`PrepareContext` renvoie un `*sql.Stmt` réutilisable, utile lorsqu'on exécute **la même requête de nombreuses fois**.

```go
stmt, err := db.PrepareContext(ctx, `SELECT name FROM users WHERE id = $1`)
if err != nil {
	return err
}
defer stmt.Close()

for _, id := range ids {
	var name string
	if err := stmt.QueryRowContext(ctx, id).Scan(&name); err != nil {
		return err
	}
	// ...
}
```

Nuance propre à `database/sql` : un `Stmt` est lié à une connexion du pool ; s'il est réutilisé sur une autre connexion, le paquet le **re-prépare** de façon transparente, ce qui peut engendrer plus d'allers-retours que prévu. En pratique : préparez explicitement pour une requête répétée dans une boucle serrée ; pour une requête ponctuelle, passez directement le SQL à `QueryContext`/`ExecContext` (certains drivers, dont pgx, gèrent en plus leur propre cache d'instructions — [§ 7.2](02-drivers.md)).

Point de sécurité essentiel : **les requêtes paramétrées protègent contre l'injection SQL**. On ne construit *jamais* une requête par concaténation de chaînes contenant une entrée utilisateur ; on passe toujours les valeurs comme paramètres (`$1`, `?`), jamais dans le texte de la requête (approfondi en [§ 16.1](../16-securite/01-owasp-go.md)). Pour les rares cas nécessitant des paramètres nommés, `sql.Named` existe, avec une syntaxe dépendante du driver ; les placeholders positionnels restent la norme.

## Transactions

Une transaction s'ouvre avec `BeginTx`, se valide avec `Commit` et s'annule avec `Rollback`. L'idiome consiste à **différer le `Rollback` juste après le `BeginTx`** : si le `Commit` réussit, le `Rollback` différé devient sans effet (il renvoie `sql.ErrTxDone`, que l'on ignore) ; à la moindre erreur, le retour anticipé déclenche l'annulation.

```go
func Transfer(ctx context.Context, db *sql.DB, from, to int64, amount int) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // sans effet si Commit a déjà réussi

	const debit = `UPDATE accounts SET balance = balance - $1 WHERE id = $2`
	if _, err := tx.ExecContext(ctx, debit, amount, from); err != nil {
		return err // annulation via le defer
	}
	const credit = `UPDATE accounts SET balance = balance + $1 WHERE id = $2`
	if _, err := tx.ExecContext(ctx, credit, amount, to); err != nil {
		return err
	}
	return tx.Commit()
}
```

Deux règles à retenir. À l'intérieur d'une transaction, on utilise **`tx.ExecContext` / `tx.QueryContext`** — jamais les méthodes de `db`, qui emprunteraient une autre connexion, hors transaction. Et l'on garde une transaction **courte** : elle immobilise une connexion et peut poser des verrous. Le niveau d'isolation et le mode lecture seule se règlent via `sql.TxOptions`, avec un support variable selon le moteur :

```go
tx, err := db.BeginTx(ctx, &sql.TxOptions{
	Isolation: sql.LevelSerializable,
	ReadOnly:  true,
})
```

## Côté IDE : GoLand et VS Code

L'enjeu ici est d'**écrire et valider le SQL incorporé dans les chaînes Go**, et de diagnostiquer les erreurs de `Scan` (nombre de colonnes ou types incompatibles).

**GoLand** peut **injecter le langage SQL** dans les chaînes de requête — automatiquement, ou via *Alt+Entrée → « Inject language or reference »* — offrant coloration, autocomplétion et inspections dès qu'une *Data Source* correspondant à votre schéma est configurée dans la fenêtre *Database*. Un *scratch* SQL permet d'éprouver une requête avant de la coller dans le code ; un point d'arrêt sur l'erreur de `Scan` (Delve) aide à repérer un décalage colonnes/champs.

**VS Code** n'interprète pas le SQL à l'intérieur des chaînes Go : on prototype la requête dans un fichier `.sql` via l'extension **SQLTools** (et son pilote), puis on la transfère dans le code ; `gopls` couvre le Go, `dlv` le débogage.

Dans les deux cas, pour exécuter le code contre une **vraie** base plutôt qu'un simulacre, les Testcontainers lancent un moteur jetable en conteneur ([§ 13.3](../13-tests-qualite/03-tests-integration.md)).

## En résumé

- `sql.Open` n'ouvre pas de connexion : bornez le pool (`SetMaxOpenConns`), validez avec `PingContext`, et gardez un unique `*sql.DB` pour toute l'application.
- Trois verbes : `ExecContext` (écritures), `QueryContext` (plusieurs lignes), `QueryRowContext` (une ligne) ; toujours en variante `...Context`.
- Parcours de `Rows` : `defer rows.Close()` **et** vérifier `rows.Err()` après la boucle ; `sql.ErrNoRows` (via `errors.Is`) est un cas métier.
- `LastInsertId` n'est pas universel (PostgreSQL → `RETURNING`) ; gérez le `NULL` avec `sql.Null…`, `sql.Null[T]` (Go 1.22) ou un pointeur.
- Requêtes **paramétrées** contre l'injection (jamais de concaténation, [§ 16.1](../16-securite/01-owasp-go.md)) ; `PrepareContext` pour une requête très répétée.
- Transactions : `defer tx.Rollback()` après `BeginTx`, `tx.*` à l'intérieur (pas `db.*`), et courtes.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [7.2 — Drivers : PostgreSQL (pgx), MySQL, SQLite](02-drivers.md)

⏭ [Drivers : PostgreSQL (pgx) ⭐, MySQL, SQLite](/07-acces-donnees/02-drivers.md)
