package cukesvhs

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	gherkin "github.com/cucumber/gherkin/go/v26"
	messages "github.com/cucumber/messages/go/v21"
	"github.com/spf13/afero"
)

// ParseFeatureDir walks a directory for .feature files and parses all scenarios into ScenarioIR.
//
// Expected: dir is a path to a directory containing .feature files; source indicates origin.
// Returns: Slice of ScenarioIR for all scenarios found; empty slice and nil error for empty directories.
// Side effects: Reads files from disk.
func ParseFeatureDir(dir string, source SourceType) ([]ScenarioIR, error) {
	return ParseFeatureDirFs(DefaultFs(), dir, source)
}

// ParseFeatureDirFs walks a directory for .feature files using the provided filesystem.
//
// Expected: fs is the filesystem to use; dir is a path to a directory; source indicates origin.
// Returns: Slice of ScenarioIR for all scenarios found; empty slice and nil error for empty directories.
// Side effects: Reads files from the provided filesystem.
func ParseFeatureDirFs(afs afero.Fs, dir string, source SourceType) ([]ScenarioIR, error) {
	exists, err := afero.DirExists(afs, dir)
	if err != nil {
		return nil, fmt.Errorf("checking directory %s: %w", dir, err)
	}
	if !exists {
		return []ScenarioIR{}, nil
	}

	var results []ScenarioIR

	err = afero.Walk(afs, dir, func(path string, info fs.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if info.IsDir() || !strings.HasSuffix(info.Name(), ".feature") {
			return nil
		}

		scenarios, parseErr := parseFeatureFileFs(afs, path, source)
		if parseErr != nil {
			return fmt.Errorf("parsing %s: %w", path, parseErr)
		}

		results = append(results, scenarios...)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking directory %s: %w", dir, err)
	}

	if results == nil {
		return []ScenarioIR{}, nil
	}

	return results, nil
}

func parseFeatureFileFs(afs afero.Fs, path string, source SourceType) ([]ScenarioIR, error) {
	path = filepath.Clean(path)

	f, err := afs.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	doc, err := gherkin.ParseGherkinDocument(f, (&messages.Incrementing{}).NewId)
	if err != nil {
		return nil, fmt.Errorf("parsing gherkin: %w", err)
	}

	if doc.Feature == nil {
		return nil, nil
	}

	return extractScenarios(doc.Feature, source), nil
}

func extractScenarios(feature *messages.Feature, source SourceType) []ScenarioIR {
	var backgroundSteps []*messages.Step

	for _, child := range feature.Children {
		if child.Background != nil {
			backgroundSteps = child.Background.Steps
			break
		}
	}

	var featureTags []string
	for _, tag := range feature.Tags {
		featureTags = append(featureTags, tag.Name)
	}

	var results []ScenarioIR

	for _, child := range feature.Children {
		if child.Scenario == nil {
			continue
		}

		scenario := child.Scenario

		var tags []string
		tags = append(tags, featureTags...)
		for _, tag := range scenario.Tags {
			tags = append(tags, tag.Name)
		}

		if len(scenario.Examples) > 0 {
			hasRows := false
			for _, ex := range scenario.Examples {
				if ex.TableHeader != nil {
					for _, row := range ex.TableBody {
						ir := buildOutlineIR(scenario, feature.Name, tags, source, backgroundSteps, ex.TableHeader, row)
						results = append(results, ir)
						hasRows = true
					}
				}
			}
			if !hasRows {
				ir := buildOutlineIR(scenario, feature.Name, tags, source, backgroundSteps, nil, nil)
				results = append(results, ir)
			}
		} else {
			results = append(results, buildScenarioIR(scenario, feature.Name, tags, source, backgroundSteps))
		}
	}

	return results
}

func buildScenarioIR(
	scenario *messages.Scenario,
	featureName string,
	tags []string,
	source SourceType,
	backgroundSteps []*messages.Step,
) ScenarioIR {
	ir := ScenarioIR{
		Name:    scenario.Name,
		Feature: featureName,
		Tags:    tags,
		Source:  source,
	}

	lastPrimaryType := "Given"

	for _, step := range backgroundSteps {
		stepType := resolveStepType(step.KeywordType, &lastPrimaryType)
		ir.SetupSteps = append(ir.SetupSteps, translateToStepIR(step.Text, stepType))
	}

	lastPrimaryType = "Given"

	for _, step := range scenario.Steps {
		stepType := resolveStepType(step.KeywordType, &lastPrimaryType)
		stepIR := translateToStepIR(step.Text, stepType)
		classifyStep(&ir, stepIR, stepType)
	}

	ir.Translatable = isScenarioTranslatable(ir)

	return ir
}

func buildOutlineIR(
	scenario *messages.Scenario,
	featureName string,
	tags []string,
	source SourceType,
	backgroundSteps []*messages.Step,
	header *messages.TableRow,
	row *messages.TableRow,
) ScenarioIR {
	ir := ScenarioIR{
		Name:    scenario.Name,
		Feature: featureName,
		Tags:    tags,
		Source:  source,
	}

	lastPrimaryType := "Given"

	for _, step := range backgroundSteps {
		stepType := resolveStepType(step.KeywordType, &lastPrimaryType)
		text := substituteExampleValues(step.Text, header, row)
		ir.SetupSteps = append(ir.SetupSteps, translateToStepIR(text, stepType))
	}

	lastPrimaryType = "Given"

	for _, step := range scenario.Steps {
		stepType := resolveStepType(step.KeywordType, &lastPrimaryType)
		text := substituteExampleValues(step.Text, header, row)
		stepIR := translateToStepIR(text, stepType)
		classifyStep(&ir, stepIR, stepType)
	}

	ir.Translatable = isScenarioTranslatable(ir)

	return ir
}

func resolveStepType(kwType messages.StepKeywordType, lastPrimary *string) string {
	switch kwType {
	case messages.StepKeywordType_CONTEXT:
		*lastPrimary = "Given"
		return "Given"
	case messages.StepKeywordType_ACTION:
		*lastPrimary = "When"
		return "When"
	case messages.StepKeywordType_OUTCOME:
		*lastPrimary = "Then"
		return "Then"
	default:
		return *lastPrimary
	}
}

func translateToStepIR(text, stepType string) StepIR {
	cmds, translatable, reason := TranslateStep(text, stepType)
	return StepIR{
		Text:                 text,
		StepType:             stepType,
		Commands:             cmds,
		Translatable:         translatable,
		UntranslatableReason: reason,
	}
}

func classifyStep(ir *ScenarioIR, step StepIR, stepType string) {
	if stepType == "Given" {
		ir.SetupSteps = append(ir.SetupSteps, step)
	} else {
		ir.DemoSteps = append(ir.DemoSteps, step)
	}
}

func substituteExampleValues(text string, header, row *messages.TableRow) string {
	if header == nil || row == nil {
		return text
	}

	for i, cell := range header.Cells {
		if i < len(row.Cells) {
			text = strings.ReplaceAll(text, "<"+cell.Value+">", row.Cells[i].Value)
		}
	}

	return text
}

func isScenarioTranslatable(ir ScenarioIR) bool {
	for _, step := range ir.SetupSteps {
		if !step.Translatable {
			return false
		}
	}

	for _, step := range ir.DemoSteps {
		if !step.Translatable {
			return false
		}
	}

	return true
}
