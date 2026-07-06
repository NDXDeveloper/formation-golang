🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 10.3 Configuration, feature flags, principes 12-factor

Un même binaire doit tourner en local, en préproduction et en production **sans être recompilé**. C'est la promesse centrale de la méthodologie *12-factor*, et le troisième volet de ce module : après avoir décidé où tracer les frontières ([§ 10.1](01-monolithe-vs-microservices.md)) et comment organiser l'intérieur ([§ 10.2](02-clean-architecture.md)), on adapte le service au monde réel — par la **configuration** (au démarrage) et par les **feature flags** (à l'exécution).

Le fil directeur reste le même qu'au [§ 10.2](02-clean-architecture.md) : commencer simple, et n'ajouter une couche que lorsqu'elle gagne sa place. La configuration comme les flags peuvent virer à l'usine à gaz ; l'enjeu est de doser.

## Le cadre : la méthodologie 12-factor

La méthodologie [*The Twelve-Factor App*](https://12factor.net) (Adam Wiggins, Heroku, 2011) énonce douze principes pour bâtir des services portables et exploitables. Son facteur le plus structurant pour nous est le **facteur III — Config** : la configuration doit être **strictement séparée du code** et vivre dans **l'environnement**. Le test décisif : pourriez-vous rendre le dépôt public à l'instant, sans divulguer le moindre secret ? Si la réponse est non, de la configuration est codée en dur là où elle ne devrait pas l'être.

Quelques autres facteurs éclairent l'architecture d'un service Go, et sont traités ailleurs dans la formation :

- **IV — Services externes** : base, cache, file sont des ressources *attachées*, interchangeables via la configuration (une URL). Passer de PostgreSQL local à une base gérée ne doit être qu'un changement de `DATABASE_URL`.
- **V — Build, release, run** : trois étapes séparées, versionnées — l'affaire du [module 15](../15-deploiement-devops/01-build-versioning.md) (build reproductible, estampillage par `ldflags`).
- **VI / IX — Processus sans état et jetables** : démarrage rapide, arrêt propre (*graceful shutdown*) — un point fort de Go, détaillé en [§ 9.2](../09-conteneurs-cloud/02-kubernetes.md).
- **XI — Logs comme flux d'événements** : écrire sur la sortie standard, ne pas gérer de fichiers de log — ce que fait naturellement `log/slog` ([§ 12.3](../12-erreurs-debogage/03-slog.md)).

Go incarne 12-factor presque sans effort : binaire statique unique (build/release/run nets), processus sans état, démarrage en millisecondes, logs vers `stdout`. Une nuance s'impose toutefois : 12-factor date de 2011, et l'ère des conteneurs en a affiné certains points. Dans un cluster Kubernetes, « la config dans l'environnement » se réalise par des *ConfigMaps* (et des *Secrets*) injectés en variables d'environnement — l'esprit du facteur III, avec l'outillage de 2026 (cf. [§ 9.2](../09-conteneurs-cloud/02-kubernetes.md)). On retient les principes, pas la lettre.

## La configuration en Go : la bibliothèque standard d'abord

L'idiome Go tient en quatre gestes : une structure `Config` **typée**, chargée **une seule fois** au démarrage, **validée** immédiatement (*fail fast*), puis **passée explicitement** aux composants — jamais une variable globale mutable (dans la droite ligne de l'injection de dépendances vue en [§ 10.2](02-clean-architecture.md)).

Pour la grande majorité des services, la bibliothèque standard suffit : `flag` pour les arguments, `os.LookupEnv` pour l'environnement, `strconv`/`time.ParseDuration` pour les conversions. Le conseil qui fait consensus en 2026 est simple : partir de la stdlib avec un chargeur validant **écrit à la main**, et ne tirer une bibliothèque que lorsqu'une de ses fonctionnalités précises le justifie.

Deux détails de la stdlib méritent l'attention : `os.Getenv("KEY")` renvoie une chaîne vide pour une variable absente, tandis que `os.LookupEnv("KEY")` renvoie `(string, bool)` et distingue donc « non défini » de « défini à vide » ; et toute valeur d'environnement est une **chaîne** au niveau de l'OS — la conversion vers `int`, `bool` ou `time.Duration` vous incombe.

```go
// internal/platform/config/config.go
package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"
)

// Config est typée, chargée une fois, validée au démarrage, puis passée explicitement.
type Config struct {
	Addr        string
	DatabaseURL string        // secret : jamais journalisé (cf. 16.2)
	LogLevel    string
	Timeout     time.Duration
}

// Load applique la précédence : défauts < variables d'environnement < flags.
func Load(args []string) (Config, error) {
	c := Config{ // 1) valeurs par défaut
		Addr:     ":8080",
		LogLevel: "info",
		Timeout:  5 * time.Second,
	}

	// 2) environnement (12-factor : la config vit dans l'environnement)
	if v, ok := os.LookupEnv("APP_ADDR"); ok {
		c.Addr = v
	}
	c.DatabaseURL = os.Getenv("DATABASE_URL")
	if v, ok := os.LookupEnv("LOG_LEVEL"); ok {
		c.LogLevel = v
	}

	// 3) flags (priorité la plus haute) sur un FlagSet LOCAL — pas d'état global.
	// ContinueOnError renvoie l'erreur au lieu d'appeler os.Exit : composable et testable.
	fs := flag.NewFlagSet("app", flag.ContinueOnError)
	fs.StringVar(&c.Addr, "addr", c.Addr, "adresse d'écoute HTTP")
	fs.StringVar(&c.LogLevel, "log-level", c.LogLevel, "niveau de log")
	fs.DurationVar(&c.Timeout, "timeout", c.Timeout, "timeout des requêtes")
	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	if err := c.validate(); err != nil { // 4) fail fast
		return Config{}, fmt.Errorf("configuration invalide : %w", err)
	}
	return c, nil
}

func (c Config) validate() error {
	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL est requis")
	}
	if c.Timeout <= 0 {
		return errors.New("timeout doit être positif")
	}
	return nil
}
```

Le `main` charge la configuration et **échoue immédiatement** si elle est invalide, avant même d'ouvrir la moindre connexion — c'est tout l'intérêt du *fail fast* :

```go
cfg, err := config.Load(os.Args[1:])
if err != nil {
	log.Fatalf("démarrage impossible : %v", err)
}
```

Deux points de vigilance idiomatiques. D'abord, `os.Setenv` **n'est pas sûr en concurrence** (il enveloppe le `setenv` de C) : c'est pourquoi `t.Setenv` refuse de s'exécuter dans un test parallèle (cf. [module 13](../13-tests-qualite/01-tests-unitaires.md)). Ensuite, pour le confort de développement local, `godotenv` charge un fichier `.env` — mais `Load()` n'écrase jamais une variable déjà définie (utilisez `Overload()` pour forcer), et un `.env` reste un artefact **de développement**, non versionné : en production, l'environnement vient de la plateforme.

**Quand tirer une bibliothèque ?** Les mécaniques de `flag` et de Cobra/Viper sont détaillées côté CLI en [§ 6.1](../06-cli-outillage/01-flag-args-env.md) et [§ 6.2](../06-cli-outillage/02-cobra-viper.md) ; ici, l'angle est architectural. Le choix se résume ainsi :

| Besoin | Approche recommandée |
|---|---|
| Service courant, config issue de l'environnement | **stdlib** (`flag` + `os.LookupEnv`) + chargeur validant écrit à la main |
| Mapping direct env → struct, avec `required` et valeurs par défaut | `caarlos0/env` ou `kelseyhightower/envconfig` (tags de struct) |
| Plusieurs sources et formats (fichiers YAML/TOML, flags, *remote*), fusion, *watch* | **koanf** (léger, modulaire, peu de dépendances) — ou **Viper** pour le couteau suisse tout-en-un |
| Données sensibles | Gestionnaire de secrets — voir la section suivante et [§ 16.2](../16-securite/02-cryptographie-tls.md) |

Un mot sur Viper : très complet, mais il tire historiquement de lourdes dépendances de *remote config* dans son cœur — d'où des binaires nettement plus gros (mesure sur un « hello config » lisant une variable d'environnement, build `-s -w` : **~4,5 Mo avec Viper**, contre ~1,9 Mo avec koanf et ~1,5 Mo en stdlib) — et son API à **singleton global** rend la configuration malaisée à injecter dans les tests — précisément l'inverse de l'idiome « pas de global » retenu au [§ 10.2](02-clean-architecture.md). On ne le tire donc pas « juste pour lire une douzaine de variables ».

## Configuration vs secrets

12-factor range secrets et configuration dans le même sac ; la pratique moderne les **distingue**. La configuration non sensible (adresse d'écoute, niveau de log, indicateurs) peut vivre en clair dans l'environnement ; les **secrets** (mots de passe de base, clés d'API, jetons) exigent un traitement à part :

- **Jamais** dans le dépôt, **jamais** dans les logs — attention à ne pas journaliser une `Config` entière qui contiendrait un mot de passe ; on redéfinit au besoin un `String()` qui masque les champs sensibles.
- Idéalement fournis par un **gestionnaire de secrets** (HashiCorp Vault, secrets managés du cloud, *Secrets* Kubernetes), souvent injectés comme variables d'environnement au démarrage.

Le durcissement (chiffrement, rotation, TLS, stockage des secrets) est le sujet de [§ 16.2](../16-securite/02-cryptographie-tls.md) ; ici, l'essentiel est de **tracer la frontière** entre config anodine et secret dès la conception.

## Feature flags

Un *feature flag* (drapeau de fonctionnalité) **découple le déploiement de la mise en service** : on livre du code « éteint », on l'active plus tard, indépendamment. C'est ce qui permet de désactiver instantanément une fonctionnalité défaillante sans redéployer, d'ouvrir une nouveauté à un sous-ensemble d'utilisateurs, ou de mener un test A/B.

Tous les flags ne se ressemblent pas — et cette distinction (d'après la taxonomie de Pete Hodgson) gouverne leur **durée de vie** :

- **Release toggles** — masquent du code en cours de livraison. À **durée de vie courte** : une fois la fonctionnalité stabilisée, on retire le flag. Les laisser traîner, c'est de la *dette de flags*.
- **Ops toggles** — interrupteurs d'exploitation (*kill switches*) pour couper une fonctionnalité coûteuse ou instable. Souvent durables.
- **Experiment toggles** — pilotent un A/B test. Temporaires, le temps de l'expérience.
- **Permission toggles** — activent des fonctions selon le profil (offre payante, bêta). Durables par nature.

### Le spectre d'implémentation, du plus simple au plus riche

Comme pour la configuration, on dose. Idiomatiquement, on définit un petit **port** (au sens de [§ 10.2](02-clean-architecture.md)) : le domaine dépend d'une abstraction `Flags`, pas d'un fournisseur particulier.

```go
// Flags est un PORT : le domaine évalue des drapeaux sans connaître leur source.
type Flags interface {
	Enabled(ctx context.Context, name string) bool
}

// staticFlags : l'implémentation la plus simple, alimentée par la configuration.
type staticFlags map[string]bool

func (s staticFlags) Enabled(_ context.Context, name string) bool { return s[name] }
```

Le domaine s'en sert sans rien savoir du dessous :

```go
func (s *Service) Checkout(ctx context.Context /* … */) error {
	if s.flags.Enabled(ctx, "new-checkout-flow") {
		return s.checkoutV2(ctx /* … */)
	}
	return s.checkoutV1(ctx /* … */)
}
```

Et le *composition root* choisit l'implémentation — statique en développement, adossée à une plateforme en production, **derrière le même port** :

```go
// dev : drapeaux statiques ; prod : un fournisseur (OpenFeature, Unleash…) satisfait le même port
svc := orders.NewService(repo, staticFlags{"new-checkout-flow": false})
```

Les paliers, du plus léger au plus outillé :

| Besoin | Approche |
|---|---|
| Interrupteur on/off changeant rarement | Booléen **statique** depuis la config (rechargement ou redéploiement) |
| Bascule dynamique sans redéploiement (kill switch, % de *rollout*) | Source **dynamique en process** (base, service de config) ou plateforme |
| Ciblage, *rollout* progressif, A/B à l'échelle, plusieurs services/langages | **Plateforme** — Unleash ou Flagsmith en open source auto-hébergeable, LaunchDarkly ou Split en SaaS |
| Éviter le verrouillage fournisseur, standardiser sur de nombreux services | **OpenFeature** comme couche d'abstraction, avec le fournisseur de votre choix derrière |

**OpenFeature** mérite une mention à part : c'est un standard **neutre** (SDK et spécification), projet CNCF en incubation depuis fin 2023, avec un SDK Go. Son modèle est celui d'OpenTelemetry pour l'observabilité — une API commune, des *providers* propres à chaque outil enfichés dessous, pour éviter le verrouillage. Mais, comme le rappellent ses propres promoteurs, OpenFeature est de la **plomberie, pas un backend** : il standardise *comment* le code évalue un drapeau ; il faut toujours une source pour décider *quelle* valeur ce drapeau prend (SaaS, solution auto-hébergée comme Unleash ou `flagd`, voire un simple fichier JSON).

### Doser, ici aussi

La proportionnalité est la même qu'au [§ 10.2](02-clean-architecture.md). Pour un service unique évaluant cinq drapeaux via un seul fournisseur, la couche d'abstraction OpenFeature ajoute plus de cérémonie que de valeur ; pour une équipe plateforme qui standardise l'évaluation sur des dizaines de services et plusieurs langages, c'est au contraire la bonne fondation — et l'établir tôt coûte bien moins cher que de la rétro-adapter. Le petit port `Flags` ci-dessus offre le meilleur des deux mondes : on démarre en statique, on branche un fournisseur le jour où le besoin apparaît, sans toucher au domaine.

Un dernier garde-fou : la **dette de flags**. Chaque *release toggle* est une branche conditionnelle et une combinaison de tests supplémentaires ; on les retire dès que la fonctionnalité est acquise. Un système criblé de drapeaux oubliés devient aussi illisible qu'un monolithe en boule de boue — le mal que ce module cherche justement à éviter.

## Configuration dynamique et rechargement

La majeure partie de la configuration est **statique** : lue une fois au démarrage. Fidèle au principe des processus jetables (facteur IX), la voie la plus simple pour reconfigurer reste de **redéployer** ou de redémarrer — cohérent avec des déploiements immuables.

Quelques réglages gagnent néanmoins à changer **à chaud**, sans redémarrage : le niveau de log, les *feature flags*, les limites de débit. On les recharge alors via un signal (`SIGHUP`), une surveillance de fichier (les *watchers* de koanf ou Viper) ou un service dédié. La règle : **statique par défaut**, dynamique seulement pour ce qui le justifie réellement — chaque source dynamique est une pièce mobile de plus à raisonner.

## Côté IDE : GoLand et VS Code

Le sujet outillage concret, ici, c'est *fournir* variables d'environnement et arguments au lancement local, dans les deux IDE :

- **GoLand** — dans *Run/Debug Configurations*, les champs **Environment variables** et **Program arguments** ; le plugin **EnvFile** charge un `.env` pour peupler l'environnement d'exécution.
- **VS Code** — dans `launch.json`, les clés `env` (variables inline), `envFile` (pointant un `.env`) et `args` (arguments du programme).

Dans les deux cas, un `.env` local reproduit l'environnement de production **en développement** uniquement ; il n'est pas versionné, et la configuration réelle provient de la plateforme. Les raccourcis correspondants sont regroupés en [annexe D](../annexes/goland-vscode/README.md).

## En résumé

- **12-factor** donne le cadre : un binaire, plusieurs environnements ; la configuration vit dans l'environnement, séparée du code. Go l'incarne presque nativement (binaire unique, sans état, logs vers `stdout`).
- **Configuration** : structure `Config` typée, chargée une fois, **validée au démarrage** (*fail fast*), passée explicitement — jamais de global mutable. La **stdlib suffit** pour la plupart des services ; on ne tire koanf ou Viper que lorsqu'une fonctionnalité précise le justifie.
- **Secrets** : à distinguer nettement de la config — jamais dans le dépôt ni dans les logs, idéalement via un gestionnaire de secrets ([§ 16.2](../16-securite/02-cryptographie-tls.md)).
- **Feature flags** : découplent déploiement et mise en service. On les place derrière un petit **port**, on démarre en statique, on branche **OpenFeature** ou une plateforme quand l'échelle l'exige — et on **retire les drapeaux périmés** pour éviter la dette de flags.

> **Pour aller plus loin** — la méthodologie de référence : [The Twelve-Factor App](https://12factor.net) ; le standard de *feature flagging* : [OpenFeature](https://openfeature.dev) (projet CNCF, SDK Go). Les mécaniques `flag`/Cobra/Viper sont approfondies en [§ 6.1](../06-cli-outillage/01-flag-args-env.md) et [§ 6.2](../06-cli-outillage/02-cobra-viper.md).

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [11 — Interopérabilité et migration](../11-interop-migration/README.md)

⏭ [Interopérabilité et migration](/11-interop-migration/README.md)
