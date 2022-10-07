package terminal

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type ChoiceModel struct {
	exit    bool
	cursor  int
	choice  string
	options []string
	prompt  string
}

func (c ChoiceModel) Init() tea.Cmd {
	return nil
}

func (c ChoiceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case tea.KeyCtrlC.String():
			c.exit = true
			return c, tea.Quit

		case tea.KeyEnter.String():
			// Send the choice on the channel and exit.
			c.choice = c.options[c.cursor]
			return c, tea.Quit

		case tea.KeyDown.String():
			c.cursor++
			if c.cursor >= len(c.options) {
				c.cursor = 0
			}

		case tea.KeyUp.String():
			c.cursor--
			if c.cursor < 0 {
				c.cursor = len(c.options) - 1
			}
		}

	}

	return c, nil
}

func (c ChoiceModel) View() string {
	s := strings.Builder{}
	s.WriteString(fmt.Sprintf("%s\n\n", c.prompt))

	for i := 0; i < len(c.options); i++ {
		if c.cursor == i {
			s.WriteString("> ")
		} else {
			s.WriteString("  ")
		}

		s.WriteString(c.options[i])
		s.WriteString("\n")
	}

	s.WriteString("\n")
	return s.String()
}

func NewChoicePrompt(prompt string, choices []string) (string, error) {
	p := tea.NewProgram(ChoiceModel{
		options: choices,
		prompt:  prompt,
	})

	m, err := p.StartReturningModel()
	if err != nil {
		return "", err
	}

	result, ok := m.(ChoiceModel)
	if !ok {
		return "", fmt.Errorf("Invalid result not a valid ChoiceModel. If you are seeing this error please file an issue.")
	}

	if result.exit == true {
		return "", fmt.Errorf("No valid option selected.")
	}

	return result.choice, nil
}
