/* ============================================================================
   Section 7.2 : Drivers — PostgreSQL (pgx), MySQL, SQLite
   Description : Le driver le plus autonome pour toucher le MÉCANISME : SQLite
                 pur Go (modernc.org/sqlite, nom « sqlite », zéro serveur,
                 zéro cgo). Import « blanc », placeholders « ? », LastInsertId
                 (supporté ici, contrairement à PostgreSQL), mode WAL +
                 SetMaxOpenConns(1) pour l'écriture concurrente. La variante
                 cgo (mattn, « sqlite3 ») et son piège du stub sont en
                 commentaire — et les DSN pgx/MySQL (parseTime !) au README.
   Fichier source : 02-drivers.md
   Lancer : go run .        (crée puis supprime ./app.db — aucun service requis)
   ============================================================================ */

package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite" // pur Go : s'enregistre sous "sqlite" (PAS "sqlite3")
	// Variante cgo — fonctionnalités maximales mais CGO_ENABLED=1 + gcc requis :
	//   import _ "github.com/mattn/go-sqlite3"        // s'enregistre sous "sqlite3"
	//   db, err := sql.Open("sqlite3", "./app.db")
	// Piège : avec CGO_ENABLED=0 le build de mattn RÉUSSIT (stub), et le binaire
	// panique au premier usage : « go-sqlite3 requires cgo to work. This is a stub ».
)

func main() {
	ctx := context.Background()

	// SQLite crée le fichier s'il n'existe pas — une base = un fichier.
	db, err := sql.Open("sqlite", "./app.db")
	if err != nil {
		panic(err)
	}
	// Démo : on ne laisse rien derrière. On ferme AVANT de supprimer, et le
	// mode WAL crée des fichiers compagnons (-wal, -shm) à retirer aussi.
	defer func() {
		db.Close()
		for _, f := range []string{"./app.db", "./app.db-wal", "./app.db-shm"} {
			os.Remove(f)
		}
	}()

	// SQLite n'accepte qu'UN écrivain à la fois : WAL + pool d'écriture à 1
	// (la recommandation de la section pour un contexte concurrent).
	if _, err := db.ExecContext(ctx, `PRAGMA journal_mode=WAL`); err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(1)

	// Placeholders « ? » (pas « $1 » : ça, c'est PostgreSQL).
	mustExec(ctx, db, `CREATE TABLE notes (id INTEGER PRIMARY KEY, texte TEXT NOT NULL)`)
	res, err := db.ExecContext(ctx, `INSERT INTO notes (texte) VALUES (?)`, "bonjour")
	if err != nil {
		panic(err)
	}

	// LastInsertId : SQLite (comme MySQL) le fournit — PostgreSQL, non.
	id, err := res.LastInsertId()
	fmt.Println("fichier app.db créé · LastInsertId =", id, "· err =", err)

	var texte string
	_ = db.QueryRowContext(ctx, `SELECT texte FROM notes WHERE id = ?`, id).Scan(&texte)
	fmt.Println("relu :", texte)
	fmt.Println("driver « sqlite » (modernc, pur Go) : compatible CGO_ENABLED=0 ✔")
}

func mustExec(ctx context.Context, db *sql.DB, q string) {
	if _, err := db.ExecContext(ctx, q); err != nil {
		panic(err)
	}
}
