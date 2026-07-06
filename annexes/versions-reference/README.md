🔝 Retour au [Sommaire](/SOMMAIRE.md)

# Annexe H — Versions Go et politique de compatibilité

Référence factuelle sur les versions de Go : cadence de publication, fenêtre de support, promesse de compatibilité et repères d'historique. La **gouvernance** du langage (proposals, processus, philosophie de la compatibilité) est traitée en [§18.1](../../18-strategie-roadmap/01-gouvernance-compatibilite.md) ; cette annexe en est le **complément chiffré**.

> Les dates et l'état du support ci-dessous ont été vérifiés sur go.dev à la rédaction (mi-2026) ; voir les [sources](#sources) en fin de page. Les versions évoluant, recoupez au besoin avec la page officielle.

---

## Cadence de publication

Go suit un rythme semestriel stable : **deux versions majeures par an**, l'une vers **février**, l'autre vers **août**, à six mois d'intervalle. Entre deux majeures, des **versions correctives** (`goX.Y.Z`) apportent les corrections de sécurité et les bugs critiques, environ une fois par mois.

Le schéma de version est `go1.MAJEURE.CORRECTIF` (ex. `go1.26.0`). Particularité de terminologie : c'est le **deuxième** nombre qui est la « majeure » — dans `go1.26.0`, la majeure est `26`. Go prévoit de **ne jamais atteindre la version 2.0**, en faisant primer la compatibilité ascendante sur les ruptures. Voir aussi le cycle de release en [§1.3](../../01-introduction-go/03-ecosysteme-go.md).

---

## Fenêtre de support

Chaque version majeure est maintenue (corrections de sécurité et bugs critiques) **jusqu'à ce que deux majeures plus récentes existent** — en pratique, les **deux dernières majeures**.

| Version | Sortie | Statut (mi-2026) |
|---|---|---|
| **Go 1.26** | février 2026 | **Courante** — supportée |
| **Go 1.25** | août 2025 | **Précédente** — supportée |
| Go 1.24 | février 2025 | Fin de support (deux majeures plus récentes) |
| Go 1.23 et antérieures | — | Non supportées |

À la rédaction, les correctifs les plus récents étaient **go1.26.4** et **go1.25.11** (2 juin 2026) ; la prochaine majeure, Go 1.27, est attendue vers **août 2026**.

**Conséquence pratique** : faites tourner une version supportée. Une version hors support ne reçoit plus de correctifs de sécurité — un risque à part entière (cf. [§15.3](../../15-deploiement-devops/03-supply-chain.md)).

---

## La promesse de compatibilité Go 1

Énoncé de référence : **un programme qui compile et s'exécute sous une version Go 1.x doit continuer de compiler et s'exécuter sous les versions Go 1.x ultérieures**, pour la spécification du langage et l'essentiel de la bibliothèque standard. Cette promesse a été tenue de Go 1.0 (2012) jusqu'à Go 1.26.

C'est ce qui rend les montées de version **peu risquées** et explique la conservatisme du versionnement (jamais de 2.0). Le *pourquoi* et le processus sont en [§18.1](../../18-strategie-roadmap/01-gouvernance-compatibilite.md).

Ce que la promesse **ne couvre pas** (à connaître) : les correctifs de **sécurité** peuvent modifier un comportement ; les comportements **non spécifiés** (ordre d'itération d'une map) ne sont pas garantis ; les **bugs** peuvent être corrigés ; les **prérequis de plateforme** évoluent (par ex. version minimale de macOS relevée d'une version à l'autre) ; enfin, certaines évolutions de comportement sont **pilotées par `GODEBUG`** (voir ci-dessous).

---

## Faire évoluer le langage sans casser le code

Go concilie compatibilité et évolution grâce à quatre mécanismes — utiles à comprendre pour maîtriser ses montées de version.

- **Directive `go` du `go.mod`** — fixe la **version de langage** du module. Les nouvelles **sémantiques de langage** sont conditionnées à cette ligne : par exemple, la variable de boucle propre à chaque itération (Go 1.22) ne s'applique que si le `go.mod` déclare `go 1.22` ou plus. Mettre à jour la *toolchain* ne change donc pas silencieusement le comportement de votre code.
- **`GODEBUG`** — les changements de comportement du *runtime* ou de la stdlib peuvent être rétablis à l'ancien via des réglages `GODEBUG` ; la ligne `go` du module sélectionne des valeurs par défaut compatibles avec sa version. (Mécanisme formalisé en Go 1.21.)
- **Gestion de la *toolchain* (Go 1.21+)** — les directives `go` et `toolchain` du `go.mod`, combinées à `GOTOOLCHAIN` (défaut `auto`), permettent à la commande `go` de **télécharger et utiliser la toolchain exacte** requise par un module. `GOTOOLCHAIN=local` désactive ce téléchargement automatique (cf. [annexe G](../faq-depannage/README.md)).
- **`GOEXPERIMENT`** — active ou désactive des fonctionnalités **expérimentales** (ex. `GOEXPERIMENT=jsonv2` pour activer la nouvelle implémentation JSON, `GOEXPERIMENT=nogreenteagc` pour revenir à l'ancien GC).

---

## Quelle version cibler ?

- `go mod init` inscrit la version **N-1** : sous Go 1.26, il écrit `go 1.25.0` (avec le `.0`) — cf. [§1.5](../../01-introduction-go/05-premier-projet.md). Ce choix conservateur maximise la compatibilité.
- Utilisez une version **supportée** et **récente** ; appliquez les correctifs (`goX.Y.Z`) et faites tourner `govulncheck` (cf. [§15.3](../../15-deploiement-devops/03-supply-chain.md)).
- Inutile de courir après la toute dernière fonctionnalité : la promesse de compatibilité rend la montée en version simple quand elle devient utile.

---

## Historique des versions (repères)

| Version | Date | Apport marquant |
|---|---|---|
| 1.0 | mars 2012 | Promesse de compatibilité Go 1 |
| 1.5 | août 2015 | Compilateur auto-hébergé, GC concurrent |
| 1.11 | août 2018 | Modules (expérimental) |
| 1.13 | septembre 2019 | Enveloppe d'erreurs (`%w`, `errors.Is`/`As`), proxy de modules |
| 1.16 | février 2021 | `embed`, modules par défaut, `io/ioutil` déprécié |
| 1.18 | mars 2022 | **Génériques**, fuzzing natif, workspaces |
| 1.20 | février 2023 | `errors.Join`, aperçu de PGO |
| 1.21 | août 2023 | `min`/`max`/`clear`, `log/slog`, `slices`/`maps`, PGO stable, gestion de toolchain |
| 1.22 | février 2024 | Variable de boucle par itération, `range` sur entier, routage `net/http` enrichi |
| 1.23 | août 2024 | Itérateurs `range`-over-func, package `unique` |
| 1.24 | février 2025 | Alias de types génériques, refonte des maps, directives `tool`, `os.Root` |
| 1.25 | août 2025 | `testing/synctest` (GA), Green Tea GC (expérimental), `GOMAXPROCS` *container-aware*, flight recorder |
| 1.26 | février 2026 | Green Tea GC par défaut, post-quantique (ML-KEM, `crypto/hpke`), refonte de `go fix` |

---

## Nouveautés récentes 1.21 → 1.26 (détail)

Consolidation des évolutions intégrées tout au long de la formation, avec le renvoi vers la section qui les exploite.

| Version | Nouveautés clés (et renvois) |
|---|---|
| **1.21** (08/2023) | Builtins `min`/`max`/`clear` ; `log/slog` (§12.3) ; packages `slices`/`maps` ; PGO stable (§14.3) ; gestion de *toolchain* et compatibilité `GODEBUG` |
| **1.22** (02/2024) | Variable de boucle par itération (§4.1, annexe B) ; `for range n` sur entier (§2.5) ; routage `net/http` par méthode et *wildcards* (§5.1) |
| **1.23** (08/2024) | Itérateurs `range`-over-func (§2.5) ; package `unique` ; améliorations des `time.Timer`/`Ticker` |
| **1.24** (02/2025) | Alias de types génériques (§3.4) ; refonte des maps (*Swiss Tables*) ; directives `tool` dans `go.mod` ; `os.Root` (§7.6) ; `testing.B.Loop` |
| **1.25** (08/2025) | `testing/synctest` **stable** (§4.6) ; Green Tea GC **expérimental** (§14.2) ; `GOMAXPROCS` *container-aware* (§9.2) ; flight recorder (§14.1) ; `encoding/json/v2` en expérimentation (§5.3) ; DWARF 5 |
| **1.26** (02/2026) | Green Tea GC **par défaut** — 10-40 % d'overhead GC en moins, gains supplémentaires sur AVX-512 (§14.2) ; **post-quantique** : ML-KEM hybride par défaut, `crypto/hpke` (§16.2) ; refonte de `go fix` (*modernizers*, `//go:fix inline`) (§13.5) ; `new(expr)` (§2.2) ; génériques auto-référentiels (§3.4) ; `slog.NewMultiHandler` (§12.3) ; `io.ReadAll` ~2× plus rapide (§7.6) ; overhead cgo réduit d'~30 % (§11.1) ; retrait de l'ancienne API `synctest` |

> Options d'opt-out utiles : `GOEXPERIMENT=nogreenteagc` (ancien GC), `GOEXPERIMENT=jsonv2` (activer JSON v2). L'`encoding/json/v2` reste **expérimental** à ce stade : à considérer en perspective, pas en production.

---

## Vérifier et gérer sa version (des deux côtés)

- **Connaître sa version** : `go version`.
- **Plusieurs versions en parallèle** : les wrappers officiels par version — `go install golang.org/dl/go1.X@latest` puis `go1.X download` — ou le mécanisme `GOTOOLCHAIN` décrit plus haut.
- **GoLand** : *Settings → Go → GOROOT* pour choisir ou télécharger le SDK ; la version du `go.mod` est affichée et gérée par l'IDE.
- **VS Code** : l'extension Go utilise le `go` du `PATH` ; commande *Go: Locate Configured Go Tools* pour vérifier l'installation.

Raccourcis et réglages IDE : [annexe D](../goland-vscode/README.md) ; installation détaillée : [§1.4](../../01-introduction-go/04-installation-outils.md).

---

## Sources

- Historique des versions : [go.dev/doc/devel/release](https://go.dev/doc/devel/release)
- Notes de version : [Go 1.24](https://go.dev/doc/go1.24) · [Go 1.25](https://go.dev/doc/go1.25) · [Go 1.26](https://go.dev/doc/go1.26)
- Gouvernance et promesse de compatibilité (dans la formation) : [§18.1](../../18-strategie-roadmap/01-gouvernance-compatibilite.md)

---

## Pour aller plus loin

- **Gouvernance, proposals, promesse de compatibilité** : [§18.1](../../18-strategie-roadmap/01-gouvernance-compatibilite.md).
- **Roadmap et évolutions à venir** : [§18.2](../../18-strategie-roadmap/02-roadmap.md).
- **Écosystème et cycle de release** : [§1.3](../../01-introduction-go/03-ecosysteme-go.md).
- **Gestion de version au quotidien / dépannage toolchain** : [annexe G](../faq-depannage/README.md).
- **Glossaire** (semver, toolchain, GODEBUG…) : [annexe F](../glossaire/README.md).

---

🔝 [Retour au sommaire](../../SOMMAIRE.md) · ⬅ [Retour au README](../../README.md)

*Dernière annexe de la formation.*

⏭ Retour au [Sommaire](/SOMMAIRE.md)
