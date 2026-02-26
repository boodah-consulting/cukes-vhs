package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	"github.com/boodah-consulting/cukes-vhs/internal/cukesvhs"
)

// writeFileFs writes data to a file using the provided filesystem.
func writeFileFs(afs afero.Fs, path string, data []byte, perm os.FileMode) error {
	return afero.WriteFile(afs, path, data, perm)
}

type scenarioWithResult struct {
	scenario cukesvhs.ScenarioIR
	result   cukesvhs.AnalysisResult
}

func parseAllScenarios(featuresDir, scenariosDir string, errOut io.Writer) ([]cukesvhs.ScenarioIR, error) {
	if _, err := cliFs.Stat(featuresDir); err != nil {
		fmt.Fprintf(errOut, "Error parsing features dir %q: %v\n", featuresDir, err)
		return nil, err
	}

	businessScenarios, err := cukesvhs.ParseFeatureDir(featuresDir, cukesvhs.SourceBusiness)
	if err != nil {
		fmt.Fprintf(errOut, "Error parsing features dir %q: %v\n", featuresDir, err)
		return nil, err
	}

	vhsOnlyScenarios, err := cukesvhs.ParseFeatureDir(scenariosDir, cukesvhs.SourceVHSOnly)
	if err != nil {
		fmt.Fprintf(errOut, "Error parsing scenarios dir %q: %v\n", scenariosDir, err)
		return nil, err
	}

	allScenarios := make([]cukesvhs.ScenarioIR, 0, len(businessScenarios)+len(vhsOnlyScenarios))
	allScenarios = append(allScenarios, businessScenarios...)
	allScenarios = append(allScenarios, vhsOnlyScenarios...)
	return allScenarios, nil
}

// filterResults selects scenarios to generate based on flags.
func filterResults(
	results []cukesvhs.AnalysisResult,
	scenarios []cukesvhs.ScenarioIR,
	all bool,
	featureFilter, scenarioFilter string,
) []scenarioWithResult {
	var out []scenarioWithResult

	resultByID := make(map[string]cukesvhs.AnalysisResult, len(results))
	for i := range results {
		r := &results[i]
		resultByID[r.ScenarioID] = *r
	}
	for i := range scenarios {
		s := &scenarios[i]
		key := cukesvhs.BuildScenarioID(s.Source, s.Feature, s.Name)
		result, ok := resultByID[key]
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

// writeScenarioTape generates and writes a tape file to {output}/{feature-slug}/{scenario-slug}.tape.
// Both business and VHS-only scenarios use the same nested layout, matching the
// GIF/ASCII output paths set by GenerateTape in the generator package.
func writeScenarioTape(scenario cukesvhs.ScenarioIR, outputDir, configSourcePath string) (string, error) {
	featureSlug := cukesvhs.Slugify(scenario.Feature)
	scenarioSlug := cukesvhs.Slugify(scenario.Name)
	tapeDir := filepath.Join(outputDir, featureSlug)

	config := cukesvhs.GeneratorConfig{
		OutputDir:        outputDir,
		ConfigSourcePath: configSourcePath,
	}

	content, err := cukesvhs.GenerateTape(scenario, config)
	if err != nil {
		return "", err
	}

	if err := cliFs.MkdirAll(tapeDir, 0o750); err != nil {
		return "", fmt.Errorf("creating output directory %q: %w", tapeDir, err)
	}

	outPath := filepath.Join(tapeDir, scenarioSlug+".tape")
	if err := writeFileFs(cliFs, outPath, []byte(content), 0o600); err != nil {
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
	exists, err := afero.DirExists(cliFs, dir)
	if err != nil {
		return nil, err
	}
	if !exists {
		return []string{}, nil
	}

	var files []string

	err = afero.Walk(cliFs, dir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if !info.IsDir() && strings.HasSuffix(path, ".ascii") {
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

	return cukesvhs.Slugify(rel)
}

// deriveGIFPath converts an .ascii path to the corresponding .gif path by replacing the extension.
//
// Expected: asciiPath ends in ".ascii".
// Returns: the same path with ".ascii" replaced by ".gif".
// Side effects: none.
func deriveGIFPath(asciiPath string) string {
	return strings.TrimSuffix(asciiPath, ".ascii") + ".gif"
}

// findASCIIFileForScenario searches outputDir recursively for a .ascii file
// matching the given scenario slug. The generator creates nested paths
// ({outputDir}/{featureSlug}/{scenarioSlug}.ascii), so a recursive search is
// necessary when only the scenario name is known.
//
// Expected: outputDir is a readable directory; scenarioSlug is a slugified scenario name.
// Returns: the path to the matching .ascii file; error if not found or ambiguous.
// Side effects: none.
func findASCIIFileForScenario(outputDir, scenarioSlug string) (string, error) {
	targetName := scenarioSlug + ".ascii"

	var matches []string

	err := afero.Walk(cliFs, outputDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if !info.IsDir() && filepath.Base(path) == targetName {
			matches = append(matches, path)
		}

		return nil
	})
	if err != nil {
		return "", fmt.Errorf("scanning output dir %q: %w", outputDir, err)
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no .ascii file found for scenario %q in %q", scenarioSlug, outputDir)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("ambiguous: scenario %q found in multiple locations: %v", scenarioSlug, matches)
	}
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
// Example: ["all", "--output", "demo"] → ["--all", "--output", "demo"].
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

type generateStats struct {
	total        int
	fromBusiness int
	fromVHSOnly  int
	warnings     int
}

type generateConfig struct {
	outputDir    string
	configSource string
	verbose      bool
	out          io.Writer
	errOut       io.Writer
}

// generateTapes processes filtered scenarios and writes tape files.
func generateTapes(filtered []scenarioWithResult, cfg generateConfig) generateStats {
	var stats generateStats

	for i := range filtered {
		entry := &filtered[i]
		scenario := entry.scenario
		result := entry.result

		if !result.Translatable {
			if cfg.verbose {
				fmt.Fprintf(cfg.out, "Skipping %q (not translatable)\n", scenario.Name)
			}
			stats.warnings++
			continue
		}

		outPath, tapeErr := writeScenarioTape(scenario, cfg.outputDir, cfg.configSource)
		if tapeErr != nil {
			fmt.Fprintf(cfg.errOut, "Error generating tape for %q: %v\n", scenario.Name, tapeErr)
			continue
		}

		fmt.Fprintf(cfg.out, "Written: %s\n", outPath)

		switch scenario.Source {
		case cukesvhs.SourceBusiness:
			stats.fromBusiness++
		case cukesvhs.SourceVHSOnly:
			stats.fromVHSOnly++
		}
	}

	stats.total = stats.fromBusiness + stats.fromVHSOnly
	return stats
}
