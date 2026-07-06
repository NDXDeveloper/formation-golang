# 💻 Formation Go (Golang) — GoLand & VS Code

![License](https://img.shields.io/badge/License-CC%20BY--NC--SA%204.0-blue.svg)  
![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8.svg)  
![GoLand](https://img.shields.io/badge/GoLand-2026-black.svg)  
![VS Code](https://img.shields.io/badge/VS%20Code-Extension%20Go-007ACC.svg)  
![Modules](https://img.shields.io/badge/Modules-18-green.svg)  
![Language](https://img.shields.io/badge/Langue-Français-blue.svg)  
![Mise à jour](https://img.shields.io/badge/Mise%20à%20jour-Juillet%202026-brightgreen.svg)

**Un guide progressif pour apprendre Go là où il excelle réellement : backend, concurrence, cloud-native et outillage CLI — du débutant au professionnel.**

<p align="center">
  <img src="https://go.dev/images/go-logo-blue.svg" alt="Go Logo" width="200"/>
</p>

---

## 📖 Table des matières

- [À propos](#-à-propos)
- [Points forts](#-points-forts)
- [Pour qui ?](#-pour-qui-)
- [Contenu](#-contenu-de-la-formation)
- [Démarrage](#-démarrage-rapide)
- [Structure](#-structure-du-projet)
- [Parcours](#-parcours-dapprentissage-suggéré)
- [Ressources](#-ressources-officielles)
- [FAQ](#-faq)
- [Licence](#-licence)
- [Contact](#-contact)

---

## 📋 À propos

Formation centrée sur **ce que Go fait réellement bien** : services backend HTTP, programmation  
concurrente (goroutines, channels), applications cloud-native (Docker, Kubernetes, serverless),  
outils en ligne de commande, gRPC et communication entre services — avec un volet complet sur le
**développement assisté par IA**.

Go est un langage **volontairement simple et stable** : la promesse de compatibilité Go 1.x  
garantit que le code écrit aujourd'hui compilera encore dans dix ans. Cette formation en tire  
parti : elle enseigne le **Go idiomatique** (gestion d'erreurs explicite, composition, petites  
interfaces) plutôt que de plaquer des habitudes venues d'autres langages. Elle couvre les deux  
environnements de travail majeurs : **GoLand** (JetBrains) et **VS Code** avec l'extension  
officielle Go.

> Cette formation ne cherche pas à promouvoir Go face à d'autres langages, mais à fournir une
> ressource sérieuse et à jour pour apprendre à construire des services robustes et performants.
> Elle n'est pas parfaite ; l'objectif est de vous faire gagner du temps.

**✨ Ce que vous y trouverez :**

- 📚 **18 modules progressifs** — du langage aux services cloud-native, en passant par les données et les tests
- 🆕 **Go 1.26** — version de référence, avec les apports récents (itérateurs, PGO, `ServeMux` moderne)
- 🧭 **Go idiomatique** — erreurs explicites, composition, interfaces implicites : penser en Go, pas en Java/C# traduit
- ⚡ **Concurrence** — goroutines, channels, `context`, patterns (worker pool, pipeline) : le cœur de Go
- 🌐 **Backend & cloud-native** — `net/http`, gRPC, Docker multi-stage, Kubernetes, serverless
- 🔧 **Outillage double** — GoLand et VS Code couverts en parallèle (débogage Delve, refactoring, tests)
- 🤖 **Développement assisté par IA** — un module complet (GitHub Copilot, Claude, ChatGPT)
- 📖 **8 annexes** — Go idiomatique, layout de projet, aide-mémoire multi-langages, FAQ
- 🧪 **Exemples exécutables** — chaque module a un dossier `exemples/` de code compilé, testé et commenté
- 🇫🇷 **En français** — parce que c'est plus accessible

**Durée estimée :** ~30-40 h de lecture théorique • 22-28 jours pour le parcours complet en pratiquant les exemples  
**Niveau :** débutant à confirmé

---

## ✨ Points forts

- ✅ **Idiomatique** — on enseigne le Go tel qu'il se pratique, pas un autre langage avec la syntaxe de Go
- ✅ **Recentrée** — backend HTTP ⭐, concurrence ⭐, cloud-native, CLI : le terrain réel de Go
- ✅ **À jour** — Go 1.26, GoLand 2026 et l'extension VS Code officielle (gopls, dlv)
- ✅ **Double IDE** — chaque sujet outillage est traité pour GoLand **et** VS Code
- ⚡ **Concurrence au centre** — un module complet et des patterns réutilisables (Module 4)
- 🤖 **IA-first** — un module complet sur le développement assisté et ses pièges en Go
- 🛠️ **Pragmatique** — stdlib d'abord, frameworks quand c'est justifié, grilles de choix (sqlc vs ORM, Gin vs stdlib)
- 🧩 **Production-ready** — tests, profilage pprof, observabilité, CI/CD, sécurité supply chain

---

## 🎯 Pour qui ?

| 👤 Profil | 📚 Modules recommandés | ⏱️ Durée estimée |
|-----------|------------------------|------------------|
| **Débutant** | 1-5, 12 | 6-8 jours |
| **Backend / API** ⭐ | 1-5, 7-9, 12-13, 15-16 | 11-13 jours |
| **CLI & Outillage** | 1-4, 6, 12-13, 15 | 7-9 jours |
| **Cloud / DevOps** | 1-4, 8-12, 15-16 | 11-13 jours |
| **IA-First** 🤖 | 1-2, 5, 17, 18 | 4-6 jours |
| **Formation complète** | 1-18 + annexes | 22-28 jours |

---

## 📚 Contenu de la formation

### Les 7 parties

| # | Partie | Modules | Niveau | Sujets clés |
|---|--------|---------|--------|-------------|
| **1** | Comprendre Go en 2026 (cadrage & langage) | 1-4 | 🌱 Débutant | Positionnement, types, erreurs, interfaces, **concurrence** ⭐ |
| **2** | Construire des applications | 5-6 | 🌿 Intermédiaire | Backend HTTP ⭐, API REST, CLI (Cobra), distribution |
| **3** | Données et persistance | 7 | 🌿 Intermédiaire | `database/sql`, pgx, sqlc vs ORM, migrations, `embed` |
| **4** | Services, communication et temps réel | 8 | 🌿 Intermédiaire | gRPC ⭐, WebSockets, messaging (NATS, Kafka) |
| **5** | Cloud-native (les forces de Go) | 9-11 | 🌳 Avancé | Docker, Kubernetes, serverless, architecture, migration vers Go |
| **6** | Qualité, performance et exploitation | 12-16 | 🌳 Avancé | slog, tests, pprof, CI/CD, sécurité |
| **7** | IA et avenir | 17-18 | 🌳 Avancé | **Développement assisté par IA** 🤖, gouvernance, roadmap |

### Les modules

1. Introduction : Go et son écosystème
2. Fondamentaux du langage
3. Types, méthodes et interfaces
4. **Concurrence — le point fort de Go** ⭐
5. **Backend HTTP — le scénario phare** ⭐
6. Applications CLI et outillage
7. Accès aux données (`database/sql`, pgx, sqlc)
8. Communication entre services (gRPC, WebSockets)
9. Conteneurs et déploiement cloud
10. Architecture de services
11. Interopérabilité et migration
12. Erreurs, débogage et journalisation (`log/slog`)
13. Tests et qualité du code
14. Performance et gestion de la mémoire (pprof)
15. Déploiement et DevOps
16. Sécurité des applications
17. **Développer en Go avec l'IA (l'ère Copilot)** 🤖
18. Stratégie, feuille de route et ressources

### 🔥 Apports récents de Go couverts

**Nouveautés Go 1.26 (version de référence) :**

- 🍵 **GC « Green Tea »** — nouveau ramasse-miettes activé par défaut : 10 à 40 % d'*overhead* GC en moins, sans changement de code
- 🔐 **Cryptographie post-quantique** — échanges TLS hybrides ML-KEM **par défaut**, et le nouveau paquet `crypto/hpke`
- 🛠️ **Modernizers `go fix`** — réécriture automatique du « Go de 2020 » vers les idiomes récents (`any`, `min`/`max`, `maps.Copy`…)
- 🧩 **`errors.AsType`, `os.Root`, `new(expr)`** — extraction d'erreurs sans pointeur de sortie, accès fichiers confiné, `new` avec valeur initiale

**Depuis Go 1.21-1.24 :**

- 🔁 **Itérateurs** — `range-over-func`, package `iter`
- 🌐 **`net/http` moderne** — routing par méthode HTTP et wildcards dans `ServeMux` (adieu les routeurs obligatoires)
- ⚡ **PGO** (Profile-Guided Optimization) — gains de performance guidés par profils de production
- 📊 **`log/slog`** — journalisation structurée dans la bibliothèque standard
- 🧪 **Fuzzing et benchmarks natifs** — `go test -fuzz`, benchstat
- 🔒 **Supply chain** — `govulncheck`, builds reproductibles, SBOM

### 📎 Les 8 annexes

- **A.** Correspondance syntaxique Go ↔ autres langages (aide-mémoire)
- **B.** **Go idiomatique** : *Effective Go* condensé et anti-patterns ⭐
- **C.** Bonnes pratiques de codage Go (+ avec l'IA 🤖)
- **D.** Raccourcis et astuces GoLand & VS Code
- **E.** Layout de projet standard commenté
- **F.** Glossaire et acronymes
- **G.** FAQ et dépannage
- **H.** Versions Go et politique de compatibilité

📖 **Sommaire détaillé** → [SOMMAIRE.md](/SOMMAIRE.md)

---

## 🚀 Démarrage rapide

### Installation de Go et des IDE

```bash
# Vérifier la version Go installée
go version

# Installer Go 1.26
# Windows / macOS / Linux : https://go.dev/dl/
# macOS : brew install go
# Linux (Debian/Ubuntu) : voir https://go.dev/doc/install

# GoLand (essai 30 jours, gratuit pour étudiants/open source)
# https://www.jetbrains.com/goland/

# VS Code + extension officielle Go
# https://code.visualstudio.com/
# ✅ Installer l'extension "Go" (golang.go), puis : Go: Install/Update Tools
```

### Créer votre premier projet Go

```bash
# Initialiser un module
mkdir mon-premier-projet && cd mon-premier-projet
go mod init github.com/votre-user/mon-premier-projet

# Créer main.go, puis :
go run .

# Compiler un binaire unique (la force de Go)
go build

# Cross-compilation (ex. binaire Linux depuis Windows/macOS)
GOOS=linux GOARCH=amd64 go build

# Lancer les tests
go test ./...

# ✅ Vous êtes prêt !
```

> 💡 Pas de runtime à installer chez l'utilisateur final : Go produit un **binaire unique
> statique**. C'est l'une des raisons de son succès pour les CLI et le cloud.

### Cloner cette formation

```bash
git clone https://github.com/NDXDeveloper/formation-golang.git
cd formation-golang
```

---

## 📁 Structure du projet

```
formation-golang/
│
├── 📄 README.md                      # Ce fichier
├── 📄 SOMMAIRE.md                    # Table des matières complète (source de vérité)
├── 📄 LICENSE                        # Licence CC BY-NC-SA 4.0
│
├── 📂 01-introduction-go/            # Écosystème, GoLand & VS Code, positionnement
├── 📂 02-fondamentaux-langage/       # Types, slices, maps, erreurs, defer
├── 📂 03-types-interfaces/           # Structs, composition, interfaces, génériques
├── 📂 04-concurrence/                # ⭐ Goroutines, channels, context, patterns
├── 📂 05-backend-http/               # ⭐ net/http, middleware, API REST, auth
├── 📂 06-cli-outillage/              # Cobra, Viper, Bubble Tea, GoReleaser
├── 📂 07-acces-donnees/              # database/sql, pgx, sqlc, migrations
├── 📂 08-communication-services/     # gRPC, WebSockets, NATS, Kafka
├── 📂 09-conteneurs-cloud/           # Docker, Kubernetes, serverless
├── 📂 10-architecture-services/      # Clean archi, 12-factor, configuration
├── 📂 11-interop-migration/          # cgo, WebAssembly, migrer vers Go
├── 📂 12-erreurs-debogage/           # Stratégies d'erreurs, Delve, slog
├── 📂 13-tests-qualite/              # testing, testify, fuzzing, linters
├── 📂 14-performance/                # pprof, GC, escape analysis, PGO
├── 📂 15-deploiement-devops/         # Build, CI/CD, govulncheck, SBOM
├── 📂 16-securite/                   # OWASP, crypto, TLS, durcissement
├── 📂 17-developpement-ia/           # 🤖 Développer avec l'IA (l'ère Copilot)
├── 📂 18-strategie-roadmap/          # Gouvernance, compatibilité, ressources
│
└── 📂 annexes/
    ├── correspondance-go-autres/     # A. Aide-mémoire multi-langages
    ├── go-idiomatique/               # B. ⭐ Effective Go condensé
    ├── bonnes-pratiques/             # C.
    ├── goland-vscode/                # D.
    ├── layout-projet/                # E.
    ├── glossaire/                    # F.
    ├── faq-depannage/                # G.
    └── versions-reference/           # H.
```

> Chaque dossier contient un `README.md` (sommaire du module) et un fichier `.md` par section,
> reliés entre eux par une navigation 🔝 Sommaire / ⏭️ section suivante — plus un sous-dossier
> `exemples/` de code exécutable, compilé et testé (avec son propre `README.md` et les sorties attendues).

---

## 🗓️ Parcours d'apprentissage suggéré

```
🌱 DÉBUTANT
│
├─ Partie 1 : Comprendre Go (langage, interfaces, concurrence)
└─ Partie 2 : Construire des applications (backend HTTP, CLI)
   │
   ▼
🌿 INTERMÉDIAIRE
│
├─ Partie 3 : Données et persistance (database/sql, sqlc)
└─ Partie 4 : Services et communication (gRPC, WebSockets)
   │
   ▼
🌳 AVANCÉ
│
├─ Partie 5 : Cloud-native (Docker, Kubernetes, architecture)
├─ Partie 6 : Qualité, performance et exploitation
└─ Partie 7 : IA et avenir (développement assisté, roadmap)

🎓 Total : ~30-40 h de lecture (22-28 jours en pratiquant les exemples, à 30 min-1 h/jour)
```

**🎯 Parcours Express (3-4 jours)** pour les pressés :
- Module 1 — Introduction et positionnement
- Module 2 — Fondamentaux (dont la gestion des erreurs)
- Module 4 — Concurrence
- Module 5 — Backend HTTP
- Module 17 — Développer avec l'IA

---

## 🔗 Ressources officielles

| Ressource | Lien |
|-----------|------|
| 📖 Documentation Go | [go.dev/doc](https://go.dev/doc/) |
| 🎓 Tour of Go (interactif) | [go.dev/tour](https://go.dev/tour/) |
| 📘 Effective Go | [go.dev/doc/effective_go](https://go.dev/doc/effective_go) |
| 📦 Index des packages | [pkg.go.dev](https://pkg.go.dev) |
| 🧭 Blog officiel Go | [go.dev/blog](https://go.dev/blog/) |
| ⚙️ GoLand | [jetbrains.com/goland](https://www.jetbrains.com/goland/) |
| ⚙️ Extension Go pour VS Code | [marketplace.visualstudio.com](https://marketplace.visualstudio.com/items?itemName=golang.go) |
| 🤖 GitHub Copilot | [github.com/features/copilot](https://github.com/features/copilot) |
| 📥 Télécharger Go | [go.dev/dl](https://go.dev/dl/) |

---

## ❓ FAQ

**Q : Faut-il GoLand (payant) ou VS Code suffit-il ?**
> VS Code + l'extension officielle Go est **gratuit et largement suffisant** pour apprendre et
> travailler. GoLand apporte un confort supérieur (refactoring, inspections, débogueur intégré).
> La formation couvre les deux ; l'Annexe D compare en détail.

**Q : Go convient-il aux débutants en programmation ?**
> Oui — c'est même un de ses atouts : peu de mots-clés, une seule façon d'écrire les boucles,
> formatage automatique (`gofmt`). Le parcours Débutant est conçu pour ça.

**Q : Pourquoi pas de module desktop / mobile ?**
> Ce n'est pas le terrain de Go. Cette formation enseigne Go **là où il excelle** : backend,
> CLI, cloud-native. Pour du desktop ou du mobile, d'autres écosystèmes sont plus adaptés.

**Q : Faut-il apprendre un framework web (Gin, Echo) ?**
> Pas d'emblée. Depuis les évolutions de `ServeMux`, la **stdlib couvre la majorité des besoins**.
> Le Module 5 enseigne d'abord `net/http`, puis donne une grille de choix pour les frameworks.

**Q : Mon code Go restera-t-il compatible dans le temps ?**
> Oui — la **promesse de compatibilité Go 1.x** garantit que le code valide continue de compiler
> avec les nouvelles versions. Détails au Module 18.

**Q : La formation couvre-t-elle le développement avec l'IA ?**
> Oui — **Module 17** complet : prompting pour du Go idiomatique, pièges fréquents (erreurs
> ignorées, code sur-abstrait), génération de tests et migration assistée.

**Q : Cette formation remplace-t-elle la documentation officielle ?**
> Non, elle la complète. C'est un guide d'apprentissage progressif, pas une référence exhaustive.

**Q : Puis-je utiliser ce contenu pour enseigner ?**
> Oui, sous licence CC BY-NC-SA 4.0 — attribution requise, usage non commercial, partage identique.

---

## 📝 Licence

Ce projet est sous licence **Creative Commons Attribution - Pas d'Utilisation Commerciale - Partage dans les Mêmes Conditions 4.0 International (CC BY-NC-SA 4.0)**.

✅ **Vous pouvez :**
- Partager — copier et redistribuer le matériel
- Adapter — remixer, transformer et créer à partir du matériel

⚠️ **Selon les conditions suivantes :**
- **Attribution** — vous devez créditer l'œuvre originale
- **Pas d'Utilisation Commerciale** — pas d'usage à des fins commerciales
- **Partage dans les Mêmes Conditions** — toute redistribution sous la même licence

📄 Voir le fichier [LICENSE](/LICENSE) pour les détails complets.

**Attribution suggérée :**
```
Formation Go (Golang) avec GoLand & VS Code par Nicolas DEOUX
https://github.com/NDXDeveloper/formation-golang
Licence CC BY-NC-SA 4.0
```

---

## 👨‍💻 Contact

**Nicolas DEOUX**
- 📧 [NDXDev@gmail.com](mailto:NDXDev@gmail.com)
- 💼 [LinkedIn](https://www.linkedin.com/in/nicolas-deoux-ab295980/)
- 🐙 [GitHub](https://github.com/NDXDeveloper)

---

## 🙏 Remerciements

Merci à :
- **L'équipe Go** chez Google et les contributeurs du projet
- **JetBrains** pour GoLand et **Microsoft** pour VS Code
- La **communauté Go** francophone et internationale
- **Anthropic**, **OpenAI** et les créateurs d'outils IA qui transforment le développement
- **Vous** qui prenez le temps d'apprendre avec cette formation

---

<div align="center">

## 💻 Bon apprentissage avec Go ! 💻

*Cette formation est un travail en cours. Elle n'est pas parfaite, mais j'espère sincèrement qu'elle vous sera utile dans votre parcours d'apprentissage.*

**[📖 Consulter le sommaire complet →](/SOMMAIRE.md)**

[![Star on GitHub](https://img.shields.io/github/stars/NDXDeveloper/formation-golang?style=social)](https://github.com/NDXDeveloper/formation-golang)
[![Follow](https://img.shields.io/github/followers/NDXDeveloper?style=social)](https://github.com/NDXDeveloper)

**[⬆ Retour en haut](#-formation-go-golang--goland--vs-code)**

*Dernière mise à jour : Juillet 2026*

</div>

