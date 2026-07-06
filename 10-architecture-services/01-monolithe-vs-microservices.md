🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 10.1 Monolithe modulaire vs microservices

La première décision d'architecture d'un service n'est pas « quel framework ? » mais « **une seule unité déployable, ou plusieurs ?** ». C'est le choix qui coûte le plus cher à défaire — et celui que l'on prend trop souvent par défaut, sous l'effet d'une mode plutôt que de contraintes réelles.

Pendant une décennie, la réponse par défaut a été « microservices ». En 2026, ce consensus s'est fissuré. Les données d'enquête sectorielles montrent une correction : d'après l'enquête annuelle de la CNCF, **42 %** des organisations ayant adopté les microservices **reconsolident** désormais certains services en unités de déploiement plus larges, et l'adoption du *service mesh* a nettement reculé sur la même période (de 18 % à 8 % entre 2023 et 2025). Une analyse de Thoughtworks va dans le même sens : plus de 40 % des organisations disent regretter au moins une partie de leurs décisions de découpage, en citant la complexité opérationnelle et le coût. Le facteur dominant en 2026 n'est plus la mise à l'échelle, mais le **coût** — de cloud comme de charge cognitive.

Attention au contresens : ce n'est pas « les microservices, c'est mal ». C'est « l'architecture est une décision de contexte, pas un dogme ». Le cas devenu emblématique — l'équipe *Video Quality Analysis* de Prime Video, qui a réduit ses coûts d'infrastructure de plus de 90 % en repliant une chaîne serverless distribuée (AWS Step Functions + Lambda) en un **processus unique** sur ECS, avec transfert de données en mémoire au lieu de passer par S3 — n'invalide pas les microservices : Netflix, à l'inverse, a éclaté son monolithe en microservices à partir de 2008 *précisément* pour gagner en mise à l'échelle et en tolérance aux pannes, et Uber en opère des milliers, à juste titre pour son échelle. Comme l'a résumé le CTO d'Amazon Werner Vogels : bâtir des systèmes évolutifs relève de la stratégie, pas de la religion — et rouvrir ses choix d'architecture est un signe de maturité.

Ce qu'il faut copier de ces équipes, ce n'est pas leur architecture : c'est leur **raisonnement**.

## Trois formes le long d'un même spectre

Le débat « mono vs micro » masque un spectre, dont le point le plus intéressant est au milieu.

- **Monolithe** — une seule unité déployable. Deux visages très différents : la *boule de boue* (« big ball of mud »), sans structure interne, où tout dépend de tout ; ou le monolithe **structuré**, où les responsabilités sont proprement séparées.
- **Monolithe modulaire** — une seule unité déployable, **mais** des modules internes aux frontières explicites et respectées. Discipline de séparation proche de celle des microservices, simplicité opérationnelle de celle d'un monolithe.
- **Microservices** — plusieurs services **déployables indépendamment**, chacun propriétaire de ses données, communiquant par le réseau.

Le point clé, souvent mal compris : *un monolithe modulaire n'est pas une boule de boue relookée*. C'est un patron d'architecture à part entière — modules bien séparés et frontières explicites (comme les microservices), déployés en un seul bloc avec des appels **en mémoire** (comme un monolithe). C'est cette combinaison qui explique son retour en grâce.

## La taxe du distribué

Chaque frontière de service que vous tracez n'est pas gratuite : c'est un appel réseau, une frontière de sérialisation, un nouveau mode de panne et une surface d'exploitation — que vous payez **pour toujours**. Le distribué introduit, de façon irréductible :

- **Latence et pannes partielles** — le réseau échoue, ralentit, se partitionne ; il faut *timeouts*, *retries*, disjoncteurs, et raisonner sur l'indisponibilité d'un pair (cf. [§ 8.1](../08-communication-services/01-consommer-api.md)).
- **Cohérence** — plus de transaction ACID à cheval sur deux services. On bascule vers la cohérence éventuelle, les *sagas*, l'idempotence — un saut de complexité considérable.
- **Observabilité distribuée** — une requête traverse N services ; sans *tracing* corrélé, le débogage devient une enquête (cf. [§ 12.4](../12-erreurs-debogage/04-observabilite.md)).
- **Déploiement et contrats** — orchestration (Kubernetes), versioning des contrats entre services, compatibilité ascendante/descendante à chaque évolution.
- **Coût** — trafic réseau, orchestration, infrastructure dupliquée, et surtout le coût humain : monter en compétence sur l'outillage distribué (Kubernetes, *mesh*, CI/CD par service) prend des mois avant que l'équipe ne livre à pleine vitesse.

La règle qui découle de cette taxe est simple : **ne tracez une frontière de service que lorsque son bénéfice dépasse clairement ce coût permanent.** Par défaut, une frontière de *module* (en mémoire) offre la même séparation logique sans en payer le prix réseau.

## Pourquoi Go rend les deux voies praticables

C'est ici que le langage change la donne — et pourquoi ce module figure dans la partie « les forces de Go ».

**Côté monolithe**, Go rend le bloc unique *agréable à faire vivre* : compilation quasi instantanée (le retour reste rapide même sur une grosse base de code), **binaire statique unique** (déploiement trivial, pas de runtime à installer), bibliothèque standard riche (serveur HTTP, JSON, SQL sans dépendances lourdes). Le principal reproche fait au monolithe — « il devient lent à compiler et à déployer » — mord beaucoup moins en Go.

**Côté microservices**, Go rend chaque service *peu coûteux à exploiter* : empreinte mémoire réduite, **démarrage à froid en quelques millisecondes**, concurrence de première classe, petits binaires qui tiennent dans des images `scratch`/`distroless` (cf. [§ 9.1](../09-conteneurs-cloud/01-docker.md)). Un service Go coûte peu à faire tourner et à mettre à l'échelle.

La conséquence est importante : **Go ne penche pas la balance à votre place.** Les deux voies sont techniquement confortables — donc la décision revient entièrement à *vos* contraintes (taille d'équipe, cadence, profil de charge), pas à une limite du langage ou de l'outillage. C'est aussi ce qui fait de Go un excellent terrain pour *démarrer* en monolithe modulaire et *extraire* plus tard, sans jamais se battre contre l'écosystème.

## Construire un monolithe modulaire en Go

### Le package est l'unité de modularité

Go n'a pas de « module » au sens d'un composant chargé à l'exécution : la frontière, c'est le **package**. La décision structurante est donc d'organiser les packages **par domaine** (*bounded context*), pas par couche technique.

L'anti-patron classique, hérité d'autres écosystèmes, est le découpage par couche : `models/`, `controllers/`, `services/`, `repositories/`. En Go, il fabrique mécaniquement des cycles d'import et des packages fourre-tout, et disperse un même domaine dans quatre dossiers. Le découpage idiomatique regroupe *tout ce qui concerne un domaine* dans un package :

```
myapp/
├── cmd/
│   └── server/
│       └── main.go            // composition root : câble les modules entre eux
├── internal/
│   ├── catalog/               // module « catalogue » (un bounded context)
│   │   ├── catalog.go         // API publique du module : types + point d'entrée exportés
│   │   ├── service.go         // logique interne (non exportée)
│   │   └── store.go           // accès aux données propres au module
│   ├── orders/                // module « commandes »
│   │   ├── orders.go
│   │   └── store.go
│   └── platform/              // briques transverses (config, log, db, http) — sans logique métier
│       ├── database/
│       └── httpx/
├── go.mod
└── go.sum
```

Ici, la granularité qui nous intéresse est la **frontière entre modules** (`catalog`, `orders`). La structure *interne* d'un module — comment y isoler le métier des E/S selon les principes ports & adaptateurs — est le sujet de la section [§ 10.2](02-clean-architecture.md). Le layout complet et commenté figure en [annexe E](../annexes/layout-projet/README.md), et les notions de `internal/`, packages et *workspaces* sont posées en [§ 3.5](../03-types-interfaces/05-organisation-code.md).

### `internal/` : une frontière que le compilateur fait respecter

Go offre un mécanisme rare : le répertoire `internal/` est une frontière **vérifiée à la compilation**. Un package situé sous `internal/` ne peut être importé que par du code enraciné dans le répertoire parent de ce `internal/`. Placer les modules sous `internal/` garantit donc qu'aucun code extérieur au dépôt ne dépendra de vos détails d'implémentation — et, avec des sous-arbres bien pensés, que les modules ne s'importent qu'aux endroits prévus.

À cela s'ajoute un second garde-fou gratuit : **le compilateur Go interdit les cycles d'import.** Ce qui passe ailleurs pour une contrainte à outiller est ici une règle du langage. Un cycle entre `catalog` et `orders` ne compile tout simplement pas, ce qui force à clarifier la direction des dépendances.

### Chaque module expose une petite surface

Un module bien conçu montre le moins possible : un point d'entrée exporté, et tout le reste non exporté (donc invisible au-dehors).

```go
// internal/catalog/catalog.go
package catalog

import (
	"context"
	"database/sql"
)

// Product est le modèle que le module expose à l'extérieur.
type Product struct {
	ID    string
	Name  string
	Price int64 // en centimes
}

// Service est le point d'entrée public du module catalogue.
type Service struct {
	store *store // non exporté : les internes restent invisibles hors du package
}

// New construit le module à partir de ses dépendances.
func New(db *sql.DB) *Service {
	return &Service{store: newStore(db)}
}

func (s *Service) Get(ctx context.Context, id string) (Product, error) {
	return s.store.byID(ctx, id) // … logique métier …
}
```

Le type `store` étant non exporté, aucun autre module ne peut le manipuler : ils passent obligatoirement par `Service`. C'est la version Go du principe « exposer une API, cacher l'implémentation » — sans mot-clé de visibilité au-delà de la simple règle majuscule/minuscule.

### Interfaces définies par le consommateur

Le point *le plus idiomatique* — et le plus utile pour la suite — est que ce n'est pas le module `catalog` qui déclare l'interface dont les autres dépendent. C'est **le consommateur** qui déclare la petite interface dont *il* a besoin :

```go
// internal/orders/orders.go
package orders

import (
	"context"

	"myapp/internal/catalog"
)

// catalogGetter est l'interface définie PAR le consommateur : orders
// ne dépend que de la poignée de méthodes qu'il utilise réellement.
type catalogGetter interface {
	Get(ctx context.Context, id string) (catalog.Product, error)
}

type Service struct {
	catalog catalogGetter
}

func New(cat catalogGetter) *Service {
	return &Service{catalog: cat}
}
```

Ce patron — petites interfaces côté consommateur + injection par le constructeur — apporte trois bénéfices d'un coup : le couplage est minimal (orders ne connaît que `Get`), le test est trivial (on injecte un faux `catalogGetter`, cf. [§ 13.2](../13-tests-qualite/02-mocks-testify.md)), et **la couture d'extraction est déjà en place** (on y revient plus bas). C'est aussi pourquoi Go n'a pas besoin de *framework* d'injection de dépendances : de simples fonctions constructeur suffisent (cf. [§ 10.2](02-clean-architecture.md)).

### Communication entre modules : appels en mémoire, pas HTTP vers soi-même

Dans un monolithe modulaire, les modules s'appellent par **appel de fonction en mémoire** à travers ces interfaces — jamais en s'envoyant des requêtes HTTP à `localhost`. C'est tout le gain d'efficacité (Prime Video repliant le réseau en transfert mémoire en est l'illustration extrême). David Heinemeier Hansson a résumé le contresens inverse d'une formule restée célèbre : remplacer des appels de méthode et des séparations de modules par des invocations réseau, *à l'intérieur d'une même équipe et d'une même application*, est presque toujours une complication gratuite.

Le câblage se fait une seule fois, dans le *composition root* (le `main`) :

```go
// cmd/server/main.go
func main() {
	db := platform.MustOpenDB(cfg)

	cat := catalog.New(db) // module catalogue
	ord := orders.New(cat) // orders dépend de catalog → injection explicite

	mux := http.NewServeMux()
	cat.RegisterRoutes(mux)
	ord.RegisterRoutes(mux)

	log.Fatal(http.ListenAndServe(cfg.Addr, mux))
}
```

Aucune magie, aucune annotation : le graphe de dépendances est un simple enchaînement de constructeurs, lisible de haut en bas.

### Propriété des données : une base, des frontières

Un monolithe modulaire peut parfaitement partager **une seule base de données** — à une condition de discipline : *chaque module est propriétaire de ses tables*. Un module ne lit jamais directement les tables d'un autre ; il passe par l'API de ce module. Concrètement, on réserve un préfixe ou un schéma par domaine (`catalog.*`, `orders.*`) et l'on s'interdit les jointures inter-domaines dans le SQL.

Cette règle a deux vertus : elle empêche le couplage sournois par la base (le pire, car invisible dans le code Go), et elle **rend l'extraction possible** — les données d'un module peuvent le suivre le jour où il devient un service. Le socle `database/sql`, le pool de connexions et les transactions sont traités au [module 7](../07-acces-donnees/README.md).

### Les pièges qui transforment un monolithe modulaire en boule de boue

- **Packages fourre-tout** `util`, `common`, `shared`, `helpers` : ils deviennent des dépôts de tout et recréent le couplage global. Préférez des packages nommés par *intention* (ce qu'ils font), pas par *nature* (« utilitaires »).
- **Fuite des types internes** hors de la frontière : exporter ce qui devrait rester privé rouvre la porte au couplage. La règle majuscule/minuscule est votre première ligne de défense.
- **Accès direct aux tables d'un autre module** : le couplage par la base contourne toutes vos belles interfaces.
- **Le monolithe « distribué »** : un système où chaque module dépend de tous les autres — on cumule alors les défauts des deux mondes. Des frontières nettes, c'est *justement* ce qui évite ce sort, que l'on reste en monolithe ou qu'on éclate plus tard.

## Tenir les frontières avec un linter

Les bonnes intentions ne tiennent pas seules ; il faut *rendre visibles* et, si possible, *bloquer* les dépendances interdites.

Au niveau du langage, on l'a vu, `internal/` et l'interdiction des cycles font déjà une partie du travail à la compilation. Pour aller plus loin, on déclare les dépendances autorisées avec un *linter* : `depguard` (intégré à **golangci-lint**, cf. [§ 13.5](../13-tests-qualite/05-linters.md)) interdit des imports précis entre modules.

```yaml
# .golangci.yml — interdire au module « catalog » de dépendre de « orders »
# (format v2 de golangci-lint, le standard depuis 2025)
version: "2"
linters:
  enable:
    - depguard
  settings:
    depguard:
      rules:
        catalog:
          files:
            - "**/internal/catalog/**"
          deny:
            - pkg: "myapp/internal/orders"
              desc: "le catalogue ne doit pas dépendre des commandes"
```

## Côté IDE : GoLand et VS Code

- **GoLand** — visualisation par **diagramme de dépendances** (menu *Diagrams*), **matrice de dépendances** (*Dependency Matrix* / DSM) pour repérer d'un coup d'œil les liens interdits ou les cycles, inspections *Analyze → Dependencies* et détection des dépendances cycliques, et **scopes** personnalisés pour cadrer les analyses sur un module.
- **VS Code** — `gopls` et l'extension Go officielle fournissent la navigation croisée (*Go to references*, *Call Hierarchy*) pour tracer qui dépend de quoi ; `golangci-lint` (donc `depguard`) remonte les violations à la sauvegarde ; des extensions de visualisation de graphe de dépendances complètent l'ensemble.

Les raccourcis et astuces correspondants sont regroupés en [annexe D](../annexes/goland-vscode/README.md).

## Quand les microservices méritent leur coût

Il existe de vraies raisons de payer la taxe du distribué. Les principales :

- **Autonomie d'équipes à l'échelle (loi de Conway)** — la raison n°1, et la plus solide : plusieurs équipes qui doivent livrer *sans se coordonner*. L'architecture finit toujours par refléter l'organisation ; si vos équipes sont structurées pour posséder des services indépendants, le découpage suit. Si elles ne le sont pas, il vous coûtera sans rien vous rendre.
- **Mise à l'échelle indépendante** — des profils de ressources hétérogènes : un composant gourmand en CPU/GPU à côté d'un composant surtout en I/O, ou un domaine sujet à des pics isolés qu'on veut scaler seul.
- **Isolation de panne (blast radius)** — un domaine critique dont la défaillance ne doit pas emporter le reste du système.
- **Hétérogénéité technologique** — un composant qui exige un autre runtime, un autre langage, ou des bibliothèques incompatibles avec le reste.
- **Contraintes réglementaires / résidence des données** — isoler un domaine pour la conformité (données de santé, localisation géographique imposée…).
- **Cycles de vie divergents** — un composant qui évolue dix fois plus vite qu'un autre et que l'on veut déployer à son propre rythme.

À l'inverse, ne sont **pas** de bonnes raisons : suivre la mode, garnir un CV, « parce que les grands le font » (vous n'avez presque jamais les problèmes d'échelle de Google), ou anticiper une croissance hypothétique. Rappel des contrepoints : Netflix a éclaté son monolithe **pour de bonnes raisons** (échelle, tolérance aux pannes) ; Shopify tient un **monolithe modulaire** à très grande échelle et n'en extrait que quelques domaines précis (paiement, détection de fraude). Deux stratégies opposées, deux contextes différents — pas de recette universelle.

## Grille de décision

| Signal / contrainte | Penche vers le **monolithe modulaire** | Penche vers les **microservices** |
|---|---|---|
| **Nombre d'équipes** | Une seule équipe, ou deux qui se coordonnent bien | Nombreuses équipes autonomes, livraisons indépendantes |
| **Cadence de déploiement** | Partagée, coordonnée sans douleur | Doit être indépendante par domaine |
| **Profil de charge** | Homogène | Hétérogène (CPU/GPU vs I/O, pics isolés) |
| **Maturité opérationnelle** (K8s, *tracing*, CI/CD par service) | Faible ou naissante | Élevée et déjà en place |
| **Tolérance à la complexité distribuée** (cohérence éventuelle, sagas) | Faible — on veut des transactions ACID simples | Assumée et maîtrisée |
| **Isolation de panne** | Acceptable en un bloc | Critique : isoler un domaine sensible |
| **Contraintes réglementaires / données** | Aucune particulière | Fortes : isolation exigée |
| **Hétérogénéité technologique** | Un seul runtime suffit | Besoins polyglottes réels |
| **Coût cloud sous contrainte** | Oui — moins d'overhead réseau/orchestration | Le surcoût est justifié par un autre bénéfice |
| **Stabilité des frontières de domaine** | Encore mouvantes (le refactor en mémoire reste facile) | Éprouvées et stables (extraction sûre) |

Lecture : plus vos réponses tombent à gauche, plus le monolithe modulaire est le bon point de départ ; une majorité à droite justifie d'extraire — souvent *quelques* services autour d'un cœur modulaire, plutôt qu'une flotte complète.

## Extraire un service : le monolithe modulaire comme rampe de lancement

L'intérêt décisif de l'approche : si les frontières sont propres, **extraire un module devient un refactor mécanique, pas une réécriture.**

La couture, c'est l'interface définie par le consommateur. Le jour où `catalog` doit devenir un service distant, `orders` n'a pas à changer : il suffit de lui injecter une *autre* implémentation de `catalogGetter` — un client réseau au lieu d'un appel en mémoire.

```go
// Avant : implémentation en mémoire (in-process)
ord := orders.New(catalog.New(db)) // catalogGetter satisfait par *catalog.Service

// Après : le catalogue est devenu un service distant.
// Un client gRPC/REST satisfait la MÊME interface catalogGetter.
ord := orders.New(catalogclient.New(conn)) // même contrat, transport réseau
```

Deux nuances accompagnent ce basculement. D'abord, le contrat de données (`catalog.Product`) laisse place à un **DTO généré** (par exemple un message protobuf), que le client mappe vers/depuis le type métier — le sujet est traité au [module 8](../08-communication-services/README.md) (gRPC, REST résilient). Ensuite, pour découper *progressivement* un système existant plutôt que d'un bloc, on applique le patron **strangler fig** : on route peu à peu le trafic vers le nouveau service, façade par façade, jusqu'à ce que l'ancien chemin puisse être retiré. La migration d'un existant Python/Java/Node vers Go, qui mobilise la même logique, est traitée en [§ 11.3](../11-interop-migration/03-migrer-vers-go.md).

Pour le reste de la chaîne : conteneuriser et déployer les services relève du [module 9](../09-conteneurs-cloud/README.md) ; structurer l'intérieur de chaque module (métier vs E/S) de la section [§ 10.2](02-clean-architecture.md) ; configurer indépendamment chaque service et piloter les *feature flags* de la section [§ 10.3](03-configuration-12factor.md).

## En résumé

- En 2026, le **monolithe modulaire** est le point de départ raisonnable pour la grande majorité des équipes : simplicité opérationnelle *et* discipline de frontières.
- En Go, le trio **package + `internal/` + interfaces définies par le consommateur** suffit à tenir ces frontières — sans framework, avec le compilateur comme garde-fou (cycles interdits, `internal/` vérifié).
- Chaque frontière de *service* est une **taxe permanente** (réseau, cohérence, exploitation) : ne la payez que contre un bénéfice clair — autonomie d'équipes, mise à l'échelle hétérogène, isolation de panne ou réglementaire.
- Des frontières propres transforment l'extraction en **refactor mécanique** : Go vous laisse démarrer simple et grandir sans vous battre contre l'outillage. Copiez le *raisonnement* des équipes citées, jamais leur architecture telle quelle.

> **Sources et repères** — le billet d'ingénierie Prime Video d'origine (mars 2023) a depuis été **retiré du web** (le blog Prime Video Tech redirige aujourd'hui ailleurs) — on le retrouve via l'Internet Archive ; sa mise en perspective sectorielle (dont Sam Newman et le débat monolithe/microservices) reste lisible chez [DevClass](https://devclass.com/2023/05/05/reduce-costs-by-90-by-moving-from-microservices-to-monolith-amazon-internal-case-study-raises-eyebrows/). Données de consolidation citées : l'enquête annuelle de la [CNCF](https://www.cncf.io/reports/) et l'analyse Thoughtworks sur les tendances d'architecture.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [10.2 — Clean architecture / hexagonale en Go (sans sur-ingénierie)](02-clean-architecture.md)

⏭ [Clean architecture / hexagonale en Go (sans sur-ingénierie)](/10-architecture-services/02-clean-architecture.md)
