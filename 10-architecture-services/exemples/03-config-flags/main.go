/* ============================================================================
   Section 10.3 : Configuration, feature flags, principes 12-factor
   Description : La section 10.3 en action, en quatre démonstrations :
                 (1) la PRÉCÉDENCE défauts < env < flags, prouvée valeur par
                 valeur, et le FAIL FAST (sans DATABASE_URL, Load échoue
                 avant d'ouvrir la moindre connexion) ; (2) godotenv — Load()
                 n'écrase JAMAIS une variable déjà définie par la plateforme,
                 Overload() force (le .env reste un artefact de dev) ;
                 (3) les FEATURE FLAGS derrière un petit PORT : le domaine
                 évalue des drapeaux sans connaître leur source (statique en
                 dev, plateforme en prod — même interface) ; (4) le
                 rechargement À CHAUD via SIGHUP (niveau de log), pour les
                 seuls réglages qui le justifient — statique par défaut.
   Fichier source : 03-configuration-12factor.md
   Lancer : go run .        (aucun service requis)
   ============================================================================ */

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/exemple/configflags/config"
)

// Flags est un PORT : le domaine évalue des drapeaux sans connaître leur source.
type Flags interface {
	Enabled(ctx context.Context, name string) bool
}

// staticFlags : l'implémentation la plus simple, alimentée par la configuration.
type staticFlags map[string]bool

func (s staticFlags) Enabled(_ context.Context, name string) bool { return s[name] }

// checkout illustre le domaine qui bascule sur un drapeau, sans rien
// savoir du fournisseur en dessous.
func checkout(ctx context.Context, flags Flags) string {
	if flags.Enabled(ctx, "new-checkout-flow") {
		return "checkout V2 (nouveau parcours)"
	}
	return "checkout V1 (parcours actuel)"
}

func main() {
	ctx := context.Background()

	fmt.Println("=== 1. Précédence défauts < env < flags, et fail fast ===")
	os.Setenv("DATABASE_URL", "postgres://localhost/app")    // fournie par la plateforme
	os.Setenv("APP_ADDR", ":9000")                           // l'env bat le défaut :8080
	cfg, err := config.Load([]string{"-log-level", "debug"}) // le flag bat tout
	fmt.Printf("addr=%s (env > défaut) · log=%s (flag > tout) · timeout=%s (défaut) · err=%v\n",
		cfg.Addr, cfg.LogLevel, cfg.Timeout, err)

	os.Unsetenv("DATABASE_URL")
	_, err = config.Load(nil)
	fmt.Println("sans DATABASE_URL → fail fast :", err)

	fmt.Println()
	fmt.Println("=== 2. godotenv : Load n'écrase pas, Overload force ===")
	os.Setenv("APP_MODE", "défini-par-la-plateforme")
	_ = godotenv.Load(".env.exemple")
	fmt.Println("après Load     :", os.Getenv("APP_MODE"), "(la plateforme a gagné)")
	_ = godotenv.Overload(".env.exemple")
	fmt.Println("après Overload :", os.Getenv("APP_MODE"), "(le .env a forcé)")

	fmt.Println()
	fmt.Println("=== 3. Feature flags derrière un port ===")
	// dev : drapeaux statiques ; prod : un fournisseur (OpenFeature, Unleash…)
	// satisferait le MÊME port — le domaine ne change pas.
	flags := staticFlags{"new-checkout-flow": false}
	fmt.Println("flag à false →", checkout(ctx, flags))
	flags["new-checkout-flow"] = true // le kill switch qu'on bascule
	fmt.Println("flag à true  →", checkout(ctx, flags))

	fmt.Println()
	fmt.Println("=== 4. Rechargement à chaud : SIGHUP (statique par défaut, dynamique si justifié) ===")
	var level atomic.Value
	level.Store("info") // le niveau nominal de production (celui du défaut)
	hup := make(chan os.Signal, 1)
	signal.Notify(hup, syscall.SIGHUP)
	go func() {
		for range hup {
			level.Store("debug") // en vrai : relire la source de configuration
		}
	}()
	fmt.Println("niveau de log :", level.Load())
	_ = syscall.Kill(os.Getpid(), syscall.SIGHUP) // l'exploitant taperait : kill -HUP <pid>
	time.Sleep(150 * time.Millisecond)
	fmt.Println("niveau de log :", level.Load(), "(changé à chaud, sans redémarrage)")
}
