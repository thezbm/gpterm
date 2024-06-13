package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gabezbm/gpterm/bot"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var (
	youStyle           = lipgloss.NewStyle().Foreground(lipgloss.Color("#ed8796")).Bold(true).Border(lipgloss.RoundedBorder(), true, true, true, true)
	botStyle           = lipgloss.NewStyle().Foreground(lipgloss.Color("#7dc4e4")).Bold(true).Border(lipgloss.RoundedBorder(), true, true, true, true)
	messageRenderer, _ = glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(120),
	)
	spinnerStyle = spinner.Spinner{
		Frames: []string{"∙∙∙∙∙", "●∙∙∙∙", "∙●∙∙∙", "∙∙●∙∙", "∙∙∙●∙", "∙∙∙∙●"},
		FPS:    time.Second / 5,
	}
)

func main() {
	bot.SetUp()
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type model struct {
	viewport viewport.Model
	textarea textarea.Model
	spinner  spinner.Model
	messages []string
}

func initialModel() model {
	vp := viewport.New(120, 35)
	vp.SetContent(`Have a chat with Bot!
Type a message and press Enter to send.`)
	vp.KeyMap = viewport.KeyMap{
		PageUp: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown"),
			key.WithHelp("pgdn", "page down"),
		),
		Up: key.NewBinding(
			key.WithKeys("ctrl+up"),
			key.WithHelp("ctrl+up", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("ctrl+down"),
			key.WithHelp("ctrl+down", "down"),
		),
	}

	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.Prompt = "┃ "
	ta.ShowLineNumbers = false
	ta.SetWidth(120)
	ta.SetHeight(5)
	ta.CharLimit = 1000
	ta.KeyMap.InsertNewline.SetEnabled(false)

	sp := spinner.New()
	sp.Spinner = spinnerStyle

	return model{
		viewport: vp,
		textarea: ta,
		spinner:  sp,
		messages: []string{},
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		vpCmd tea.Cmd
		taCmd tea.Cmd
		spCmd tea.Cmd
	)
	m.viewport, vpCmd = m.viewport.Update(msg)
	m.textarea, taCmd = m.textarea.Update(msg)

	type botMsg struct {
		content string
	}
	type errMsg error

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if !m.textarea.Focused() {
				return m, nil
			}
			m.textarea.Blur()
			input := m.textarea.Value()
			m.textarea.Reset()
			renderedInput, err := messageRenderer.Render(input)
			if err != nil {
				log.Fatal(err)
			}
			m.messages = append(m.messages, youStyle.Render("You: ")+renderedInput)
			m.messages = append(m.messages, botStyle.Render(m.spinner.View())+"\n")
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()
			return m, tea.Batch(func() tea.Msg { return botMsg{bot.Ask(input)} }, m.spinner.Tick, vpCmd, taCmd)
		}
	case botMsg:
		renderedOutput, err := messageRenderer.Render(msg.content)
		if err != nil {
			log.Fatal(err)
		}
		m.messages[len(m.messages)-1] = botStyle.Render(fmt.Sprintf("Bot (%s): ", bot.GetModel())) + renderedOutput
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()
		m.textarea.Focus()
	case spinner.TickMsg:
		if m.textarea.Focused() {
			return m, nil
		}
		m.spinner, spCmd = m.spinner.Update(msg)
		m.messages[len(m.messages)-1] = botStyle.Render(m.spinner.View())+"\n"
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		return m, spCmd
	case errMsg:
		log.Fatal(msg)
		return m, nil
	}

	return m, tea.Batch(vpCmd, taCmd)
}

func (m model) View() string {
	return fmt.Sprintf(
		"\n\n%s\n\n%s",
		m.viewport.View(),
		m.textarea.View(),
	) + "\n\n"
}
