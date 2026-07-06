🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 6.2 Cobra + Viper (commandes, sous-commandes, configuration)

Le package `flag` ([§ 6.1](01-flag-args-env.md)) suffit tant qu'un outil se résume à quelques options et un ou deux arguments. Dès qu'apparaissent un **arbre de sous-commandes** (`app serve`, `app config set`…), une **aide et une complétion générées**, ou la nécessité de **fusionner plusieurs sources de configuration** (drapeaux + variables d'environnement + fichier), deux bibliothèques mûres prennent le relais : **Cobra** pour la structure de commandes, **Viper** pour la configuration. C'est le socle de `kubectl`, `hugo`, `helm` ou de la CLI GitHub (`gh`), tous bâtis sur Cobra, la bibliothèque de référence pour les CLI Go modernes.

Fidèle au principe « stdlib avant frameworks », cette section ne les sort pas par défaut : envelopper un outil de trente lignes dans Cobra + Viper est une sur-ingénierie. On les adopte quand la **complexité de l'interface** (sous-commandes, complétion) ou la **richesse de la configuration** (fusion multi-source) le justifie réellement. À l'heure où ces lignes sont écrites (2026), Cobra en est à la série v1.10 et Viper à la série v1.x ; une v2 de Viper est annoncée mais pas encore publiée.

```console
$ go get github.com/spf13/cobra@latest
$ go get github.com/spf13/viper@latest
```

## Cobra : l'arbre de commandes

Cobra structure une application autour de trois notions : les **commandes** (les actions, `app VERBE NOM`), les **arguments** (ce sur quoi elles agissent) et les **drapeaux** (les modificateurs). Les drapeaux sont fournis par `pflag`, qui offre la syntaxe POSIX/GNU (`--port`, forme courte `-p`) et, contrairement au package `flag` de la stdlib, **autorise l'entremêlement** des drapeaux et des arguments positionnels — le piège d'ordre vu au [§ 6.1](01-flag-args-env.md) disparaît.

### Anatomie d'une commande

Une commande est une valeur `cobra.Command`. Le champ **`RunE`** (variante de `Run` qui renvoie une `error`) est l'idiome à privilégier : il ramène la gestion d'erreurs dans le flux habituel plutôt que d'imposer un `os.Exit` au fond du traitement.

```go
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		// Cobra a déjà écrit l'erreur sur stderr ; on fixe le code de sortie.
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:          "monoutil",
		Short:        "Un outil d'exemple",
		Long:         "monoutil illustre Cobra : commandes, sous-commandes et drapeaux.",
		Version:      "1.4.0",
		SilenceUsage: true, // évite le pavé d'usage à chaque erreur renvoyée par RunE
	}
	root.AddCommand(newServeCmd())
	return root
}
```

Deux choix idiomatiques ici. D'abord, des **fonctions constructrices** (`newRootCmd`, `newServeCmd`) plutôt que des variables globales et des `init()` : le graphe de commandes est reconstruit à la demande, sans état partagé — plus lisible et surtout testable (voir plus bas). Le générateur `cobra-cli` produit, lui, le style « globales + `init()` » ; les deux fonctionnent, mais l'approche par constructeurs vieillit mieux. Ensuite, **`SilenceUsage: true`** sur la racine : sans lui, Cobra imprime tout le message d'usage à chaque erreur métier, ce qui noie l'information utile. Le réglage est hérité par les sous-commandes.

### Une sous-commande avec drapeaux et validation d'arguments

```go
func newServeCmd() *cobra.Command {
	var port int
	var verbose bool

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Démarre le serveur",
		Args:  cobra.NoArgs, // refuse tout argument positionnel
		RunE: func(cmd *cobra.Command, args []string) error {
			if port < 1 || port > 65535 {
				return fmt.Errorf("port hors plage : %d", port)
			}
			fmt.Printf("écoute sur :%d (verbose=%t)\n", port, verbose)
			return nil
		},
	}

	cmd.Flags().IntVar(&port, "port", 8080, "port d'écoute")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "sortie détaillée")
	return cmd
}
```

`cmd.Flags()` déclare des drapeaux **locaux** à `serve`. `BoolVarP` ajoute une forme courte (`-v`). Le champ `Args` branche un **validateur** parmi ceux fournis : `cobra.NoArgs`, `cobra.ExactArgs(n)`, `cobra.MinimumNArgs(n)`, `cobra.RangeArgs(min, max)`, `cobra.OnlyValidArgs`, combinables avec `cobra.MatchAll(...)`. La validation se fait ainsi **avant** `RunE`, de façon déclarative.

### Drapeaux persistants vs locaux

Un drapeau déclaré sur `PersistentFlags()` est **hérité par toutes les sous-commandes** ; c'est le bon endroit pour les options transverses (fichier de configuration, verbosité globale).

```go
root.PersistentFlags().String("config", "", "chemin d'un fichier de configuration")
// disponible sur « monoutil », « monoutil serve », etc.
```

### Ce que Cobra offre d'office

Sans code supplémentaire, Cobra fournit : la reconnaissance de `-h` / `--help` à tous les niveaux, un drapeau `--version` dès que le champ `Version` est renseigné, des **suggestions** en cas de faute de frappe (« did you mean… »), et une **complétion shell** générée pour bash, zsh, fish et PowerShell via une commande `completion` ajoutée automatiquement. Ces fonctionnalités — drapeaux POSIX, aide automatique, autocomplétion multi-shell, pages de manuel — font partie du socle de Cobra.

```console
$ monoutil completion zsh > "${fpath[1]}/_monoutil"   # complétion zsh
$ monoutil --version
monoutil version 1.4.0
```

### Échafaudage optionnel : `cobra-cli`

Le générateur historique vit désormais dans un dépôt distinct et s'installe séparément. Il crée un squelette (`init`) puis ajoute des commandes (`add`).

```console
$ go install github.com/spf13/cobra-cli@latest
$ cobra-cli init          # génère main.go + cmd/root.go
$ cobra-cli add serve     # ajoute cmd/serve.go
```

C'est un confort, pas une obligation : le code généré adopte le style « globales + `init()` », et beaucoup d'équipes préfèrent écrire leurs commandes à la main pour garder la maîtrise de la structure.

## Viper : la configuration multi-source

Viper agit comme un **registre unique** de configuration : il agrège plusieurs sources et expose des accesseurs typés (`GetInt`, `GetString`, `GetBool`…). Il lit les formats courants — JSON, TOML, YAML, INI, envfile, entre autres — et cible explicitement les applications 12-factor ([§ 10.3](../10-architecture-services/03-configuration-12factor.md)).

### Ordre de précédence

Viper applique une priorité fixe, du plus fort au plus faible : appel explicite à `Set`, puis **drapeau**, puis **variable d'environnement**, puis **fichier de configuration**, puis magasin clé/valeur distant, puis **valeur par défaut**. C'est exactement la précédence « drapeau > environnement > défaut » esquissée au [§ 6.1](01-flag-args-env.md), étendue à un fichier de configuration au milieu.

### Fichier de configuration

```go
v := viper.New() // instance dédiée plutôt que le singleton global du package

v.SetConfigName("monoutil") // nom de base, sans extension
v.SetConfigType("yaml")     // utile si le fichier n'a pas d'extension explicite
v.AddConfigPath(".")
v.AddConfigPath("$HOME/.config/monoutil")

if err := v.ReadInConfig(); err != nil {
	var notFound viper.ConfigFileNotFoundError
	if !errors.As(err, &notFound) {
		return fmt.Errorf("lecture de la configuration : %w", err)
	}
	// Aucun fichier trouvé : on continue avec défauts + environnement + drapeaux.
}
```

L'absence de fichier n'est pas une erreur fatale : on la distingue via `errors.As` sur `viper.ConfigFileNotFoundError` (l'idiome d'erreurs typées du [§ 2.9](../02-fondamentaux-langage/09-gestion-erreurs.md)) pour ne remonter que les vraies pannes de lecture.

### Variables d'environnement

```go
v.SetEnvPrefix("MONOUTIL")                                   // clé « port » → MONOUTIL_PORT
v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_")) // « db.host » → MONOUTIL_DB_HOST
v.AutomaticEnv()                                             // liaison automatique clé ↔ env
```

Après `AutomaticEnv`, tout `v.GetString("db.host")` consulte `MONOUTIL_DB_HOST` sans liaison manuelle.

### Lier les drapeaux Cobra à Viper

Le pont entre les deux mondes est `BindPFlag`, qui associe une clé Viper à un drapeau `pflag` de Cobra. On peut lier un drapeau à la fois, ou tout un `FlagSet`.

```go
_ = v.BindPFlag("port", cmd.Flags().Lookup("port"))
// ou, pour tout lier d'un coup :
_ = v.BindPFlags(cmd.Flags())
```

Nuance importante : un drapeau lié n'écrase l'environnement et le fichier **que s'il a été explicitement fourni** sur la ligne de commande. Non renseigné, sa valeur par défaut reste la source de priorité la plus basse. On obtient donc en pratique : drapeau explicite > environnement > fichier > défaut du drapeau.

### Vers une configuration typée

Plutôt que de disséminer des `v.GetString(...)` dans le code, on **dé-sérialise** la configuration dans une structure via `Unmarshal` (balises `mapstructure`).

```go
type Config struct {
	Port    int    `mapstructure:"port"`
	Verbose bool   `mapstructure:"verbose"`
	DB      DBConf `mapstructure:"db"`
}
type DBConf struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

var cfg Config
if err := v.Unmarshal(&cfg); err != nil {
	return fmt.Errorf("configuration invalide : %w", err)
}
```

### Points de vigilance

- **Singleton global.** Les fonctions de paquet (`viper.SetConfigName`, `viper.GetInt`…) opèrent sur une instance globale unique. Pour une bibliothèque, un test isolé ou plusieurs configurations distinctes, préférez une instance créée par `viper.New()`.
- **Clés insensibles à la casse.** `Port` et `port` désignent la même clé ; à garder en tête pour éviter les surprises.
- **Pas de sûreté en concurrence.** Des lectures et écritures concurrentes sur un même Viper peuvent provoquer un panic : synchronisez l'accès (paquet `sync`) si nécessaire.
- **Poids.** Viper tire de nombreuses dépendances transitives. Pour un besoin modeste (quelques drapeaux + variables d'environnement), la stdlib du [§ 6.1](01-flag-args-env.md) reste plus légère — un choix parfaitement idiomatique.

## Assembler Cobra + Viper

Le câblage se fait proprement dans un `PersistentPreRunE` sur la racine, exécuté avant toute sous-commande : il initialise Viper (fichier + environnement), puis chaque commande lie ses drapeaux et lit ses valeurs via l'instance partagée.

```go
func newRootCmd() *cobra.Command {
	v := viper.New()

	root := &cobra.Command{
		Use:          "monoutil",
		Version:      "1.4.0",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig(v, cmd)
		},
	}
	root.PersistentFlags().String("config", "", "chemin d'un fichier de configuration")
	root.AddCommand(newServeCmd(v))
	return root
}

func initConfig(v *viper.Viper, cmd *cobra.Command) error {
	if f, _ := cmd.Flags().GetString("config"); f != "" {
		v.SetConfigFile(f) // chemin explicite fourni par --config
	} else {
		v.SetConfigName("monoutil")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.config/monoutil")
	}

	v.SetEnvPrefix("MONOUTIL")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
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
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			port := v.GetInt("port") // défaut < fichier < env < drapeau explicite
			if port < 1 || port > 65535 {
				return fmt.Errorf("port hors plage : %d", port)
			}
			fmt.Printf("écoute sur :%d\n", port)
			return nil
		},
	}
	cmd.Flags().Int("port", 8080, "port d'écoute")
	_ = v.BindPFlag("port", cmd.Flags().Lookup("port"))
	return cmd
}
```

Avec ce montage, `monoutil serve` lit `port` dans cet ordre : `--port` s'il est passé, sinon `MONOUTIL_PORT`, sinon la clé `port` du fichier `monoutil.yaml`, sinon `8080`.

## Tester une commande

Le style par constructeurs porte ici ses fruits. Cobra permet d'exécuter une commande en mémoire : on injecte les arguments avec `SetArgs`, on redirige les sorties avec `SetOut` / `SetErr`, puis on appelle `Execute`. Comme chaque test reconstruit un arbre neuf (`newRootCmd()`), aucun état global ne fuit d'un test à l'autre — un atout que les variables globales rendraient fragile (voir le [module 13](../13-tests-qualite/README.md)).

```go
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
```

## Côté IDE : GoLand et VS Code

Les mécaniques de base (passer arguments et variables d'environnement) sont celles du [§ 6.1](01-flag-args-env.md) ; l'élément nouveau ici est le **débogage d'une sous-commande précise** et la découverte du fichier de configuration selon le répertoire de travail.

**GoLand** — configuration *Go Build* : placez la sous-commande et ses drapeaux dans *Program arguments*, p. ex. `serve --port 9000` ; renseignez *Environment* (`MONOUTIL_PORT=9000`) et surtout *Working directory* sur la racine du projet, sinon `AddConfigPath(".")` ne trouvera pas `monoutil.yaml`. Un point d'arrêt dans le `RunE` de `serve` est atteint directement via Delve. Astuce : dupliquez la configuration (une par sous-commande fréquemment déboguée) pour basculer d'un clic.

**VS Code** — `.vscode/launch.json` (extension Go + `dlv`) : la sous-commande vit dans le tableau `args`, la configuration d'environnement dans `env` ou un `envFile`.

```jsonc
{
  "name": "monoutil serve",
  "type": "go",
  "request": "launch",
  "mode": "debug",
  "program": "${workspaceFolder}",
  "cwd": "${workspaceFolder}",
  "args": ["serve", "--port", "9000"],
  "env": { "MONOUTIL_PORT": "9000" }
}
```

La clé `cwd` fixe le répertoire de travail (même rôle que le *Working directory* de GoLand). En ligne de commande, l'équivalent reste `go run . serve --port 9000` — les arguments de la sous-commande se placent après le `.`.

## En résumé

- Cobra structure l'application en commandes / sous-commandes ; `RunE` (plutôt que `Run`) ramène les erreurs dans le flux normal, et `SilenceUsage: true` évite le pavé d'usage à chaque erreur.
- `Flags()` = drapeaux locaux, `PersistentFlags()` = hérités ; les validateurs `cobra.*Args` valident en amont de `RunE`. Aide, `--version`, suggestions et complétion shell sont fournis d'office (via `pflag`, qui autorise l'entremêlement drapeaux/arguments).
- Viper agrège fichier + environnement + drapeaux avec une précédence fixe (`Set` > drapeau > env > fichier > défaut) ; `BindPFlag` fait le pont, un drapeau ne l'emportant que s'il est explicitement fourni.
- Préférez `viper.New()` au singleton global, `Unmarshal` vers une structure typée, et gardez en tête l'insensibilité à la casse et l'absence de sûreté en concurrence.
- Ne dégainez Cobra + Viper que lorsque la complexité le justifie : pour un petit outil, la stdlib du [§ 6.1](01-flag-args-env.md) est plus légère et tout aussi idiomatique.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [6.3 — TUI avec Bubble Tea (notions)](03-tui-bubbletea.md)

⏭ [TUI avec Bubble Tea (notions)](/06-cli-outillage/03-tui-bubbletea.md)
