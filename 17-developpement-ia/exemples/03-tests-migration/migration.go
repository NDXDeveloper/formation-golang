/* ============================================================================
   Section 17.3 §2 : Migration assistée (idiomatique, pas littéral)
   Description : La correspondance clé d'une migration Python/Java → Go :
                 l'EXCEPTION devient une ERREUR RETOURNÉE. Là où le Python
                 lèverait `raise NotFoundError(uid)`, le Go idiomatique renvoie
                 `(nil, ErrNotFound)` — une sentinelle testable par errors.Is.
                 Le store est ici en mémoire pour rester autonome ; dans un
                 vrai service, ce serait `db.QueryRowContext` + `sql.ErrNoRows`.
   Fichier source : 17.3 (03-tests-migration-ia.md)
   ============================================================================ */

package main

import (
	"context"
	"errors"
	"fmt"
)

// ErrNotFound : sentinelle, l'équivalent Go de la NotFoundError de la source.
var ErrNotFound = errors.New("utilisateur introuvable")

type User struct {
	ID   int
	Name string
}

// store en mémoire (tiendrait la place de la base dans un vrai service).
var users = map[int]string{1: "alice", 2: "bob"}

// LoadUser : traduction idiomatique de la fonction Python `load_user`.
// Python : `if row is None: raise NotFoundError(uid)`.
// Go     : l'absence devient une erreur wrappée testable par errors.Is.
func LoadUser(ctx context.Context, uid int) (*User, error) {
	name, ok := users[uid]
	if !ok {
		return nil, fmt.Errorf("utilisateur %d : %w", uid, ErrNotFound)
	}
	return &User{ID: uid, Name: name}, nil
}
