/* ============================================================================
   Section 10.3 : Configuration, feature flags, principes 12-factor
   Description : Le chargeur de configuration du cours — l'idiome en quatre
                 gestes : une structure Config TYPÉE, chargée UNE FOIS,
                 VALIDÉE immédiatement (fail fast), puis passée explicitement.
                 Load applique la précédence « défauts < variables
                 d'environnement < flags » sur un FlagSet LOCAL (pas d'état
                 global, ContinueOnError = composable et testable). La stdlib
                 suffit : flag + os.LookupEnv (qui distingue « non défini »
                 de « défini à vide », contrairement à os.Getenv).
   Fichier source : 03-configuration-12factor.md
   ============================================================================ */

package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"
)

// Config est typée, chargée une fois, validée au démarrage, puis passée explicitement.
type Config struct {
	Addr        string
	DatabaseURL string // secret : jamais journalisé (cf. § 16.2)
	LogLevel    string
	Timeout     time.Duration
}

// Load applique la précédence : défauts < variables d'environnement < flags.
func Load(args []string) (Config, error) {
	c := Config{ // 1) valeurs par défaut
		Addr:     ":8080",
		LogLevel: "info",
		Timeout:  5 * time.Second,
	}

	// 2) environnement (12-factor : la config vit dans l'environnement)
	if v, ok := os.LookupEnv("APP_ADDR"); ok {
		c.Addr = v
	}
	c.DatabaseURL = os.Getenv("DATABASE_URL")
	if v, ok := os.LookupEnv("LOG_LEVEL"); ok {
		c.LogLevel = v
	}

	// 3) flags (priorité la plus haute) sur un FlagSet LOCAL — pas d'état global.
	// ContinueOnError renvoie l'erreur au lieu d'appeler os.Exit : composable et testable.
	fs := flag.NewFlagSet("app", flag.ContinueOnError)
	fs.StringVar(&c.Addr, "addr", c.Addr, "adresse d'écoute HTTP")
	fs.StringVar(&c.LogLevel, "log-level", c.LogLevel, "niveau de log")
	fs.DurationVar(&c.Timeout, "timeout", c.Timeout, "timeout des requêtes")
	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	if err := c.validate(); err != nil { // 4) fail fast
		return Config{}, fmt.Errorf("configuration invalide : %w", err)
	}
	return c, nil
}

func (c Config) validate() error {
	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL est requis")
	}
	if c.Timeout <= 0 {
		return errors.New("timeout doit être positif")
	}
	return nil
}
