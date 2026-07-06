/* ============================================================================
   Section 7.1 : database/sql (pool, requêtes préparées, transactions)
   Description : La section entière, exécutable contre un PostgreSQL en
                 container — pool borné + PingContext, les trois verbes
                 (Exec/Query/QueryRow ...Context), Next/Scan/Err/Close,
                 sql.ErrNoRows → sentinelle métier, RowsAffected,
                 INSERT ... RETURNING (pas de LastInsertId en PostgreSQL —
                 démontré), NULL (NullString et sql.Null[T] 1.22), requête
                 préparée en boucle, transaction : defer tx.Rollback(),
                 commit puis rollback réel sur contrainte violée
   Fichier source : 01-database-sql.md
   Prérequis : le container PostgreSQL du README (port 5433), puis :
               DATABASE_URL="postgres://postgres:secret@127.0.0.1:5433/formation?sslmode=disable" go run .
   ============================================================================ */

package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // driver PostgreSQL (pgx), enregistré sous "pgx"
)

// La sentinelle MÉTIER : l'appelant teste errors.Is(err, ErrNotFound),
// sans dépendre de database/sql.
var ErrNotFound = errors.New("utilisateur introuvable")

type User struct {
	ID   int64
	Name string
}

// Open : sql.Open ne se connecte PAS (pool paresseux) — on borne le pool
// puis on valide réellement la connexion avec PingContext.
func Open(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("ouverture : %w", err)
	}
	db.SetMaxOpenConns(25)                 // borne le nombre total de connexions
	db.SetMaxIdleConns(25)                 // inactives conservées (idéalement = max)
	db.SetConnMaxLifetime(5 * time.Minute) // recyclage périodique
	db.SetConnMaxIdleTime(1 * time.Minute)
	if err := db.PingContext(ctx); err != nil { // vérifie *réellement* la connexion
		db.Close()
		return nil, fmt.Errorf("connexion : %w", err)
	}
	return db, nil
}

// ListActiveUsers : le squelette Next/Scan/Err/Close — rows.Err() est
// l'oubli le plus insidieux (Next renvoie false à la fin ET sur erreur).
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

// GetUser : une ligne au plus — sql.ErrNoRows est un CAS MÉTIER, traduit
// en sentinelle, jamais laissé remonter brut.
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

// InsertUser : PostgreSQL n'a pas LastInsertId → RETURNING + QueryRow.
func InsertUser(ctx context.Context, db *sql.DB, name string) (int64, error) {
	const q = `INSERT INTO users (name, active) VALUES ($1, $2) RETURNING id`
	var id int64
	err := db.QueryRowContext(ctx, q, name, true).Scan(&id)
	return id, err
}

// Transfer : l'idiome transactionnel — defer tx.Rollback() JUSTE après
// BeginTx (sans effet après un Commit réussi), tx.* à l'intérieur (jamais db.*).
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

func main() {
	ctx := context.Background()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL manquante — lancez le container du README d'abord")
		os.Exit(1)
	}

	db, err := Open(ctx, dsn)
	if err != nil {
		fmt.Fprintln(os.Stderr, "erreur :", err)
		os.Exit(1)
	}
	defer db.Close()

	// Schéma de démonstration (IF NOT EXISTS : relançable), nettoyé à chaque run.
	mustExec(ctx, db, `CREATE TABLE IF NOT EXISTS users (
		id BIGSERIAL PRIMARY KEY, name TEXT NOT NULL, nickname TEXT,
		active BOOLEAN NOT NULL DEFAULT true)`)
	mustExec(ctx, db, `CREATE TABLE IF NOT EXISTS accounts (
		id BIGINT PRIMARY KEY, balance INT NOT NULL CHECK (balance >= 0))`)
	mustExec(ctx, db, `TRUNCATE users, accounts`)

	fmt.Println("=== INSERT ... RETURNING (l'idiome PostgreSQL) ===")
	id1, _ := InsertUser(ctx, db, "Ada")
	id2, _ := InsertUser(ctx, db, "Alan")
	mustExec(ctx, db, `INSERT INTO users (name, active) VALUES ('Inactif', false)`)
	fmt.Println("ids générés :", id1, id2)

	fmt.Println("=== LastInsertId ? Pas en PostgreSQL — la preuve ===")
	res, _ := db.ExecContext(ctx, `INSERT INTO users (name) VALUES ('X')`)
	if _, err := res.LastInsertId(); err != nil {
		fmt.Println("LastInsertId →", err) // « not supported by this driver »
	}

	fmt.Println("=== QueryContext : Next/Scan/Err (3 actifs sur 4) ===")
	users, err := ListActiveUsers(ctx, db)
	fmt.Println("actifs :", len(users), "· premier :", users[0].Name, "· err =", err)

	fmt.Println("=== QueryRowContext : trouvé, puis ErrNoRows → sentinelle ===")
	u, _ := GetUser(ctx, db, id1)
	fmt.Println("GetUser(id1) :", u.Name)
	_, err = GetUser(ctx, db, 999999)
	fmt.Println("id inconnu → ErrNotFound :", errors.Is(err, ErrNotFound))

	fmt.Println("=== NULL : NullString et sql.Null[T] (Go 1.22) ===")
	var nick sql.NullString
	_ = db.QueryRowContext(ctx, `SELECT nickname FROM users WHERE id = $1`, id1).Scan(&nick)
	var nick2 sql.Null[string]
	_ = db.QueryRowContext(ctx, `SELECT nickname FROM users WHERE id = $1`, id1).Scan(&nick2)
	fmt.Println("NullString.Valid :", nick.Valid, "· sql.Null[string].Valid :", nick2.Valid)

	fmt.Println("=== Requête préparée (réutilisée en boucle) ===")
	stmt, _ := db.PrepareContext(ctx, `SELECT name FROM users WHERE id = $1`)
	defer stmt.Close()
	for _, id := range []int64{id1, id2} {
		var name string
		_ = stmt.QueryRowContext(ctx, id).Scan(&name)
		fmt.Println("  stmt →", name)
	}

	fmt.Println("=== Transaction : commit, puis ROLLBACK réel (CHECK violé) ===")
	mustExec(ctx, db, `INSERT INTO accounts (id, balance) VALUES (1, 100), (2, 50)`)
	fmt.Println("Transfer(1→2, 30) :", Transfer(ctx, db, 1, 2, 30))
	fmt.Println("soldes :", soldes(ctx, db), "(attendu [70 80])")
	err = Transfer(ctx, db, 1, 99, 1000) // CHECK balance >= 0 → le débit casse
	fmt.Println("Transfer(1→99, 1000) échoue :", err != nil)
	fmt.Println("soldes après rollback :", soldes(ctx, db), "(intacts)")
}

func mustExec(ctx context.Context, db *sql.DB, q string) {
	if _, err := db.ExecContext(ctx, q); err != nil {
		fmt.Fprintln(os.Stderr, "erreur :", err)
		os.Exit(1)
	}
}

func soldes(ctx context.Context, db *sql.DB) [2]int {
	var s [2]int
	_ = db.QueryRowContext(ctx, `SELECT balance FROM accounts WHERE id = 1`).Scan(&s[0])
	_ = db.QueryRowContext(ctx, `SELECT balance FROM accounts WHERE id = 2`).Scan(&s[1])
	return s
}
