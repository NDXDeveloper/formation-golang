/* ============================================================================
   Section 17.1 : Prompting efficace pour du Go idiomatique
   Description : Le code qu'un BON prompt produit — celui qu'on n'a pas à
                 réécrire. La demande cadrée (« signature imposée, stdlib
                 uniquement, erreurs explicites wrappées avec %w, jamais de
                 panic, aucun état global ») donne ce LoadConfig idiomatique,
                 à comparer au générique vague qu'aurait produit « écris une
                 fonction qui lit un fichier de config ». Le fichier
                 AGENTS.md voisin montre les instructions persistantes qui
                 évitent de répéter ces contraintes à chaque prompt.
   Fichier source : 17.1 (01-prompting-go.md)
   Lancer : go run .
   ============================================================================ */

package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Addr    string `json:"addr"`
	Timeout int    `json:"timeout"`
}

// LoadConfig lit un fichier JSON et le désérialise dans *Config.
// Retourne une erreur wrappée si le fichier est absent ou mal formé.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("lecture config %q : %w", path, err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config %q : %w", path, err)
	}
	return &cfg, nil
}

func main() {
	// écrit un fichier de démonstration, le charge, puis nettoie
	const path = "config.json"
	_ = os.WriteFile(path, []byte(`{"addr":":8080","timeout":30}`), 0o644)
	defer os.Remove(path)

	cfg, err := LoadConfig(path)
	if err != nil {
		fmt.Println("erreur :", err)
		return
	}
	fmt.Printf("config chargée : Addr=%s Timeout=%d\n", cfg.Addr, cfg.Timeout)

	// cas d'erreur : le %w préserve la chaîne (errors.Is sur os.ErrNotExist).
	if _, err := LoadConfig("absent.json"); err != nil {
		fmt.Println("fichier absent → erreur wrappée :", err)
	}
}
