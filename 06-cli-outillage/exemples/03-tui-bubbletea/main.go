/* ============================================================================
   Section 6.3 : TUI avec Bubble Tea (notions)
   Description : Le compteur de la section — l'architecture Elm au complet :
                 un Model (l'état), Update(msg) qui CALCULE l'état suivant
                 (récepteur par valeur : pas de mutation partagée), View()
                 qui rend une chaîne, tea.Quit pour finir. API v1 (la plus
                 répandue) ; la logique se teste SANS terminal (voir
                 model_test.go — c'est la vertu de TEA).
   Fichier source : 03-tui-bubbletea.md
   Lancer : go run .        (exige un VRAI terminal ; ↑/↓ ou k/j, q pour quitter)
   Tester : go test ./...   (aucun TTY nécessaire)
   ============================================================================ */

package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// Le Model : TOUT l'état de l'application vit ici.
type model struct {
	count int
}

// Init : la commande à lancer au démarrage (aucune ici).
func (m model) Init() tea.Cmd {
	return nil
}

// Update : le cœur de TEA — reçoit un message (événement), renvoie le
// PROCHAIN modèle et une éventuelle commande. Fonction quasi pure :
// les effets de bord vivent ailleurs, dans des tea.Cmd.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg: // un événement clavier
		switch msg.String() {
		case "up", "k":
			m.count++ // on modifie la COPIE (récepteur valeur)…
		case "down", "j":
			m.count--
		case "q", "ctrl+c":
			return m, tea.Quit // commande spéciale : arrêter le programme
		}
	}
	return m, nil // … et on la renvoie : c'est le nouvel état
}

// View : rend l'état en chaîne — Bubble Tea l'affiche à chaque changement.
func (m model) View() string {
	return fmt.Sprintf("Compteur : %d\n\n↑/↓ pour ajuster · q pour quitter\n", m.count)
}

func main() {
	// NewProgram prend la main sur le terminal (TTY requis) et fait tourner
	// la boucle messages → Update → View jusqu'à tea.Quit.
	if _, err := tea.NewProgram(model{}).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "erreur :", err)
		os.Exit(1)
	}
}
