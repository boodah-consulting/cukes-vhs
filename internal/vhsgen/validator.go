package vhsgen

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ValidationStatus represents the outcome of a single scenario validation.
type ValidationStatus string

const (
	// ValidationPass indicates the current ASCII output matches the golden baseline.
	ValidationPass ValidationStatus = "PASS"
	// ValidationFail indicates the current ASCII output differs from the golden baseline.
	ValidationFail ValidationStatus = "FAIL"
	// ValidationNew indicates no golden baseline existed; the current output was saved as the new baseline.
	ValidationNew ValidationStatus = "NEW"
)

// ValidationResult holds the outcome of validating a single scenario.
type ValidationResult struct {
	Scenario   string
	ASCIIPath  string
	GoldenPath string
	Status     ValidationStatus
	Diff       string
}

// ansiEscape matches ANSI terminal escape sequences.
var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*[mGHFJK]`)

// minimalGIF is a valid 1×1 transparent GIF (43 bytes) used as a placeholder.
var minimalGIF = []byte{
	0x47, 0x49, 0x46, 0x38, 0x39, 0x61, // GIF89a
	0x01, 0x00, 0x01, 0x00, // 1x1 canvas
	0x80, 0x00, 0x00, // GCT flag, 2 colors
	0x00, 0x00, 0x00, // Color 0: black
	0xff, 0xff, 0xff, // Color 1: white
	0x21, 0xf9, 0x04, 0x01, 0x00, 0x00, 0x00, 0x00, // GCE
	0x2c, 0x00, 0x00, 0x00, 0x00, // Image descriptor
	0x01, 0x00, 0x01, 0x00, 0x00, // 1x1, no LCT
	0x02, 0x02, 0x44, 0x01, 0x00, // LZW min code + data
	0x3b, // Trailer
}

// stripANSI removes ANSI escape codes from s.
func stripANSI(s string) string {
	return ansiEscape.ReplaceAllString(s, "")
}

// normalise strips ANSI codes and trims trailing whitespace to produce a canonical form.
func normalise(s string) string {
	s = stripANSI(s)
	lines := strings.Split(s, "\n")

	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t\r")
	}

	return strings.Join(lines, "\n")
}

// ValidateScenario compares the ASCII file at currentASCIIPath against the golden baseline for scenario.
//
// Expected: goldenDir is a writable directory; currentASCIIPath is a readable .ascii file.
// Returns: ValidationNew (saving as baseline) when no prior baseline exists;
// ValidationFail with diff when content differs; ValidationPass otherwise.
// Side effects: may create a new baseline directory and placeholder GIF when no baseline exists.
func ValidateScenario(goldenDir, scenario, currentASCIIPath string) (ValidationResult, error) {
	currentASCIIPath = filepath.Clean(currentASCIIPath)

	result := ValidationResult{
		Scenario:  scenario,
		ASCIIPath: currentASCIIPath,
	}

	goldenASCII, exists, err := GetBaseline(goldenDir, scenario)
	if err != nil {
		return result, fmt.Errorf("looking up baseline for %q: %w", scenario, err)
	}

	if !exists {
		placeholderGIF, err := ensurePlaceholderGIF(goldenDir, scenario)
		if err != nil {
			return result, fmt.Errorf("creating placeholder GIF for %q: %w", scenario, err)
		}

		if err := SaveBaseline(goldenDir, scenario, currentASCIIPath, placeholderGIF); err != nil {
			return result, fmt.Errorf("saving new baseline for %q: %w", scenario, err)
		}

		savedASCII, _, saveErr := GetBaseline(goldenDir, scenario)
		if saveErr != nil {
			return result, fmt.Errorf("retrieving saved baseline for %q: %w", scenario, saveErr)
		}

		result.GoldenPath = savedASCII
		result.Status = ValidationNew

		return result, nil
	}

	result.GoldenPath = goldenASCII
	goldenASCII = filepath.Clean(goldenASCII)

	currentData, err := os.ReadFile(currentASCIIPath)
	if err != nil {
		return result, fmt.Errorf("reading current ASCII %q: %w", currentASCIIPath, err)
	}

	goldenData, err := os.ReadFile(goldenASCII)
	if err != nil {
		return result, fmt.Errorf("reading golden ASCII %q: %w", goldenASCII, err)
	}

	currentNorm := normalise(string(currentData))
	goldenNorm := normalise(string(goldenData))

	if currentNorm == goldenNorm {
		result.Status = ValidationPass
		return result, nil
	}

	result.Status = ValidationFail
	result.Diff = buildDiff(goldenNorm, currentNorm)

	return result, nil
}

// ValidateAll walks outputDir recursively for .ascii files and validates each against the golden baselines.
//
// Expected: outputDir is a readable directory; goldenDir is writable for new baselines.
// Returns: all ValidationResults collected regardless of PASS/FAIL; a non-nil error only for directory scanning failures.
// Side effects: may create new baseline entries for NEW scenarios.
func ValidateAll(goldenDir, outputDir string) ([]ValidationResult, error) {
	files, err := collectASCIIFiles(outputDir)
	if err != nil {
		return nil, fmt.Errorf("scanning output directory %q: %w", outputDir, err)
	}

	results := make([]ValidationResult, 0, len(files))

	for _, asciiPath := range files {
		scenario := deriveScenario(outputDir, asciiPath)

		result, validateErr := ValidateScenario(goldenDir, scenario, asciiPath)
		if validateErr != nil {
			result.Scenario = scenario
			result.ASCIIPath = asciiPath
			result.Status = ValidationFail
			result.Diff = validateErr.Error()
		}

		results = append(results, result)
	}

	return results, nil
}

// collectASCIIFiles returns all .ascii files found recursively under dir.
func collectASCIIFiles(dir string) ([]string, error) {
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

	return files, nil
}

// deriveScenario converts an ASCII file path to a slugified scenario name by stripping the outputDir prefix and .ascii suffix.
func deriveScenario(outputDir, asciiPath string) string {
	rel := asciiPath

	if r, err := filepath.Rel(outputDir, asciiPath); err == nil {
		rel = r
	}

	rel = strings.TrimSuffix(rel, ".ascii")
	rel = strings.ReplaceAll(rel, string(filepath.Separator), "-")

	return Slugify(rel)
}

// ensurePlaceholderGIF creates an empty placeholder GIF to satisfy SaveBaseline's GIF path requirement.
func ensurePlaceholderGIF(goldenDir, scenario string) (string, error) {
	placeholderDir := filepath.Join(goldenDir, ".placeholders")
	if err := os.MkdirAll(placeholderDir, 0o750); err != nil {
		return "", fmt.Errorf("creating placeholder dir: %w", err)
	}

	gifPath := filepath.Join(placeholderDir, Slugify(scenario)+".gif")
	if err := os.WriteFile(gifPath, minimalGIF, 0o600); err != nil {
		return "", fmt.Errorf("writing placeholder GIF: %w", err)
	}

	return gifPath, nil
}

// buildDiff generates a unified-style text diff between golden and current with 3-line context.
func buildDiff(golden, current string) string {
	goldenLines := strings.Split(golden, "\n")
	currentLines := strings.Split(current, "\n")

	var sb strings.Builder

	sb.WriteString("--- golden\n")
	sb.WriteString("+++ current\n")

	diffs := computeLineDiffs(goldenLines, currentLines)
	writeDiffHunks(&sb, diffs)

	return sb.String()
}

type diffLine struct {
	kind rune
	text string
}

// computeLineDiffs produces a flat sequence of diffLine entries using longest-common-subsequence.
func computeLineDiffs(golden, current []string) []diffLine {
	m := len(golden)
	n := len(current)

	lcs := buildLCS(golden, current, m, n)

	var diffs []diffLine

	gi, ci := 0, 0

	for _, common := range lcs {
		for gi < common.goldenIdx {
			diffs = append(diffs, diffLine{'-', golden[gi]})
			gi++
		}

		for ci < common.currentIdx {
			diffs = append(diffs, diffLine{'+', current[ci]})
			ci++
		}

		diffs = append(diffs, diffLine{' ', golden[gi]})
		gi++
		ci++
	}

	for gi < m {
		diffs = append(diffs, diffLine{'-', golden[gi]})
		gi++
	}

	for ci < n {
		diffs = append(diffs, diffLine{'+', current[ci]})
		ci++
	}

	return diffs
}

type commonLine struct {
	goldenIdx  int
	currentIdx int
}

// buildLCS returns the longest common subsequence as a list of index pairs.
func buildLCS(golden, current []string, m, n int) []commonLine {
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if golden[i-1] == current[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] >= dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	result := make([]commonLine, 0, dp[m][n])

	i, j := m, n
	for i > 0 && j > 0 {
		if golden[i-1] == current[j-1] {
			result = append(result, commonLine{i - 1, j - 1})
			i--
			j--
		} else if dp[i-1][j] >= dp[i][j-1] {
			i--
		} else {
			j--
		}
	}

	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}

	return result
}

const diffContextLines = 3

// openHunk returns the hunk start position, anchoring it at i minus context lines when a new hunk begins.
func openHunk(i, hunkStart int, inHunk bool) int {
	if inHunk {
		return hunkStart
	}

	start := i - diffContextLines
	if start < 0 {
		start = 0
	}

	return start
}

// writeDiffHunks writes the diff lines grouped into hunks with 3-line context on each side.
func writeDiffHunks(sb *strings.Builder, diffs []diffLine) {
	type hunkRange struct {
		start int
		end   int
	}

	var hunks []hunkRange

	inHunk := false
	hunkStart := 0
	consecutiveContext := 0

	for i, d := range diffs {
		if d.kind != ' ' {
			hunkStart = openHunk(i, hunkStart, inHunk)
			inHunk = true
			consecutiveContext = 0

			continue
		}

		if !inHunk {
			continue
		}

		consecutiveContext++

		if consecutiveContext > 2*diffContextLines {
			end := i - diffContextLines
			hunks = append(hunks, hunkRange{hunkStart, end})
			inHunk = false
			consecutiveContext = 0
		}
	}

	if inHunk {
		end := len(diffs)
		hunks = append(hunks, hunkRange{hunkStart, end})
	}

	for _, h := range hunks {
		sb.WriteString("@@ ... @@\n")
		for _, d := range diffs[h.start:h.end] {
			sb.WriteRune(d.kind)
			sb.WriteString(d.text)
			sb.WriteByte('\n')
		}
	}
}
