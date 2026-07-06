🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 17.4 Assistants intégrés : GoLand AI Assistant, Copilot dans VS Code

Les deux environnements de cette formation ont chacun leur assistant natif : côté **GoLand**, l'AI Assistant et l'agent **Junie** ; côté **VS Code**, **GitHub Copilot**. En 2026, les deux ont convergé vers le *même* jeu de capacités — ce qui les distingue tient désormais à l'écosystème et à la profondeur d'intégration dans l'IDE, plus qu'à une liste de fonctions.

C'est aussi le domaine le plus mouvant de toute la formation. Les noms de fonctions, les modèles disponibles et les tarifs ci-dessous sont un **instantané de 2026** : recoupez-les avec la documentation courante. Ce qui dure, c'est la *forme* — le socle décrit ci-dessous.

## 1. Un socle de capacités désormais commun

Pour ne pas dépendre du vocabulaire d'un outil, voici les catégories que **les deux** proposent, du moins autonome au plus autonome :

- **Complétion** — suggestions en ligne (*ghost text*) et « suggestions d'édition suivante » (prédire le *prochain* endroit à modifier). Rappel de [§ 17.1](01-prompting-go.md) : la complétion se pilote par le **contexte**, pas par les fichiers d'instructions.
- **Chat** — poser une question, faire expliquer un bout de code.
- **Édition** — décrire une modification sur des fichiers choisis ; **le diff est présenté avant d'être appliqué**.
- **Agent** — autonome : planifier → éditer → lancer tests et commandes → lire les erreurs → recommencer. Les deux offrent désormais une étape de **plan à valider** avant que l'agent ne touche au code.
- **Instructions de projet** — `AGENTS.md` (convention transverse) et le fichier propre à chaque outil, versionnés ([§ 17.1](01-prompting-go.md)).
- **Choix du modèle** — plusieurs fournisseurs (OpenAI, Anthropic, Google…), souvent avec un modèle **local** possible ; sélection manuelle ou automatique.
- **MCP** — brancher l'assistant sur des outils et données externes (base de données, services).

Cette échelle — complétion, chat, édition, agent — prolonge la distinction *assistant/agent* de [§ 17.1](01-prompting-go.md).

## 2. Copilot dans VS Code

Copilot combine **complétion** (*ghost text*) et **suggestions d'édition suivante** (NES), qui anticipent la modification d'après, y compris ailleurs dans le fichier. Le chat s'articule en trois modes : **Ask** (questions), **Edit** (on désigne des fichiers, on décrit le changement, on relit le diff) et **Agent** (autonome). En mode agent, Copilot détermine les fichiers à modifier, lance des commandes du terminal et des tests, **réagit aux erreurs de compilation et de linter et se corrige en boucle** — pour du Go, cela veut dire qu'il lit la sortie de `go build`, `go vet` et `go test` et itère. Un mode **Plan** permet de valider un plan avant l'exécution.

Le pilotage passe par `.github/copilot-instructions.md` et `AGENTS.md` ([§ 17.1](01-prompting-go.md)), par des *prompt files* réutilisables, et par MCP. La **revue de code Copilot** relit les *pull requests* en tenant compte de ces instructions. Côté modèle, Copilot est multi-modèle (OpenAI, Anthropic, Google…), avec sélection automatique et *bring your own key*.

Pour Go, l'association gagnante est **Copilot + l'extension Go officielle** (gopls) : l'IA écrit, gopls et la chaîne Go vérifient — les modernizers et le linting à la sauvegarde de [§ 17.2](02-pieges-ia.md) s'affichent en direct. Un réglage utile dans les *settings* :

```json
{
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package"
}
```

## 3. GoLand : AI Assistant + Junie

L'**AI Assistant** offre complétion, chat, édition multi-fichiers, explication de code, **génération de tests** et messages de commit. Sa particularité : il est **ancré dans le moteur d'analyse de l'IDE** (types résolus, références de symboles, graphes d'appels, structure des tests), si bien que ses suggestions contredisent rarement les inspections de GoLand. C'est la force propre de JetBrains — l'IA raisonne sur le code *compris*, pas seulement sur du texte.

**Junie** est l'agent autonome (sorti de l'EAP et rendu disponible — avec un palier gratuit — courant 2025, puis officiellement hors bêta à la mi-2026). Il suit une boucle **plan → exécution → vérification** : il produit d'abord un plan (stocké dans `.junie/plans`, versionnable) que vous approuvez, puis écrit le code en s'appuyant sur l'index sémantique, le **lanceur de tests** et le **débogueur** de l'IDE — face à un bug, il *ouvre le débogueur* plutôt que d'empiler des logs. Il lit ses règles dans `AGENTS.md` (ou `.junie/guidelines.md`), se connecte à MCP, et laisse choisir le modèle (y compris local).

Pensez au fichier **`.aiignore`** (même logique que `.gitignore`) pour tenir secrets et code vendorisé hors du contexte de l'IA — un garde-fou concret, à rapprocher de la gestion des secrets ([§ 16.2](../16-securite/02-cryptographie-tls.md)) :

```text
# .aiignore
.env*
secrets/
vendor/
```

Note : ces fonctions IA sont un module payant qui s'ajoute à l'IDE (un palier gratuit couvre la complétion locale ; le chat cloud et Junie consomment des crédits). À vérifier selon votre édition.

## 4. Le même socle Go dessous

Quel que soit l'IDE, l'assistant se pose **au-dessus de la chaîne Go et de gopls**. C'est la thèse du module, une dernière fois : l'IA propose ; `gopls`, `go vet`, `staticcheck` / golangci-lint, `go test -race` et les modernizers de `go fix` disposent. Et les meilleurs agents *lisent la sortie de cette chaîne et se corrigent* — ce qui explique pourquoi les vérifications mécaniques de Go rendent l'agentique relativement sûre en Go ([§ 17.2](02-pieges-ia.md), [§ 17.3](03-tests-migration-ia.md)).

Conséquence pratique : le réglage qui compte le plus n'est pas le fournisseur d'IA, c'est d'avoir la chaîne Go **branchée** (gopls actif, linters à la sauvegarde, tests exécutables). Les deux IDE le font nativement — c'est d'ailleurs le socle d'outillage installé dès le [module 1](../01-introduction-go/04-installation-outils.md).

## 5. Choisir, combiner, garder la main

Ce n'est pas un choix binaire. Il dépend surtout :

- de l'**écosystème** : centré sur GitHub (issues, PR, CI) → Copilot ; intégration JetBrains profonde, refactorings, débogueur → GoLand + Junie ;
- des besoins de **modèle et de gouvernance** : `.aiignore`, *bring your own key*, modèles locaux, rétention des données ;
- du **modèle de coût**.

Beaucoup d'équipes **combinent** (complétion dans l'un, agent dans l'autre, ou un agent en terminal à côté). Pour la grille de choix des IDE eux-mêmes, voir [§ 1.6.2](../01-introduction-go/06.2-goland-vs-vscode.md) ; pour les raccourcis au quotidien, l'[annexe D](../annexes/goland-vscode/README.md).

La constante, quel que soit l'outil : **relire le diff**, garder la chaîne Go comme porte de sortie, et ne pas laisser se dissoudre la responsabilité de celui qui fusionne ([§ 17.3](03-tests-migration-ia.md)). Et réexaminer son choix régulièrement : le paysage bouge tous les quelques mois.

## En résumé

- Les deux assistants natifs de la formation — **GoLand** (AI Assistant + **Junie**) et **Copilot** dans **VS Code** — ont convergé vers un socle commun : complétion, chat, édition (diff), agent (plan → exécute → vérifie), instructions `AGENTS.md`, choix du modèle, MCP.
- **Copilot / VS Code** : modes Ask / Edit / Agent (+ Plan), multi-modèle, à associer à l'extension Go (gopls) pour la vérification.
- **GoLand** : AI Assistant ancré dans le moteur d'analyse de l'IDE ; **Junie** planifie, exécute, lance les tests et *pilote le débogueur* ; `.aiignore` pour les secrets.
- **Dessous, le même socle Go** : l'IA propose, la chaîne (gopls, `vet`, `staticcheck`, `-race`, `go fix`) dispose — et les agents s'en servent pour se corriger.
- **Choisir** par écosystème, modèle/gouvernance et coût ; **combiner** librement ; **garder la main** (relire le diff, responsabilité humaine) ; recouper l'état courant, car tout évolue vite.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [18. Stratégie, feuille de route et ressources](../18-strategie-roadmap/README.md)

⏭ [Stratégie, feuille de route et ressources](/18-strategie-roadmap/README.md)
