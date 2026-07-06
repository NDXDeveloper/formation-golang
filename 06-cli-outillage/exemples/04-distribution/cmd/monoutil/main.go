/* ============================================================================
   Section 6.4 : Distribution — binaire unique, cross-compilation, GoReleaser
   Description : L'outil à distribuer — les DEUX approches d'estampillage de
                 la section : variables remplies par -ldflags "-X ..." (ce que
                 fait GoReleaser avec {{ .Version }}), et lecture des
                 métadonnées incorporées via runtime/debug.ReadBuildInfo
                 (sans aucun ldflags). Layout ./cmd/monoutil comme dans les
                 commandes du chapitre. Builds : voir le README (statique,
                 -s -w, -trimpath, GOOS/GOARCH, GoReleaser, Docker).
   Fichier source : 04-distribution.md
   ============================================================================ */

package main

import (
	"fmt"
	"runtime/debug"
)

// Remplies au build par : -ldflags "-X main.version=… -X main.commit=… -X main.date=…"
// (GoReleaser les alimente depuis le tag Git : version affichée = version publiée).
var (
	version = "dev"     // remplacé au build par -X main.version=...
	commit  = "none"    // ... -X main.commit=...
	date    = "unknown" // ... -X main.date=...
)

// buildVersion lit la version du module INCORPORÉE par la toolchain
// (depuis Go 1.18) : « v1.2.3 » après go install …@v1.2.3, sinon « (devel) ».
func buildVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return "dev"
}

func main() {
	fmt.Printf("monoutil %s (commit %s, date %s)\n", version, commit, date)
	fmt.Println("version du module (ReadBuildInfo) :", buildVersion())
}
