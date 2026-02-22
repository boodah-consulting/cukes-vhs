package vhsgen

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

const (
	baselineASCIIFile = "baseline.txt"
	baselineGIFFile   = "baseline.gif"
)

// BaselineInfo holds metadata about a stored golden baseline.
type BaselineInfo struct {
	Scenario  string
	ASCIIPath string
	GIFPath   string
	ModTime   time.Time
}

// SaveBaseline copies the ASCII and GIF files into the golden directory under a slug derived from scenario.
//
// Expected: asciiPath and gifPath must be readable files; goldenDir must be writable.
// Returns: error if directory creation or file copy fails; nil on success.
// Side effects: creates {goldenDir}/{scenario-slug}/ and writes baseline.txt + baseline.gif.
func SaveBaseline(goldenDir, scenario, asciiPath, gifPath string) error {
	dir := baselineDir(goldenDir, scenario)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating baseline dir %q: %w", dir, err)
	}

	if err := copyFile(asciiPath, filepath.Join(dir, baselineASCIIFile)); err != nil {
		return fmt.Errorf("copying ASCII baseline for %q: %w", scenario, err)
	}

	if err := copyFile(gifPath, filepath.Join(dir, baselineGIFFile)); err != nil {
		return fmt.Errorf("copying GIF baseline for %q: %w", scenario, err)
	}

	return nil
}

// GetBaseline returns the ASCII path and whether both baseline files exist for the scenario.
//
// Expected: goldenDir is the root golden directory; scenario is the scenario name.
// Returns: asciiPath and exists=true if both baseline files are present; exists=false and err=nil if absent.
// Side effects: none.
func GetBaseline(goldenDir, scenario string) (asciiPath string, exists bool, err error) {
	dir := baselineDir(goldenDir, scenario)

	if _, statErr := os.Stat(dir); os.IsNotExist(statErr) {
		return "", false, nil
	}

	ascii := filepath.Join(dir, baselineASCIIFile)
	gif := filepath.Join(dir, baselineGIFFile)

	if _, statErr := os.Stat(ascii); os.IsNotExist(statErr) {
		return "", false, nil
	}

	if _, statErr := os.Stat(gif); os.IsNotExist(statErr) {
		return "", false, nil
	}

	return ascii, true, nil
}

// UpdateBaseline overwrites the golden baseline for scenario with new files.
//
// Expected: asciiPath and gifPath must be readable; goldenDir must be writable.
// Returns: error if the update fails; nil on success.
// Side effects: replaces existing baseline.txt and baseline.gif for the scenario.
func UpdateBaseline(goldenDir, scenario, asciiPath, gifPath string) error {
	return SaveBaseline(goldenDir, scenario, asciiPath, gifPath)
}

// ListBaselines returns metadata for every stored baseline under goldenDir.
//
// Expected: goldenDir may be an empty or non-existent directory.
// Returns: slice of BaselineInfo (one per scenario with both baseline files); empty slice and nil error when none exist.
// Side effects: none.
func ListBaselines(goldenDir string) ([]BaselineInfo, error) {
	entries, err := os.ReadDir(goldenDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []BaselineInfo{}, nil
		}

		return nil, fmt.Errorf("reading golden dir %q: %w", goldenDir, err)
	}

	var results []BaselineInfo

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		ascii := filepath.Join(goldenDir, entry.Name(), baselineASCIIFile)
		gif := filepath.Join(goldenDir, entry.Name(), baselineGIFFile)

		asciiInfo, statErr := os.Stat(ascii)
		if os.IsNotExist(statErr) {
			continue
		}

		if statErr != nil {
			return nil, fmt.Errorf("stat baseline for %q: %w", entry.Name(), statErr)
		}

		if _, statErr := os.Stat(gif); statErr != nil {
			if os.IsNotExist(statErr) {
				continue
			}

			return nil, fmt.Errorf("stat GIF baseline for %q: %w", entry.Name(), statErr)
		}

		results = append(results, BaselineInfo{
			Scenario:  entry.Name(),
			ASCIIPath: ascii,
			GIFPath:   gif,
			ModTime:   asciiInfo.ModTime(),
		})
	}

	if results == nil {
		return []BaselineInfo{}, nil
	}

	return results, nil
}

func baselineDir(goldenDir, scenario string) string {
	return filepath.Join(goldenDir, Slugify(scenario))
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening source %q: %w", src, err)
	}
	defer func() { _ = in.Close() }()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("creating destination %q: %w", dst, err)
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copying %q to %q: %w", src, dst, err)
	}

	return out.Close()
}
