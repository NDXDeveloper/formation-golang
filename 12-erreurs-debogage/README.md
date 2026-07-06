🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 12. Erreurs, débogage et journalisation

On ne peut ni assurer la qualité, ni régler la performance, ni exploiter sereinement un service qu'on ne *voit* pas. Ce module ouvre la partie « Qualité, performance et exploitation » par la brique qui les conditionne toutes : rendre le programme **lisible** — dans ses échecs, pendant le développement, et en production.

Le fil conducteur est l'**explicitude**, marque de fabrique de Go. Là où d'autres langages laissent les défaillances se propager en silence par des exceptions, Go en fait des **valeurs** que l'on traite au grand jour ; là où les logs sont souvent des chaînes libres, la journalisation structurée en fait des **données** exploitables ; et l'observabilité rend enfin lisible ce qu'un service fait une fois en production. Trois registres complémentaires d'une même exigence : savoir ce que le programme fait.

Le module suit une progression naturelle, du plus proche du code au plus proche de l'exploitation : représenter et gérer les échecs, inspecter l'état d'un programme en développement, puis le rendre observable en production. Les erreurs sous-tendent le reste — une bonne stratégie d'erreurs est ce qui donne du sens aux logs et aux traces.

## 🎯 Objectifs du module

À l'issue de ce module, vous serez capable de :

- porter la gestion d'erreurs **à l'échelle d'une application** : taxonomie proportionnée (opaque / sentinelle / typée), enrobage `%w` discipliné, « gérer une erreur **une seule fois** », conversion propre à la frontière (*Problem Details*) ;
- **déboguer avec Delve** dans GoLand comme VS Code : points d'arrêt conditionnels, watchpoints, vue des goroutines, code optimisé, débogage distant ;
- journaliser en **structuré avec `log/slog`** ⭐ : handlers texte/JSON, niveaux dynamiques, attributs et groupes, logger contextuel, `MultiHandler` (Go 1.26) ;
- rendre un service **observable** : traces OpenTelemetry propagées (W3C TraceContext), métriques Prometheus (`/metrics`, RED/USE, cardinalité), health checks — et corréler les trois signaux.

## 🗺️ Les quatre registres

### 12.1 — [Stratégies d'erreurs à l'échelle d'une application](01-strategies-erreurs.md)

**De l'idiome à la stratégie.** Les fondamentaux — `error` comme valeur, `errors.Is`/`errors.As`, l'enrobage avec `%w`, erreurs sentinelles et personnalisées — sont posés en [§ 2.9](../02-fondamentaux-langage/09-gestion-erreurs.md). Cette section les porte à l'échelle d'une application : quelle **taxonomie** d'erreurs adopter, *où* enrichir de contexte (en remontant) et *où* traiter (à la frontière), comment convertir une erreur interne en réponse de transport (statut HTTP, code gRPC, *Problem Details*, cf. [§ 5.5](../05-backend-http/05-api-rest-complete.md)) sans fuiter les détails internes, et la discipline cardinale : **gérer ou propager une erreur, mais ne la journaliser qu'une fois**.

### 12.2 — [Débogage avec Delve (dans GoLand et VS Code)](02-debogage-delve.md)

**Inspecter un programme en marche.** Delve (`dlv`) est le débogueur natif de Go : points d'arrêt (y compris conditionnels), pas-à-pas, inspection des variables et des expressions, vue des **goroutines** et des piles, débogage des tests, et débogage à distance (conteneurs, Kubernetes). Le sujet est traité pour les deux environnements de la formation — le débogueur intégré de **GoLand** (Delve sous le capot) et l'extension Go de **VS Code** (`dlv` piloté par `launch.json`) — avec un mot sur le débogage du code concurrent (cf. [§ 4.6](../04-concurrence/06-tester-code-concurrent.md)) et sur le bon arbitrage entre débogueur et journalisation. Raccourcis en [annexe D](../annexes/goland-vscode/README.md).

### 12.3 — [`log/slog` — journalisation structurée](03-slog.md) ⭐

**Des logs qui sont des données.** `log/slog` (depuis Go 1.21) est la réponse de la bibliothèque standard à la journalisation **structurée** : des paires clé-valeur exploitables par la machine plutôt que des chaînes libres — indispensable à l'observabilité et cohérent avec le principe *12-factor* des logs vers la sortie standard ([§ 10.3](../10-architecture-services/03-configuration-12factor.md)). On y verra les *handlers* (`TextHandler` en développement, `JSONHandler` en production), les niveaux, les attributs et groupes, la propagation par le contexte, et l'écriture de *handlers* personnalisés. Nouveauté à intégrer : Go 1.26 ajoute `slog.NewMultiHandler()` (diffusion d'un même log vers plusieurs *handlers*). L'esprit reste stdlib d'abord : `slog` remplace, pour l'essentiel des besoins, l'ancien paquet `log` et les journaux tiers.

### 12.4 — [Observabilité (OpenTelemetry, métriques Prometheus, health checks)](04-observabilite.md)

**Rendre la production lisible.** L'observabilité, c'est comprendre l'état interne d'un système à partir de ses sorties, selon trois signaux : les logs (12.3), les **métriques** et les **traces**. On y situera **OpenTelemetry**, le standard neutre (CNCF) d'instrumentation — même logique qu'OpenFeature pour les *feature flags* : une API commune, des *backends* interchangeables — ; les métriques **Prometheus** (compteurs, jauges, histogrammes, point d'accès `/metrics`) ; et les *health checks* (sondes *liveness*/*readiness* de Kubernetes, cf. [§ 9.2](../09-conteneurs-cloud/02-kubernetes.md)). À distinguer du profilage en profondeur (`pprof`, traces d'exécution), qui relève de [§ 14.1](../14-performance/01-pprof.md).

## 📋 Positionnement et prérequis

- Ce module **ouvre la partie 6** (Qualité, performance et exploitation). Il fait suite à la partie 5 (Cloud-native, [module 9](../09-conteneurs-cloud/README.md) à [11](../11-interop-migration/README.md)) et fonde la suite : on ne peut ni **tester** ([module 13](../13-tests-qualite/README.md)), ni **optimiser** ([module 14](../14-performance/README.md)), ni sécuriser un service qu'on ne sait pas observer.
- Les erreurs (12.1) s'appuient sur les fondamentaux de [§ 2.9](../02-fondamentaux-langage/09-gestion-erreurs.md) et se prolongent côté API en [§ 5.5](../05-backend-http/05-api-rest-complete.md) (*Problem Details*).
- La journalisation (12.3) matérialise le principe *12-factor* des logs vers `stdout` ([§ 10.3](../10-architecture-services/03-configuration-12factor.md)) et alimente l'observabilité (12.4).
- L'observabilité (12.4) rejoint les sondes Kubernetes ([§ 9.2](../09-conteneurs-cloud/02-kubernetes.md)), se distingue du profilage de [§ 14.1](../14-performance/01-pprof.md), et prend tout son sens en systèmes distribués — le *tracing* corrélé entre services évoqué au [module 8](../08-communication-services/README.md), précieux notamment pendant une migration progressive ([§ 11.3](../11-interop-migration/03-migrer-vers-go.md)).
- Côté outillage, le débogage (12.2) est traité pour **GoLand comme VS Code** ; les raccourcis sont regroupés en [annexe D](../annexes/goland-vscode/README.md).

## 💡 Le fil rouge du module

 Une même exigence, déclinée à trois moments de la vie d'un programme : à la compilation, l'erreur est une **valeur** qu'on gère explicitement, pas une exception qui déroule en silence ; au développement, Delve rend l'état **inspectable** ; en production, logs structurés et observabilité rendent le comportement **lisible**. Voir, avant de juger — c'est la condition de tout ce qui suit dans cette partie.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [12.1 — Stratégies d'erreurs à l'échelle d'une application](01-strategies-erreurs.md)

⏭ [Stratégies d'erreurs à l'échelle d'une application](/12-erreurs-debogage/01-strategies-erreurs.md)
