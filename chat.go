package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func renderLine(ol outputLine, w int) string {
	switch ol.kind {
	case kindInput:
		dollarPrefix := lipgloss.NewStyle().
			Foreground(lipgloss.Color(tipOrange)).
			Render("$ ")
		text := lipgloss.NewStyle().
			Foreground(pureWhite).
			Render(ol.content)
		return lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(subtleGrey).
			PaddingLeft(2).
			PaddingRight(2).
			Width(w - 1).
			Render(dollarPrefix + text)

	case kindPlain:
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			PaddingLeft(2).
			Render(wordWrap(ol.content, w-3))

	case kindGray:
		return renderSideBox(ol.content, colorGray, w)

	case kindBlue:
		return renderSideBox(ol.content, colorBlue, w)

	case kindRed:
		return renderSideBox(ol.content, colorRed, w)
	}
	return ol.content
}

func renderSideBox(content string, accent lipgloss.Color, w int) string {
	inner := w - 6
	if inner < 10 {
		inner = 10
	}
	body := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Render(wordWrap(content, inner))

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(accent).
		PaddingLeft(3).
		PaddingRight(2).
		Width(w - 1).
		Render(body)
}

func (m *model) rebuildViewport() {
	if !m.vpReady {
		return
	}
	vw := m.vp.Width
	var parts []string
	for _, ol := range m.lines {
		parts = append(parts, renderLine(ol, vw))
		parts = append(parts, "")
	}
	m.vp.SetContent(strings.Join(parts, "\n"))
	m.vp.GotoBottom()
}

func (m *model) execCommand(input string) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return
	}

	switch parts[0] {

	case "/exit":
		// handled in Update

	case "/new":
		m.lines = nil
		m.lines = append(m.lines, outputLine{kindGray, "Session cleared."})

	case "/help":
		m.lines = append(m.lines, outputLine{kindBlue,
			"Commands\n" +
				"─────────────────────────────────────────\n" +
				"/new           Clear session\n" +
				"/help          Show this help\n" +
				"/exit          Exit\n" +
				"─────────────────────────────────────────\n" +
				"Type / to open the command menu."})

	// TODO: add other commands here

	default:
		m.lines = append(m.lines, outputLine{kindRed,
			fmt.Sprintf("'%s' is not a valid command. Type /help for the command list.", parts[0])})
	}
}

func (m model) updateChat(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		vpH := chatViewportHeight(m.height, m.showCommands)
		if !m.vpReady {
			m.vp = viewport.New(m.width, vpH)
			m.vpReady = true
		} else {
			m.vp.Width = m.width
			m.vp.Height = vpH
		}
		m.rebuildViewport()
		return m, nil

	case tea.MouseMsg:
		if m.vpReady {
			var cmd tea.Cmd
			m.vp, cmd = m.vp.Update(msg)
			return m, cmd
		}

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

		if m.showCommands {
			var cmd tea.Cmd
			m, cmd, _ = handleCmdListKey(m, msg)
			m.rebuildViewport()
			return m, cmd
		}

		switch msg.Type {
		case tea.KeyEsc:
			if m.setupStep != setupNone {
				m.setupStep = setupNone
				m.setupName = ""
				m.textInput.Placeholder = "Run a command or type /help to get started…"
				m.textInput.SetValue("")
				m.lines = append(m.lines, outputLine{kindGray, "Cancelled."})
				m.rebuildViewport()
				return m, nil
			}
			m.mode = modeHome
			m.textInput.SetValue("")
			m.showCommands = false
			return m, nil

		case tea.KeyUp:
			if m.vpReady {
				m.vp.LineUp(3)
			}
			return m, nil

		case tea.KeyDown:
			if m.vpReady {
				m.vp.LineDown(3)
			}
			return m, nil

		case tea.KeyPgUp:
			if m.vpReady {
				m.vp.HalfViewUp()
			}
			return m, nil

		case tea.KeyPgDown:
			if m.vpReady {
				m.vp.HalfViewDown()
			}
			return m, nil

		case tea.KeyEnter:
			val := strings.TrimSpace(m.textInput.Value())
			if val == "" {
				return m, nil
			}
			if val == "/exit" {
				return m, tea.Quit
			}
			m.lines = append(m.lines, outputLine{kindInput, val})
			if strings.HasPrefix(val, "/") {
				m.execCommand(val)
			} else {
				m.lines = append(m.lines, outputLine{kindGray,
					fmt.Sprintf("'%s' is not a command. Commands start with /, type /help for the list.", val)})
			}
			m.textInput.SetValue("")
			m.showCommands = false
			m.rebuildViewport()
			return m, nil
		}

		var cmd tea.Cmd
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

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) viewChat() string {
	w := m.width

	// ── Sağ üst: versiyon / güncelleme bildirimi ──
	var topBar string
	if UpdateAvailable(m.latestVersion) {
		topBar = lipgloss.NewStyle().
			Width(w).
			Align(lipgloss.Right).
			Foreground(tipOrange).
			Render("  ↑ update available: " + m.latestVersion + " ")
	} else {
		topBar = lipgloss.NewStyle().
			Width(w).
			Align(lipgloss.Right).
			Foreground(subtleGrey).
			Render(titaniumVersion + " ")
	}
	topBarH := lipgloss.Height(topBar)

	// inputBar her zaman sabit altta
	inputBar := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(accentBlue).
		Padding(1, 4).
		Background(bgDark).
		Width(w).
		Render(m.textInput.View())
	inputBarH := lipgloss.Height(inputBar)

	// Viewport kalan alanı doldurur
	vpH := m.height - topBarH - inputBarH
	if vpH < 1 {
		vpH = 1
	}
	if m.vpReady && m.vp.Height != vpH {
		m.vp.Height = vpH
	}

	vpView := ""
	if m.vpReady {
		vpView = m.vp.View()
	}

	// cmdList varsa viewport'un altına yapışık, inputBar'ın hemen üstünde göster
	if m.showCommands {
		// Ekran yüksekliğinin 1/3'ünü geçmesin, en az 1, en fazla cmdListMaxVisible
		maxVisible := (m.height - inputBarH) / 3
		if maxVisible < 1 {
			maxVisible = 1
		}
		if maxVisible > cmdListMaxVisible {
			maxVisible = cmdListMaxVisible
		}

		cmdListView := renderCmdListFixed(m.filteredCmds, m.cmdSelected, true, w, maxVisible)
		cmdListH := lipgloss.Height(cmdListView)

		// Viewport'un son cmdListH satırını sil, yerine cmdList koy
		vpLines := strings.Split(vpView, "\n")
		if len(vpLines) > cmdListH {
			vpView = strings.Join(vpLines[:len(vpLines)-cmdListH], "\n")
		} else {
			vpView = ""
		}
		return lipgloss.JoinVertical(lipgloss.Left,
			topBar,
			vpView,
			cmdListView,
			inputBar,
		)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		topBar,
		vpView,
		inputBar,
	)
}

// chatViewportHeight — geriye dönük uyumluluk için, enterChat'te kullanılır
func chatViewportHeight(totalH int, showCmds bool) int {
	inputBarH := 3
	h := totalH - inputBarH
	if h < 1 {
		h = 1
	}
	return h
}

func (m *model) enterChat(firstInput string) {
	m.mode = modeChat
	vpH := chatViewportHeight(m.height, false)
	m.vp = viewport.New(m.width, vpH)
	m.vp.Width = m.width
	m.vpReady = true

	m.lines = []outputLine{
		{kindGray, buildStartupStatus()},
	}

	if UpdateAvailable(m.latestVersion) {
		m.lines = append(m.lines, outputLine{kindRed,
			fmt.Sprintf("Update available: %s → %s\nOlder versions receive less support. Update for fewer bugs and new features.", titaniumVersion, m.latestVersion)})
	}

	m.lines = append(m.lines, outputLine{kindInput, firstInput})
	if strings.HasPrefix(firstInput, "/") {
		m.execCommand(firstInput)
	}

	m.textInput.SetValue("")
	m.showCommands = false
	m.rebuildViewport()
}
