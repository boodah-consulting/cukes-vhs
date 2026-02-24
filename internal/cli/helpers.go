package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/boodah-consulting/cukesvhs/internal/vhsgen"
)

type scenarioWithResult struct {
	scenario vhsgen.ScenarioIR
	result   vhsgen.AnalysisResult
}

func parseAllScenarios(featuresDir, scenariosDir string, errOut io.Writer) ([]vhsgen.ScenarioIR, error) {
	if _, err := os.Stat(featuresDir); err != nil {
		fmt.Fprintf(errOut, "Error parsing features dir %q: %v\n", featuresDir, err)
		return nil, err
	}

	businessScenarios, err := vhsgen.ParseFeatureDir(featuresDir, vhsgen.SourceBusiness)
	if err != nil {
		fmt.Fprintf(errOut, "Error parsing features dir %q: %v\n", featuresDir, err)
		return nil, err
	}

	vhsOnlyScenarios, err := vhsgen.ParseFeatureDir(scenariosDir, vhsgen.SourceVHSOnly)
	if err != nil {
		fmt.Fprintf(errOut, "Error parsing scenarios dir %q: %v\n", scenariosDir, err)
		return nil, err
	}

	allScenarios := make([]vhsgen.ScenarioIR, 0, len(businessScenarios)+len(vhsOnlyScenarios))
	allScenarios = append(allScenarios, businessScenarios...)
	allScenarios = append(allScenarios, vhsOnlyScenarios...)
	return allScenarios, nil
}

// filterResults selects scenarios to generate based on flags.
func filterResults(
	results []vhsgen.AnalysisResult,
	scenarios []vhsgen.ScenarioIR,
	all bool,
	featureFilter, scenarioFilter string,
) []scenarioWithResult {
	var out []scenarioWithResult

	resultByName := make(map[string]vhsgen.AnalysisResult, len(results))
	for i := range results {
		r := &results[i]
		resultByName[r.ScenarioName] = *r
	}

	for i := range scenarios {
		s := &scenarios[i]
		result, ok := resultByName[s.Name]
		if !ok {
			continue
		}

		if !all {
			if featureFilter != "" && !strings.EqualFold(s.Feature, featureFilter) {
				continue
			}
			if scenarioFilter != "" && !strings.EqualFold(s.Name, scenarioFilter) {
				continue
			}
		}

		out = append(out, scenarioWithResult{scenario: *s, result: result})
	}

	return out
}

// writeScenarioTape generates and writes a tape file with source-aware routing:
// Business tapes → {output}/{feature-slug}/{scenario-slug}.tape.
// VHS-only tapes → {output}/scenarios/{subdirectory}/{scenario-slug}.tape.
func writeScenarioTape(scenario vhsgen.ScenarioIR, outputDir, configSourcePath string) (string, error) {
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
		ConfigSourcePath: configSourcePath,
	}

	content, err := vhsgen.GenerateTape(scenario, config)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(tapeDir, 0o755); err != nil {
		return "", fmt.Errorf("creating output directory %q: %w", tapeDir, err)
	}

	outPath := filepath.Join(tapeDir, scenarioSlug+".tape")
	if err := os.WriteFile(outPath, []byte(content), 0o600); err != nil {
		return "", fmt.Errorf("writing tape file %q: %w", outPath, err)
	}

	return outPath, nil
}

// collectOutputASCIIFiles returns all .ascii files found recursively under dir.
//
// Expected: dir is a readable directory (or non-existent, returning empty slice).
// Returns: slice of absolute .ascii file paths; empty slice when dir is empty or missing; non-nil error on walk failure.
// Side effects: none.
func collectOutputASCIIFiles(dir string) ([]string, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return []string{}, nil
	}

	var files []string

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if !d.IsDir() && strings.HasSuffix(path, ".ascii") {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if files == nil {
		return []string{}, nil
	}

	return files, nil
}

// deriveScenarioName converts an ASCII output path to a scenario slug by stripping the outputDir prefix and .ascii suffix.
//
// Expected: outputDir is the root output directory; asciiPath is an absolute path under outputDir.
// Returns: a slugified scenario name derived from the relative path.
// Side effects: none.
func deriveScenarioName(outputDir, asciiPath string) string {
	rel := asciiPath

	if r, err := filepath.Rel(outputDir, asciiPath); err == nil {
		rel = r
	}

	rel = strings.TrimSuffix(rel, ".ascii")
	rel = strings.ReplaceAll(rel, string(filepath.Separator), "-")

	return vhsgen.Slugify(rel)
}

// deriveGIFPath converts an .ascii path to the corresponding .gif path by replacing the extension.
//
// Expected: asciiPath ends in ".ascii".
// Returns: the same path with ".ascii" replaced by ".gif".
// Side effects: none.
func deriveGIFPath(asciiPath string) string {
	return strings.TrimSuffix(asciiPath, ".ascii") + ".gif"
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// normaliseArgs rewrites bare positional keywords to flag form so that
// Go's flag package (which stops at the first non-flag arg) parses them.
// Example: ["all", "--output", "demo"] → ["--all", "--output", "demo"]
func normaliseArgs(args []string) []string {
	normalised := make([]string, len(args))
	copy(normalised, args)
	for i, arg := range normalised {
		if arg == "all" {
			normalised[i] = "--all"
		}
	}
	return normalised
}
