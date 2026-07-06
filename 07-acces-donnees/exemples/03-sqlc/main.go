/* ============================================================================
   Section 7.3 : sqlc (SQL typé généré) vs ORM — le défaut idiomatique
   Description : Le flux sqlc complet de la section — db/schema.sql et
                 db/queries.sql (annotées « -- name: ... :one ») sont la
                 SOURCE DE VÉRITÉ ; « sqlc generate » (directive go:generate
                 ci-dessous) produit db/gen/ : types, méthodes typées et
                 l'interface Querier (emit_interface). Ici on UTILISE ce code
                 généré avec un pool pgx : zéro réflexion, types vérifiés à
                 la compilation. (GORM et Ent, les deux autres familles de la
                 grille, restent illustrés dans le .md.)
   Fichier source : 03-sqlc-vs-orm.md
   Prérequis : le container PostgreSQL du README, puis :
               DATABASE_URL="postgres://postgres:secret@127.0.0.1:5433/formation?sslmode=disable" go run .
   Régénérer : go generate ./...   (nécessite sqlc : cf. README)
   ============================================================================ */

//go:generate sqlc generate

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/exemple/sqlcdemo/db/gen"
)

func main() {
	ctx := context.Background()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL manquante — lancez le container du README d'abord")
		os.Exit(1)
	}

	pool, err := pgxpool.New(ctx, dsn) // le pool natif pgx du § 7.2
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	// Le schéma (en vrai : appliqué par une migration, § 7.4).
	if _, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS authors (
		id BIGSERIAL PRIMARY KEY, name TEXT NOT NULL, bio TEXT)`); err != nil {
		panic(err)
	}

	// db.New accepte *pgx.Conn ou *pgxpool.Pool — l'API GÉNÉRÉE par sqlc.
	q := db.New(pool)

	// CreateAuthor : signature et types produits depuis le SQL annoté.
	created, err := q.CreateAuthor(ctx, db.CreateAuthorParams{
		Name: "Kernighan",
		Bio:  pgtype.Text{String: "co-auteur de K&R", Valid: true}, // NULL-able → pgtype.Text
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("créé : id=%d name=%s\n", created.ID, created.Name)

	// GetAuthor : → (db.Author, error), zéro Scan manuel, zéro réflexion.
	author, err := q.GetAuthor(ctx, created.ID)
	if err != nil {
		panic(err)
	}
	fmt.Printf("relu : %s · bio valide : %t\n", author.Name, author.Bio.Valid)
	fmt.Println("sqlc : le SQL est la source de vérité, le Go est généré et typé ✔")
}
