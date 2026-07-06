🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 10. Architecture de services

Une fois le service conteneurisé et déployé ([module 9](../09-conteneurs-cloud/README.md)), reste la question qui décide de sa longévité : quelle **forme** lui donner ? Où tracer ses frontières, comment organiser ce qu'elles enferment, et comment le faire vivre dans des environnements variés sans jamais le recompiler.

## 🎯 Objectifs du module

À l'issue de ce module, vous serez capable de :

- choisir **où tracer les frontières de déploiement** — monolithe modulaire par défaut, microservices quand leur « taxe » est réellement payée de retour — à l'aide d'une grille de décision ;
- construire un **monolithe modulaire** en Go : packages par domaine, `internal/`, interfaces définies par le consommateur, *composition root* — avec l'extraction comme simple refactor ;
- organiser l'intérieur d'un module selon **ports & adaptateurs**, à la façon Go (interfaces implicites, aucun framework d'injection) et **sans sur-ingénierie** — doser la structure selon la complexité réelle ;
- appliquer les principes **12-factor** : configuration typée validée au démarrage (*fail fast*), stdlib d'abord, secrets à part ;
- découpler déploiement et mise en service avec des **feature flags** proportionnés (du booléen statique à OpenFeature).

L'architecture, en Go, est d'abord une discipline de **retenue**. Le langage a été pensé pour rendre gérables de très grandes bases de code — non pas en multipliant les mécanismes d'abstraction, mais en les **limitant** volontairement : pas d'héritage, des interfaces implicites et minuscules, une seule façon de formater le code, des dépendances explicites. Cette philosophie déteint sur l'architecture : les bons choix ne sont pas ceux qui cochent le plus de cases d'un catalogue de patterns, mais ceux qui préservent la **clarté** et l'**exploitabilité**. *Clear is better than clever* vaut autant pour un fichier de trente lignes que pour un système de trente services.

Cette partie s'intitule « Cloud-native : les forces de Go », et l'architecture est précisément l'endroit où ces forces se convertissent en liberté de choix. Compilation quasi instantanée, binaire statique unique, empreinte mémoire réduite, démarrage à froid en quelques millisecondes, concurrence de première classe : ce sont exactement les propriétés qui rendent *à la fois* un monolithe agréable à faire vivre *et* une flotte de microservices économiquement viable. Autrement dit, Go ne **tranche pas** la décision d'architecture à votre place — il rend les deux voies praticables, ce qui remet le choix là où il doit être : sur vos contraintes réelles (taille d'équipe, cadence de livraison, besoins de mise à l'échelle), pas sur une mode.

Le module suit le fil des trois décisions qui façonnent tout service : **où** tracer ses frontières de déploiement, **comment** organiser l'intérieur de ces frontières, et **comment** l'adapter au monde réel sans le reconstruire.

## 🗺️ Plan du module

| Section | Sujet | En bref |
|---------|-------|---------|
| 10.1 | Monolithe modulaire vs microservices | **Où** tracer les frontières : la taxe du distribué, la grille de décision, l'extraction comme refactor. |
| 10.2 | Clean architecture / hexagonale | **Comment** organiser l'intérieur : ports & adaptateurs à la Go, en dosant la structure. |
| 10.3 | Configuration, feature flags, 12-factor | **Comment** adapter au réel : config validée au démarrage, flags derrière un port. |

## 🧭 Les trois décisions

### 10.1 — [Monolithe modulaire vs microservices](01-monolithe-vs-microservices.md)

**Où tracer les frontières de déploiement.** Le monolithe modulaire — un seul binaire déployable, mais des modules internes aux frontières nettes — est le point de départ raisonnable pour la grande majorité des équipes. On verra pourquoi, comment les packages et le répertoire `internal/` en font respecter les cloisons, et à partir de quel seuil les microservices méritent réellement leur « taxe » : latence réseau, cohérence des données distribuées, observabilité éclatée, complexité de déploiement. Le mot d'ordre : ne pas commencer par des microservices « au cas où ».

### 10.2 — [Clean architecture / hexagonale en Go (sans sur-ingénierie)](02-clean-architecture.md)

**Comment organiser l'intérieur des frontières.** Ports & adaptateurs (architecture hexagonale) et Clean Architecture partagent une même idée : isoler la logique métier des détails d'entrées/sorties — base de données, HTTP, frameworks — en faisant pointer les dépendances vers le cœur. On en verra la traduction *idiomatique* en Go : interfaces définies par le consommateur, injection par simple fonction constructeur, aucun framework de DI. Et surtout le garde-fou annoncé par le titre — proportionner la structure à la complexité, plutôt qu'empiler cinq couches et vingt interfaces pour un CRUD. *A little copying is better than a little dependency.*

### 10.3 — [Configuration, feature flags, principes 12-factor](03-configuration-12factor.md)

**Comment adapter le service au réel.** Un même binaire doit tourner en local, en préproduction et en production sans être recompilé : c'est le cœur de la méthodologie *12-factor*, que les services Go incarnent naturellement. On couvrira la configuration (variables d'environnement, flags, fichiers, ordre de précédence, approche stdlib avant Viper), sa validation au démarrage selon le principe *fail fast*, et les *feature flags* pour découpler le déploiement de la mise en service : déploiements progressifs, interrupteurs d'arrêt d'urgence, du simple booléen en mémoire aux plateformes dédiées.

## 📋 Positionnement et prérequis

- Ce module **prolonge** le [module 9](../09-conteneurs-cloud/README.md) (conteneurs et déploiement cloud) : là où le 9 plaçait le service *dans* un cluster, le 10 décide de sa forme *en amont*. Il **précède** le [module 11](../11-interop-migration/README.md) (interopérabilité et migration), consacré au passage d'un existant Python / Java / Node vers Go.
- Les frontières évoquées en [§ 10.1](01-monolithe-vs-microservices.md) s'appuient sur les packages, `internal/` et le layout standard vus en **[§ 3.5](../03-types-interfaces/05-organisation-code.md)**.
- La communication *entre* services découpés relève du **[module 8](../08-communication-services/README.md)** (REST, gRPC) : ce module-ci décide *s'il faut* découper, pas *comment* les services se parlent.
- Côté exploitation : la configuration injectée par Kubernetes (ConfigMaps, variables d'environnement) est traitée en **[§ 9.2](../09-conteneurs-cloud/02-kubernetes.md)** ; le versioning à la compilation (`ldflags`) en **[§ 15.1](../15-deploiement-devops/01-build-versioning.md)** ; la gestion des **secrets** — à distinguer de la simple configuration — en **[§ 16.2](../16-securite/02-cryptographie-tls.md)**.
- Côté outillage, GoLand (diagrammes et matrice de dépendances) comme VS Code (`gopls`, extension Go officielle) aident à *visualiser* et à faire respecter les frontières entre packages — un appui précieux pour tenir une architecture modulaire honnête. Les raccourcis correspondants sont regroupés en **[annexe D](../annexes/goland-vscode/README.md)**.

## 💡 Anti-pattern à garder en tête tout au long du module

Le plus grand risque architectural en Go n'est pas de *sous*-structurer, mais de **transposer telles quelles** les habitudes d'écosystèmes plus cérémonieux : couches spéculatives « pour plus tard », interfaces à implémentation unique, usines à fabriques. La bonne question n'est jamais « quel pattern appliquer ? » mais « quelle est la structure *la plus simple* qui tienne la complexité *réelle* d'aujourd'hui ? ». Les anti-patterns idiomatiques sont recensés en **[annexe B](../annexes/go-idiomatique/README.md)**.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [10.1 — Monolithe modulaire vs microservices](01-monolithe-vs-microservices.md)

⏭ [Monolithe modulaire vs microservices](/10-architecture-services/01-monolithe-vs-microservices.md)
