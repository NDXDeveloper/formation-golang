/* ============================================================================
   Section 3.5 : Organisation du code (workspaces)
   Description : La bibliothèque « en cours de développement » — jamais publiée ;
                 app la voit grâce au go.work (modifiez Message() et relancez
                 app : la modification est prise immédiatement)
   Fichier source : 05-organisation-code.md
   ============================================================================ */

// Package lib est la bibliothèque locale de la démonstration go work.
package lib

// Message identifie la version utilisée.
func Message() string { return "version LOCALE de lib, vue via go.work" }
