🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 1. Introduction : Go et son écosystème

> **Partie 1 — Comprendre Go en 2026 (cadrage & langage)**
> Module d'ouverture : on pose le décor avant d'écrire la moindre ligne de logique métier.

Bienvenue dans le premier module de cette formation. Avant de plonger dans la syntaxe, il faut répondre à trois questions simples : **qu'est-ce que Go**, **à quoi sert-il réellement en 2026**, et **comment mettre en place un environnement de travail productif** — sous GoLand comme sous VS Code.

Go (souvent écrit « Golang », notamment pour les moteurs de recherche) est un langage compilé, à typage statique et doté d'un ramasse-miettes, conçu chez Google autour d'une idée directrice : **la simplicité au service de la lisibilité et de la maintenance à grande échelle**. En 2026, il s'est imposé comme une valeur sûre pour trois familles d'usages — le **backend et les API**, les **outils en ligne de commande (CLI)** et le **cloud-native** : Docker, Kubernetes, Terraform ou encore Prometheus sont eux-mêmes écrits en Go.

Ce module ne cherche pas encore à vous rendre productif à l'écriture — c'est l'objet des modules suivants. Il vise à vous donner une **carte mentale claire** : d'où vient Go, comment fonctionne son outillage, et surtout **quand** le choisir plutôt qu'un autre langage.

## 🎯 Objectifs du module

À l'issue de ce module, vous serez capable de :

- situer Go dans le paysage technologique de 2026 et identifier ses terrains de prédilection ;
- expliquer la philosophie du langage (simplicité, lisibilité, stabilité) et ce qu'elle implique au quotidien ;
- décrire la toolchain (`go build`, `go run`, `go test`, `go vet`, `gofmt`) et le cycle de release semestriel ;
- installer et configurer un environnement de développement complet, **sous GoLand et sous VS Code** ;
- créer et exécuter un premier projet à partir de `go mod init` ;
- appliquer une **grille de décision** pour choisir — ou écarter — Go en connaissance de cause.

## 📋 Prérequis

Aucune connaissance préalable de Go n'est requise. Une expérience de la programmation dans un autre langage (Python, Java, JavaScript, C#…) et une aisance minimale avec le terminal suffisent à suivre confortablement. L'installation détaillée des outils est traitée en section 1.4.

## 🧭 Contexte de version

Cette formation cible **Go 1.26** (juin 2026), la version stable la plus récente. Go suit un **rythme de publication semestriel** — une version majeure environ tous les six mois — assorti d'une **promesse de compatibilité** pour la branche Go 1.x : le code idiomatique écrit aujourd'hui continuera de compiler demain. Les apports récents du langage et de l'outillage sont signalés par le repère 🆕 tout au long des sections.

## 🗺️ Plan du module

| §   | Section | En bref |
|-----|---------|---------|
| 1.1 | [Qu'est-ce que Go et à quoi il sert réellement](01-quest-ce-que-go.md) | Backend, CLI, cloud-native : les usages concrets, sans marketing. |
| 1.2 | [Histoire et philosophie du langage](02-histoire-philosophie.md) | Les origines chez Google et les principes de conception (simplicité, lisibilité). |
| 1.3 | [L'écosystème Go](03-ecosysteme-go.md) | Toolchain, modules et cycle de release semestriel. |
| 1.4 | [Installation et outils](04-installation-outils.md) | GoLand, VS Code (extension officielle, `gopls`, `dlv`) et la CLI de base. |
| 1.5 | [Premier projet pas à pas](05-premier-projet.md) | `go mod init`, un « hello world » exécuté de bout en bout. |
| 1.6 | [**Positionnement 2026 : quand choisir Go**](06-positionnement-2026.md) ⭐ | La grille de décision du module — à connaître avant de s'engager. |

La section 1.6 se décline en deux volets complémentaires :

- 1.6.1 — [Go vs Rust / Python / Java / C# : quand choisir quoi](06.1-go-vs-autres-langages.md)
- 1.6.2 — [GoLand vs VS Code : forces et limites de chaque IDE](06.2-goland-vs-vscode.md)

## 💡 Comment aborder ce module

- **Approche idiomatique d'abord.** Toute la formation privilégie le Go idiomatique et la bibliothèque standard avant les frameworks. Ce module en pose l'état d'esprit : gardez-le en tête dès le premier exemple.
- **Deux IDE, à parité.** Chaque point d'outillage est traité pour **GoLand** *et* pour **VS Code**. Choisissez celui qui vous convient : rien dans la suite ne dépend d'un éditeur en particulier.
- **Des repères visuels.** ⭐ signale un contenu essentiel, 🆕 un apport récent du langage ou de l'outillage, 🤖 un point lié au développement assisté par IA.

---

[🔝 Sommaire](../SOMMAIRE.md) · [⏭️ Section suivante : 1.1 Qu'est-ce que Go](01-quest-ce-que-go.md)

⏭ [Qu'est-ce que Go et à quoi il sert réellement (backend, CLI, cloud-native)](/01-introduction-go/01-quest-ce-que-go.md)
