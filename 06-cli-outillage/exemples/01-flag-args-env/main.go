/* ============================================================================
   Section 6.1 : flag, os.Args, variables d'environnement
   Description : Le squelette idiomatique complet d'un outil stdlib — os.Args,
                 drapeaux typés (Int, Bool, Duration), drapeau répétable
                 (flag.Func), type auto-validé (flag.TextVar + netip),
                 précédence drapeau > environnement > défaut (envIntOr),
                 patron run(ctx, args) error avec signal.NotifyContext
                 (Ctrl-C propre), résultats sur stdout / diagnostics sur stderr
   Fichier source : 01-flag-args-env.md
   Essayer :  go run . -port 9000 -verbose -timeout 2s fichier.txt
              PORT=9999 go run .          (env < drapeau : cf. README)
              go run . -H "Accept: json" -H "X: y" -addr ::1
   ============================================================================ */

package main

import (
	"context"
	"flag"
	"fmt"
	"net/netip"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// La configuration de l'outil : remplie par les drapeaux (XxxVar).
type config struct {
	Port    int
	Verbose bool
	Timeout time.Duration
	Addr    netip.Addr // validée par netip via flag.TextVar
	Headers []string   // drapeau -H répétable via flag.Func
}

// envIntOr lit un entier dans l'environnement, sinon renvoie le défaut.
// Utilisée comme DÉFAUT du drapeau : un -port explicite l'écrase —
// c'est la précédence idiomatique drapeau > environnement > défaut.
func envIntOr(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok { // LookupEnv : distingue absente de vide
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func main() {
	// Ctrl-C (SIGINT) ou SIGTERM annulent ctx : l'outil s'interrompt proprement.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Toute la logique vit dans run() : os.Exit n'exécute PAS les defer,
	// donc on ne l'appelle qu'ici, une fois run() (et ses defer) terminés.
	if err := run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "erreur :", err) // diagnostics → stderr
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	// Un FlagSet en ContinueOnError : les erreurs de parsing nous reviennent
	// (testable), au lieu du os.Exit(2) du FlagSet global ExitOnError.
	fs := flag.NewFlagSet("mytool", flag.ContinueOnError)

	var cfg config
	fs.IntVar(&cfg.Port, "port", envIntOr("PORT", 8080), "port d'écoute") // défaut < $PORT < -port
	fs.BoolVar(&cfg.Verbose, "verbose", false, "sortie détaillée")
	fs.DurationVar(&cfg.Timeout, "timeout", 5*time.Second, "délai maximal") // "2s", "1m30s"…

	// Drapeau RÉPÉTABLE : chaque -H passe par cette fonction de parsing,
	// qui peut refuser une occurrence en renvoyant une erreur.
	fs.Func("H", "en-tête « clé: valeur » (répétable)", func(s string) error {
		if !strings.Contains(s, ":") {
			return fmt.Errorf("format invalide : %q", s)
		}
		cfg.Headers = append(cfg.Headers, s)
		return nil
	})

	// TextVar : tout type qui sait se dé-sérialiser depuis du texte
	// (encoding.TextUnmarshaler) devient un drapeau auto-validé.
	cfg.Addr = netip.MustParseAddr("127.0.0.1")
	fs.TextVar(&cfg.Addr, "addr", cfg.Addr, "adresse d'écoute")

	// Un usage personnalisé — sur la sortie d'erreur, convention Unix.
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage : %s [options] <fichier...>\n\nOptions :\n", fs.Name())
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err // inclut flag.ErrHelp : -h/-help remontent proprement
	}

	// Après Parse : fs.Args() = les positionnels (APRÈS les drapeaux — le
	// package flag s'arrête au premier argument qui n'est pas un drapeau).
	fmt.Printf("port=%d verbose=%t timeout=%s addr=%s\n", cfg.Port, cfg.Verbose, cfg.Timeout, cfg.Addr)
	if len(cfg.Headers) > 0 {
		fmt.Println("en-têtes :", cfg.Headers)
	}
	fmt.Println("arguments positionnels :", fs.Args())

	// Un travail interruptible : select entre la fin du minuteur et Ctrl-C.
	if cfg.Verbose {
		select {
		case <-time.After(50 * time.Millisecond): // le « travail »
			fmt.Println("travail terminé")
		case <-ctx.Done(): // Ctrl-C pendant le travail
			return ctx.Err()
		}
	}
	return nil
}
