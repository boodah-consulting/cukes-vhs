package cukesvhs

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// CommandTiming holds the result of benchmarking a single command.
type CommandTiming struct {
	Command  string
	Duration time.Duration
	Runs     int
}

// BenchmarkCommand runs a shell command the specified number of times and
// returns the average execution duration.
func BenchmarkCommand(command string, runs int) (CommandTiming, error) {
	if command == "" {
		return CommandTiming{}, errors.New("command must not be empty")
	}
	if runs < 1 {
		return CommandTiming{}, errors.New("runs must be at least 1")
	}

	var total time.Duration
	for range runs {
		start := time.Now()
		cmd := exec.Command("sh", "-c", command)
		if err := cmd.Run(); err != nil {
			return CommandTiming{}, fmt.Errorf("executing command %q: %w", command, err)
		}
		total += time.Since(start)
	}

	return CommandTiming{
		Command:  command,
		Duration: total / time.Duration(runs),
		Runs:     runs,
	}, nil
}

// ExtractCommands identifies executable CLI commands from When steps. It looks
// for Type commands whose argument text begins with "./" or contains a path
// separator, indicating a CLI invocation rather than plain text input.
func ExtractCommands(steps []StepIR) []string {
	var commands []string
	for _, step := range steps {
		if step.StepType != "When" {
			continue
		}
		for _, cmd := range step.Commands {
			if cmd.Type != Type {
				continue
			}
			if len(cmd.Args) < 2 {
				continue
			}
			text := cmd.Args[len(cmd.Args)-1]
			if isCLIInvocation(text) {
				commands = append(commands, text)
			}
		}
	}
	return commands
}

// BenchmarkScenario benchmarks all extractable CLI commands found in a
// scenario's demo steps.
func BenchmarkScenario(scenario ScenarioIR, runs int) (map[string]CommandTiming, error) {
	if runs < 1 {
		return nil, errors.New("runs must be at least 1")
	}

	commands := ExtractCommands(scenario.DemoSteps)
	results := make(map[string]CommandTiming, len(commands))

	for _, command := range commands {
		timing, err := BenchmarkCommand(command, runs)
		if err != nil {
			return nil, fmt.Errorf("benchmarking scenario %q command %q: %w", scenario.Name, command, err)
		}
		results[command] = timing
	}

	return results, nil
}

func isCLIInvocation(text string) bool {
	if strings.HasPrefix(text, "./") {
		return true
	}
	if strings.HasPrefix(text, "/") {
		return true
	}
	return false
}
