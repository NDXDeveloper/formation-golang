/* ============================================================================
   Section 3.5 : Organisation du code : packages, internal/, layout, workspaces
   Description : Package d'implémentation placé sous internal/ — importable
                 UNIQUEMENT par le code de ce module (github.com/acme/app/…) ;
                 depuis tout autre module, l'import serait refusé à la
                 compilation (« use of internal package … not allowed »)
   Fichier source : 05-organisation-code.md
   ============================================================================ */

// Package store donne accès aux données (implémentation privée du module).
package store

// Store donne accès aux données (champs omis pour l'exemple).
type Store struct{}

// New construit un Store prêt à l'emploi.
func New() *Store { return &Store{} }

// Greet prouve que le package interne est bien utilisable depuis cmd/api.
func (s *Store) Greet() string { return "store interne : accessible depuis cmd/api (même module)" }
