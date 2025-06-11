package spinner

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SpinnerModel represents the loading spinner model
type SpinnerModel struct {
	spinner spinner.Model
	message string
	done    bool
	err     error
	result  interface{}
}

// Init initializes the spinner
func (m SpinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update handles messages
func (m SpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
	case doneMsg:
		m.done = true
		m.result = msg.result
		m.err = msg.err
		return m, tea.Quit
	case updateMsg:
		m.message = msg.message
		return m, nil
	}
	return m, nil
}

// View renders the spinner
func (m SpinnerModel) View() string {
	if m.done {
		if m.err != nil {
			return lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("✗ " + m.message + " failed")
		}
		return lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("✓ " + m.message + " complete")
	}
	return fmt.Sprintf("%s %s", m.spinner.View(), m.message)
}

type doneMsg struct {
	result interface{}
	err    error
}

type updateMsg struct {
	message string
}

// RunWithSpinner executes a function while showing a spinner
func RunWithSpinner(message string, fn func() (interface{}, error)) (interface{}, error) {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	m := SpinnerModel{
		spinner: s,
		message: message,
	}

	// Create program
	p := tea.NewProgram(m)

	// Run the function in a goroutine
	done := make(chan struct{})
	var result interface{}
	var err error

	go func() {
		result, err = fn()
		p.Send(doneMsg{result: result, err: err})
		close(done)
	}()

	// Run the spinner
	if _, err := p.Run(); err != nil {
		return nil, err
	}

	// Wait for the function to complete
	<-done

	// Add a small delay to show the completion message
	time.Sleep(100 * time.Millisecond)

	return result, err
}

// UpdateMessage sends a message update to the spinner
func UpdateMessage(p *tea.Program, message string) {
	p.Send(updateMsg{message: message})
}

// SequentialSpinner runs multiple operations with a spinner that updates its message
type SequentialSpinner struct {
	program  *tea.Program
	model    SpinnerModel
	finished bool
}

// NewSequentialSpinner creates a new sequential spinner
func NewSequentialSpinner() *SequentialSpinner {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	model := SpinnerModel{
		spinner: s,
		message: "Initializing...",
	}

	return &SequentialSpinner{
		model: model,
	}
}

// Start begins the spinner
func (ss *SequentialSpinner) Start() {
	ss.program = tea.NewProgram(ss.model)
	go func() {
		ss.program.Run()
	}()
	time.Sleep(50 * time.Millisecond) // Minimal delay to ensure it starts
}

// Update changes the spinner message
func (ss *SequentialSpinner) Update(message string) {
	if ss.program != nil && !ss.finished {
		ss.program.Send(updateMsg{message: message})
		time.Sleep(50 * time.Millisecond) // Minimal delay for update
	}
}

// Stop stops the spinner with success or error
func (ss *SequentialSpinner) Stop(err error) {
	if ss.program != nil && !ss.finished {
		ss.finished = true
		ss.program.Send(doneMsg{err: err})
		time.Sleep(100 * time.Millisecond) // Reduced delay for smoother transition
		ss.program.Quit()
	}
}
