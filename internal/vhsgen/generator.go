package vhsgen

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	defaultConfigSourcePath = "demos/vhs/config.tape"
	defaultSleepDuration    = "2s"
	manualStepMarker        = "Manual step needed"
)

var forbiddenPatterns = []string{"rm -rf", "DELETE", "DROP"}

// GenerateTape produces VHS tape file content from a ScenarioIR and GeneratorConfig.
//
// Expected: scenario with populated setup/demo steps, config with OutputDir set.
// Returns: rendered tape content as string, or error if rendering or validation fails.
// Side effects: None.
func GenerateTape(scenario ScenarioIR, config GeneratorConfig) (string, error) {
	configSourcePath := config.ConfigSourcePath
	if configSourcePath == "" {
		configSourcePath = defaultConfigSourcePath
	}

	sleepDuration := config.SleepDuration
	if sleepDuration == "" {
		sleepDuration = defaultSleepDuration
	}

	featureSlug := slugify(scenario.Feature)
	scenarioSlug := slugify(scenario.Name)

	data := TapeData{
		FeatureName:      scenario.Feature,
		ScenarioName:     scenario.Name,
		GIFPath:          filepath.Join(config.OutputDir, featureSlug, scenarioSlug+".gif"),
		ASCIIPath:        filepath.Join(config.OutputDir, featureSlug, scenarioSlug+".ascii"),
		ConfigSourcePath: configSourcePath,
		SetupCommands:    renderSteps(scenario.SetupSteps, sleepDuration),
		DemoCommands:     renderSteps(scenario.DemoSteps, sleepDuration),
	}

	result, err := RenderTape(data)
	if err != nil {
		return "", fmt.Errorf("generating tape for %q: %w", scenario.Name, err)
	}

	if err := validateOutput(result); err != nil {
		return "", err
	}

	return result, nil
}

// WriteTape generates tape content and writes it to a file.
//
// Expected: scenario and config with valid OutputDir.
// Returns: error if generation or file writing fails.
// Side effects: Creates directories and writes tape file to disk.
func WriteTape(scenario ScenarioIR, config GeneratorConfig) error {
	content, err := GenerateTape(scenario, config)
	if err != nil {
		return err
	}

	featureSlug := slugify(scenario.Feature)
	scenarioSlug := slugify(scenario.Name)
	dir := filepath.Join(config.OutputDir, featureSlug)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating output directory %q: %w", dir, err)
	}

	outPath := filepath.Join(dir, scenarioSlug+".tape")

	if err := os.WriteFile(outPath, []byte(content), 0o600); err != nil {
		return fmt.Errorf("writing tape file %q: %w", outPath, err)
	}

	return nil
}

func renderSteps(steps []StepIR, sleepDuration string) string {
	var lines []string
	lastHadCommands := false

	for _, step := range steps {
		if !step.Translatable {
			lines = append(lines, fmt.Sprintf(
				"# [%s] — %s (%s)",
				manualStepMarker, step.Text, step.UntranslatableReason,
			))
			lastHadCommands = false

			continue
		}

		if len(step.Commands) == 0 {
			continue
		}

		if lastHadCommands {
			lines = append(lines, "Sleep "+sleepDuration)
		}

		for _, cmd := range step.Commands {
			lines = append(lines, renderCommand(cmd))
		}

		lastHadCommands = true
	}

	return strings.Join(lines, "\n")
}

func renderCommand(cmd VHSCommand) string {
	switch cmd.Type {
	case Type:
		if len(cmd.Args) >= 2 {
			return fmt.Sprintf("Type@%s %q", cmd.Args[0], cmd.Args[1])
		}

		if len(cmd.Args) == 1 {
			return fmt.Sprintf("Type %q", cmd.Args[0])
		}

		return "Type"
	case Sleep:
		if len(cmd.Args) >= 1 {
			return "Sleep " + cmd.Args[0]
		}

		return "Sleep 1s"
	default:
		if len(cmd.Args) >= 1 {
			return fmt.Sprintf("%s %s", string(cmd.Type), cmd.Args[0])
		}

		return string(cmd.Type)
	}
}

func validateOutput(content string) error {
	for _, pattern := range forbiddenPatterns {
		if strings.Contains(content, pattern) {
			return fmt.Errorf("generated tape contains forbidden pattern: %q", pattern)
		}
	}

	return nil
}

var (
	slugStripRe    = regexp.MustCompile(`[^a-z0-9-]`)
	slugCollapseRe = regexp.MustCompile(`-{2,}`)
)

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	s = slugStripRe.ReplaceAllString(s, "")
	s = slugCollapseRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")

	return s
}
