# 📚 Sommaire — Formation Go (Golang) avec GoLand & VS Code
**Du débutant au professionnel · Go 1.26 · Juin 2026**

> **Légende** : ⭐ essentiel · 🆕 apport récent du langage ou de l'outillage · 🤖 développement assisté par IA

---

## 🎯 Parcours de formation

| Parcours | Modules | Durée estimée |
|----------|---------|---------------|
| **Débutant** | 1-5, 12 | 6-8 jours |
| **Backend / API** ⭐ | 1-5, 7-9, 12-13, 15-16 | 11-13 jours |
| **CLI & Outillage** | 1-4, 6, 12-13, 15 | 7-9 jours |
| **Cloud / DevOps** | 1-4, 8-12, 15-16 | 11-13 jours |
| **IA-First** 🤖 | 1-2, 5, 17, 18 | 4-6 jours |
| **Formation complète** | 1-18 + annexes | 22-28 jours |

---

## **Partie 1 — Comprendre Go en 2026 (cadrage & langage)**

### 1. [Introduction : Go et son écosystème](01-introduction-go/README.md)
- 1.1 [Qu'est-ce que Go et à quoi il sert réellement (backend, CLI, cloud-native)](01-introduction-go/01-quest-ce-que-go.md)
- 1.2 [Histoire et philosophie du langage (simplicité, lisibilité)](01-introduction-go/02-histoire-philosophie.md)
- 1.3 [L'écosystème Go (toolchain, modules, cycle de release semestriel)](01-introduction-go/03-ecosysteme-go.md)
- 1.4 [Installation et outils](01-introduction-go/04-installation-outils.md)
    - GoLand (inspections, refactoring, débogueur intégré)
    - VS Code + extension officielle Go (gopls, dlv)
    - CLI : `go build`, `go run`, `go test`, `go vet`, `gofmt`
- 1.5 [Premier projet pas à pas (`go mod init`, hello world)](01-introduction-go/05-premier-projet.md)
- 1.6 [**Positionnement 2026 : quand choisir Go (grille de décision)**](01-introduction-go/06-positionnement-2026.md) ⭐
    - [Go vs Rust / Python / Java / C# : quand choisir quoi](01-introduction-go/06.1-go-vs-autres-langages.md)
    - [GoLand vs VS Code : forces et limites de chaque IDE](01-introduction-go/06.2-goland-vs-vscode.md)

### 2. [Fondamentaux du langage](02-fondamentaux-langage/README.md)
- 2.1 [Structure d'un programme, packages, visibilité (majuscule = exporté)](02-fondamentaux-langage/01-structure-packages.md)
- 2.2 [Types de données, variables, constantes, `iota`, zéro-values](02-fondamentaux-langage/02-types-variables.md)
- 2.3 [Fonctions (retours multiples, valeurs nommées, variadiques, closures)](02-fondamentaux-langage/03-fonctions.md)
- 2.4 [Structures conditionnelles (`if` avec initialisation, `switch`, `type switch`)](02-fondamentaux-langage/04-conditions.md)
- 2.5 [Boucles (`for` unique, `range`, itérateurs `range-over-func`)](02-fondamentaux-langage/05-boucles.md) 🆕
- 2.6 [Chaînes, runes, UTF-8 (`strings`, `strconv`, `fmt`)](02-fondamentaux-langage/06-chaines.md)
- 2.7 [Tableaux, slices (capacité, `append`, pièges) et maps](02-fondamentaux-langage/07-slices-maps.md) ⭐
- 2.8 [Pointeurs (sans arithmétique), sémantique valeur vs référence](02-fondamentaux-langage/08-pointeurs.md)
- 2.9 [**Gestion des erreurs — l'idiome Go**](02-fondamentaux-langage/09-gestion-erreurs.md) ⭐
    - `error`, `errors.Is`/`errors.As`, wrapping `%w`
    - Erreurs sentinelles et erreurs personnalisées
- 2.10 [`defer`, `panic`, `recover` (et quand ne PAS les utiliser)](02-fondamentaux-langage/10-defer-panic-recover.md)

### 3. [Types, méthodes et interfaces](03-types-interfaces/README.md)
- 3.1 [Structs, méthodes, receveurs valeur vs pointeur](03-types-interfaces/01-structs-methodes.md)
- 3.2 [Composition plutôt qu'héritage (embedding)](03-types-interfaces/02-composition-embedding.md)
- 3.3 [**Interfaces implicites — le cœur du design Go**](03-types-interfaces/03-interfaces.md) ⭐
    - Petites interfaces (`io.Reader`, `io.Writer`), satisfaction implicite
- 3.4 [Génériques (contraintes, `any`, `comparable` — quand les utiliser ou pas)](03-types-interfaces/04-generiques.md)
- 3.5 [Organisation du code : packages, `internal/`, layout standard, workspaces (`go work`)](03-types-interfaces/05-organisation-code.md)

### 4. [Concurrence — le point fort de Go](04-concurrence/README.md) ⭐
- 4.1 [Goroutines et le scheduler](04-concurrence/01-goroutines.md)
- 4.2 [Channels (buffered/unbuffered, `select`, fermeture)](04-concurrence/02-channels.md)
- 4.3 [Synchronisation (`sync.WaitGroup`, `Mutex`, `Once`, `errgroup`)](04-concurrence/03-synchronisation.md)
- 4.4 [**`context.Context` (annulation, timeout, propagation)**](04-concurrence/04-context.md) ⭐
- 4.5 [Patterns : worker pool, fan-in/fan-out, pipeline](04-concurrence/05-patterns-concurrence.md)
- 4.6 [Tester le code concurrent : détecteur de races (`-race`) et `testing/synctest`](04-concurrence/06-tester-code-concurrent.md) 🆕

---

## **Partie 2 — Construire des applications**

### 5. [Backend HTTP — le scénario phare](05-backend-http/README.md) ⭐
- 5.1 [`net/http` : serveur, handlers, `ServeMux` (routing par méthode et wildcards)](05-backend-http/01-net-http.md)
- 5.2 [Middleware (chaînage, logging, recovery, CORS)](05-backend-http/02-middleware.md)
- 5.3 [JSON (`encoding/json`, validation, DTO)](05-backend-http/03-json.md)
- 5.4 [Frameworks : stdlib vs Gin / Echo / Chi (grille de choix)](05-backend-http/04-frameworks.md)
- 5.5 [API REST complète (structure, versioning, gestion d'erreurs, Problem Details)](05-backend-http/05-api-rest-complete.md)
- 5.6 [Authentification (JWT, OAuth 2.0 / OIDC, sessions)](05-backend-http/06-authentification.md)
- 5.7 [Documentation OpenAPI / Swagger](05-backend-http/07-openapi.md)

### 6. [Applications CLI et outillage](06-cli-outillage/README.md)
- 6.1 [`flag`, `os.Args`, variables d'environnement](06-cli-outillage/01-flag-args-env.md)
- 6.2 [Cobra + Viper (commandes, sous-commandes, configuration)](06-cli-outillage/02-cobra-viper.md)
- 6.3 [TUI avec Bubble Tea (notions)](06-cli-outillage/03-tui-bubbletea.md)
- 6.4 [Distribution : binaire unique, cross-compilation, GoReleaser](06-cli-outillage/04-distribution.md) ⭐

---

## **Partie 3 — Données et persistance**

### 7. [Accès aux données](07-acces-donnees/README.md) ⭐
- 7.1 [`database/sql` (pool de connexions, requêtes préparées, transactions)](07-acces-donnees/01-database-sql.md)
- 7.2 [Drivers : PostgreSQL (pgx) ⭐, MySQL, SQLite](07-acces-donnees/02-drivers.md)
- 7.3 [sqlc (SQL typé généré) vs ORM (GORM, Ent) — grille de choix](07-acces-donnees/03-sqlc-vs-orm.md)
- 7.4 [Migrations de schéma (golang-migrate, goose)](07-acces-donnees/04-migrations.md)
- 7.5 [NoSQL et cache : MongoDB, Redis](07-acces-donnees/05-nosql-redis.md)
- 7.6 [Fichiers et E/S (`io`, `bufio`, `os`, `embed`)](07-acces-donnees/06-fichiers-io.md) ⭐

---

## **Partie 4 — Services, communication et temps réel**

### 8. [Communication entre services](08-communication-services/README.md)
- 8.1 [Consommer des API REST (`http.Client`, timeouts, retries, résilience)](08-communication-services/01-consommer-api.md)
- 8.2 [**gRPC (protobuf, streaming, interceptors)**](08-communication-services/02-grpc.md) ⭐
- 8.3 [WebSockets et Server-Sent Events](08-communication-services/03-websockets-sse.md)
- 8.4 [Messaging : NATS, Kafka (notions)](08-communication-services/04-messaging.md)

---

## **Partie 5 — Cloud-native (les forces de Go)**

### 9. [Conteneurs et déploiement cloud](09-conteneurs-cloud/README.md) ⭐
- 9.1 [Dockerfile multi-stage, images distroless / scratch](09-conteneurs-cloud/01-docker.md)
- 9.2 [Kubernetes : probes, configuration, graceful shutdown](09-conteneurs-cloud/02-kubernetes.md)
- 9.3 [Serverless (AWS Lambda, Cloud Run)](09-conteneurs-cloud/03-serverless.md)

### 10. [Architecture de services](10-architecture-services/README.md)
- 10.1 [Monolithe modulaire vs microservices](10-architecture-services/01-monolithe-vs-microservices.md)
- 10.2 [Clean architecture / hexagonale en Go (sans sur-ingénierie)](10-architecture-services/02-clean-architecture.md)
- 10.3 [Configuration, feature flags, principes 12-factor](10-architecture-services/03-configuration-12factor.md)

### 11. [Interopérabilité et migration](11-interop-migration/README.md)
- 11.1 [cgo (quand l'éviter), FFI](11-interop-migration/01-cgo-ffi.md)
- 11.2 [WebAssembly (WASI)](11-interop-migration/02-webassembly.md)
- 11.3 [Migrer un service Python / Java / Node vers Go : stratégies](11-interop-migration/03-migrer-vers-go.md)

---

## **Partie 6 — Qualité, performance et exploitation**

### 12. [Erreurs, débogage et journalisation](12-erreurs-debogage/README.md)
- 12.1 [Stratégies d'erreurs à l'échelle d'une application](12-erreurs-debogage/01-strategies-erreurs.md)
- 12.2 [Débogage avec Delve (dans GoLand et VS Code)](12-erreurs-debogage/02-debogage-delve.md)
- 12.3 [**`log/slog` — journalisation structurée**](12-erreurs-debogage/03-slog.md) ⭐
- 12.4 [Observabilité (OpenTelemetry, métriques Prometheus, health checks)](12-erreurs-debogage/04-observabilite.md)

### 13. [Tests et qualité du code](13-tests-qualite/README.md)
- 13.1 [Package `testing`, table-driven tests ⭐, sous-tests](13-tests-qualite/01-tests-unitaires.md)
- 13.2 [Mocks par interfaces, testify, `httptest`](13-tests-qualite/02-mocks-testify.md)
- 13.3 [Tests d'intégration (Testcontainers, bases de données)](13-tests-qualite/03-tests-integration.md)
- 13.4 [Fuzzing natif et benchmarks (`go test -bench`)](13-tests-qualite/04-fuzzing-benchmarks.md)
- 13.5 [Linters : `go vet`, staticcheck, golangci-lint](13-tests-qualite/05-linters.md)
- 13.6 [Couverture de code ; génération de tests par IA](13-tests-qualite/06-couverture-tests-ia.md) 🤖

### 14. [Performance et gestion de la mémoire](14-performance/README.md)
- 14.1 [**Profilage avec pprof (CPU, heap, goroutines)**](14-performance/01-pprof.md) ⭐
- 14.2 [Le GC de Go, allocations, escape analysis](14-performance/02-gc-allocations.md)
- 14.3 [Optimisations : `sync.Pool`, préallocation, PGO](14-performance/03-optimisations-pgo.md) 🆕
- 14.4 [Benchmarking rigoureux (benchstat)](14-performance/04-benchmarking.md)

### 15. [Déploiement et DevOps](15-deploiement-devops/README.md)
- 15.1 [Build reproductible, versioning (ldflags), `go.mod` / `go.sum`](15-deploiement-devops/01-build-versioning.md)
- 15.2 [CI/CD (GitHub Actions, GitLab CI)](15-deploiement-devops/02-cicd.md)
- 15.3 [Sécurité de la supply chain : `govulncheck`, SBOM](15-deploiement-devops/03-supply-chain.md)

### 16. [Sécurité des applications](16-securite/README.md)
- 16.1 [OWASP appliqué à Go (injection, XSS et `html/template`, validation d'entrées)](16-securite/01-owasp-go.md)
- 16.2 [Cryptographie (`crypto/*`), TLS, gestion des secrets](16-securite/02-cryptographie-tls.md)
- 16.3 [Durcissement des services HTTP (headers, rate limiting, timeouts)](16-securite/03-durcissement-http.md)

---

## **Partie 7 — IA et avenir**

### 17. [Développer en Go avec l'IA (l'ère Copilot)](17-developpement-ia/README.md) 🤖
- 17.1 [Copilot, Claude, ChatGPT : prompting efficace pour du Go idiomatique](17-developpement-ia/01-prompting-go.md)
- 17.2 [Pièges de l'IA en Go (code non idiomatique, erreurs ignorées, sur-abstraction)](17-developpement-ia/02-pieges-ia.md)
- 17.3 [Génération de tests, migration assistée, revue de code par IA](17-developpement-ia/03-tests-migration-ia.md)
- 17.4 [Assistants intégrés : GoLand AI Assistant, Copilot dans VS Code](17-developpement-ia/04-assistants-ide.md) 🆕

### 18. [Stratégie, feuille de route et ressources](18-strategie-roadmap/README.md)
- 18.1 [Gouvernance du langage, proposals, promesse de compatibilité Go 1.x](18-strategie-roadmap/01-gouvernance-compatibilite.md)
- 18.2 [Roadmap et évolutions récentes du langage](18-strategie-roadmap/02-roadmap.md)
- 18.3 [Communauté, veille et ressources pour continuer](18-strategie-roadmap/03-communaute-ressources.md)

---

## 📎 Annexes

- **A.** [Correspondance syntaxique Go ↔ autres langages (aide-mémoire)](annexes/correspondance-go-autres/README.md)
- **B.** [Go idiomatique : *Effective Go* condensé et anti-patterns](annexes/go-idiomatique/README.md) ⭐
- **C.** [Bonnes pratiques de codage Go (+ avec l'IA 🤖)](annexes/bonnes-pratiques/README.md)
- **D.** [Raccourcis et astuces GoLand & VS Code](annexes/goland-vscode/README.md)
- **E.** [Layout de projet standard commenté](annexes/layout-projet/README.md)
- **F.** [Glossaire et acronymes](annexes/glossaire/README.md)
- **G.** [FAQ et dépannage](annexes/faq-depannage/README.md)
- **H.** [Versions Go et politique de compatibilité](annexes/versions-reference/README.md)

---

[⬅ Retour au README](README.md) · Formation Go (Golang) par **Nicolas DEOUX** · Licence [CC BY-NC-SA 4.0](LICENSE)


