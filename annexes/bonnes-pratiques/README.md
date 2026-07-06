🔝 Retour au [Sommaire](/SOMMAIRE.md)

# Annexe C — Bonnes pratiques de codage Go (+ avec l'IA 🤖)

L'[annexe B](../go-idiomatique/README.md) décrit **à quoi ressemble** du Go idiomatique — le style et les idiomes du langage. Cette annexe se place au niveau du dessus : les **pratiques et habitudes d'ingénierie** qui font tenir un projet Go dans la durée (dépendances, documentation, tests, sécurité, automatisation, revue), suivies d'un volet dédié au **travail avec un assistant IA**.

> **Comment lire cette annexe.** Ce n'est ni un rappel des idiomes (→ [annexe B](../go-idiomatique/README.md)) ni un substitut aux chapitres dédiés. Elle consolide le *quoi faire* sous forme de repères actionnables, et renvoie systématiquement au chapitre concerné pour le *pourquoi* et le *comment*. À bookmarker et à parcourir avant une revue ou une mise en production.

Exemples de code **inline** uniquement. Apports récents signalés 🆕. Cible : **Go 1.26**, stdlib avant frameworks.

---

## Discipline de code au quotidien

- **Petites unités, responsabilité unique.** Des fonctions courtes, une intention par fonction ; on extrait dès qu'un bloc mérite un nom.
- **Lisibilité avant astuce.** *Clear is better than clever* : le code est lu bien plus souvent qu'il n'est écrit.
- **Commentaires : le *pourquoi*, pas le *quoi*.** Le code dit ce qu'il fait ; le commentaire explique une décision, un compromis, un contexte non évident.
- **Conventions `TODO` / `FIXME`.** Utiles si elles sont traçables (idéalement avec un identifiant de ticket) ; sinon elles pourrissent.
- **Laisser l'outillage trancher le style.** Pas de débat de formatage : `gofmt` décide (cf. [annexe B](../go-idiomatique/README.md) et [§1.4](../../01-introduction-go/04-installation-outils.md)).

---

## Hygiène des dépendances et du module

La sobriété est une bonne pratique en soi : chaque dépendance est une surface d'attaque, une dette de maintenance et un risque de rupture.

- **La stdlib d'abord.** Beaucoup de besoins (HTTP, JSON, journalisation, tests) sont couverts sans dépendance externe. *A little copying is better than a little dependency.*
- **Évaluer avant d'ajouter.** Une dépendance se juge sur : activité et maintenance, **poids transitif** (ce qu'elle tire avec elle), qualité de l'API, **licence**, et surtout : la stdlib ne suffit-elle pas ?
- **Garder `go.mod` propre.** `go mod tidy` après chaque changement de dépendances ; ne jamais éditer `go.sum` à la main.
- **Épingler et vérifier.** Les versions sont figées dans `go.mod`/`go.sum` ; `go mod verify` contrôle l'intégrité. Le *vendoring* (`go mod vendor`) reste pertinent pour des builds hermétiques ou hors-ligne.
- **Comprendre l'arbre.** `go mod why <pkg>` et `go mod graph` expliquent *pourquoi* une dépendance est là.
- **Mettre à jour de façon maîtrisée**, pas en masse et à l'aveugle : une dépendance à la fois, tests à l'appui.

Voir aussi [§15.3 (supply chain)](../../15-deploiement-devops/03-supply-chain.md) pour `govulncheck` et le SBOM.

---

## Documentation et découvrabilité

- **Commentaires de doc** commençant par le nom de l'élément, en phrases complètes (rappel en [annexe B](../go-idiomatique/README.md)) ; ils sont rendus par `go doc` et pkg.go.dev.
- **Exemples testables** : une fonction `Example_...` sert à la fois de **documentation** et de **test** (vérifiée par `// Output:`).

```go
func ExampleGreet() {
    fmt.Println(Greet("Ada"))
    // Output: Bonjour Ada
}
```

- **README de projet et de package** : à quoi ça sert, comment l'utiliser, comment le construire et le tester.
- **Documenter la valeur zéro** : est-elle utilisable, ou faut-il un constructeur ? Le préciser évite bien des `panic`.

Ressources de découvrabilité (pkg.go.dev, `go doc`) en [§18.3](../../18-strategie-roadmap/03-communaute-ressources.md).

---

## Discipline de test

Le détail est au [chapitre 13](../../13-tests-qualite/README.md) ; les habitudes à retenir :

- **Tester le comportement, pas l'implémentation.** Un test qui casse au moindre refactor teste la mauvaise chose.
- **Tables de cas** (*table-driven tests*) avec sous-tests nommés : le style de test par défaut en Go (cf. [§13.1](../../13-tests-qualite/01-tests-unitaires.md)).
- **La couverture est un indicateur, pas un objectif.** 100 % de lignes ne garantit pas 100 % de cas utiles (cf. [§13.6](../../13-tests-qualite/06-couverture-tests-ia.md)).
- **`-race` systématique** sur le code concurrent : `go test -race ./...` (cf. [§4.6](../../04-concurrence/06-tester-code-concurrent.md)).
- **Isoler du réseau et de l'horloge** : interfaces pour les dépendances externes, `httptest` pour le HTTP, temps injecté ou virtuel.

---

## Erreurs, journaux et observabilité : les habitudes

- **Enrichir et remonter, loguer une fois.** On ajoute du contexte avec `%w` et on remonte ; le log se fait **à la frontière** (main, handler), pas à chaque étage (détail en [annexe B](../go-idiomatique/README.md) et [§12.1](../../12-erreurs-debogage/01-strategies-erreurs.md)).
- **Journalisation structurée** avec `log/slog` : des paires clé/valeur exploitables par la machine, pas des chaînes formatées à la main (cf. [§12.3](../../12-erreurs-debogage/03-slog.md)).
- **Rendre le service observable** : health checks, métriques, traces dès la conception, pas après coup (cf. [§12.4](../../12-erreurs-debogage/04-observabilite.md)).

```go
// ✅ log structuré, à la frontière, avec contexte
slog.Error("échec traitement commande", "commande_id", id, "err", err)
```

---

## Coder avec la sécurité en tête

Les détails sont au [chapitre 16](../../16-securite/01-owasp-go.md) ; les réflexes de codage :

- **Valider toute entrée externe** (requêtes, fichiers, variables d'environnement) : ne jamais faire confiance à ce qui vient du dehors (cf. [§16.1](../../16-securite/01-owasp-go.md)).
- **Aucun secret en dur, ni dans le dépôt.** Configuration par l'environnement/coffre-fort (cf. [§10.3](../../10-architecture-services/03-configuration-12factor.md), [§16.2](../../16-securite/02-cryptographie-tls.md)).
- 🆕 **Accès fichiers confiné** avec `os.Root` (Go 1.24) pour empêcher les évasions de répertoire (*path traversal*), au lieu de concaténer des chemins à la main (cf. [§7.6](../../07-acces-donnees/06-fichiers-io.md)).
- **Timeouts et `context` partout** sur les E/S réseau : pas d'appel sans délai (cf. [§4.4](../../04-concurrence/04-context.md), [§16.3](../../16-securite/03-durcissement-http.md)).
- **Scanner les vulnérabilités connues** : `govulncheck ./...` n'alerte que si du code *réellement appelé* est affecté — à intégrer au flux (cf. [§15.3](../../15-deploiement-devops/03-supply-chain.md)).

---

## Performance : mesurer avant d'optimiser

- **Profiler d'abord.** Aucune optimisation à l'aveugle : `pprof` (CPU, tas, goroutines) désigne les vrais points chauds (cf. [§14.1](../../14-performance/01-pprof.md)).
- **Benchmarks rigoureux.** `go test -bench`, comparaison avec `benchstat` pour distinguer un vrai gain du bruit (cf. [§14.4](../../14-performance/04-benchmarking.md)).
- **Éviter les micro-optimisations prématurées** qui compliquent le code sans bénéfice mesuré.

---

## Le filet automatique (pré-commit et CI)

Une grande partie de la qualité se vérifie **sans intervention humaine**. La suite à faire tourner, en local et en CI :

| Étape | Commande | Rôle |
|---|---|---|
| Format + imports | `gofmt` / `goimports` | Style unique, imports rangés |
| Analyse rapide | `go vet ./...` | Constructions suspectes |
| Analyse avancée | `golangci-lint run` | Agrège staticcheck, vet, etc. |
| Tests + courses | `go test -race ./...` | Correction + sûreté concurrente |
| Vulnérabilités | `govulncheck ./...` | CVE sur le code appelé |

- **Pré-commit local** : hook Git ou « actions à l'enregistrement » de l'IDE ; des orchestrateurs externes (`pre-commit`, `lefthook`) existent, mais les outils `go` restent le cœur.
- **CI qui rejoue tout** : la CI est l'autorité, pas la machine du développeur (cf. [§15.2](../../15-deploiement-devops/02-cicd.md)).
- 🆕 **Rester à jour** : `go fix` (Go 1.26) et ses « modernizers » réécrivent d'anciens motifs en idiomes récents ; utile en revue de dette (cf. [§13.5](../../13-tests-qualite/05-linters.md)).

**Câblage IDE (des deux côtés) :**

- *GoLand* : **Settings → Tools → Actions on Save** (reformater, optimiser les imports) ; inspections intégrées pour le reste ; configuration de test avec la case **« Enable data race detection »** ; intégration `golangci-lint` configurable.
- *VS Code* : `"editor.formatOnSave": true`, `"editor.codeActionsOnSave": { "source.organizeImports": true }`, `"go.lintTool": "golangci-lint"`, `"go.lintOnSave": "package"`, et `-race` via la configuration de test.

---

## Revue de code : viser l'essentiel

- **Laisser l'outillage gérer le style.** Une revue qui débat de mise en forme perd son temps : `gofmt` et les linters ont déjà tranché.
- **La revue humaine se concentre sur ce que les outils ne voient pas** : justesse, conception, lisibilité, sécurité, qualité des tests, gestion des erreurs.
- **Petites *pull requests*.** Plus une PR est petite, meilleure — et plus honnête — est la revue.

---

## 🤖 Coder du Go avec un assistant IA

Copilot, Claude, ChatGPT et les assistants intégrés accélèrent l'écriture, mais changent la nature du travail : on passe d'auteur à **relecteur d'un brouillon à vérifier**. Le [chapitre 17](../../17-developpement-ia/README.md) traite le sujet en profondeur ; voici les pratiques à graver.

### Le principe : brouillon d'abord, vérité jamais

Une sortie d'IA est une **proposition**, pas une référence. Bonne nouvelle propre à Go : la **boucle de vérification est peu coûteuse** (compilation rapide, `go vet` et tests intégrés, détecteur de courses). En pratique : `go build` → `go test -race` → `golangci-lint` → **relecture**. La plupart des erreurs d'IA tombent dès la compilation ou les tests.

### Checklist de relecture du Go généré par IA

Les modes d'échec typiques (code non idiomatique, erreurs ignorées, sur-abstraction) sont détaillés en [§17.2](../../17-developpement-ia/02-pieges-ia.md). À vérifier systématiquement :

- **Idiome** — erreurs explicites, petites interfaces, composition ; pas d'abstractions inventées « au cas où ». L'IA a tendance à sur-architecturer.
- **Erreurs** — aucune n'est avalée (`result, _ :=`) ni oubliée ; chacune est vérifiée et enrichie.
- **Concurrence** — pas de goroutine sans condition d'arrêt (fuite) ; validé sous `-race`.
- **API et paquets hallucinés** — le paquet **existe-t-il** (pkg.go.dev) ? La signature est-elle exacte ? Go évolue vite : l'IA propose parfois des symboles inexistants ou d'anciennes API.
- **Code périmé** — l'IA imite le corpus d'hier. Trois exemples fréquents :

```go
// ❌ ioutil est déprécié depuis Go 1.16
data, _ := ioutil.ReadFile(path)
// ✅ moderne, et l'erreur est traitée
data, err := os.ReadFile(path)

// ❌ dépendance externe que la stdlib remplace depuis Go 1.13
return pkgerrors.Wrap(err, "lecture") // github.com/pkg/errors
// ✅ stdlib
return fmt.Errorf("lecture : %w", err)

// ❌ interface{} recopié d'anciens exemples
func h(v interface{}) { /* ... */ }
// ✅ any (alias depuis Go 1.18) — ou mieux, un type concret / un générique
func h(v any) { /* ... */ }
```

- **Sécurité** — entrées validées, aucun secret en dur, pas de motif dangereux glissé sans qu'on le remarque.
- **Contexte** — `ctx` propagé sur les E/S, pas ignoré.

### Prompter pour du Go idiomatique

Le prompting efficace est détaillé en [§17.1](../../17-developpement-ia/01-prompting-go.md). Les leviers spécifiques à Go :

- **Donner la version cible** (« Go 1.26 ») pour éviter les idiomes datés.
- **Exiger la stdlib d'abord** et justifier toute dépendance externe.
- **Demander les tests** (table-driven) en même temps que le code.
- **Préciser le style** : erreurs enrichies avec `%w`, petites interfaces côté consommateur, pas de sur-abstraction.
- **Fournir du contexte réel** (signatures, types existants) plutôt qu'une demande abstraite.

### Génération de tests assistée

Utile pour dégrossir, mais **relire les cas** : l'IA couvre volontiers le chemin heureux et oublie les bords (erreurs, limites, entrées vides). La couverture générée mesure des lignes, pas la pertinence des cas (cf. [§13.6](../../13-tests-qualite/06-couverture-tests-ia.md) et [§17.3](../../17-developpement-ia/03-tests-migration-ia.md)).

### Assistants intégrés (des deux côtés)

- *GoLand* : **AI Assistant** (complétion, génération, explication, refactor guidé).
- *VS Code* : **Copilot** et assistants équivalents en extension.

Détails et comparaison en [§17.4](../../17-developpement-ia/04-assistants-ide.md).

### La responsabilité reste humaine

L'IA propose ; **vous signez**. Le code fusionné est le vôtre, avec sa dette et ses bugs. Ne validez que ce que vous comprenez et que la suite d'outils a passé au vert.

---

## Aide-mémoire : la checklist en bref

**Avant de committer**

- [ ] `gofmt`/`goimports` passés (idéalement à l'enregistrement)
- [ ] `go vet ./...` propre
- [ ] `golangci-lint run` propre
- [ ] `go test -race ./...` au vert
- [ ] `go mod tidy` si les dépendances ont bougé
- [ ] `govulncheck ./...` sans alerte
- [ ] Aucune erreur ignorée, aucun secret en dur

**Conception**

- [ ] Stdlib avant toute dépendance ; dépendance justifiée
- [ ] Petites interfaces, côté consommateur ; pas d'abstraction prématurée
- [ ] Valeur zéro documentée ; entrées externes validées
- [ ] Doc comments + au moins un exemple testable sur l'API publique

**Code généré par IA**

- [ ] Compilé, testé, `-race`, linté **avant** relecture
- [ ] Idiome, erreurs, concurrence, sécurité vérifiés
- [ ] Paquets/API confirmés sur pkg.go.dev ; pas d'API périmée (`ioutil`, `pkg/errors`, `interface{}`…)
- [ ] Compris et assumé par un humain

---

## Pour aller plus loin

- **Idiomes et anti-patterns du langage** : [annexe B](../go-idiomatique/README.md).
- **Raccourcis et astuces au quotidien** (GoLand & VS Code) : [annexe D](../goland-vscode/README.md).
- **Développer avec l'IA**, en détail : [chapitre 17](../../17-developpement-ia/README.md).
- **Tests et qualité** : [chapitre 13](../../13-tests-qualite/README.md) · **Sécurité** : [chapitre 16](../../16-securite/01-owasp-go.md) · **Supply chain** : [§15.3](../../15-deploiement-devops/03-supply-chain.md).
- **Ressources et veille** : [§18.3](../../18-strategie-roadmap/03-communaute-ressources.md).

---

🔝 [Retour au sommaire](../../SOMMAIRE.md) · ⏭️ [Annexe D — Raccourcis et astuces GoLand & VS Code](../goland-vscode/README.md)


⏭ [Raccourcis et astuces GoLand & VS Code](/annexes/goland-vscode/README.md)
