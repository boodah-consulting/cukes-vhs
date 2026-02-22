package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/boodah-consulting/cukes-vhs/internal/vhsgen"
)

// generateModel is the Bubble Tea model for the generate pipeline.
// It orchestrates parsing, tape generation, and rendering stages.
//
// Expected: created via NewGenerateModel with directory paths and config.
// Returns: a tea.Model suitable for tea.NewProgram.
type generateModel struct {
	progress     ProgressModel
	featuresDir  string
	scenariosDir string
	outputDir    string
	configSource string
	timeoutSec   int
	scenarios    []vhsgen.ScenarioIR
	err          error
}

// parsedMsg carries the parsed scenarios from the parsing stage.
type parsedMsg struct {
	scenarios []vhsgen.ScenarioIR
}

// generatedMsg signals that a single tape was generated successfully.
type generatedMsg struct {
	name string
	path string
}

// renderedMsg carries the render results from the rendering stage.
type renderedMsg struct {
	results []vhsgen.RenderResult
}

// errMsg carries an error from any pipeline stage.
type errMsg struct {
	err error
}

// Error implements the error interface for errMsg.
//
// Expected: a non-nil err field.
// Returns: the underlying error message.
func (e errMsg) Error() string {
	return e.err.Error()
}

// NewGenerateModel creates a new generate model for the tape generation
// pipeline.
//
// Expected: featuresDir and scenariosDir are readable directories;
// outputDir is a writable directory; configSource is the path to the
// config source file; timeoutSec is the render timeout in seconds.
// Returns: a generateModel ready for use with tea.NewProgram.
// Side effects: none.
func NewGenerateModel(
	featuresDir, scenariosDir, outputDir, configSource string,
	timeoutSec int,
) generateModel {
	return generateModel{
		progress:     NewProgressModel(nil),
		featuresDir:  featuresDir,
		scenariosDir: scenariosDir,
		outputDir:    outputDir,
		configSource: configSource,
		timeoutSec:   timeoutSec,
	}
}

// Init implements tea.Model. Fires the initial parse command.
//
// Expected: called by the Bubble Tea runtime on programme start.
// Returns: a command that triggers scenario parsing.
// Side effects: none.
func (m generateModel) Init() tea.Cmd {
	return m.parseCmd
}

// Update implements tea.Model. Routes messages through the generate
// pipeline stages: parse → generate (one-by-one) → render.
//
// Expected: msg is a tea.Msg from the Bubble Tea runtime.
// Returns: the updated model and any command to execute.
// Side effects: none (side effects happen in commands).
func (m generateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}

	case parsedMsg:
		m.scenarios = msg.scenarios
		names := make([]string, len(m.scenarios))
		for i := range m.scenarios {
			names[i] = m.scenarios[i].Name
		}
		m.progress = NewProgressModel(names)
		m.progress.UpdateStage(Generating)
		if len(m.scenarios) > 0 {
			return m, m.generateOneCmd(0)
		}
		m.progress.Completed()
		return m, tea.Quit

	case generatedMsg:
		m.progress.UpdateScenario(msg.name, Done)
		// Find the next scenario to generate.
		nextIdx := m.findNextPending()
		if nextIdx >= 0 {
			return m, m.generateOneCmd(nextIdx)
		}
		// All generated; move to rendering.
		m.progress.UpdateStage(Rendering)
		return m, m.renderCmd

	case renderedMsg:
		for i := range msg.results {
			r := &msg.results[i]
			name := deriveScenarioName(r.TapePath)
			if r.Success {
				m.progress.UpdateScenario(name, Done)
			} else {
				m.progress.UpdateScenario(name, Failed)
			}
		}
		m.progress.Completed()
		return m, tea.Quit

	case errMsg:
		m.err = msg.err
		m.progress.SetError(msg.err)
		return m, tea.Quit
	}

	return m, nil
}

// View implements tea.Model.
//
// Expected: called by the Bubble Tea runtime to render the display.
// Returns: the progress view as a string.
// Side effects: none.
func (m generateModel) View() string {
	return m.progress.View()
}

// parseCmd parses feature files from both business and VHS-only directories.
//
// Expected: m.featuresDir and m.scenariosDir are readable directories.
// Returns: a parsedMsg with all translatable scenarios, or an errMsg.
// Side effects: reads the filesystem.
func (m generateModel) parseCmd() tea.Msg {
	if _, err := os.Stat(m.featuresDir); err != nil {
		return errMsg{fmt.Errorf("checking features directory: %w", err)}
	}

	businessScenarios, err := vhsgen.ParseFeatureDir(m.featuresDir, vhsgen.SourceBusiness)
	if err != nil {
		return errMsg{fmt.Errorf("parsing business features: %w", err)}
	}

	var vhsOnlyScenarios []vhsgen.ScenarioIR
	if _, statErr := os.Stat(m.scenariosDir); statErr == nil {
		parsed, parseErr := vhsgen.ParseFeatureDir(m.scenariosDir, vhsgen.SourceVHSOnly)
		if parseErr != nil {
			return errMsg{fmt.Errorf("parsing VHS-only scenarios: %w", parseErr)}
		}
		vhsOnlyScenarios = parsed
	}

	allScenarios := make([]vhsgen.ScenarioIR, 0, len(businessScenarios)+len(vhsOnlyScenarios))
	allScenarios = append(allScenarios, businessScenarios...)
	allScenarios = append(allScenarios, vhsOnlyScenarios...)

	// Filter to translatable scenarios only.
	results := vhsgen.AnalyseScenarios(allScenarios)
	translatableNames := make(map[string]bool, len(results))
	for i := range results {
		r := &results[i]
		if r.Translatable {
			translatableNames[r.ScenarioName] = true
		}
	}

	filtered := make([]vhsgen.ScenarioIR, 0, len(translatableNames))
	for i := range allScenarios {
		s := &allScenarios[i]
		if translatableNames[s.Name] {
			filtered = append(filtered, *s)
		}
	}

	return parsedMsg{scenarios: filtered}
}

// generateOneCmd returns a command that generates a single tape file for
// the scenario at the given index.
//
// Expected: idx is a valid index into m.scenarios.
// Returns: a tea.Cmd that produces a generatedMsg or errMsg.
// Side effects: the returned command writes files to the filesystem.
func (m generateModel) generateOneCmd(idx int) tea.Cmd {
	scenario := m.scenarios[idx]
	outputDir := m.outputDir
	configSource := m.configSource

	return func() tea.Msg {
		featureSlug := vhsgen.Slugify(scenario.Feature)
		scenarioSlug := vhsgen.Slugify(scenario.Name)

		var tapeDir string
		switch scenario.Source {
		case vhsgen.SourceVHSOnly:
			tapeDir = filepath.Join(outputDir, "scenarios", featureSlug)
		default:
			tapeDir = filepath.Join(outputDir, featureSlug)
		}

		config := vhsgen.GeneratorConfig{
			OutputDir:        outputDir,
			ConfigSourcePath: configSource,
		}

		content, err := vhsgen.GenerateTape(scenario, config)
		if err != nil {
			return errMsg{fmt.Errorf("generating tape for %s: %w", scenario.Name, err)}
		}

		if err := os.MkdirAll(tapeDir, 0o755); err != nil {
			return errMsg{fmt.Errorf("creating tape directory: %w", err)}
		}

		outPath := filepath.Join(tapeDir, scenarioSlug+".tape")
		if err := os.WriteFile(outPath, []byte(content), 0o600); err != nil {
			return errMsg{fmt.Errorf("writing tape file: %w", err)}
		}

		return generatedMsg{name: scenario.Name, path: outPath}
	}
}

// renderCmd renders all generated tapes via the VHS renderer.
//
// Expected: m.outputDir contains generated .tape files.
// Returns: a renderedMsg with results, or an errMsg.
// Side effects: spawns VHS child processes.
func (m generateModel) renderCmd() tea.Msg {
	renderer := vhsgen.NewRenderer()
	timeout := time.Duration(m.timeoutSec) * time.Second

	results, err := renderer.RenderAll(m.outputDir, timeout)
	if err != nil {
		return errMsg{fmt.Errorf("rendering tapes: %w", err)}
	}

	return renderedMsg{results: results}
}

// findNextPending returns the index of the next scenario with Pending status,
// or -1 if none remain.
//
// Expected: m.scenarios is populated.
// Returns: the index of the next pending scenario, or -1.
// Side effects: none.
func (m generateModel) findNextPending() int {
	for i := range m.progress.scenarios {
		s := &m.progress.scenarios[i]
		if s.status == Pending {
			return i
		}
	}
	return -1
}

// RunGenerate runs the generate pipeline with TUI.
//
// Expected: all directory paths are valid; timeoutSec > 0.
// Returns: nil on success, or an error if the TUI programme fails.
// Side effects: runs a full-screen TUI, reads and writes files.
func RunGenerate(
	featuresDir, scenariosDir, outputDir, configSource string,
	timeoutSec int,
) error {
	model := NewGenerateModel(
		featuresDir, scenariosDir, outputDir,
		configSource, timeoutSec,
	)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("running generate TUI: %w", err)
	}

	return nil
}

// deriveScenarioName extracts the scenario slug from a tape file path
// by stripping the .tape extension and returning the last path component.
//
// Expected: tapePath ends with ".tape".
// Returns: the scenario slug portion of the path.
// Side effects: none.
func deriveScenarioName(tapePath string) string {
	base := filepath.Base(tapePath)
	ext := filepath.Ext(base)
	if ext != "" {
		base = base[:len(base)-len(ext)]
	}
	return base
}
