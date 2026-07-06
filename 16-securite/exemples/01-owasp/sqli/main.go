/* ============================================================================
   Section 16.1 : Injection SQL (A05:2025)
   Description : La démonstration de la section, sur une VRAIE base PostgreSQL.
                 La même charge d'injection « x' OR '1'='1 » est envoyée deux
                 fois : (❌) concaténée dans la requête, elle contourne le
                 filtre et remonte TOUTES les lignes ; (✅) passée par un
                 placeholder $1, elle est traitée comme une valeur inerte et ne
                 correspond à rien. C'est le driver, pas vous, qui assemble la
                 requête finale.
   Fichier source : 16.1 (01-owasp-go.md)
   Prérequis : Postgres via docker-compose (voir le README) sur 127.0.0.1:5433.
   Lancer : docker compose up -d  puis  go run .
   ============================================================================ */

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	dsn := "postgres://app:app@127.0.0.1:5433/app?sslmode=disable"
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("connexion (Postgres démarré ? `docker compose up -d`) : %v", err)
	}
	defer pool.Close()

	// Schéma + données de démonstration (idempotent).
	if _, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS users (
		id serial PRIMARY KEY, name text, email text);
		TRUNCATE users;
		INSERT INTO users(name,email) VALUES ('alice','a@x.fr'),('bob','b@x.fr');`); err != nil {
		log.Fatalf("init : %v", err)
	}

	const attaque = "x' OR '1'='1" // charge d'injection SQL classique

	// ❌ Concaténation VULNÉRABLE : la donnée devient du code SQL.
	var nVuln int
	vuln := fmt.Sprintf("SELECT count(*) FROM users WHERE name = '%s'", attaque)
	if err := pool.QueryRow(ctx, vuln).Scan(&nVuln); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("❌ concaténation : %d lignes remontées — l'injection matche TOUT\n", nVuln)

	// ✅ Placeholder $1 : la donnée reste une valeur inerte.
	var nSafe int
	if err := pool.QueryRow(ctx,
		"SELECT count(*) FROM users WHERE name = $1", attaque).Scan(&nSafe); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ placeholder $1 : %d ligne — la charge est traitée comme un nom littéral\n", nSafe)
}
