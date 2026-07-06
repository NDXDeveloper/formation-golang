/* ============================================================================
   Section 12.1 : Stratégies d'erreurs à l'échelle d'une application
   Description : La taxonomie d'erreurs du cours — une erreur SENTINELLE
                 (valeur exportée, condition connue, testée par errors.Is) et
                 une erreur TYPÉE (transporte des données, extraite par
                 errors.As). Ces deux formes font partie du contrat du
                 package : opaque par défaut, sentinelle/type seulement quand
                 un appelant a besoin de brancher dessus.
   Fichier source : 01-strategies-erreurs.md
   ============================================================================ */

package store

import (
	"errors"
	"fmt"
)

// ErrNotFound est une sentinelle : une condition connue, partie du contrat.
var ErrNotFound = errors.New("store: enregistrement introuvable")

// ValidationError est une erreur typée : elle transporte des données
// exploitables par l'appelant (le champ fautif et le message).
type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("champ %q invalide : %s", e.Field, e.Msg)
}
