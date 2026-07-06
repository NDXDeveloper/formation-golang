🔝 Retour au [Sommaire](/SOMMAIRE.md)

# Annexe F — Glossaire et acronymes

Ce glossaire réunit les termes et sigles employés tout au long de la formation, avec une définition courte et un renvoi vers la section qui les traite en détail. Les entrées sont **regroupées par thème** et classées **par ordre alphabétique** à l'intérieur de chaque groupe.

> **Notation.** Les renvois de la forme `§X.Y` désignent les sections du [sommaire](../../SOMMAIRE.md) (`§4.2` = chapitre 4, section 2). Ils ne sont pas hypertextes ici pour garder le glossaire lisible ; les liens principaux sont regroupés en [fin de page](#pour-aller-plus-loin). Pour la correspondance de syntaxe avec d'autres langages, voir l'[annexe A](../correspondance-go-autres/README.md) ; pour les idiomes, l'[annexe B](../go-idiomatique/README.md).

---

## Acronymes et sigles

| Sigle | Développé | En bref |
|---|---|---|
| ACID | *Atomicity, Consistency, Isolation, Durability* | Propriétés d'une transaction fiable (§7.1) |
| API | *Application Programming Interface* | Interface de programmation exposée par un service ou une bibliothèque |
| AWS | *Amazon Web Services* | Plateforme cloud d'Amazon (§9.3) |
| CI/CD | *Continuous Integration / Continuous Delivery (ou Deployment)* | Intégration et livraison/déploiement continus (§15.2) |
| CLI | *Command-Line Interface* | Application en ligne de commande (chap. 6) |
| CORS | *Cross-Origin Resource Sharing* | Contrôle des requêtes inter-origines côté navigateur (§5.2) |
| CSP | *Communicating Sequential Processes* | Modèle de concurrence par échange de messages (chap. 4) |
| CVE | *Common Vulnerabilities and Exposures* | Identifiant public d'une vulnérabilité (§15.3) |
| DTO | *Data Transfer Object* | Objet dédié au transfert de données à la frontière d'une API (§5.3) |
| E/S | *Entrées/Sorties (I/O)* | Lectures/écritures fichiers, réseau… (§7.6) |
| FFI | *Foreign Function Interface* | Appel de code écrit dans un autre langage (§11.1) |
| GC | *Garbage Collector* | Ramasse-miettes (gestion automatique de la mémoire) (§14.2) |
| gRPC | *gRPC Remote Procedure Calls* | Framework RPC hautes performances (protobuf, HTTP/2), issu de Google (§8.2) |
| HPKE | *Hybrid Public Key Encryption* | Chiffrement hybride à clé publique (`crypto/hpke`, Go 1.26) (§16.2) |
| HTTP(S) | *HyperText Transfer Protocol (Secure)* | Protocole du web (chap. 5) |
| IDE | *Integrated Development Environment* | Environnement de développement (GoLand, VS Code) (§1.6.2) |
| JSON | *JavaScript Object Notation* | Format d'échange texte (§5.3) |
| JWT | *JSON Web Token* | Jeton signé pour l'authentification sans état (§5.6) |
| k8s | *Kubernetes* | Orchestrateur de conteneurs (§9.2) |
| ML-KEM | *Module-Lattice-Based Key-Encapsulation Mechanism* | Échange de clés post-quantique (hybride par défaut en Go 1.26) (§16.2) |
| OAuth | *Open Authorization* | Cadre d'autorisation déléguée (§5.6) |
| OIDC | *OpenID Connect* | Couche d'identité au-dessus d'OAuth 2.0 (§5.6) |
| ORM | *Object-Relational Mapping* | Correspondance objet↔relationnel (§7.3) |
| OTel | *OpenTelemetry* | Standard d'observabilité : traces, métriques, logs (§12.4) |
| OWASP | *Open Worldwide Application Security Project* | Référentiel de sécurité applicative (§16.1) |
| PGO | *Profile-Guided Optimization* | Optimisation guidée par un profil d'exécution (§14.3) |
| REST | *Representational State Transfer* | Style d'architecture d'API sur HTTP (§5.5) |
| RPC | *Remote Procedure Call* | Appel de procédure distante (§8.2) |
| SBOM | *Software Bill of Materials* | Inventaire des composants d'un logiciel (§15.3) |
| SSE | *Server-Sent Events* | Flux d'événements unidirectionnel serveur→client (§8.3) |
| SQL | *Structured Query Language* | Langage de requêtes relationnelles (§7.1) |
| TLS | *Transport Layer Security* | Chiffrement de la couche transport (§16.2) |
| TUI | *Terminal / Text User Interface* | Interface en mode texte (§6.3) |
| URL / URI | *Uniform Resource Locator / Identifier* | Adresse / identifiant d'une ressource |
| UTF-8 | *Unicode Transformation Format 8-bit* | Encodage des chaînes en Go (§2.6) |
| WASI | *WebAssembly System Interface* | Interface système pour WebAssembly (§11.2) |
| WASM | *WebAssembly* | Format binaire portable exécutable hors navigateur (§11.2) |
| XSS | *Cross-Site Scripting* | Injection de scripts côté client (§16.1) |
| YAML | *YAML Ain't Markup Language* | Format de configuration lisible (k8s, CI…) |

---

## Langage & fondamentaux

| Terme | Définition |
|---|---|
| **any** | Alias de `interface{}` (Go 1.18) ; « ne dit rien » quand un type concret suffirait (§3.4) |
| **append** | Ajoute des éléments à un slice, avec réallocation si la capacité est dépassée (§2.7) |
| **byte** | Alias de `uint8` ; un octet (§2.6) |
| **closure (fermeture)** | Fonction capturant des variables de son environnement (§2.3) |
| **comma-ok** | Idiome `v, ok := …` testant la présence (map), le succès (assertion) ou l'ouverture (canal) |
| **comparable** | Contrainte des types utilisables avec `==` en générique (§3.4) |
| **composition** | Assemblage de types par embedding, à la place de l'héritage (§3.2) |
| **contrainte (constraint)** | Interface bornant les types acceptés par un paramètre générique (§3.4) |
| **defer** | Diffère un appel jusqu'au retour de la fonction (ordre LIFO) ; libère les ressources (§2.10) |
| **embedding (incorporation)** | Inclusion d'un type dans un autre, qui en promeut champs et méthodes (§3.2) |
| **error** | Interface standard (`Error() string`) ; en Go, les erreurs sont des valeurs (§2.9) |
| **escape analysis** | Analyse du compilateur décidant pile ou tas pour une variable (§14.2) |
| **exporté / non exporté** | La casse décide : `Majuscule` = public, `minuscule` = privé au package (§2.1) |
| **fonction de première classe** | Les fonctions sont des valeurs : assignables, passables en argument (§2.3) |
| **génériques (generics)** | Types et fonctions paramétrés par des types (depuis Go 1.18) (§3.4) |
| **interface** | Ensemble de signatures de méthodes, satisfaite **implicitement** (structurellement) (§3.3) |
| **iota** | Compteur d'un bloc `const`, incrémenté à chaque ligne ; sert à énumérer (§2.2) |
| **itérateur (range-over-func)** | Fonction parcourable par `for range` (Go 1.23) (§2.5) |
| **jeu de méthodes (method set)** | Méthodes associées à un type ; diffère entre `T` et `*T` (§3.1) |
| **map** | Table de hachage clé→valeur ; ordre de parcours non déterministe (§2.7) |
| **méthode** | Fonction attachée à un type via un receveur (§3.1) |
| **new** | Alloue une variable et renvoie son pointeur ; 🆕 accepte une expression d'initialisation depuis Go 1.26 (§2.8) |
| **nil** | Valeur zéro des pointeurs, slices, maps, canaux, fonctions et interfaces (§2.8) |
| **package** | Unité d'organisation du code : un dossier = un package (§2.1, §3.5) |
| **panic** | Interruption du flux normal ; réservée à l'irrécupérable (§2.10) |
| **pointeur (pointer)** | Adresse d'une variable ; **sans arithmétique** en Go (§2.8) |
| **receveur (receiver)** | Cible d'une méthode (`func (r T) M()`), en valeur ou en pointeur (§3.1) |
| **recover** | Reprend la main après un `panic`, uniquement dans un `defer` (§2.10) |
| **rune** | Alias de `int32` ; un point de code Unicode (§2.6) |
| **erreur sentinelle (sentinel error)** | Erreur exportée comparable (`io.EOF`), testée via `errors.Is` (§2.9) |
| **slice** | Vue redimensionnable sur un tableau : pointeur, longueur, capacité (§2.7) |
| **struct** | Type agrégat de champs nommés ; il n'y a pas de classe en Go (§3.1) |
| **tableau (array)** | Séquence de **taille fixe** ; type valeur, copié à l'affectation (§2.7) |
| **type assertion** | `x.(T)` : extrait la valeur concrète d'une interface (§2.4) |
| **type switch** | `switch v := x.(type)` : aiguille selon le type dynamique (§2.4) |
| **variadique (variadic)** | Paramètre acceptant un nombre variable d'arguments (`nums ...int`) (§2.3) |
| **valeur zéro (zero value)** | Valeur par défaut d'un type non initialisé (`0`, `""`, `false`, `nil`) (§2.2) |
| **wrapping (enveloppe)** | `fmt.Errorf("… : %w", err)` conserve la chaîne d'erreurs (§2.9) |

---

## Concurrence

| Terme | Définition |
|---|---|
| **canal (channel)** | Conduit typé de communication et de synchronisation entre goroutines (§4.2) |
| **canal tamponné / non tamponné** | Tamponné (`make(chan T, n)`) stocke n éléments ; non tamponné synchronise par rendez-vous (§4.2) |
| **context.Context** | Transporte annulation, délais et valeurs à travers les appels (§4.4) |
| **data race** | Accès concurrent à une donnée dont au moins un écrit, sans synchronisation (§4.6) |
| **deadlock (interblocage)** | Goroutines bloquées en attente mutuelle |
| **errgroup** | Coordonne un groupe de goroutines avec propagation d'erreur et annulation (§4.3) |
| **fan-in / fan-out** | Fusionner plusieurs sources en une / distribuer le travail à plusieurs goroutines (§4.5) |
| **goroutine** | Fonction exécutée de façon concurrente par le runtime (pas un thread OS) ; lancée par `go f()` (§4.1) |
| **fuite de goroutine (leak)** | Goroutine qui ne se termine jamais faute de condition d'arrêt (§4.1) |
| **mutex** | Verrou d'exclusion mutuelle (`sync.Mutex`) (§4.3) |
| **ordonnanceur (scheduler)** | Répartit les goroutines sur les threads système (modèle M:N) (§4.1) |
| **pipeline** | Chaîne d'étapes reliées par des canaux (§4.5) |
| **race condition** | Comportement dépendant de l'ordonnancement d'accès concurrents ; détecté par `-race` (§4.6) |
| **select** | Attend simultanément plusieurs opérations de canal (§4.2) |
| **synctest** | `testing/synctest` : test de code concurrent en temps virtuel, stable en Go 1.25 (§4.6) |
| **WaitGroup** | Compteur attendant la fin d'un ensemble de goroutines (§4.3) |
| **worker pool** | Goroutines consommant des tâches depuis un canal (§4.5) |

---

## Web, API & communication

| Terme | Définition |
|---|---|
| **Chi / Echo / Gin** | Routeurs et frameworks HTTP tiers, alternatives à la stdlib (§5.4) |
| **arrêt propre (graceful shutdown)** | Arrêt du serveur en laissant les requêtes en cours se terminer (§9.2) |
| **handler** | Fonction traitant une requête HTTP (§5.1) |
| **interceptor** | Équivalent gRPC d'un middleware (§8.2) |
| **middleware** | Fonction enveloppant un handler pour ajouter un comportement (logging, auth, CORS) (§5.2) |
| **Problem Details (RFC 9457)** | Format JSON standard pour représenter une erreur HTTP (§5.5) |
| **protobuf (Protocol Buffers)** | Format de sérialisation binaire, base de gRPC (§8.2) |
| **ServeMux** | Routeur HTTP de la bibliothèque standard (§5.1) |
| **WebSocket** | Canal bidirectionnel persistant établi sur une connexion HTTP (§8.3) |

---

## Données & persistance

| Terme | Définition |
|---|---|
| **embed (`//go:embed`)** | Incorpore des fichiers statiques directement dans le binaire (§7.6) |
| **Ent / GORM** | ORM Go (mapping objet-relationnel) (§7.3) |
| **golang-migrate / goose** | Outils de migration de schéma de base (§7.4) |
| **migration** | Évolution versionnée d'un schéma de base (§7.4) |
| **os.Root** | Accès fichiers **confiné** à un répertoire (Go 1.24), contre les évasions de chemin (§7.6) |
| **pgx** | Driver PostgreSQL performant, recommandé (§7.2) |
| **pool de connexions (connection pool)** | Réserve de connexions réutilisables gérée par `database/sql` (§7.1) |
| **requête préparée (prepared statement)** | Requête pré-compilée et paramétrée (§7.1) |
| **sqlc** | Génère du code Go **typé** à partir de requêtes SQL (§7.3) |

---

## Cloud-native, architecture & exploitation

| Terme | Définition |
|---|---|
| **12-factor** | Méthodologie d'applications cloud-natives (config par l'environnement, etc.) (§10.3) |
| **clean architecture** | Organisation en couches, dépendances orientées vers le domaine (§10.2) |
| **cross-compilation** | Compiler pour un autre OS/architecture via `GOOS`/`GOARCH` (§1.5, §6.4) |
| **distroless** | Image conteneur minimale, sans shell ni gestionnaire de paquets (§9.1) |
| **feature flag** | Interrupteur activant une fonctionnalité sans redéploiement (§10.3) |
| **GoReleaser** | Outil de build et de publication multi-plateformes (§6.4) |
| **hexagonale (ports & adapters)** | Architecture isolant le domaine des entrées/sorties (§10.2) |
| **multi-stage build** | Dockerfile en plusieurs étapes : compilation puis image finale légère (§9.1) |
| **probe (liveness / readiness)** | Sonde Kubernetes vérifiant la santé d'un conteneur (§9.2) |
| **scratch** | Image conteneur **vide**, base d'un binaire statique (§9.1) |
| **serverless** | Exécution sans gestion de serveurs (AWS Lambda, Cloud Run) (§9.3) |

---

## Tests, qualité & performance

| Terme | Définition |
|---|---|
| **benchmark** | Mesure de performance (`func BenchmarkX(b *testing.B)`) (§13.4) |
| **benchstat** | Comparaison statistique de résultats de benchmarks (§14.4) |
| **flight recorder / execution trace** | Enregistrement d'événements d'exécution pour diagnostic (Go 1.25), en complément de pprof (§14.1) |
| **fuzzing** | Test à entrées aléatoires cherchant des bugs (`go test -fuzz`) (§13.4) |
| **Green Tea GC** | Nouveau ramasse-miettes, **activé par défaut en Go 1.26** (10-40 % d'overhead GC en moins) ; opt-out `GOEXPERIMENT=nogreenteagc` (§14.2) |
| **httptest** | Utilitaires de la stdlib pour tester du code HTTP (§13.2) |
| **mock** | Double de test simulant une dépendance via une interface (§13.2) |
| **pprof** | Profilage CPU, tas (heap) et goroutines (§14.1) |
| **sync.Pool** | Cache d'objets réutilisables pour réduire les allocations (§14.3) |
| **table-driven test** | Test paramétré par une table de cas, avec sous-tests (§13.1) |
| **Testcontainers** | Dépendances réelles (bases…) en conteneurs jetables pour les tests d'intégration (§13.3) |
| **testify** | Bibliothèque tierce d'assertions et de mocks (§13.2) |

---

## Outillage & écosystème Go

| Terme | Définition |
|---|---|
| **Bubble Tea** | Framework de TUI (interfaces en mode texte) (§6.3) |
| **Cobra / Viper** | Commandes et sous-commandes CLI / gestion de configuration (§6.2) |
| **Delve (`dlv`)** | Débogueur Go, intégré à GoLand et VS Code (§12.2) |
| **gofmt / goimports** | Formatage du code ; `goimports` gère aussi les imports (§1.4) |
| **golangci-lint** | Agrégateur de linters, standard de fait (§13.5) |
| **go.mod / go.sum** | Fichiers de module : dépendances déclarées et empreintes vérifiées (§15.1) |
| **GOMAXPROCS** | Nombre max de threads exécutant du code Go ; **container-aware** depuis Go 1.25 (§9.2) |
| **GOOS / GOARCH** | Cibles système et architecture de compilation (§1.5) |
| **gopls** | Serveur de langage Go : complétion, navigation, diagnostics (§1.4) |
| **GOPATH** | Ancien répertoire de travail/cache, héritage d'avant les modules (§1.4) |
| **GOROOT** | Répertoire d'installation de Go (§1.4) |
| **go vet** | Analyse statique intégrée signalant des constructions suspectes (§13.5) |
| **go.work (workspace)** | Développement multi-module en local (§3.5) |
| **govulncheck** | Détecte les vulnérabilités connues affectant le code **réellement appelé** (§15.3) |
| **ldflags** | Options de l'éditeur de liens (ex. injecter la version au build) (§15.1) |
| **modernizers (`go fix`)** | Réécritures automatiques vers des idiomes récents (Go 1.26) (§13.5) |
| **module** | Unité de versionnement et de distribution du code (`go.mod`) (§1.3) |
| **runtime** | Bibliothèque d'exécution intégrée au binaire (scheduler, GC, mémoire) (§4.1) |
| **semver (Semantic Versioning)** | Versionnement `MAJEUR.MINEUR.CORRECTIF` des modules (§18.1, annexe H) |
| **staticcheck** | Suite d'analyse statique avancée (§13.5) |
| **stdlib (bibliothèque standard)** | Paquets livrés avec Go ; à privilégier avant les frameworks |
| **toolchain (chaîne d'outils)** | L'ensemble des outils `go` : build, test, vet, fmt… (§1.3) |
| **vendoring** | Copie des dépendances dans `vendor/` pour des builds hermétiques (annexes C et E) |

---

## Sécurité

| Terme | Définition |
|---|---|
| **HPKE (`crypto/hpke`)** | Chiffrement hybride à clé publique, nouveau paquet en Go 1.26 (§16.2) |
| **ML-KEM** | Mécanisme d'échange de clés **post-quantique** ; hybride par défaut en Go 1.26 (§16.2) |
| **rate limiting (limitation de débit)** | Borne le nombre de requêtes pour protéger un service (§16.3) |
| **secret** | Donnée sensible (clé, mot de passe) ; jamais en dur ni dans le dépôt (§16.2, §10.3) |
| **supply chain (chaîne d'approvisionnement)** | Sécurité des dépendances et du processus de build (§15.3) |

---

## Pour aller plus loin

- **Correspondance de syntaxe** avec Python, Java, C#, Rust : [annexe A](../correspondance-go-autres/README.md).
- **Idiomes et anti-patterns** du langage : [annexe B](../go-idiomatique/README.md).
- **Versions Go et politique de compatibilité** : [annexe H](../versions-reference/README.md).
- **Ressources, veille et communauté** : [§18.3](../../18-strategie-roadmap/03-communaute-ressources.md).
- Toutes les sections sont accessibles depuis le [sommaire](../../SOMMAIRE.md).

---

🔝 [Retour au sommaire](../../SOMMAIRE.md) · ⏭️ [Annexe G — FAQ et dépannage](../faq-depannage/README.md)


⏭ [FAQ et dépannage](/annexes/faq-depannage/README.md)
