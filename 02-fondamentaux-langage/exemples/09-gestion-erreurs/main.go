/* ============================================================================
   Section 2.9 : Gestion des erreurs — l'idiome Go
   Description : Créer des erreurs (errors.New, fmt.Errorf), ajouter du contexte
                 avec le wrapping %w, inspecter avec errors.Is (sentinelle, à
                 travers la chaîne) et errors.As (erreur personnalisée typée)
   Fichier source : 09-gestion-erreurs.md
   ============================================================================ */

package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

// Le motif de base : message fixe → errors.New ; message formaté → fmt.Errorf.
func withdraw(balance, amount int) (int, error) {
	if amount > balance {
		return 0, errors.New("solde insuffisant")
	}
	if amount < 0 {
		return 0, fmt.Errorf("montant invalide : %d", amount)
	}
	return balance - amount, nil
}

// Le wrapping %w : on AJOUTE du contexte (le chemin) tout en PRÉSERVANT
// l'erreur d'origine dans une chaîne — errors.Is pourra la retrouver.
// (%v, lui, ne ferait que formater le texte : le lien serait perdu.)
type Config struct{}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("lecture de %s : %w", path, err)
	}
	_ = data
	return &Config{}, nil
}

// Une SENTINELLE : une valeur d'erreur exportée, que l'appelant reconnaît.
// Convention de nommage : ErrXxx.
var ErrNotFound = errors.New("ressource introuvable")

type Item struct{ ID int }

func find(id int) (*Item, error) {
	// On l'enveloppe : seule errors.Is (qui traverse la chaîne) la retrouvera.
	return nil, fmt.Errorf("id %d : %w", id, ErrNotFound)
}

// Une erreur PERSONNALISÉE : un type qui porte des données structurées.
// Il « est » une error parce qu'il a la méthode Error() string.
type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("champ %q : %s", e.Field, e.Msg)
}

func validate() error { return &ValidationError{Field: "email", Msg: "requis"} }

func main() {
	fmt.Println("=== Créer et retourner des erreurs ===")
	solde, err := withdraw(100, 30)
	fmt.Println("withdraw(100, 30) →", solde, err) // nil = succès
	_, err = withdraw(10, 20)
	fmt.Println("withdraw(10, 20)  → erreur :", err)
	_, err = withdraw(5, -1)
	fmt.Println("withdraw(5, -1)   → erreur :", err)

	fmt.Println("=== Wrapping %w (sur une vraie erreur d'E/S) ===")
	_, err = loadConfig("/chemin/inexistant.conf")
	fmt.Println("erreur enrichie :", err) // notre contexte + l'erreur système
	// errors.Is REMONTE la chaîne des %w jusqu'à trouver fs.ErrNotExist.
	fmt.Println("errors.Is(err, fs.ErrNotExist) →", errors.Is(err, fs.ErrNotExist))

	fmt.Println("=== Sentinelle : errors.Is, jamais == ===")
	_, err = find(42)
	fmt.Println("errors.Is(err, ErrNotFound) →", errors.Is(err, ErrNotFound))
	// == compare l'erreur ENVELOPPE, pas la sentinelle au fond de la chaîne :
	fmt.Println("err == ErrNotFound          →", err == ErrNotFound, "(le wrap casse ==)")

	fmt.Println("=== Erreur personnalisée : errors.As ===")
	err = validate()
	var verr *ValidationError
	// As cherche dans la chaîne une erreur DU TYPE demandé et l'extrait :
	// on accède alors à ses champs structurés.
	if errors.As(err, &verr) {
		fmt.Println("champ en cause :", verr.Field)
	}
	fmt.Println("message :", err)
}
