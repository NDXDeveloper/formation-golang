🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 17. Développer en Go avec l'IA (l'ère Copilot)

Écrire du Go en 2026, c'est le plus souvent écrire du Go *assisté*. Ce que GitHub Copilot a inauguré en 2021 — la complétion ligne à ligne — a mûri en assistance **agentique** : des outils qui prennent une intention, modifient plusieurs fichiers, lancent les tests et itèrent seuls pendant quelques minutes. La question n'est plus *si* l'on utilise l'IA, mais *comment* l'utiliser bien : pour produire du Go **idiomatique**, et non du Java (ou du Python) déguisé en syntaxe Go.

Ce module n'est pas un panorama d'outils de plus. Il repose sur une conviction : **l'IA propose, l'outillage Go dispose.**

## Pourquoi ce module est spécifique à Go

Deux faits, en tension, gouvernent le couple Go + IA :

1. **Go est simple et régulier.** Sa petite surface syntaxique et ses conventions strictes font que les modèles génèrent du Go plausible avec aisance. C'est autant un piège qu'un cadeau : *plausible* n'est pas *idiomatique*. L'IA produit avec assurance des erreurs ignorées (`_`), des interfaces inutiles, de la sur-abstraction, des `panic` là où un `error` s'impose — précisément les travers que ce module apprend à repérer ([§ 17.2](02-pieges-ia.md)).

2. **L'outillage Go est un filet de sécurité exceptionnel** pour du code généré. `gofmt` clôt tout débat de style ; le compilateur et son typage fort attrapent des classes entières d'erreurs ; `go vet`, `staticcheck` / `golangci-lint`, les tests, le détecteur de races (`-race`) et `govulncheck` vérifient **mécaniquement** ce que l'IA propose. Là où un langage dynamique laisserait passer, la chaîne Go recale — c'est ce qui rend l'assistance IA particulièrement sûre en Go, à condition de s'en servir.

Le corollaire est exigeant : **c'est vous qui possédez la qualité idiomatique.** Un modèle ne sait pas qu'il vient d'écrire du Go maladroit ; vous, oui — si vous connaissez la langue (modules 1 à 4, [annexe B](../annexes/go-idiomatique/README.md)). L'IA accélère un jugement ; elle ne le remplace pas.

## Ce que ce module couvre — et ce qu'il ne duplique pas

Le fil est pratique : bien formuler ses demandes, connaître les pièges, déléguer les tâches vérifiables, puis brancher tout cela dans l'IDE.

Deux recoupements sont assumés plutôt que répétés :

- La **génération de tests par IA** est abordée dans le contexte des tests en [§ 13.6](../13-tests-qualite/06-couverture-tests-ia.md) ; la [§ 17.3](03-tests-migration-ia.md) la traite comme un sujet à part entière (avec migration et revue).
- Les **idiomes** que l'IA écorne sont définis aux modules 2 à 4 et à l'[annexe B](../annexes/go-idiomatique/README.md) ; l'outillage qui les rattrape, aux modules 13, 15 et 16. Les bonnes pratiques propres à l'IA sont réunies en [annexe C](../annexes/bonnes-pratiques/README.md).

## 🎯 Objectifs du module

À l'issue de ce module, vous saurez :

- Formuler des demandes qui produisent du Go idiomatique plutôt que générique.
- Reconnaître et corriger les modes d'échec récurrents de l'IA en Go.
- Déléguer à l'IA la génération de tests, la migration assistée et la revue — en vérifiant systématiquement.
- Configurer et exploiter les assistants intégrés (GoLand AI Assistant, Copilot dans VS Code).

## 📋 Prérequis

Une aisance réelle avec le Go idiomatique (modules 1 à 4) est indispensable : on ne juge pas ce qu'on ne maîtrise pas. Les modules qualité ([13](../13-tests-qualite/README.md), [15](../15-deploiement-devops/README.md), [16](../16-securite/README.md)) sont le complément naturel, puisque leur outillage est le filet décrit plus haut.

## 🗺️ Contenu du module

### 17.1 · [Copilot, Claude, ChatGPT : prompting efficace pour du Go idiomatique](01-prompting-go.md)
Comment cadrer une demande — contexte, contraintes, exemples — pour obtenir du Go conforme aux conventions plutôt qu'un plus petit dénominateur commun multi-langage.

### 17.2 · [Pièges de l'IA en Go](02-pieges-ia.md)
Les travers récurrents : code non idiomatique, erreurs ignorées, sur-abstraction. Les reconnaître, puis les redresser.

### 17.3 · [Génération de tests, migration assistée, revue de code par IA](03-tests-migration-ia.md)
Les tâches où l'IA rend le plus de service *parce qu'elles sont vérifiables* : produire des tests table-driven, migrer d'un langage vers Go, mener une première passe de revue.

### 17.4 · [Assistants intégrés : GoLand AI Assistant, Copilot dans VS Code](04-assistants-ide.md) 🆕
Brancher l'assistance dans l'IDE, des deux côtés du double outillage de cette formation.

## Un domaine qui bouge vite

Le paysage — modèles, agents, tarifs, parts de marché — change de mois en mois. Ce module privilégie donc les **principes durables** (comment formuler, quoi vérifier, où l'IA achoppe en Go) ; les intégrations concrètes sont en [§ 17.4](04-assistants-ide.md), à recouper avec l'état courant. Une distinction, elle, tient dans le temps :

- un **assistant** suggère du code dans l'éditeur (complétion, chat) ;
- un **agent** reçoit un objectif, modifie plusieurs fichiers, lance les tests et itère.

La plupart des flux de travail de 2026 combinent les deux — et, dans tous les cas, la **revue humaine** du code produit reste non négociable.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [17.1 Prompting efficace pour du Go idiomatique](01-prompting-go.md)

⏭ [Copilot, Claude, ChatGPT : prompting efficace pour du Go idiomatique](/17-developpement-ia/01-prompting-go.md)
