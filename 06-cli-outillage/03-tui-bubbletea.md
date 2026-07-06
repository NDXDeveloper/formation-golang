🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 6.3 TUI avec Bubble Tea (notions)

Cobra et Viper ([§ 6.2](02-cobra-viper.md)) organisent l'invocation et la configuration d'un outil. Un autre besoin, plus rare, est celui d'une **interface texte interactive** (TUI, *Terminal User Interface*) : un tableau de bord qui se rafraîchit, un assistant pas à pas, une liste filtrable au clavier. En Go, la référence est **Bubble Tea**, le framework de l'éditeur Charm — « la façon amusante, fonctionnelle et à état de construire des applications de terminal », fondée sur l'architecture Elm. Il est largement adopté en production, chez Microsoft Azure, AWS, Cockroach Labs, Ubuntu ou NVIDIA, et figure parmi les projets Go les plus étoilés.

Cette section reste au niveau des **notions** : le modèle mental, un exemple minimal et l'écosystème, de quoi décider si une TUI est pertinente et savoir par où commencer — sans viser l'exhaustivité.

## Quand une TUI se justifie (et quand non)

Fidèle au principe « stdlib avant frameworks », la première question est : *ai-je vraiment besoin d'interactivité plein écran ?* Le plus souvent, non.

- Pour **afficher** un résultat, `fmt` et `text/tabwriter` (colonnes alignées) suffisent, et la sortie reste redirigeable (`| grep`, `> fichier`).
- Pour une **saisie ponctuelle** (un choix, une confirmation, un mot de passe), lire sur `os.Stdin` ou utiliser **Huh** (formulaires de terminal de Charm) est bien plus léger qu'une TUI complète.
- Bubble Tea prend tout son sens quand l'affichage est **réellement interactif et dynamique** : navigation au clavier, mise à jour continue, plusieurs panneaux, sélection dans une longue liste.

Une TUI est une dépendance et une complexité supplémentaires : on l'introduit à bon escient, pas par défaut.

## L'architecture Elm : Model, Update, View

Bubble Tea structure un programme autour de **l'architecture Elm** (*The Elm Architecture*, TEA), qui sépare proprement l'état, les transitions et le rendu. Trois éléments suffisent à la décrire :

- un **`Model`** : la structure qui contient tout l'état de l'application ;
- **`Init`** : renvoie une éventuelle commande initiale à exécuter au démarrage ;
- **`Update(msg)`** : reçoit un **message** (un événement), fait évoluer l'état et renvoie le nouveau modèle plus une éventuelle commande ;
- **`View()`** : produit, à partir du modèle, la **chaîne** à afficher.

Deux types complètent le tableau : `tea.Msg` (un `any`, n'importe quel événement) et `tea.Cmd` (une fonction `func() tea.Msg` représentant un **effet de bord** — E/S réseau, minuteur — dont le résultat reviendra plus tard sous forme de message). Le point clé, et la vertu de ce modèle : `Update` est une fonction quasi pure (état d'entrée → état de sortie), tandis que **tous les effets de bord sont isolés dans des commandes**. Cela rend la logique lisible et testable, indépendamment du rendu.

## Un exemple minimal

Un compteur ajustable au clavier illustre le cycle complet. On utilise ici l'API v1 (voir l'encadré « État 2026 » plus bas pour la v2).

```go
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	count int
}

func (m model) Init() tea.Cmd {
	return nil // aucune commande initiale
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg: // un événement clavier
		switch msg.String() {
		case "up", "k":
			m.count++
		case "down", "j":
			m.count--
		case "q", "ctrl+c":
			return m, tea.Quit // commande spéciale : arrêter le programme
		}
	}
	return m, nil // on renvoie le modèle (éventuellement modifié) et aucune commande
}

func (m model) View() string {
	return fmt.Sprintf("Compteur : %d\n\n↑/↓ pour ajuster · q pour quitter\n", m.count)
}

func main() {
	if _, err := tea.NewProgram(model{}).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "erreur :", err)
		os.Exit(1)
	}
}
```

Le récepteur par **valeur** et le fait qu'`Update` **renvoie** le prochain modèle traduisent l'esprit de TEA : on ne mute pas un état partagé, on calcule l'état suivant. `tea.NewProgram(model{}).Run()` prend la main sur le terminal, appelle `Update` à chaque message et `View` à chaque rafraîchissement, jusqu'à `tea.Quit`.

## Messages, commandes et effets de bord

Tout ce qui n'est pas instantané passe par une **commande**. Une commande est une fonction exécutée par Bubble Tea *hors* du cycle `Update` (dans une goroutine) ; sa valeur de retour est réinjectée comme message.

```go
type statusMsg int
type errMsg struct{ err error }

func fetchStatus(url string) tea.Cmd {
	return func() tea.Msg {
		resp, err := http.Get(url) // travail bloquant, isolé du rendu
		if err != nil {
			return errMsg{err}
		}
		defer resp.Body.Close()
		return statusMsg(resp.StatusCode)
	}
}
```

`Update` traite ensuite `statusMsg` ou `errMsg` comme n'importe quel autre événement. Autour de ce mécanisme, Bubble Tea fournit des commandes utilitaires — `tea.Quit` (quitter), `tea.Batch` (lancer plusieurs commandes de front), `tea.Sequence` (les enchaîner), `tea.Tick` (un minuteur) — et des messages système comme `tea.WindowSizeMsg` (redimensionnement du terminal), indispensable pour adapter la mise en page.

Conséquence pratique : puisque la TUI occupe le terminal, **on ne peut pas déboguer avec `fmt.Println`**. Bubble Tea expose `tea.LogToFile`, précisément parce qu'on ne peut pas écrire sur le terminal occupé par l'interface ; on journalise dans un fichier que l'on suit en parallèle (`tail -f debug.log`).

## L'écosystème Charm

Bubble Tea se limite au moteur ; le reste vient de bibliothèques compagnes que l'on assemble par composition :

- **Bubbles** (`github.com/charmbracelet/bubbles`) : des composants prêts à l'emploi, chacun étant lui-même un petit modèle Bubble Tea — champ de saisie, zone de texte, *spinner*, barre de progression, *viewport*, tableau, et une liste riche avec filtrage, pagination, aide et messages de statut.
- **Lip Gloss** (`github.com/charmbracelet/lipgloss`) : la mise en forme et la mise en page — couleurs, bordures, marges, alignement, assemblage de blocs — à la manière d'une feuille de style pour le terminal.
- **Huh** (`github.com/charmbracelet/huh`) : des formulaires interactifs clés en main, pour les cas où l'on veut de l'interactivité sans écrire une TUI complète.

D'autres outils Charm gravitent autour (Glamour pour le rendu Markdown en terminal, Wish pour servir des applications Bubble Tea via SSH, Gum pour scripter de l'interactivité en shell, VHS pour enregistrer des démonstrations), mais ils dépassent le cadre de ces notions.

## État 2026 : Bubble Tea v1 et v2

C'est le point de vigilance du moment. **Deux versions majeures coexistent** :

- **v1** — `github.com/charmbracelet/bubbletea` : stable, de très loin la plus utilisée, et la base de la quasi-totalité des tutoriels et exemples existants. C'est celle des extraits ci-dessus.
- **v2** — publiée en 2026, avec un **changement de chemin de module** : le code source reste sur `github.com/charmbracelet/bubbletea`, mais l'import devient `charm.land/bubbletea/v2`.

La v2 conserve **le même modèle mental** (Model / Update / View) mais apporte surtout un moteur de rendu réécrit (« Cursed Renderer »), la sortie synchronisée (le mode terminal « 2026 » — un numéro de mode DEC, rien à voir avec l'année) qui met à jour le terminal de façon atomique pour éliminer scintillements et déchirures, la détection étendue du clavier (modificateurs, relâchement de touches) et le presse-papiers natif (OSC52, y compris via SSH). Côté migration, quelques types de messages passent d'alias à structures : la gestion clavier se fait désormais via `tea.KeyPressMsg` (là où v1 utilise `tea.KeyMsg`). Les bibliothèques Bubbles et Lip Gloss suivent avec des lignes v2 alignées (Lip Gloss a par ailleurs atteint la v1.0).

En pratique, pour un nouveau projet en 2026 : la v2 est le choix d'avenir, mais **vérifiez la compatibilité des composants Bubbles/Lip Gloss** que vous visez et attendez-vous à trouver la plupart des ressources encore rédigées pour v1. Les concepts de cette section restent valables dans les deux cas.

## Côté IDE : GoLand et VS Code

Une TUI a une exigence particulière : elle a besoin d'un **vrai terminal (TTY)**. Or les panneaux d'exécution intégrés des IDE ne sont pas toujours de véritables terminaux, ce qui casse le rendu et la capture des touches.

**GoLand** — la voie sûre est de lancer la TUI dans le **Terminal intégré** (`go run .`), qui est un vrai TTY. Selon la version, la configuration d'exécution propose une option « *Emulate terminal in output console* » (héritée de la plateforme IntelliJ, longtemps absente des configurations Go — vérifiez sa présence) qui fournit un pseudo-terminal dans la console. Le débogueur reste utilisable (points d'arrêt dans `Update`), mais comme l'interface occupe l'écran, la bonne pratique est de tracer via `tea.LogToFile` et de suivre le fichier dans un second onglet Terminal.

**VS Code** — dans `.vscode/launch.json`, forcez l'exécution dans un terminal réel plutôt que la *Debug Console* :

```jsonc
{
  "name": "monoutil (TUI)",
  "type": "go",
  "request": "launch",
  "mode": "debug",
  "program": "${workspaceFolder}",
  "console": "integratedTerminal"
}
```

`"console": "integratedTerminal"` (ou `"externalTerminal"`) garantit un TTY ; la *Debug Console* par défaut n'en est pas un et perturbe les applications interactives. Ici encore, `tea.LogToFile` + `tail -f` est la voie la plus confortable pour observer l'état pendant l'exécution.

## En résumé

- Une TUI n'est justifiée que pour de l'interactivité réelle ; pour afficher, `fmt`/`tabwriter` suffisent, et pour une saisie simple, `os.Stdin` ou Huh sont plus légers.
- Bubble Tea applique l'architecture Elm : un `Model` d'état, `Update(msg)` qui calcule l'état suivant, `View()` qui rend une chaîne ; `tea.Msg` = événement, `tea.Cmd` = effet de bord isolé renvoyant un message.
- L'écosystème se compose : Bubbles (composants), Lip Gloss (style/layout), Huh (formulaires).
- On ne débogue pas une TUI avec `fmt.Println` : `tea.LogToFile` + suivi de fichier, et un IDE configuré pour fournir un vrai TTY.
- État 2026 : v1 (`github.com/charmbracelet/bubbletea`) domine ; v2 (`charm.land/bubbletea/v2`) est sortie, même modèle mais rendu réécrit et quelques renommages de messages (`KeyMsg` → `KeyPressMsg`).

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [6.4 — Distribution : binaire unique, cross-compilation, GoReleaser](04-distribution.md)

⏭ [Distribution : binaire unique, cross-compilation, GoReleaser](/06-cli-outillage/04-distribution.md)
