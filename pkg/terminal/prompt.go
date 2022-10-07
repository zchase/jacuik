package terminal

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg struct{}
type errMsg error

type TextPromptModel struct {
	exit        bool
	defaultName string
	prompt      string
	textInput   textinput.Model
	err         error
}

func (m TextPromptModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m TextPromptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.exit = true
			return m, tea.Quit
		case tea.KeyEnter:
			return m, tea.Quit
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m TextPromptModel) View() string {
	return fmt.Sprintf(
		"%s\n\n%s\n\n(press enter to accept default [%s])",
		m.prompt,
		m.textInput.View(),
		m.defaultName,
	) + "\n"
}

func NewTextPrompt(prompt, defaultName string) (string, error) {
	input := textinput.New()
	input.Placeholder = defaultName
	input.Focus()
	input.CharLimit = 128
	input.Width = 20

	p := tea.NewProgram(TextPromptModel{
		defaultName: defaultName,
		prompt:      prompt,
		textInput:   input,
		err:         nil,
	})

	m, err := p.StartReturningModel()
	if err != nil {
		return "", err
	}

	result, ok := m.(TextPromptModel)
	if !ok {
		return "", fmt.Errorf("Invalid result not a valid TextPromptModel. If you are seeing this error please file an issue.")
	}

	if result.exit == true {
		return "", fmt.Errorf("No value supplied.")
	}

	resultText := result.textInput.Value()
	if resultText == "" {
		resultText = defaultName
	}

	return resultText, nil
}
