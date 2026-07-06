/* ============================================================================
   Section 12.3 : log/slog — journalisation structurée
   Description : slog de bout en bout — le MÊME code, deux sorties (TextHandler
                 lisible en dev, JSONHandler machine en prod) ; ReplaceAttr qui
                 MASQUE un secret avant qu'il n'atteigne la sortie ; LevelVar
                 pour changer le niveau À CHAUD (sans redémarrer) ; groupes et
                 logger contextuel With ; Enabled qui court-circuite ; et la
                 nouveauté Go 1.26 slog.NewMultiHandler (fan-out « au mieux »,
                 erreurs agrégées par errors.Join). Le handler personnalisé
                 est validé contre la spécification par testing/slogtest
                 (voir wrap_test.go).
   Fichier source : 03-slog.md
   Lancer : go run .        (aucun service requis)
   ============================================================================ */

package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
)

// dropTime rend la sortie déterministe (on retire l'horodatage variable).
func dropTime(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey {
		return slog.Attr{}
	}
	return a
}

// puitsEnPanne : un handler dont Handle échoue TOUJOURS — pour prouver la
// diffusion « au mieux » et l'agrégation errors.Join du MultiHandler.
type puitsEnPanne struct{}

func (puitsEnPanne) Enabled(context.Context, slog.Level) bool  { return true }
func (puitsEnPanne) Handle(context.Context, slog.Record) error { return errors.New("puits HS") }
func (p puitsEnPanne) WithAttrs([]slog.Attr) slog.Handler      { return p }
func (p puitsEnPanne) WithGroup(string) slog.Handler           { return p }

func main() {
	ctx := context.Background()

	// ===== Le même appel, deux handlers (dev vs prod) =====
	text := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{ReplaceAttr: dropTime}))
	jsonMask := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(g []string, a slog.Attr) slog.Attr {
			a = dropTime(g, a)
			if a.Key == "password" { // masquer un secret avant la sortie
				return slog.String("password", "***")
			}
			return a
		}}))
	text.Info("connexion réussie", "user", "alice", "id", 42)
	jsonMask.Info("connexion réussie", "user", "alice", "id", 42, "password", "hunter2")

	// ===== LevelVar : niveau modifiable à l'exécution (rechargement à chaud) =====
	var lv slog.LevelVar // Info par défaut
	dyn := slog.New(slog.NewTextHandler(os.Stdout,
		&slog.HandlerOptions{Level: &lv, ReplaceAttr: dropTime}))
	dyn.Debug("invisible tant que le niveau est Info")
	lv.Set(slog.LevelDebug) // ex. sur SIGHUP en production
	dyn.Debug("visible : niveau passé à Debug à chaud")

	// ===== Groupes + logger contextuel With =====
	reqLog := text.With("request_id", "r-7") // porté sur chaque log suivant
	reqLog.Info("requête traitée",
		slog.Group("request", slog.String("method", "GET"), slog.Int("status", 200)))

	// ===== Enabled court-circuite avant de construire le Record =====
	if !text.Enabled(ctx, slog.LevelDebug) {
		text.Info("Debug est éteint : on évite le coût de construction du log")
	}

	// ===== Go 1.26 : MultiHandler — diffusion « au mieux », erreurs agrégées =====
	var sb strings.Builder
	sain := slog.NewTextHandler(&sb, &slog.HandlerOptions{ReplaceAttr: dropTime})
	multi := slog.NewMultiHandler(sain, puitsEnPanne{})
	rec := slog.Record{Level: slog.LevelInfo, Message: "login"}
	rec.AddAttrs(slog.String("user", "alice"))
	err := multi.Handle(ctx, rec)
	text.Info("MultiHandler : le puits sain a reçu → " + strings.TrimSpace(sb.String()))
	text.Info("MultiHandler : Handle renvoie l'erreur du puits HS (Join)", "err", err.Error())
}
