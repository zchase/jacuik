package terminal

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

type ProgressBarStep struct {
	Order int
	Label string
	Value float64
}

type ProgressBarArgs struct {
	Padding  int
	MaxWidth int
	Steps    []ProgressBarStep
}

type ProgressBar struct {
	CurrentStep int
	Steps       []ProgressBarStep

	element *tea.Program
}

func (p *ProgressBar) Done() {
	pbFlag := make(chan bool)
	p.element.Send(progressBarFinished(pbFlag))

	<-pbFlag

	// UX to make the finish less abrupt.
	time.Sleep(time.Millisecond * 500)
	p.element.Send(closeProgressBar(1))
	time.Sleep(time.Millisecond * 500)
}

func (p *ProgressBar) IncrementStep() {
	p.CurrentStep++
	currentStep := p.Steps[p.CurrentStep]
	p.element.Send(currentStep)
}

type programCreationResult struct {
	program *tea.Program
	err     error
}

func NewProgressBar(args ProgressBarArgs) (*ProgressBar, error) {
	if args.Padding == 0 {
		args.Padding = 2
	}

	if args.MaxWidth == 0 {
		args.MaxWidth = 80
	}

	var steps []ProgressBarStep
	for i, s := range args.Steps {
		if s.Value == 0 {
			s.Value = float64(1) / float64(len(args.Steps))
		}

		if s.Order == 0 {
			s.Order = i + 1
		}

		steps = append(steps, s)
	}

	program := make(chan programCreationResult)
	go func() {
		element := progressBarElement{
			padding:    args.Padding,
			maxWidth:   args.MaxWidth,
			progress:   progress.New(progress.WithDefaultGradient()),
			totalSteps: len(steps),
			data:       steps[0],
		}

		p := tea.NewProgram(element)
		program <- programCreationResult{
			program: p,
		}

		err := p.Start()
		program <- programCreationResult{
			program: p,
			err:     err,
		}
	}()

	programResult := <-program
	if programResult.err != nil {
		return nil, programResult.err
	}

	pb := &ProgressBar{
		Steps:   steps,
		element: programResult.program,
	}

	return pb, nil
}

type progressBarElement struct {
	progress   progress.Model
	padding    int
	maxWidth   int
	totalSteps int

	data ProgressBarStep
}

type progressBarFinished chan bool
type completeProgressBar chan bool
type closeProgressBar int

func (p progressBarElement) Init() tea.Cmd {
	return p.progress.IncrPercent(p.data.Value)
}

func (e progressBarElement) View() string {
	pad := strings.Repeat(" ", e.padding)
	data := e.data

	return "\n" +
		pad + e.progress.View() + "\n" +
		pad + fmt.Sprintf("%v/%v: %s\n\n", data.Order, e.totalSteps, data.Label) +
		pad + "Press any ctr+c to quit"
}

func (m progressBarElement) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if tea.KeyCtrlC.String() == msg.String() {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - m.padding*2 - 4
		if m.progress.Width > m.maxWidth {
			m.progress.Width = m.maxWidth
		}
		return m, nil

	case ProgressBarStep:
		m.data = msg
		cmd := m.progress.IncrPercent(m.data.Value)
		return m, cmd

	case progressBarFinished:
		cmd := m.progress.SetPercent(1)
		return m, tea.Batch(closeProgressBarListener(msg), cmd)

	case completeProgressBar:
		if m.progress.Percent() == 1.0 {
			msg <- true
		}

		return m, nil

	case closeProgressBar:
		return m, tea.Quit

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	}

	return m, nil
}

func closeProgressBarListener(pbFlag chan bool) tea.Cmd {
	return tea.Tick(time.Millisecond*250, func(t time.Time) tea.Msg {
		return completeProgressBar(pbFlag)
	})
}
