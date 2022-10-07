package terminal

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/events"
	"github.com/zchase/jacuik/pkg/infrastructure"
)

type exitResourceView int

type resourceViewError struct {
	err error
}

type ResourceView struct {
	EventHandler chan<- events.EngineEvent
	program      *tea.Program
}

type resourceOutputUpdate struct {
	Hash  string
	Value infrastructure.ResourceOutput
	Count int
	Order int
}

type resourceOutputRender resourceOutputUpdate

func (v *ResourceView) Start() error {
	return v.program.Start()
}

func (v *ResourceView) Update(hash string, resource infrastructure.ResourceOutput) {
	v.program.Send(resourceOutputUpdate{
		Hash:  hash,
		Value: resource,
	})
}

func NewView(handler func(writer io.Writer) error) *ResourceView {
	model := resourceViewModel{
		resources:             make(map[string]resourceOutputUpdate),
		resourceListener:      make(chan infrastructure.ResourceOutput),
		resourceActionHandler: handler,
	}

	program := tea.NewProgram(model)

	return &ResourceView{
		EventHandler: make(chan<- events.EngineEvent),
		program:      program,
	}
}

type resourceViewModel struct {
	resourceListener      chan infrastructure.ResourceOutput
	resources             map[string]resourceOutputUpdate
	resourceActionHandler func(writer io.Writer) error
	updateQueue           []infrastructure.ResourceOutput
}

func watchForEvents(event chan infrastructure.ResourceOutput) tea.Cmd {
	return func() tea.Msg {
		return <-event
	}
}

func processEvents(updates []infrastructure.ResourceOutput) tea.Cmd {
	return tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
		if len(updates) == 0 {
			return nil
		}

		resourcesToUpdateSize := len(updates)
		if resourcesToUpdateSize > 1 {
			resourcesToUpdateSize = 1
		}

		return processResourceUpdates(resourcesToUpdateSize)
	})
}

func checkEventQueueForExit() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return pulumiProgramFinished(1)
	})
}

func (r resourceViewModel) Init() tea.Cmd {
	infraOutput := infrastructure.InfrastructureOutput{
		WriteChannel: r.resourceListener,
	}

	handler := func() tea.Msg {
		err := r.resourceActionHandler(infraOutput)
		if err != nil {
			return pulumiProgramError(err.Error())
		}

		return pulumiProgramFinished(1)
	}

	return tea.Batch(handler, watchForEvents(r.resourceListener))
}

type processResourceUpdates int
type pulumiProgramFinished int
type pulumiProgramError string

func (r resourceViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if tea.KeyCtrlC.String() == msg.String() {
			return r, tea.Quit
		}
	case infrastructure.ResourceOutput:
		startProcessing := len(r.updateQueue) == 0
		r.updateQueue = append(r.updateQueue, msg)

		if startProcessing {
			return r, tea.Batch(watchForEvents(r.resourceListener), processEvents(r.updateQueue))
		}

		return r, tea.Batch(watchForEvents(r.resourceListener))
	case processResourceUpdates:
		resourcesToUpdate := r.updateQueue[:msg]
		r.updateQueue = r.updateQueue[msg:]

		for _, u := range resourcesToUpdate {
			count := 1
			order := len(r.resources)
			res, ok := r.resources[u.Hash]
			if ok {
				count = res.Count + 1
				order = res.Order
			}

			r.resources[u.Hash] = resourceOutputUpdate{
				Count: count,
				Order: order,
				Value: infrastructure.ResourceOutput{
					URN:    u.URN,
					Status: u.Status,
				},
			}
		}

		return r, processEvents(r.updateQueue)

	case pulumiProgramFinished:
		if len(r.updateQueue) > 0 {
			return r, checkEventQueueForExit()
		}

		return r, tea.Quit

	// TODO: proper error handling
	case pulumiProgramError:
		fmt.Println(msg)
		return r, tea.Quit

	case exitResourceView:
		return r, tea.Quit
	}

	return r, nil
}

func renderContent(resources map[string]resourceOutputUpdate) string {
	s := strings.Builder{}
	s.WriteString("    Services\n\n")

	var sortedResources []resourceOutputUpdate
	for _, v := range resources {
		sortedResources = append(sortedResources, v)
	}
	sort.SliceStable(sortedResources, func(x, y int) bool {
		one := sortedResources[x]
		two := sortedResources[y]
		return one.Order < two.Order
	})

	for _, v := range sortedResources {
		resource := v
		s.WriteString(fmt.Sprintf("        %s %v %s\n", resource.Value.Status, resource.Count, resource.Value.URN))
	}

	s.WriteString("\n\nPress ctr+c to exit\n\n")
	return s.String()
}

func (r resourceViewModel) View() string {
	content := renderContent(r.resources)
	s := fmt.Sprintf("%s\n", content)
	return s
}
