/* ============================================================================
   Section 12.3 : log/slog — journalisation structurée
   Description : Le conseil de prudence du cours — écrire un Handle correct est
                 subtil ; le paquet testing/slogtest VALIDE un handler contre
                 la spécification slog. Ici, un handler qui ENVELOPPE un
                 JSONHandler (l'approche recommandée : envelopper plutôt que
                 réécrire) est passé à slogtest.TestHandler.
   Fichier source : 03-slog.md
   Lancer : go test ./...
   ============================================================================ */

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"testing/slogtest"
)

// prefixHandler enveloppe un JSONHandler en laissant tout passer : il doit
// donc satisfaire la spécification slog.Handler, ce que slogtest vérifie.
type prefixHandler struct{ slog.Handler }

func (p prefixHandler) Handle(ctx context.Context, r slog.Record) error {
	return p.Handler.Handle(ctx, r)
}

func TestHandlerConforme(t *testing.T) {
	var buf bytes.Buffer
	h := prefixHandler{slog.NewJSONHandler(&buf, nil)}

	results := func() []map[string]any {
		var out []map[string]any
		for _, line := range bytes.Split(buf.Bytes(), []byte("\n")) {
			if len(line) == 0 {
				continue
			}
			var m map[string]any
			if err := json.Unmarshal(line, &m); err != nil {
				t.Fatal(err)
			}
			out = append(out, m)
		}
		return out
	}
	if err := slogtest.TestHandler(h, results); err != nil {
		t.Fatalf("handler non conforme à la spécification slog : %v", err)
	}
}
