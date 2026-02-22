// Package tui provides Bubble Tea-based terminal user interface components
// for the cukes-vhs CLI, providing interactive progress display during
// tape generation and rendering pipelines.
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Stage represents the current pipeline stage.
//
// Expected: used as an enum via the defined constants.
// Returns: the stage value.
type Stage int

const (
	// Parsing indicates the feature-file parsing stage.
	Parsing Stage = iota
	// Generating indicates the tape generation stage.
	Generating
	// Rendering indicates the VHS rendering stage.
	Rendering
	// Validating indicates the golden-file validation stage.
	Validating
	// Complete indicates all stages have finished.
	Complete
)

// String returns the human-readable name of the stage.
//
// Expected: a valid Stage constant.
// Returns: the stage name as a string.
func (s Stage) String() string {
	switch s {
	case Parsing:
		return "Parsing"
	case Generating:
		return "Generating"
	case Rendering:
		return "Rendering"
	case Validating:
		return "Validating"
	case Complete:
		return "Complete"
	default:
		return "Unknown"
	}
}

// ScenarioStatus represents the processing status of a single scenario.
//
// Expected: used as an enum via the defined constants.
// Returns: the status value.
type ScenarioStatus int

const (
	// Pending indicates the scenario has not yet started processing.
	Pending ScenarioStatus = iota
	// InProgress indicates the scenario is currently being processed.
	InProgress
	// Done indicates the scenario completed successfully.
	Done
	// Failed indicates the scenario encountered an error.
	Failed
)

// scenarioEntry holds the name and current status of a scenario in the
// progress display.
type scenarioEntry struct {
	name   string
	status ScenarioStatus
}

// ProgressModel is the shared Bubble Tea model for displaying pipeline
// progress across generate and run commands.
//
// Expected: created via NewProgressModel with a list of scenario names.
// Returns: a model ready for use with tea.NewProgram or embedded in a
// parent model.
type ProgressModel struct {
	stage     Stage
	scenarios []scenarioEntry
	progress  progress.Model
	current   int
	total     int
	done      bool
	err       error
	width     int
}

// Styles for the TUI components.
var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
	errorStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF4444"))
	normalStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#DDDDDD"))
)

// NewProgressModel creates a new progress model initialised with the given
// scenario names. All scenarios start in Pending status.
//
// Expected: scenarioNames contains the display names for each scenario;
// may be nil or empty for a progress-only display.
// Returns: a ProgressModel ready for embedding or standalone use.
// Side effects: none.
func NewProgressModel(scenarioNames []string) ProgressModel {
	entries := make([]scenarioEntry, len(scenarioNames))
	for i := range scenarioNames {
		entries[i] = scenarioEntry{
			name:   scenarioNames[i],
			status: Pending,
		}
	}

	prog := progress.New(
		progress.WithDefaultGradient(),
		progress.WithoutPercentage(),
	)

	return ProgressModel{
		stage:     Parsing,
		scenarios: entries,
		progress:  prog,
		total:     len(scenarioNames),
	}
}

// Init implements tea.Model.
//
// Expected: called by the Bubble Tea runtime on programme start.
// Returns: nil (no initial command).
// Side effects: none.
func (m ProgressModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
//
// Expected: msg is a tea.Msg from the Bubble Tea runtime.
// Returns: the updated model and any command to execute.
// Side effects: none.
func (m ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.progress.Width = msg.Width - 4
		return m, nil
	}
	return m, nil
}

// View implements tea.Model.
//
// Expected: called by the Bubble Tea runtime to render the display.
// Returns: the full TUI view as a string.
// Side effects: none.
func (m ProgressModel) View() string {
	var b strings.Builder

	// Title bar with stage.
	title := titleStyle.Render(fmt.Sprintf("cukes-vhs  [%s]", m.stage))
	b.WriteString(title)
	b.WriteString("\n\n")

	// Progress bar.
	if m.total > 0 {
		frac := float64(m.current) / float64(m.total)
		bar := m.progress.ViewAs(frac)
		b.WriteString(bar)
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  %d/%d scenarios", m.current, m.total))
		b.WriteString("\n\n")
	}

	// Scenario list with status indicators.
	for i := range m.scenarios {
		s := &m.scenarios[i]
		var indicator string
		var style lipgloss.Style

		switch s.status {
		case Done:
			indicator = "✓"
			style = successStyle
		case Failed:
			indicator = "✗"
			style = errorStyle
		case InProgress:
			indicator = "●"
			style = normalStyle
		default:
			indicator = "○"
			style = normalStyle
		}

		line := fmt.Sprintf("  %s %s", indicator, s.name)
		b.WriteString(style.Render(line))
		b.WriteString("\n")
	}

	// Error display.
	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n")
	}

	return b.String()
}

// UpdateStage advances the pipeline to the given stage.
//
// Expected: stage is a valid Stage constant.
// Returns: nothing.
// Side effects: mutates the model's stage field.
func (m *ProgressModel) UpdateStage(stage Stage) {
	m.stage = stage
}

// UpdateScenario updates the status of the named scenario.
//
// Expected: name matches a scenario previously passed to NewProgressModel;
// status is a valid ScenarioStatus constant.
// Returns: nothing.
// Side effects: mutates the matching scenario's status and increments
// the current counter for Done or Failed statuses.
func (m *ProgressModel) UpdateScenario(name string, status ScenarioStatus) {
	for i := range m.scenarios {
		s := &m.scenarios[i]
		if s.name == name {
			s.status = status
			if status == Done || status == Failed {
				m.current++
			}
			break
		}
	}
}

// Completed marks the progress model as done and sets the stage to Complete.
//
// Expected: called when all pipeline work has finished.
// Returns: nothing.
// Side effects: sets done to true and stage to Complete.
func (m *ProgressModel) Completed() {
	m.done = true
	m.stage = Complete
}

// SetError records an error on the progress model.
//
// Expected: err is the error to display.
// Returns: nothing.
// Side effects: stores the error for display in View.
func (m *ProgressModel) SetError(err error) {
	m.err = err
}
