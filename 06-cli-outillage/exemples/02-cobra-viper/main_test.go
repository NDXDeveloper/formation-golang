/* ============================================================================
   Section 6.2 : Cobra + Viper — tester une commande
   Description : L'exécution IN-MEMORY de la section : SetArgs injecte la
                 ligne de commande, SetOut/SetErr capturent les sorties,
                 Execute déroule l'arbre — sans binaire ni sous-processus.
                 Chaque test reconstruit un arbre neuf (newRootCmd) : aucun
                 état global ne fuit d'un test à l'autre.
   Fichier source : 02-cobra-viper.md
   Lancer : go test ./...
   ============================================================================ */

package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestServe_RefusePortInvalide(t *testing.T) {
	root := newRootCmd()
	root.SetArgs([]string{"serve", "--port", "70000"})
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)

	if err := root.Execute(); err == nil {
		t.Fatal("un port hors plage aurait dû échouer")
	}
}

func TestServe_RefuseArgumentPositionnel(t *testing.T) {
	root := newRootCmd() // arbre NEUF : rien ne fuit du test précédent
	root.SetArgs([]string{"serve", "inattendu"})
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)

	// cobra.NoArgs rejette le positionnel AVANT même d'entrer dans RunE.
	if err := root.Execute(); err == nil {
		t.Fatal("cobra.NoArgs aurait dû refuser l'argument")
	}
}

func TestVersion(t *testing.T) {
	root := newRootCmd()
	root.SetArgs([]string{"--version"})
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)

	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "1.4.0") {
		t.Fatalf("sortie --version : %q", out.String())
	}
}
