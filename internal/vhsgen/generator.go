package vhsgen

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

//go:embed config.tape
var defaultConfigContent string

const (
	defaultConfigSourcePath = "config/config.tape"
	defaultSleepDuration    = "2s"
	manualStepMarker        = "Manual step needed"
)

// resolveConfigPath returns the config path to use, a warning message (if any), and a cleanup function.
// If customPath is non-empty and the file exists, use it (cleanup is a no-op, warning is empty).
// Otherwise, write embedded config to a unique temp file and return that path.
// If customPath was provided but not found, a warning message is returned.
// The caller must call the cleanup function to remove any temp file created.
func resolveConfigPath(customPath string) (string, string, func(), error) {
	var warning string
	if customPath != "" {
		if _, err := os.Stat(customPath); err == nil {
			return customPath, "", func() {}, nil
		}
		warning = fmt.Sprintf("Warning: config file not found at %s, using embedded default. Run 'cukes-vhs init' to create one.\n", customPath)
	}

	// Fallback to embedded config in a unique temp file
	f, err := os.CreateTemp("", "vhsgen-*.tape")
	if err != nil {
		return "", "", func() {}, fmt.Errorf("creating temp config file: %w", err)
	}

	tmpPath := f.Name()

	if _, err := f.Write([]byte(defaultConfigContent)); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)

		return "", "", func() {}, fmt.Errorf("writing embedded config: %w", err)
	}

	if err := f.Close(); err != nil {
		_ = os.Remove(tmpPath)

		return "", "", func() {}, fmt.Errorf("closing temp config file: %w", err)
	}

	return tmpPath, warning, func() { _ = os.Remove(tmpPath) }, nil
}

// forbiddenPatterns returns patterns that must not appear in generated tape content.
// Returns a fresh slice each call to prevent mutation of shared state.
func forbiddenPatterns() []string {
	return []string{"rm -rf", "DELETE", "DROP"}
}

// GenerateTape produces VHS tape file content from a ScenarioIR and GeneratorConfig.
//
// Expected: scenario with populated setup/demo steps, config with OutputDir set.
// Returns: rendered tape content as string, or error if rendering or validation fails.
// Side effects: May create a temporary config file (cleaned up before returning).
func GenerateTape(scenario ScenarioIR, config GeneratorConfig) (string, error) {
	configSourcePath := config.ConfigSourcePath
	if configSourcePath == "" {
		configSourcePath = defaultConfigSourcePath
	}

	// Resolve config with fallback; clean up any temp file when done
	resolvedConfigPath, _, cleanup, err := resolveConfigPath(configSourcePath)
	if err != nil {
		return "", err
	}
	defer cleanup()

	sleepDuration := config.SleepDuration
	if sleepDuration == "" {
		sleepDuration = defaultSleepDuration
	}

	featureSlug := Slugify(scenario.Feature)
	scenarioSlug := Slugify(scenario.Name)

	data := TapeData{
		FeatureName:      scenario.Feature,
		ScenarioName:     scenario.Name,
		GIFPath:          filepath.Join(config.OutputDir, featureSlug, scenarioSlug+".gif"),
		ASCIIPath:        filepath.Join(config.OutputDir, featureSlug, scenarioSlug+".ascii"),
		ConfigSourcePath: resolvedConfigPath,
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

	featureSlug := Slugify(scenario.Feature)
	scenarioSlug := Slugify(scenario.Name)
	dir := filepath.Join(config.OutputDir, featureSlug)

	if err := os.MkdirAll(dir, 0o750); err != nil {
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
	for _, pattern := range forbiddenPatterns() {
		if strings.Contains(content, pattern) {
			return fmt.Errorf("generated tape contains forbidden pattern: %q", pattern)
		}
	}

	return nil
}

// slugStripRe and slugCollapseRe are compiled once at package init and never
// reassigned thereafter. *regexp.Regexp is safe for concurrent reads.
var (
	slugStripRe    = regexp.MustCompile(`[^a-z0-9-]`)
	slugCollapseRe = regexp.MustCompile(`-{2,}`)
)

// Slugify converts a string to a URL-safe slug.
func Slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	s = slugStripRe.ReplaceAllString(s, "")
	s = slugCollapseRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")

	return s
}

// DefaultConfig returns the embedded default VHS config content.
// This can be used to generate a config file for customisation.
func DefaultConfig() string {
	return defaultConfigContent
}
