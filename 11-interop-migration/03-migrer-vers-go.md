🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 11.3 Migrer un service Python / Java / Node vers Go : stratégies

Troisième lisière du module, et la plus fréquente en pratique : remplacer un service existant par du Go. Les forces de Go — binaire statique, cross-compilation, démarrage rapide, faible empreinte, concurrence — en font une **cible de migration** idéale. Mais migrer est un coût : on ne s'y engage que contre un bénéfice qu'on sait **nommer et mesurer**.

Cette section ne détaille pas la syntaxe d'arrivée (le reste de la formation s'en charge) : elle porte sur les *stratégies* — pourquoi migrer et quand s'en abstenir, comment procéder sans big bang, comment faire cohabiter Go et l'existant, et surtout comment éviter le piège n°1, écrire du Go qui n'en est pas.

## Pourquoi migrer — et quand s'en abstenir

Le *pull* de Go est réel, et c'est celui de toute la partie « Cloud-native » : meilleures performances et empreinte mémoire réduite, **binaire statique unique** (déploiement trivial, cf. [§ 9.1](../09-conteneurs-cloud/01-docker.md) et [§ 15.1](../15-deploiement-devops/01-build-versioning.md)), démarrage à froid en millisecondes (idéal en *serverless*), concurrence de première classe, typage statique qui déplace des bugs de l'exécution vers la compilation, et simplicité opérationnelle. Les études de cas publiques le confirment : plusieurs équipes ont migré leurs backends **critiques en performance** de Python vers Go pour gagner en concurrence et en vitesse d'exécution, tout en simplifiant l'onboarding grâce à la lisibilité du langage.

Mais la retenue est aussi importante que l'élan. On **ne migre pas** quand :

- **le système marche, sans douleur claire** — un service stable et peu touché ne justifie pas le risque ;
- **l'équipe n'a pas d'expertise Go** — la courbe d'apprentissage s'ajoute au risque de migration ;
- **l'écosystème source est décisif** : le ML et la *data-science* restent le terrain de Python, où Go n'a pas d'équivalent mûr. On migre alors les services *autour* (API, passerelles, traitements performance-critiques), pas le cœur analytique ;
- **le coût dépasse le bénéfice nommable** : sans métrique cible (latence, coût d'infra, empreinte), la migration relève de la mode, pas de l'ingénierie.

La bonne question n'est jamais « faut-il passer à Go ? » mais « quel bénéfice précis, mesurable, justifie le coût et le risque de cette migration ? ».

## Le principe cardinal : incrémental, pas *big bang*

La réécriture d'un bloc (*big-bang rewrite*) est le grand anti-patron : risque maximal, **perte de connaissance métier** enfouie dans l'ancien code, « effet second système » (on ré-empile toute la complexité rêvée), longue période sans aucune valeur livrée, et bascule unique impossible à annuler. Joel Spolsky en a fait un classique du genre : réécrire de zéro est l'erreur stratégique que l'on ne devrait presque jamais commettre.

L'alternative tient en trois mots : **incrémental, réversible, observable**. On avance par petites tranches, chacune validée en production réelle avant de retirer l'ancien chemin.

## Stratégie phare : le *strangler fig*

C'est le patron déjà croisé en [§ 10.1](../10-architecture-services/01-monolithe-vs-microservices.md), appliqué ici au remplacement d'un existant. Une **façade** (un proxy, souvent une passerelle ou un *reverse proxy*) se place devant le système hérité et route chaque capacité, une tranche à la fois, vers le nouveau service Go ; l'ancien chemin n'est retiré qu'une fois le nouveau **prouvé**. Le nom vient du figuier étrangleur, qui pousse autour de son hôte jusqu'à le remplacer.

Trois phases : **intercepter** (mettre la façade devant l'existant), **remplacer progressivement** (router les tranches migrées vers Go, le reste vers l'existant), **achever** (retirer l'ancien système).

En Go, la bibliothèque standard suffit à écrire cette façade — et Go, rapide et concurrent, fait aussi un excellent proxy :

```go
// Façade « strangler » : route la tranche migrée vers le service Go,
// le reste vers l'existant — avec la seule bibliothèque standard.
func main() {
	legacy, _ := url.Parse("http://legacy.internal:9000")
	legacyProxy := httputil.NewSingleHostReverseProxy(legacy)

	mux := http.NewServeMux()
	mux.Handle("/api/orders/", newOrdersHandler()) // migré : servi par Go
	mux.Handle("/", legacyProxy)                    // le reste : encore l'existant

	log.Fatal(http.ListenAndServe(":8080", mux))
}
```

Le patron ne se réduit pas au routage. Trois disciplines le distinguent d'un bricolage :

- **Choisir la frontière stable** — quelle capacité forme un tout cohérent, avec des dépendances externes minimales.
- **Choisir la première tranche** — un service à *faible risque, forte valeur, frontières nettes* : on veut une victoire précoce qui crée l'élan.
- **Écrire des critères de sortie** — *à quelles conditions* l'ancien chemin sera retiré. Sans eux, l'ancien code survit « au cas où » et l'on finit par maintenir deux systèmes pour la même fonction, indéfiniment.

La bascule elle-même se pilote finement : on route d'abord un **pourcentage** du trafic vers Go (déploiement canari), puis on monte en charge — un usage direct des *feature flags* de [§ 10.3](../10-architecture-services/03-configuration-12factor.md), avec retour arrière instantané.

## Faire cohabiter Go et l'existant

Pendant la transition, Go et le système hérité doivent se parler. L'interface se définit par un **contrat explicite** — gRPC/protobuf ou OpenAPI (cf. [module 8](../08-communication-services/README.md)) — que les deux côtés respectent, indépendamment de leur langage.

Le point sensible est la **donnée**. Deux systèmes qui écrivent les mêmes tables, c'est le couplage sournois par la base (cf. [§ 10.1](../10-architecture-services/01-monolithe-vs-microservices.md)). On vise plutôt qu'une capacité migrée **possède ses propres tables**, l'autre côté passant par son API ; à défaut, on encadre les accès partagés par des écritures ombres et des validations, le temps de la bascule.

## Sécuriser la bascule

Quatre garde-fous transforment une migration risquée en série de pas sûrs :

- **Tests de caractérisation** (*golden tests*) : capturer le comportement de l'existant *avant* de migrer, puis exécuter les mêmes cas contre l'implémentation Go pour **prouver la parité** (cf. [module 13](../13-tests-qualite/README.md)).
- **Exécution parallèle / trafic fantôme** : faire tourner Go à côté de l'existant sur du trafic réel et **comparer les sorties** (diff) avant de basculer.
- **Feature flags** pour contrôler la bascule et revenir en arrière en un instant ([§ 10.3](../10-architecture-services/03-configuration-12factor.md)).
- **Observabilité** pour comparer latence et taux d'erreur entre l'ancien et le nouveau ([§ 12.4](../12-erreurs-debogage/04-observabilite.md)).

## Le vrai piège : écrire du Go idiomatique, pas une translittération

L'erreur la plus commune n'est pas technique, elle est culturelle : porter le code **ligne à ligne** produit du « Java-en-Go », du « Python-en-Go », du « Node-en-Go » — du code qui compile mais qui se bat contre le langage. On ne translittère pas ; on **ré-exprime le domaine** dans les idiomes de Go.

```go
// « Java/Python-en-Go » : émuler les exceptions avec panic/recover — à fuir
func getUser(id string) User {
	u, err := store.find(id)
	if err != nil {
		panic(err) // ce n'est pas ainsi qu'on gère une erreur en Go
	}
	return u
}

// Go idiomatique : l'erreur est une valeur, retournée explicitement
func getUser(id string) (User, error) {
	return store.find(id)
}
```

Le décalage à opérer diffère selon la source :

- **Depuis Python** — types explicites plutôt que *duck typing* ; erreurs retournées plutôt qu'exceptions ; goroutines plutôt qu'`asyncio` ou le GIL (nuance : le CPython *free-threaded* fait évoluer ce dernier point, cf. [§ 1.6.1](../01-introduction-go/06.1-go-vs-autres-langages.md)) ; pas de décorateurs ni de *comprehensions* magiques ; la stdlib avant les paquets. Rappel : Python reste souvent le bon choix pour le ML/data — on migre les services performants, pas tout.
- **Depuis Java** — composition plutôt qu'héritage ([§ 10.2](../10-architecture-services/02-clean-architecture.md)) ; petites interfaces plutôt que hiérarchies de classes ; erreurs plutôt qu'exceptions ; **pas** de framework d'injection (de simples constructeurs, [§ 10.2](../10-architecture-services/02-clean-architecture.md)) ; goroutines plutôt que pools de threads. Piège majeur : transposer la sur-abstraction « entreprise » (fabriques, couches profondes) — exactement le travers combattu en 10.2, recensé en [annexe B](../annexes/go-idiomatique/README.md).
- **Depuis Node.js** — goroutines et concurrence structurée plutôt que boucle d'événements, *callbacks* et promesses ; types statiques plutôt que dynamiques ; erreurs explicites plutôt que `try/catch` et rejets ; la stdlib plutôt qu'« un paquet npm pour tout ». Piège : recréer des chaînes de promesses, ou l'inflation de dépendances.

Le fil commun est le même partout : viser le **Go idiomatique** ([annexe B](../annexes/go-idiomatique/README.md)). L'IA accélère la migration mais tend justement à produire du « X-en-Go » — d'où l'importance d'une relecture exigeante ; la migration assistée par IA est traitée en [§ 17.3](../17-developpement-ia/03-tests-migration-ia.md), ses pièges en [§ 17.2](../17-developpement-ia/02-pieges-ia.md).

## Grille de décision

| Situation | Stratégie |
|---|---|
| Système critique et intégré, en production, à moderniser | **Strangler fig** (façade + tranches + critères de sortie) |
| Service isolé, petit, à faible risque | Réécriture ciblée (*greenfield*) acceptable |
| Besoin de forte confiance avant bascule | **Exécution parallèle / trafic fantôme**, puis canari |
| Cœur ML / *data-science* | **Garder Python** ; migrer les services autour |
| Système qui marche, sans douleur mesurable | **Ne pas migrer** |

## Côté IDE : GoLand et VS Code

Le temps de la transition, le dépôt est **polyglotte** (Go + Python/Java/Node) ; les deux environnements le gèrent.

- **GoLand** — refactorisation et navigation Go de premier ordre ; le code source des autres langages s'ouvre via les plugins JetBrains correspondants (ou l'IDE polyvalent). La génération des contrats (protobuf via `go:generate`, cf. [§ 8.2](../08-communication-services/02-grpc.md)) s'intègre au flux de build.
- **VS Code** — nativement polyglotte : extensions Go, Python, JS et Java cohabitent dans un même *workspace* ; les tâches (`tasks.json`) pilotent la génération des contrats et les *runtimes* de chaque côté.

Les raccourcis correspondants sont regroupés en [annexe D](../annexes/goland-vscode/README.md).

## En résumé

- Go est une **cible de migration** idéale (binaire unique, démarrage rapide, concurrence) — mais on ne migre que contre un **bénéfice nommé et mesuré**, et l'on garde Python pour le ML/data.
- **Incrémental, pas *big bang*.** La réécriture d'un bloc est l'anti-patron ; le ***strangler fig*** (façade + tranches, prolongement de [§ 10.1](../10-architecture-services/01-monolithe-vs-microservices.md)) offre une bascule réversible et observable — à condition de définir la frontière stable et des **critères de sortie**.
- **Cohabitation** par contrats explicites (gRPC/OpenAPI, [module 8](../08-communication-services/README.md)) et propriété claire des données ; **bascule sécurisée** par tests de caractérisation, trafic fantôme, *feature flags* ([§ 10.3](../10-architecture-services/03-configuration-12factor.md)) et observabilité ([§ 12.4](../12-erreurs-debogage/04-observabilite.md)).
- Le piège n°1 est d'écrire une **translittération** (« Java/Python/Node-en-Go ») : on ré-exprime le domaine en **Go idiomatique** ([annexe B](../annexes/go-idiomatique/README.md)), sans quoi la migration reproduit les défauts qu'elle prétendait corriger.

> **Pour aller plus loin** — le patron d'origine : *Strangler Fig Application* de Martin Fowler ; des retours d'expérience Go (dont des migrations Python → Go) : [études de cas Go](https://go.dev/solutions/case-studies). L'avertissement fondateur contre la réécriture de zéro est de Joel Spolsky (*Things You Should Never Do*).

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [12 — Erreurs, débogage et journalisation](../12-erreurs-debogage/README.md)

⏭ [Erreurs, débogage et journalisation](/12-erreurs-debogage/README.md)
