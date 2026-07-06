/* ============================================================================
   Section 16.2 : Cryptographie (crypto/*), TLS, gestion des secrets
   Description : Orchestration des démonstrations de la section : les primitives
                 (crypto.go), le TLS post-quantique et HPKE (tls.go), et le
                 masquage des secrets dans les logs (ci-dessous). Un type
                 Secret qui implémente slog.LogValuer rend une fuite accidentelle
                 STRUCTURELLEMENT impossible : la valeur ne s'imprime jamais en
                 clair ; l'accès explicite passe par Reveal().
   Fichier source : 16.2 (02-cryptographie-tls.md)
   Lancer : go run .
   ============================================================================ */

package main

import (
	"log/slog"
	"os"
)

// §3.3 Secret : masqué dans les logs, révélé seulement à la demande.
type Secret string

func (Secret) String() string       { return "[REDACTED]" }
func (Secret) LogValue() slog.Value { return slog.StringValue("[REDACTED]") }
func (s Secret) Reveal() string     { return string(s) }

func main() {
	demoPrimitives()
	demoTLSPostQuantique()
	demoHPKE()

	// §3.3 : la clé n'apparaît jamais en clair dans le log.
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	}))
	log.Info("config chargée", "api_key", Secret("sk-tres-secret"))
}
