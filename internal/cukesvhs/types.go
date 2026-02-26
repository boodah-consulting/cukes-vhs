package cukesvhs

// SourceType indicates where a scenario originates.
type SourceType string

const (
	// SourceBusiness indicates a scenario from business requirements.
	SourceBusiness SourceType = "business"
	// SourceVHSOnly indicates a scenario specific to VHS tape testing.
	SourceVHSOnly SourceType = "vhs-only"
)

// VHSCommandType represents a VHS tape command type.
type VHSCommandType string

const (
	// Type represents the Type command.
	Type VHSCommandType = "Type"
	// Down represents the Down arrow key command.
	Down VHSCommandType = "Down"
	// Up represents the Up arrow key command.
	Up VHSCommandType = "Up"
	// Enter represents the Enter key command.
	Enter VHSCommandType = "Enter"
	// Escape represents the Escape key command.
	Escape VHSCommandType = "Escape"
	// Tab represents the Tab key command.
	Tab VHSCommandType = "Tab"
	// Sleep represents a sleep/delay command.
	Sleep VHSCommandType = "Sleep"
	// Hide represents a hide command.
	Hide VHSCommandType = "Hide"
	// Show represents a show command.
	Show VHSCommandType = "Show"
	// Screenshot represents a screenshot command.
	Screenshot VHSCommandType = "Screenshot"
	// Source represents a source command.
	Source VHSCommandType = "Source"
	// Output represents an output command.
	Output VHSCommandType = "Output"
	// CtrlC represents the Ctrl+C keyboard command.
	CtrlC VHSCommandType = "Ctrl+C"
)

// VHSCommand represents a single VHS tape command.
type VHSCommand struct {
	Type VHSCommandType
	Args []string
}

// StepIR represents an intermediate representation of a Gherkin step.
type StepIR struct {
	Text string
	// StepType is one of "Given", "When", or "Then".
	StepType             string
	Commands             []VHSCommand
	Translatable         bool
	UntranslatableReason string
}

// ScenarioIR represents an intermediate representation of a Gherkin scenario.
type ScenarioIR struct {
	Name         string
	Feature      string
	Tags         []string
	Source       SourceType
	SetupSteps   []StepIR
	DemoSteps    []StepIR
	Translatable bool
}

// GeneratorConfig holds configuration for the VHS generator.
type GeneratorConfig struct {
	OutputDir        string
	ConfigSourcePath string
	SleepDuration    string
}

// AnalysisResult represents the result of analysing a scenario.
type AnalysisResult struct {
	// ScenarioID is a composite key combining source, feature, and scenario name
	// to uniquely identify a scenario across features. Format: source/feature/name (slugified).
	ScenarioID          string
	ScenarioName        string
	Feature             string
	Translatable        bool
	UntranslatableSteps []StepIR
	Source              SourceType
	Warnings            []string
	Errors              []string
}

// BuildScenarioID constructs a unique composite key for a scenario from its
// source type, feature name, and scenario name. Each component is slugified
// for path safety. The format is "source/feature/name".
func BuildScenarioID(source SourceType, feature, name string) string {
	return Slugify(string(source)) + "/" + Slugify(feature) + "/" + Slugify(name)
}

// ParamConstraint defines constraints on a step parameter.
type ParamConstraint struct {
	Type   string
	Values []string
}

// StepPattern represents a pattern for matching and translating steps.
type StepPattern struct {
	Pattern  string
	Type     string
	Category string
	Params   map[string]ParamConstraint
	Example  string
}
