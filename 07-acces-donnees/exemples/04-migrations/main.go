/* ============================================================================
   Section 7.4 : Migrations de schéma (golang-migrate, goose)
   Description : Le bloc « bibliothèque + migrations embarquées » de la
                 section, complet — //go:embed sur migrations/*.sql, source
                 iofs, m.Up() avec migrate.ErrNoChange filtré par errors.Is
                 (relancer ne change rien : idempotent). L'artefact est
                 AUTONOME : les migrations voyagent dans le binaire (§ 6.4).
                 (goose, l'alternative avec migrations en Go, et les CLI
                 restent illustrés dans le .md — validés par ailleurs.)
   Fichier source : 04-migrations.md
   Prérequis : le container PostgreSQL du README, puis :
               DATABASE_URL="postgres://postgres:secret@127.0.0.1:5433/formation?sslmode=disable" go run .
   ============================================================================ */

package main

import (
	"embed"
	"errors"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // moteur cible
	"github.com/golang-migrate/migrate/v4/source/iofs"         // source : un fs.FS
)

//go:embed migrations/*.sql
var migrationsFS embed.FS // les .up.sql / .down.sql voyagent DANS le binaire

// Migrate applique les migrations manquantes ; « déjà à jour » n'est pas
// une erreur (ErrNoChange, filtrée) — on peut l'appeler à chaque démarrage.
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

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL manquante — lancez le container du README d'abord")
		os.Exit(1)
	}

	fmt.Println("1er passage  :", orOK(Migrate(dsn)), "(migrations appliquées)")
	fmt.Println("2e  passage  :", orOK(Migrate(dsn)), "(ErrNoChange filtré : déjà à jour)")
	fmt.Println("la table de suivi « schema_migrations » tient le registre dans la base ✔")
}

func orOK(err error) string {
	if err != nil {
		return "ERREUR : " + err.Error()
	}
	return "OK"
}
