package main

import (
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Home Update ────────────────────────────────────────────────────────────────
func (m model) updateHome(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			if m.showCommands && len(m.filteredCmds) > 0 {
				sel := m.filteredCmds[m.cmdSelected]
				m.textInput.SetValue(sel.name + " ")
				m.textInput.CursorEnd()
				m.showCommands = false
				// textInput.Update çağrılmadan döndür — liste tekrar açılmasın
				return m, nil
			}

			val := strings.TrimSpace(m.textInput.Value())
			if val == "" {
				return m, nil
			}
			if val == "/exit" {
				return m, tea.Quit
			}
			m.showCommands = false
			m.enterChat(val)
			return m, nil
		}

		if m.showCommands {
			var listCmd tea.Cmd
			m, listCmd, _ = handleCmdListKey(m, msg)
			return m, listCmd
		}

	case tea.MouseMsg:
		if m.showCommands && msg.Action == tea.MouseActionRelease {
			listLen := len(m.filteredCmds)
			if listLen > 0 {
				mainUIHeight := 14 + listLen
				startY := (m.height-1)/2 - mainUIHeight/2
				listTop := startY + 3
				clickedRow := msg.Y - listTop
				if clickedRow >= 0 && clickedRow < listLen {
					sel := m.filteredCmds[clickedRow]
					m.textInput.SetValue(sel.name + " ")
					m.textInput.CursorEnd()
					m.showCommands = false
					m.filteredCmds = commands
					return m, nil
				}
			}
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)

	nv := m.textInput.Value()
	if shouldShowCommands(nv) {
		m.showCommands = true
		m.filteredCmds = filterCommands(nv)
		if m.cmdSelected >= len(m.filteredCmds) {
			m.cmdSelected = 0
		}
	} else {
		m.showCommands = false
	}

	return m, cmd
}

// ── Home View ──────────────────────────────────────────────────────────────────
func (m model) viewHome() string {
	dw := 80
	if m.width < 85 {
		dw = m.width - 5
	}
	if dw < 30 {
		dw = 30
	}

	// ── Logo ──
	asciiLogo := " _____  _  _              _                 \n" +
		" |_   _|| || |_ __ _ _ __ (_)_   _ _ __ ___  \n" +
		"   | |  | || __/ _` | '_ \\| | | | | '_ ` _ \\ \n" +
		"   | |  | || || (_| | | | | | |_| | | | | | |\n" +
		"   |_|  |_| \\__\\__,_|_| |_|_|\\__,_|_| |_| |_|\n"
	topBar := lipgloss.NewStyle().
		Foreground(subtleGrey).
		Width(dw).
		Align(lipgloss.Center).
		Render(asciiLogo)

	// ── Engine/Git ──
	engineLabel := lipgloss.NewStyle().Foreground(accentBlue).Bold(true).Background(bgDark).Render("ENGINE  ")
	gitLabel := lipgloss.NewStyle().Foreground(pureWhite).Background(bgDark).Render("Git  ")
	var gitState string
	if m.gitVersion != "" {
		activeLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Background(bgDark).Render("● active")
		gitState = lipgloss.JoinHorizontal(lipgloss.Top, engineLabel, gitLabel,
			lipgloss.NewStyle().Foreground(subtleGrey).Background(bgDark).Render(m.gitVersion+"  "), activeLabel)
	} else {
		gitState = lipgloss.JoinHorizontal(lipgloss.Top, engineLabel, gitLabel,
			lipgloss.NewStyle().Foreground(colorRed).Background(bgDark).Render("◌ not found"))
	}

	// ── InputBox ──
	m.textInput.Placeholder = "Run a command or type /help to get started…"
	mainBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(accentBlue).
		Padding(2, 4).
		Background(bgDark).
		Width(dw).
		Render(m.textInput.View() + "\n\n" + gitState)

	shortcuts := lipgloss.NewStyle().Width(dw).Align(lipgloss.Right).Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Foreground(pureWhite).Render("esc"),
			lipgloss.NewStyle().Foreground(subtleGrey).Render(" quit"),
		),
	)

	tipLine := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Foreground(tipOrange).Render(" ● Tip "),
		lipgloss.NewStyle().Foreground(subtleGrey).Render("Type "),
		lipgloss.NewStyle().Foreground(pureWhite).Bold(true).Render("/init"),
		lipgloss.NewStyle().Foreground(subtleGrey).Render(" to get started."),
	)

	// ── Komut listesi ──
	logoH := lipgloss.Height(topBar)
	logoPad := 2
	maxVisible := logoH + logoPad - 2
	if maxVisible < 1 {
		maxVisible = 1
	}
	if maxVisible > cmdListMaxVisible {
		maxVisible = cmdListMaxVisible
	}
	cmdSlot := ""
	if m.showCommands {
		cmdSlot = renderCmdListFixed(m.filteredCmds, m.cmdSelected, true, dw, maxVisible)
	}
	cmdSlotH := lipgloss.Height(cmdSlot)

	// ── Ortalanacak içerik grubu: logo + [cmdSlot] + inputBox ──
	// cmdSlot logo alanına overlay yapacak — logo+logoPad satırları üzerine yazar
	logoAndGap := lipgloss.JoinVertical(lipgloss.Left, topBar, strings.Repeat("\n", logoPad-1))
	logoAndGapLines := strings.Split(logoAndGap, "\n")

	if m.showCommands && cmdSlotH > 0 {
		cmdLines := strings.Split(strings.TrimRight(cmdSlot, "\n"), "\n")
		// cmdSlot'u logo bloğunun sonundan geriye doğru yaz
		insertAt := len(logoAndGapLines) - len(cmdLines)
		if insertAt < 0 {
			insertAt = 0
		}
		for i, cl := range cmdLines {
			idx := insertAt + i
			if idx < len(logoAndGapLines) {
				logoAndGapLines[idx] = cl
			}
		}
	}
	logoBlock := strings.Join(logoAndGapLines, "\n")

	// Ortalanacak grup: logoBlock + mainBox + shortcuts + tip
	group := lipgloss.JoinVertical(lipgloss.Left,
		logoBlock,
		mainBox,
		shortcuts,
		lipgloss.NewStyle().MarginTop(1).Width(dw).Align(lipgloss.Center).Render(tipLine),
	)

	// ── Tam ekran yerleşim ──
	totalH := m.height - 1 // footer satırı

	// group ortada, version sağ altta, footer sol altta
	// Place ile group'u ortala
	centeredGroup := lipgloss.Place(m.width, totalH, lipgloss.Center, lipgloss.Center, group)

	// ── Footer: sol path, sağ version ──
	path, _ := os.Getwd()
	footerPath := lipgloss.NewStyle().Foreground(subtleGrey).Bold(true).Render(path)

	versionLabel := titaniumVersion
	if UpdateAvailable(m.latestVersion) {
		versionLabel = titaniumVersion +
			lipgloss.NewStyle().Foreground(tipOrange).Render("  ↑ update available: "+m.latestVersion)
	}
	versionStr := lipgloss.NewStyle().Foreground(subtleGrey).Render(versionLabel)

	repoStatus := ""
	if IsGitRepo() {
		if branch := GitBranch(); branch.OK {
			label := branch.Output
			if IsProtectedBranch(label) {
				label = "⚠ " + label
			}
			repoStatus = lipgloss.NewStyle().Foreground(accentBlue).Render("  git:" + label)
		}
	}
	footerLeft := footerPath + repoStatus
	padLen := m.width - lipgloss.Width(footerLeft) - lipgloss.Width(versionStr)
	if padLen < 0 {
		padLen = 0
	}
	footer := footerLeft + strings.Repeat(" ", padLen) + versionStr

	return centeredGroup + "\n" + footer
}
