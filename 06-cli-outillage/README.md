🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 6. Applications CLI et outillage

Aux côtés du backend HTTP, l'**outillage en ligne de commande** (CLI) est l'autre grand domaine où Go s'est imposé. Si vous utilisez déjà `docker`, `kubectl`, `terraform`, `gh` (GitHub CLI), `hugo` ou `helm`, vous manipulez au quotidien des programmes écrits en Go. Ce n'est pas un hasard : les propriétés du langage en font un choix de premier plan pour l'outillage, du petit script utilitaire à la CLI distribuée à des milliers d'utilisateurs.

Ce module part de la **bibliothèque standard** — souvent suffisante — puis introduit les briques qui deviennent utiles à mesure que l'outil grandit : un framework de commandes, une interface texte interactive, et l'automatisation de la distribution multi-plateforme.

## 🎯 Objectifs du module

À l'issue de ce module, vous serez capable de :

- lire arguments, drapeaux et variables d'environnement avec la seule stdlib (`os.Args`, `flag`, `LookupEnv`) — pièges compris ;
- structurer un outil testable (patron `run() error`, `FlagSet`, précédence **drapeau > environnement > défaut**) ;
- bâtir une CLI à **sous-commandes** avec Cobra et une configuration **multi-source** avec Viper — et savoir quand ne pas les dégainer ;
- décider si une **TUI** se justifie, et comprendre l'architecture Elm de Bubble Tea (Model / Update / View) ;
- livrer un **binaire statique** (`CGO_ENABLED=0`), le **cross-compiler** (`GOOS`/`GOARCH`) et **industrialiser la release** avec GoReleaser.

## Pourquoi Go excelle pour les CLI

- **Un binaire autonome.** `go build` produit un exécutable unique, généralement lié statiquement, sans runtime ni dépendances à installer côté utilisateur. On copie le fichier, il s'exécute — rien à voir avec le triptyque interpréteur + dépendances + environnement virtuel.
- **Cross-compilation triviale.** Cibler Linux, macOS ou Windows (et plusieurs architectures) se résume à fixer `GOOS` et `GOARCH` avant le build, sans chaîne de compilation croisée à installer (démonstration au [§ 1.5](../01-introduction-go/05-premier-projet.md), industrialisation au [§ 6.4](04-distribution.md)).
- **Démarrage instantané.** Pas de machine virtuelle ni de phase de préchauffage : le temps de lancement se compte en millisecondes, ce qui compte pour un outil invoqué des centaines de fois par jour ou au cœur d'un script.
- **Une stdlib riche.** `flag`, `os`, `io`, `bufio`, `encoding/json`, `text/tabwriter`… couvrent l'essentiel des besoins d'un outil en ligne de commande sans la moindre dépendance externe.
- **La concurrence intégrée.** Goroutines et `context.Context` ([module 4](../04-concurrence/README.md)) permettent d'écrire simplement des outils qui parallélisent des requêtes, traitent des flux ou honorent proprement un `Ctrl-C`.

C'est cette combinaison — distribution simple, portabilité, démarrage rapide — qui explique la prépondérance de Go dans l'outillage cloud-native et DevOps moderne.

## 🗺️ Plan du module

| Section | Sujet | En bref |
|---------|-------|---------|
| 6.1 | `flag`, `os.Args`, variables d'environnement | La boîte à outils standard : lire arguments, options et configuration d'environnement sans dépendance. |
| 6.2 | Cobra + Viper | Le duo de facto pour les CLI ambitieuses : commandes et sous-commandes, aide et complétion générées, configuration multi-source. |
| 6.3 | TUI avec Bubble Tea *(notions)* | Construire des interfaces texte interactives selon l'architecture Elm ; aperçu de l'écosystème Charm. |
| **6.4** ⭐ | Distribution : binaire unique, cross-compilation, GoReleaser | Industrialiser la livraison : builds multi-plateformes reproductibles et releases automatisées. |

## 🧭 Fil conducteur : la stdlib d'abord, le framework ensuite

Fidèle à l'esprit du langage, ce module applique le principe **« stdlib avant frameworks »**. Le package `flag` de la bibliothèque standard suffit pour une grande part des outils : un ou deux drapeaux, quelques arguments, et c'est réglé sans dépendance.

On adopte **Cobra** lorsque le besoin le justifie réellement : arborescence de sous-commandes façon `git remote add`, aide contextuelle et complétion shell générées, gestion homogène des drapeaux hérités. À l'inverse, envelopper un script de trente lignes dans un framework de commandes est une sur-ingénierie ; savoir *ne pas* le faire fait aussi partie du métier.

**Bubble Tea** répond à un autre besoin — une interface interactive plein écran (barres de progression, listes navigables, formulaires) — et n'a de sens que si l'outil est véritablement interactif. Enfin, la **distribution** ([§ 6.4](04-distribution.md)) capitalise sur la portabilité de Go : c'est là que le « simple binaire » devient un avantage concret face aux autres écosystèmes.

## 📋 Prérequis

Ce module suppose acquis les fondamentaux (modules 1 à 4) : structure d'un programme et packages, fonctions et gestion des erreurs, structs et interfaces, ainsi que les notions de concurrence (`context.Context` pour l'annulation propre d'un traitement). Les manipulations de fichiers et d'E/S (`io`, `bufio`, `os`) sont approfondies au [§ 7.6](../07-acces-donnees/06-fichiers-io.md), mais leurs bases suffisent ici.

## Côté IDE : GoLand et VS Code

Tester une CLI, c'est le plus souvent l'exécuter **avec des arguments et des variables d'environnement**. Les deux IDE le permettent via leurs configurations d'exécution :

- **GoLand** : configuration *Go Build/Run* → champ *Program arguments* pour les drapeaux et arguments, section *Environment* pour les variables ; le débogueur intégré (Delve) s'attache à ces mêmes configurations.
- **VS Code** : fichier `.vscode/launch.json`, clés `"args"` (tableau d'arguments) et `"env"` (variables d'environnement) ; débogage via l'extension Go et `dlv`.

Chaque section IDE de ce module précise, le cas échéant, les réglages propres à chaque environnement.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [6.1 — `flag`, `os.Args`, variables d'environnement](01-flag-args-env.md)

⏭ [`flag`, `os.Args`, variables d'environnement](/06-cli-outillage/01-flag-args-env.md)
