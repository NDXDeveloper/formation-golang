/* ============================================================================
   Section 17.2 : Pièges de l'IA en Go
   Description : Vérifie les versions ✅ des pièges. En particulier, un bon test
                 échoue si l'on casse le code (état d'esprit du mutation
                 testing, cf. §17.3) : asString doit rendre une erreur sur un
                 type inattendu, writeFileBon doit vraiment écrire.
   Fichier source : 17.2 (02-pieges-ia.md)
   Lancer : go test ./...
   ============================================================================ */

package pieges

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	n, err := parse("42")
	if err != nil || n != 42 {
		t.Fatalf("parse(42) = %d, %v", n, err)
	}
	if _, err := parse("pas un nombre"); err == nil {
		t.Error("parse devrait rendre une erreur, pas paniquer")
	}
}

func TestAsString(t *testing.T) {
	if s, err := asString("ok"); err != nil || s != "ok" {
		t.Fatalf("asString(string) = %q, %v", s, err)
	}
	if _, err := asString(42); err == nil {
		t.Error("asString(int) devrait rendre une erreur (assertion à deux valeurs)")
	}
}

func TestWriteFileBon(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.txt")
	if err := writeFileBon(path, []byte("données")); err != nil {
		t.Fatalf("writeFileBon : %v", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "données" {
		t.Errorf("contenu = %q, want 'données'", got)
	}
}
