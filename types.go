package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	subtleGrey = lipgloss.Color("241")
	pureWhite  = lipgloss.Color("255")
	accentBlue = lipgloss.Color("33")
	tipOrange  = lipgloss.Color("208")
	bgDark     = lipgloss.Color("234")
	colorGray  = lipgloss.Color("240")
	colorBlue  = lipgloss.Color("33")
	colorRed   = lipgloss.Color("196")
	logoStyle  = lipgloss.NewStyle().Foreground(subtleGrey).Bold(true)
)

type command struct {
	name string
	desc string
}

type outputKind int

const (
	kindInput outputKind = iota
	kindPlain
	kindGray
	kindBlue
	kindRed
)

type outputLine struct {
	kind    outputKind
	content string
}

type appMode int

const (
	modeHome appMode = iota
	modeChat
)

var commands = []command{
	{"/branch", "Show active branch"},
	{"/cleanup", "Delete merged feature branch"},
	{"/commit", "Commit staged changes"},
	{"/exit", "Exit the app"},
	{"/finish", "Finish feature and merge"},
	{"/help", "List all commands"},
	{"/init", "Initialize Titanium in this repo"},
	{"/log", "Show recent commits"},
	{"/new", "Clear session"},
	{"/review", "Review uncommitted changes"},
	{"/setup", "Set git identity (name & email)"},
	{"/stage", "Stage all changes"},
	{"/start", "Start a new feature branch"},
	{"/status", "Git status + Titanium info"},
}

type setupStep int

const (
	setupNone setupStep = iota
	setupName
	setupEmail
	setupCommitMsg
)

type model struct {
	mode          appMode
	textInput     textinput.Model
	width         int
	height        int
	showCommands  bool
	cmdSelected   int
	filteredCmds  []command
	lines         []outputLine
	vp            viewport.Model
	vpReady       bool
	setupStep     setupStep
	setupName     string
	gitVersion    string
	latestVersion string // fetched from GitHub
}

func newTextInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Run a command or type /help to get started…"
	ti.Focus()
	ti.Prompt = ""
	ti.TextStyle = lipgloss.NewStyle().Foreground(pureWhite).Background(bgDark)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(subtleGrey).Background(bgDark)
	return ti
}

// versionCheckMsg — async version check result
type versionCheckMsg struct {
	latest string
	err    error
}

func doVersionCheck() tea.Msg {
	latest, err := checkLatestVersion()
	return versionCheckMsg{latest: latest, err: err}
}

func initialModel() model {
	gitVer := runGit("--version")
	shortVer := ""
	if gitVer.OK {
		raw := strings.TrimPrefix(strings.TrimSpace(gitVer.Output), "git version ")
		shortVer = strings.Fields(raw)[0]
	}

	return model{
		mode:         modeHome,
		textInput:    newTextInput(),
		filteredCmds: commands,
		gitVersion:   shortVer,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, doVersionCheck)
}

func filterCommands(input string) []command {
	if input == "" || input == "/" {
		return commands
	}
	query := strings.ToLower(strings.TrimPrefix(input, "/"))
	if query == "" {
		return commands
	}
	var result []command
	for _, c := range commands {
		nameCore := strings.TrimPrefix(strings.ToLower(c.name), "/")
		if strings.Contains(nameCore, query) {
			result = append(result, c)
		}
	}
	return result
}

func shouldShowCommands(val string) bool {
	return strings.HasPrefix(val, "/")
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func wordWrap(s string, width int) string {
	if width <= 0 {
		return s
	}
	var out []string
	for _, line := range strings.Split(s, "\n") {
		words := strings.Fields(line)
		if len(words) == 0 {
			out = append(out, "")
			continue
		}
		cur := ""
		for _, w := range words {
			if cur == "" {
				cur = w
			} else if len(cur)+1+len(w) <= width {
				cur += " " + w
			} else {
				out = append(out, cur)
				cur = w
			}
		}
		if cur != "" {
			out = append(out, cur)
		}
	}
	return strings.Join(out, "\n")
}

func handleCmdListKey(m model, msg tea.KeyMsg) (model, tea.Cmd, bool) {
	switch msg.Type {
	case tea.KeyEsc:
		m.showCommands = false
		m.filteredCmds = commands
		m.cmdSelected = 0
		return m, nil, true
	case tea.KeyUp:
		if m.cmdSelected > 0 {
			m.cmdSelected--
		}
		return m, nil, true
	case tea.KeyDown:
		if m.cmdSelected < len(m.filteredCmds)-1 {
			m.cmdSelected++
		}
		return m, nil, true
	case tea.KeyEnter:
		if len(m.filteredCmds) > 0 {
			sel := m.filteredCmds[m.cmdSelected]
			m.textInput.SetValue(sel.name + " ")
			m.textInput.CursorEnd()
			m.showCommands = false
			m.filteredCmds = commands
			m.cmdSelected = 0
		}
		return m, nil, true
	case tea.KeyBackspace:
		val := m.textInput.Value()
		if len([]rune(val)) <= 1 {
			m.textInput.SetValue("")
			m.showCommands = false
			m.filteredCmds = commands
			m.cmdSelected = 0
			return m, nil, true
		}
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		nv := m.textInput.Value()
		m.filteredCmds = filterCommands(nv)
		if m.cmdSelected >= len(m.filteredCmds) {
			m.cmdSelected = maxInt(0, len(m.filteredCmds)-1)
		}
		m.showCommands = shouldShowCommands(nv)
		return m, cmd, true
	default:
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		nv := m.textInput.Value()
		m.filteredCmds = filterCommands(nv)
		if m.cmdSelected >= len(m.filteredCmds) {
			m.cmdSelected = 0
		}
		m.showCommands = shouldShowCommands(nv)
		return m, cmd, true
	}
}

// logoStyle kullanımı derleyici uyarısını önlemek için
var _ = logoStyle

// cmdListMaxVisible — komut listesinde aynı anda gösterilecek max satır
const cmdListMaxVisible = 6

// renderCmdListFixed — her zaman sabit yükseklikte alan kaplar; kayma önlenir.
func renderCmdListFixed(filteredCmds []command, cmdSelected int, show bool, width int, maxRows int) string {
	if !show {
		return ""
	}

	if len(filteredCmds) == 0 {
		notFound := lipgloss.NewStyle().
			Foreground(subtleGrey).
			Render("no such command")
		return lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(subtleGrey).
			Padding(1, 4).
			Background(bgDark).
			Width(width).
			Render(notFound)
	}

	start := 0
	if cmdSelected >= maxRows {
		start = cmdSelected - maxRows + 1
	}
	end := start + maxRows
	if end > len(filteredCmds) {
		end = len(filteredCmds)
		start = end - maxRows
		if start < 0 {
			start = 0
		}
	}
	visible := filteredCmds[start:end]
	visSelected := cmdSelected - start

	ciw := width - 9
	if ciw < 10 {
		ciw = 10
	}

	var rows []string
	for i, c := range visible {
		sel := i == visSelected
		var rowBg, nFg, dFg lipgloss.Color
		if sel {
			rowBg, nFg, dFg = tipOrange, pureWhite, pureWhite
		} else {
			rowBg, nFg, dFg = bgDark, pureWhite, subtleGrey
		}
		ns := fmt.Sprintf("%-12s", c.name)
		da := ciw - 14
		if da < 5 {
			da = 5
		}
		dt := c.desc
		if rr := []rune(dt); len(rr) > da {
			dt = string(rr[:da-1]) + "…"
		}
		nr := lipgloss.NewStyle().Foreground(nFg).Background(rowBg).Bold(true).Render(ns)
		gr := lipgloss.NewStyle().Background(rowBg).Render("  ")
		dr := lipgloss.NewStyle().Foreground(dFg).Background(rowBg).Render(dt)
		pad := maxInt(0, ciw-12-2-lipgloss.Width(dt))
		tr := lipgloss.NewStyle().Background(rowBg).Render(strings.Repeat(" ", pad))
		rows = append(rows, nr+gr+dr+tr)
	}

	// Scroll göstergesi — sadece maxRows'u aşınca göster
	scrollLine := ""
	if len(filteredCmds) > maxRows {
		shown := fmt.Sprintf("%d/%d", end, len(filteredCmds))
		scrollLine = "\n" + lipgloss.NewStyle().Foreground(subtleGrey).Render("  ↑↓ "+shown)
	}

	inner := strings.Join(rows, "\n") + scrollLine

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(subtleGrey).
		Padding(1, 4).
		Background(bgDark).
		Width(width).
		Render(inner)
}
