// Package cukesvhs provides the core library for converting Gherkin BDD scenarios
// into VHS tape files for automated terminal recordings.
//
// # Overview
//
// cukesvhs is designed to bridge Behaviour-Driven Development (BDD) specifications
// with terminal recording automation. It parses Gherkin feature files, translates
// step definitions into VHS commands, and generates tape files that can be rendered
// by [charmbracelet/vhs] into animated GIF demonstrations.
//
// The package supports two source types:
//   - Business scenarios from standard .feature files
//   - VHS-only scenarios specifically written for terminal demonstrations
//
// # Architecture
//
// The package is organised into several subsystems:
//
//   - Parser: Reads .feature files and extracts scenarios into intermediate representation
//   - Mapping: Translates Gherkin step text into VHS commands via pattern matching
//   - Generator: Renders tape file content from templates with scenario data
//   - Renderer: Invokes the VHS binary to produce GIF and ASCII outputs
//   - Validator: Compares rendered outputs against golden baselines
//   - Analyser: Reports on scenario translatability with warnings and errors
//
// # Pipeline Flow
//
// The typical conversion pipeline follows these stages:
//
//  1. Parse: [ParseFeatureDir] walks directories for .feature files
//  2. Translate: Each step is matched against patterns via [TranslateStep]
//  3. Analyse: [AnalyseScenarios] checks translatability and reports issues
//  4. Generate: [GenerateTape] or [WriteTape] produces tape file content
//  5. Render: [Renderer.RenderTape] invokes VHS to create outputs
//  6. Validate: [ValidateScenario] compares against golden baselines
//
// # Key Types
//
// The intermediate representation types form the core data model:
//
//   - [ScenarioIR]: Represents a complete Gherkin scenario with setup and demo steps
//   - [StepIR]: Represents a single step with its VHS command translation
//   - [VHSCommand]: A single VHS tape command (Type, Enter, Sleep, etc.)
//   - [AnalysisResult]: Outcome of analysing a scenario for translatability
//   - [ValidationResult]: Outcome of comparing rendered output against baseline
//
// # Example Usage
//
// Parse scenarios from a feature directory:
//
//	scenarios, err := cukesvhs.ParseFeatureDir("features/", cukesvhs.SourceBusiness)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Generate tape content for a scenario:
//
//	config := cukesvhs.GeneratorConfig{
//	    OutputDir:     "demos/",
//	    SleepDuration: "2s",
//	}
//	content, err := cukesvhs.GenerateTape(scenario, config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Render and validate against golden baseline:
//
//	renderer := cukesvhs.NewRenderer("")
//	result, err := renderer.RenderTape("demo.tape", 120*time.Second)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	validation, err := cukesvhs.ValidateScenario("golden/", "demo", result.ASCIIPath)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Validation: %s\n", validation.Status)
//
// # Step Pattern Matching
//
// The mapping system uses regular expressions to match Gherkin step text.
// Use [ListTranslatablePatterns] to discover available patterns:
//
//	patterns := cukesvhs.ListTranslatablePatterns()
//	for _, p := range patterns {
//	    fmt.Printf("%s: %s\n", p.Category, p.Example)
//	}
//
// Supported step categories include:
//   - navigation: Menu selection, arrow keys, Enter, Escape
//   - input: Text entry with configurable typing speed
//   - setup: Given steps that establish preconditions (no VHS output)
//
// # Golden Baseline Testing
//
// The validator supports golden baseline testing for regression detection.
// When no baseline exists, the current output becomes the new baseline.
// Use [SaveBaseline], [GetBaseline], and [UpdateBaseline] for manual
// baseline management.
//
// # Thread Safety
//
// All public functions are safe for concurrent use. Pattern matchers and
// compiled regular expressions are initialised once at package load time.
// The [Renderer] type is safe for concurrent use across multiple goroutines.
//
// [charmbracelet/vhs]: https://github.com/charmbracelet/vhs
package cukesvhs
