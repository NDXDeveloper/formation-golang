/* ============================================================================
   Section 13.1 : Package testing, table-driven tests et sous-tests
   Description : Tout l'outillage de test de la stdlib, en un fichier —
                 le test TABLE-DRIVEN avec sous-tests t.Run (l'idiome Go), le
                 helper générique marqué t.Helper() (le rapport d'échec pointe
                 l'appelant), Error vs Fatal, le cycle de vie (t.Cleanup,
                 t.TempDir), t.Setenv, testing.Short(), le patron GOLDEN FILE
                 (régénérable avec -update), les fonctions Example exécutées
                 (comparées à // Output:), et TestMain (setup/teardown du
                 package). AUCUNE dépendance : go test suffit.
   Fichier source : 01-tests-unitaires.md
   Lancer : go test -v ./...   ·   go test -short ./...   ·   go test -update ./...
   ============================================================================ */

package stringutil

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// assertEqual : helper générique. t.Helper() fait que l'échec est rapporté
// à la LIGNE D'APPEL, pas à l'intérieur du helper.
func assertEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

// Le test TABLE-DRIVEN ⭐ : décrire les cas, itérer, un sous-test par cas.
func TestReverse(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"chaîne vide", "", ""},
		{"un caractère", "a", "a"},
		{"mot ASCII", "Golang", "gnaloG"},
		{"caractère accentué", "café", "éfac"}, // vérifie le raisonnement en runes
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertEqual(t, Reverse(tc.in), tc.want)
		})
	}
}

// Error rapporte PLUSIEURS problèmes en une exécution (contrairement à Fatal).
func TestSomme(t *testing.T) {
	assertEqual(t, Somme(2, 3), 5)
	assertEqual(t, Somme(-1, 1), 0)
}

// Cycle de vie : t.TempDir (effacé auto) + t.Cleanup (après le test).
func TestAvecFichier(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(path, []byte("bonjour"), 0o644); err != nil {
		t.Fatalf("écriture : %v", err) // Fatal : inutile de continuer
	}
	t.Cleanup(func() { t.Log("cleanup exécuté après le test") })
	b, _ := os.ReadFile(path)
	assertEqual(t, string(b), "bonjour")
}

// t.Setenv : fixe une variable et la restaure (incompatible t.Parallel).
func TestSetenv(t *testing.T) {
	t.Setenv("APP_MODE", "test")
	assertEqual(t, os.Getenv("APP_MODE"), "test")
}

// testing.Short() : écarter les tests lents d'une boucle rapide.
func TestLent(t *testing.T) {
	if testing.Short() {
		t.Skip("ignoré en mode court (-short)")
	}
	t.Log("test long exécuté")
}

// Patron GOLDEN FILE : comparer à un fichier de référence, régénérable.
var update = flag.Bool("update", false, "met à jour les fichiers .golden")

func TestRender(t *testing.T) {
	got := Render("bonjour")
	golden := filepath.Join("testdata", "render.golden")
	if *update {
		if err := os.WriteFile(golden, got, 0o644); err != nil {
			t.Fatalf("écriture golden : %v", err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("lecture golden : %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("sortie différente du golden ; relancer avec -update pour régénérer")
	}
}

// Example : documentation exécutable (la sortie est comparée à // Output:).
func ExampleReverse() {
	fmt.Println(Reverse("Golang"))
	// Output: gnaloG
}

// Unordered output : compare sans tenir compte de l'ordre (parcours de map…).
func Example_unordered() {
	for _, s := range []string{"b", "a"} {
		fmt.Println(s)
	}
	// Unordered output:
	// a
	// b
}

// TestMain : mise en place partagée par tout le package.
func TestMain(m *testing.M) {
	fmt.Println("[setup] avant tous les tests du package")
	code := m.Run() // exécute toutes les fonctions TestXxx
	fmt.Println("[teardown] après tous les tests")
	os.Exit(code) // n'exécute PAS les defer : teardown avant, explicitement
}
