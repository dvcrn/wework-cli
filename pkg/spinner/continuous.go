package spinner

import (
	"fmt"
	"os"
	"sync"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// ContinuousSpinner maintains a single spinner throughout multiple operations
type ContinuousSpinner struct {
	program    *tea.Program
	mu         sync.Mutex
	finished   bool
	messages   chan interface{}
	isTerminal bool
	noSpinner  bool
}

type continuousModel struct {
	spinner spinner.Model
	message string
	done    bool
	success bool
}

func (m continuousModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m continuousModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.done = true
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case updateMsg:
		m.message = msg.message
		return m, nil
	case doneMsg:
		m.done = true
		m.success = msg.err == nil
		return m, tea.Quit
	}
	return m, nil
}

func (m continuousModel) View() string {
	if m.done {
		if m.success {
			return lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("✓ " + m.message)
		}
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("✗ " + m.message)
	}
	return fmt.Sprintf("%s %s", m.spinner.View(), m.message)
}

// NewContinuousSpinner creates a spinner that stays active across multiple operations
func NewContinuousSpinner() *ContinuousSpinner {
	return &ContinuousSpinner{
		messages:   make(chan interface{}, 100),
		isTerminal: term.IsTerminal(int(os.Stdout.Fd())),
		noSpinner:  os.Getenv("WEWORK_NO_SPINNER") == "true" || os.Getenv("NO_SPINNER") == "true",
	}
}

// SetNoSpinner disables the spinner even if running in a terminal
func (cs *ContinuousSpinner) SetNoSpinner(noSpinner bool) {
	cs.noSpinner = noSpinner
}

// Start begins the continuous spinner
func (cs *ContinuousSpinner) Start(initialMessage string) {
	if !cs.isTerminal || cs.noSpinner {
		fmt.Println(initialMessage)
		return
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	model := continuousModel{
		spinner: s,
		message: initialMessage,
	}

	cs.program = tea.NewProgram(model)

	go func() {
		cs.program.Run()
	}()
}

// Update changes the spinner message without stopping it
func (cs *ContinuousSpinner) Update(message string) {
	if !cs.isTerminal || cs.noSpinner {
		fmt.Println(message)
		return
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.program != nil && !cs.finished {
		cs.program.Send(updateMsg{message: message})
	}
}

// Success stops the spinner with a success message
func (cs *ContinuousSpinner) Success(message string) {
	if !cs.isTerminal || cs.noSpinner {
		fmt.Printf("✓ %s\n", message)
		return
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.program != nil && !cs.finished {
		cs.finished = true
		cs.program.Send(updateMsg{message: message})
		cs.program.Send(doneMsg{err: nil})
	}
}

// Error stops the spinner with an error message
func (cs *ContinuousSpinner) Error(message string) {
	if !cs.isTerminal || cs.noSpinner {
		fmt.Printf("✗ %s\n", message)
		return
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.program != nil && !cs.finished {
		cs.finished = true
		cs.program.Send(updateMsg{message: message})
		cs.program.Send(doneMsg{err: fmt.Errorf(message)})
	}
}

// WithContinuousSpinner runs a series of operations with a continuous spinner
func WithContinuousSpinner(operations func(*ContinuousSpinner) error) error {
	cs := NewContinuousSpinner()
	cs.Start("Initializing...")

	err := operations(cs)

	if err != nil {
		cs.Error(err.Error())
	}

	return err
}

// WithContinuousSpinnerConfig runs operations with optional spinner disabling
func WithContinuousSpinnerConfig(noSpinner bool, operations func(*ContinuousSpinner) error) error {
	cs := NewContinuousSpinner()
	cs.SetNoSpinner(noSpinner)
	cs.Start("Initializing...")

	err := operations(cs)

	if err != nil {
		cs.Error(err.Error())
	}

	return err
}
