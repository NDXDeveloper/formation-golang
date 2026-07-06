/* ============================================================================
   Section 6.2 : Cobra + Viper (commandes, sous-commandes, configuration)
   Description : L'assemblage complet de la section — commandes par
                 CONSTRUCTEURS (pas de globales ni d'init()), RunE +
                 SilenceUsage, validation cobra.NoArgs, drapeau persistant
                 --config, câblage Viper dans PersistentPreRunE (fichier +
                 env préfixé MONOUTIL_ + BindPFlag) : la précédence
                 drapeau > env > fichier > défaut, observable (cf. README)
   Fichier source : 02-cobra-viper.md
   Essayer :  go run . serve            puis --version, serve --port 9000,
              MONOUTIL_PORT=7000 go run . serve,   go run . sevre (suggestion)
   Tester  :  go test ./...             (exécution in-memory, sans binaire)
   ============================================================================ */

package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		// Cobra a déjà écrit l'erreur sur stderr ; on fixe le code de sortie.
		os.Exit(1)
	}
}

// newRootCmd construit l'arbre de commandes À LA DEMANDE : chaque appel rend
// un arbre neuf, sans état partagé — c'est ce qui rend les tests hermétiques.
func newRootCmd() *cobra.Command {
	v := viper.New() // instance dédiée : jamais le singleton global du package

	root := &cobra.Command{
		Use:          "monoutil",
		Short:        "Un outil d'exemple Cobra + Viper",
		Version:      "1.4.0", // active --version automatiquement
		SilenceUsage: true,    // pas de pavé d'usage à chaque erreur de RunE
		// Exécuté avant TOUTE sous-commande : le bon endroit pour initialiser
		// la configuration (fichier + environnement).
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig(v, cmd)
		},
	}
	// Drapeau PERSISTANT : hérité par toutes les sous-commandes.
	root.PersistentFlags().String("config", "", "chemin d'un fichier de configuration")
	root.AddCommand(newServeCmd(v))
	return root
}

// initConfig câble Viper : fichier (explicite via --config, sinon recherche
// monoutil.yaml) puis variables d'environnement préfixées MONOUTIL_.
func initConfig(v *viper.Viper, cmd *cobra.Command) error {
	if f, _ := cmd.Flags().GetString("config"); f != "" {
		v.SetConfigFile(f) // chemin explicite : pas de recherche
	} else {
		v.SetConfigName("monoutil") // cherche monoutil.yaml…
		v.SetConfigType("yaml")
		v.AddConfigPath(".") // … dans le répertoire courant
	}

	v.SetEnvPrefix("MONOUTIL")                                   // clé « port » ↔ MONOUTIL_PORT
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_")) // « db.host » ↔ MONOUTIL_DB_HOST
	v.AutomaticEnv()                                             // liaison automatique

	if err := v.ReadInConfig(); err != nil {
		// L'ABSENCE de fichier n'est pas une erreur (défauts + env suffisent) ;
		// toute autre panne de lecture (YAML invalide…) remonte, elle.
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return fmt.Errorf("lecture de la configuration : %w", err)
		}
	}
	return nil
}

func newServeCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Démarre le serveur",
		Args:  cobra.NoArgs, // validation déclarative, AVANT RunE
		RunE: func(cmd *cobra.Command, args []string) error {
			// Une seule lecture, toutes sources confondues :
			// --port explicite > MONOUTIL_PORT > monoutil.yaml > défaut 8080.
			port := v.GetInt("port")
			if port < 1 || port > 65535 {
				return fmt.Errorf("port hors plage : %d", port)
			}
			fmt.Printf("écoute sur :%d\n", port)
			return nil
		},
	}
	cmd.Flags().Int("port", 8080, "port d'écoute")
	// Le pont Cobra → Viper : le drapeau ne l'emporte que s'il est FOURNI.
	_ = v.BindPFlag("port", cmd.Flags().Lookup("port"))
	return cmd
}
