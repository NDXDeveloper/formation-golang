🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 6.1 `flag`, `os.Args`, variables d'environnement

Avant tout framework, Go fournit dans sa bibliothèque standard tout le nécessaire pour lire ce que l'utilisateur passe à un programme : les **arguments bruts** (`os.Args`), les **drapeaux/options** (`flag`) et la **configuration d'environnement** (`os.Getenv` / `os.LookupEnv`). Pour une grande part des outils — un ou deux drapeaux, quelques arguments — cette trousse suffit, sans la moindre dépendance. On adopte un framework (Cobra, [§ 6.2](02-cobra-viper.md)) seulement quand l'outil grossit ; savoir se contenter de la stdlib est ici la démarche idiomatique.

## `os.Args` : les arguments bruts

`os.Args` est un `[]string` disponible dès le lancement. La première case est le nom (chemin) du programme ; les suivantes sont les arguments transmis.

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("programme :", os.Args[0])
	fmt.Println("arguments :", os.Args[1:])
}
```

```console
$ go run . un deux trois
programme : /tmp/go-build.../exe/main   # avec « go run », un binaire temporaire
arguments : [un deux trois]
```

`os.Args` contient toujours au moins un élément (`os.Args[0]`). On l'utilise directement pour des besoins minimalistes ou de la logique très spécifique, mais dès qu'il y a des **options** (`-verbose`, `--port=…`), le package `flag` évite de réécrire à la main l'analyse de la ligne de commande.

## Le package `flag` : options et drapeaux

### Définir et lire des drapeaux

Chaque fonction `flag.Xxx(nom, défaut, usage)` déclare un drapeau et renvoie un **pointeur** vers sa valeur. On appelle `flag.Parse()` **une seule fois**, après toutes les déclarations et avant toute lecture ; les valeurs sont alors renseignées.

```go
package main

import (
	"flag"
	"fmt"
	"time"
)

func main() {
	port := flag.Int("port", 8080, "port d'écoute")
	verbose := flag.Bool("verbose", false, "sortie détaillée")
	timeout := flag.Duration("timeout", 5*time.Second, "délai maximal")

	flag.Parse()

	fmt.Printf("port=%d verbose=%t timeout=%s\n", *port, *verbose, *timeout)
	fmt.Println("arguments positionnels :", flag.Args())
}
```

```console
$ go run . -port 9000 -verbose -timeout 2s fichier.txt
port=9000 verbose=true timeout=2s
arguments positionnels : [fichier.txt]
```

`flag.Duration` (via `time.ParseDuration`) et les autres types typés (`Int`, `Int64`, `Uint`, `Float64`, `String`, `Bool`) montrent l'intérêt du package : la **conversion et la validation de base** sont gratuites. Une variante `flag.XxxVar(&cible, …)` lie le drapeau à une variable existante — pratique pour remplir directement une structure de configuration :

```go
type config struct {
	Port    int
	Verbose bool
}

var cfg config
flag.IntVar(&cfg.Port, "port", 8080, "port d'écoute")
flag.BoolVar(&cfg.Verbose, "verbose", false, "sortie détaillée")
```

### Arguments positionnels et un piège classique

Après analyse, `flag.Args()` renvoie les arguments **non-drapeaux** (`flag.Arg(i)`, `flag.NArg()` pour l'accès indexé). Point crucial et souvent surprenant : le package `flag` **s'arrête au premier argument qui n'est pas un drapeau** et, contrairement à `getopt` GNU, il **ne réordonne pas** la ligne de commande. Les drapeaux doivent donc précéder les arguments positionnels.

```console
# drapeaux AVANT les positionnels : OK
$ mytool -port 9000 fichier.txt
# → flag.Args() == ["fichier.txt"]

# positionnel AVANT les drapeaux : -port N'EST PAS interprété
$ mytool fichier.txt -port 9000
# → flag.Args() == ["fichier.txt", "-port", "9000"]
```

### Syntaxe des drapeaux (et le cas des booléens)

Les formes `-flag`, `--flag`, `-flag=valeur` et `-flag valeur` (pour les non-booléens) sont équivalentes. Les **booléens** obéissent à une règle à part : ils s'activent par simple présence, et pour les mettre à `false` il faut la forme collée `=` — la forme séparée est interprétée comme un argument positionnel.

```console
-verbose            # true
-verbose=false      # false
-verbose false      # PIÈGE : "false" devient un argument positionnel, pas la valeur du drapeau
```

### Aide et message d'usage

Les drapeaux `-h` / `-help` sont fournis automatiquement (sauf si vous définissez vous-même un `-h`) : ils affichent l'usage généré à partir des descriptions. On personnalise l'en-tête via `flag.Usage`, et `flag.PrintDefaults()` liste les drapeaux déclarés. Par convention Unix, ce texte part sur la **sortie d'erreur** — c'est la destination par défaut de `flag` (`flag.CommandLine.Output()`).

```go
flag.Usage = func() {
	fmt.Fprintf(flag.CommandLine.Output(),
		"Usage : %s [options] <fichier>\n\nOptions :\n", os.Args[0])
	flag.PrintDefaults()
}
```

### Types personnalisés : `flag.Value`, `flag.Func`, `flag.TextVar`

Pour un drapeau **répétable** ou d'un type non couvert, `flag.Func` (Go 1.16) est la voie la plus concise : on fournit une fonction de parsing qui reçoit chaque occurrence et renvoie une `error` en cas d'entrée invalide.

```go
var headers []string
flag.Func("H", "en-tête « clé: valeur » (répétable)", func(s string) error {
	if !strings.Contains(s, ":") {
		return fmt.Errorf("format invalide : %q", s)
	}
	headers = append(headers, s)
	return nil
})
// $ mytool -H "Accept: application/json" -H "Authorization: Bearer x"
```

`flag.TextVar` (Go 1.19) branche directement tout type qui implémente `encoding.TextUnmarshaler` — par exemple une adresse `net/netip` :

```go
addr := netip.MustParseAddr("127.0.0.1")
flag.TextVar(&addr, "addr", addr, "adresse d'écoute")
// $ mytool -addr ::1   → parsing et validation assurés par netip.Addr
```

Pour une option réutilisée dans plusieurs outils, on implémente l'interface `flag.Value` (méthodes `String() string` et `Set(string) error`) sur un type dédié, puis on l'enregistre avec `flag.Var`. (`flag.BoolFunc`, Go 1.21, joue le même rôle que `flag.Func` pour les drapeaux booléens à effet de bord.)

### Erreurs de parsing et sous-commandes : les `FlagSet`

`flag.CommandLine` — le `FlagSet` implicite manipulé par `flag.Int`, `flag.Parse`, etc. — est configuré en `ExitOnError` : sur une erreur d'analyse il appelle `os.Exit(2)`, et sur `-help` il quitte avec le code `0`. Pour **garder la main** (gérer l'erreur soi-même, écrire un test), on crée un `FlagSet` en `ContinueOnError` :

```go
fs := flag.NewFlagSet("serve", flag.ContinueOnError)
port := fs.Int("port", 8080, "port d'écoute")
if err := fs.Parse(args); err != nil {
	// err == flag.ErrHelp si -h/-help a été demandé
	return err
}
_ = port
```

Un `FlagSet` par sous-commande est aussi la façon **manuelle** de bâtir une CLI façon `git commit` / `git push` : on regarde le premier argument, puis on parse le reste avec le `FlagSet` correspondant. C'est exactement ce que Cobra automatise (arbre de commandes, aide et complétion générées) — voir [§ 6.2](02-cobra-viper.md).

## Variables d'environnement

### Lire : `Getenv` vs `LookupEnv`

`os.Getenv` renvoie `""` si la variable est absente **ou** définie vide — les deux cas sont indiscernables. Quand la distinction compte, `os.LookupEnv` renvoie un second retour booléen.

```go
level := os.Getenv("LOG_LEVEL") // "" si absente OU vide

if v, ok := os.LookupEnv("LOG_LEVEL"); ok {
	level = v // on sait ici que la variable est *définie* (même à "")
}
```

Le reste de l'API complète le tableau : `os.Setenv` / `os.Unsetenv` pour modifier l'environnement du processus, `os.Environ()` pour l'obtenir en entier (tranche de `"CLÉ=VALEUR"`). Un petit assistant rend la lecture avec valeur par défaut lisible :

```go
func envOr(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}
```

### Précédence : drapeau > environnement > défaut

La convention (proche des principes 12-factor, [§ 10.3](../10-architecture-services/03-configuration-12factor.md)) veut qu'un **drapeau explicite l'emporte sur l'environnement**, lui-même prioritaire sur la valeur par défaut codée en dur. L'approche la plus simple consiste à alimenter le *défaut* du drapeau depuis l'environnement : un drapeau réellement passé écrase alors cette valeur.

```go
func envIntOr(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		} // valeur invalide : on retombe sur le défaut (ou on remonte une erreur)
	}
	return def
}

port := flag.Int("port", envIntOr("PORT", 8080), "port d'écoute")
// défaut 8080  <  $PORT  <  -port (drapeau explicite)
```

Pour distinguer précisément « drapeau réellement fourni » de « valeur par défaut », `flag.Visit` ne parcourt que les drapeaux effectivement passés (là où `flag.VisitAll` les parcourt tous) :

```go
set := map[string]bool{}
flag.Visit(func(f *flag.Flag) { set[f.Name] = true })
// set["port"] == true uniquement si -port a été passé sur la ligne de commande
```

## Assembler : une configuration idiomatique

### Le patron `run() error` et les codes de sortie

`os.Exit` **n'exécute pas les `defer`** (fermeture de fichiers, `flush` de journaux…). L'idiome consiste donc à isoler toute la logique dans une fonction `run() error` et à n'appeler `os.Exit` que dans `main`, une fois les `defer` déroulés. Par convention, un programme rend `0` en cas de succès et un code non nul en cas d'échec (`flag` utilise `2` pour un mauvais usage).

```go
package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

type config struct {
	Port    int
	Timeout time.Duration
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "erreur :", err) // diagnostics → stderr
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("mytool", flag.ContinueOnError)

	var cfg config
	fs.IntVar(&cfg.Port, "port", envIntOr("PORT", 8080), "port d'écoute")
	fs.DurationVar(&cfg.Timeout, "timeout", 5*time.Second, "délai maximal")

	if err := fs.Parse(args); err != nil {
		return err // ErrHelp inclus : -h/-help gérés proprement
	}

	// ... logique réelle ; les defer s'exécutent normalement à la sortie de run
	fmt.Println("configuration :", cfg) // résultats → stdout
	return nil
}
```

Ce squelette combine les trois briques : arguments (`os.Args[1:]` passés à `run`), drapeaux (le `FlagSet`) et environnement (`envIntOr` en défaut). Il reste testable — on appelle `run([]string{"-port", "0"})` dans un test — précisément parce qu'il n'appelle pas `os.Exit`. Et quand l'outil exécute un travail long, on étend le squelette en `run(ctx, args)` avec `signal.NotifyContext` ([§ 4.4](../04-concurrence/04-context.md)) : un `Ctrl-C` annule alors le contexte et l'outil s'interrompt proprement — le même motif que l'arrêt d'un serveur ([§ 5.1](../05-backend-http/01-net-http.md)).

### Entrées/sorties standard

Trois flux sont toujours disponibles : `os.Stdin`, `os.Stdout`, `os.Stderr`. La règle Unix, respectée par les exemples ci-dessus, sépare **résultats exploitables** (vers `stdout`, donc redirigeables et « pipe-ables ») et **diagnostics/usage/erreurs** (vers `stderr`). Écrire les messages d'erreur sur `stdout` casserait un `mytool | autre-commande` ; `fmt.Fprintln(os.Stderr, …)` est la bonne cible.

## Côté IDE : GoLand et VS Code

Tester une CLI, c'est l'exécuter avec des arguments **et** des variables d'environnement. Les deux IDE le permettent sans quitter le débogueur.

**GoLand** — *Run | Edit Configurations… | + | Go Build* :

- *Program arguments* : la ligne de drapeaux et positionnels, p. ex. `-port 9000 -verbose fichier.txt` ;
- *Environment* : les variables, séparées par `;`, p. ex. `LOG_LEVEL=debug;PORT=9000` ;
- *Working directory* : la racine du module (utile si l'outil lit des fichiers relatifs) ;
- le débogueur intégré (Delve) réutilise cette configuration — les points d'arrêt s'appliquent tels quels.

**VS Code** — fichier `.vscode/launch.json` (extension Go + `dlv`) :

```jsonc
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Lancer mytool",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}",
      "args": ["-port", "9000", "-verbose", "fichier.txt"],
      "env": { "LOG_LEVEL": "debug", "PORT": "9000" },
      "envFile": "${workspaceFolder}/.env"
    }
  ]
}
```

La clé `args` est un **tableau** d'arguments, `env` un dictionnaire de variables, et `envFile` charge en plus un fichier `.env` — commode pour ne pas versionner les valeurs locales. En ligne de commande, on obtient l'équivalent avec `go run . -port 9000 fichier.txt` : attention, les drapeaux **du programme** se placent *après* le `.` ; dans `go run -race . -port 9000`, `-race` s'adresse à `go`, pas à l'outil.

## En résumé

- `os.Args` donne les arguments bruts ; `flag` gère options, conversions typées et usage généré, sans dépendance.
- `flag.Parse()` une fois, après les déclarations ; les drapeaux doivent **précéder** les positionnels, et `-bool false` est un piège (utiliser `-bool=false`).
- `flag.Func` / `flag.TextVar` couvrent les drapeaux répétables ou typés ; un `FlagSet` en `ContinueOnError` rend le parsing testable et prépare les sous-commandes ([§ 6.2](02-cobra-viper.md)).
- `LookupEnv` distingue absente et vide ; précédence idiomatique **drapeau > environnement > défaut**.
- Isoler la logique dans `run() error` pour que les `defer` s'exécutent (et `run(ctx, args)` + `signal.NotifyContext` pour honorer `Ctrl-C`) ; résultats sur `stdout`, diagnostics sur `stderr`.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [6.2 — Cobra + Viper (commandes, sous-commandes, configuration)](02-cobra-viper.md)

⏭ [Cobra + Viper (commandes, sous-commandes, configuration)](/06-cli-outillage/02-cobra-viper.md)
