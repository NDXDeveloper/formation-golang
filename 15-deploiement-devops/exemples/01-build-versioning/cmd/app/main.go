/* ============================================================================
   Section 15.1 : Build reproductible, versioning (ldflags), go.mod / go.sum
   Description : Un binaire qui connaît sa propre version, par les DEUX voies
                 complémentaires de la section. (1) runtime/debug.ReadBuildInfo
                 récupère GRATUITEMENT l'info VCS estampillée par la chaîne
                 d'outils (révision, état modifié) quand on compile depuis un
                 dépôt Git. (2) Les variables `version`/`commit`, surchargées à
                 l'édition de liens par `-ldflags -X`, portent une étiquette de
                 release qui ne se déduit pas du VCS. Rappel : `-X` n'agit que
                 sur une VARIABLE de type chaîne (jamais une constante).
   Fichier source : 01-build-versioning.md
   Lancer : go run ./cmd/app   (ou voir le Makefile : make build / make release)
   ============================================================================ */

package main

import (
	"fmt"
	"runtime/debug"
)

// Surchargées au link : -X main.version=... -X main.commit=...
var (
	version = "dev"
	commit  = "none"
)

// buildInfo lit l'information VCS estampillée automatiquement par le compilateur.
func buildInfo() (rev string, dirty bool) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.modified":
			dirty = s.Value == "true"
		}
	}
	return
}

func main() {
	rev, dirty := buildInfo()
	fmt.Printf("version  = %s\n", version)
	fmt.Printf("commit   = %s\n", commit)
	fmt.Printf("vcs.rev  = %s (dirty=%t)\n", rev, dirty)
}
