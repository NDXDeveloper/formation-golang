/* ============================================================================
   Section 6.3 : TUI avec Bubble Tea — la testabilité de TEA
   Description : Update est une fonction quasi pure : on la teste SANS aucun
                 terminal, en fabriquant les messages clavier à la main.
                 (+1, +1, −1 → 1 ; « q » → la commande tea.Quit est renvoyée.)
   Fichier source : 03-tui-bubbletea.md
   Lancer : go test ./...
   ============================================================================ */

package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// key fabrique un tea.KeyMsg comme le ferait le terminal.
func key(s string) tea.KeyMsg {
	if s == "up" {
		return tea.KeyMsg{Type: tea.KeyUp}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func TestUpdateSansTerminal(t *testing.T) {
	var m tea.Model = model{}

	m, _ = m.Update(key("up")) // +1
	m, _ = m.Update(key("k"))  // +1
	m, _ = m.Update(key("j"))  // −1

	if got := m.(model).count; got != 1 {
		t.Fatalf("count = %d, attendu 1", got)
	}
	if !strings.Contains(m.View(), "Compteur : 1") {
		t.Fatalf("View() = %q", m.View())
	}
}

func TestQuitRenvoieLaCommande(t *testing.T) {
	var m tea.Model = model{}
	_, cmd := m.Update(key("q"))
	if cmd == nil {
		t.Fatal("« q » devrait renvoyer tea.Quit")
	}
}
