package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.mode == modeChat {
			return m.updateChat(msg)
		}
		return m, nil

	case versionCheckMsg:
		if msg.err == nil {
			m.latestVersion = msg.latest
			if m.mode == modeChat && UpdateAvailable(m.latestVersion) {
				m.lines = append(m.lines, outputLine{kindRed,
					fmt.Sprintf("Update available: %s → %s\nOlder versions receive less support. Update for fewer bugs and new features.", titaniumVersion, m.latestVersion)})
				// viewport'u yeniden oluştur
				vw := m.vp.Width
				var parts []string
				for _, ol := range m.lines {
					parts = append(parts, renderLine(ol, vw))
					parts = append(parts, "")
				}
				m.vp.SetContent(strings.Join(parts, "\n"))
				m.vp.GotoBottom()
			}
		}
		return m, nil
	}

	if m.mode == modeChat {
		return m.updateChat(msg)
	}
	return m.updateHome(msg)
}

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	if m.mode == modeChat {
		return m.viewChat()
	}
	return m.viewHome()
}

func main() {
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Hata: %v", err)
	}
}
