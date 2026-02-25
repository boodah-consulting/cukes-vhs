package cukesvhs

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const defaultRenderTimeout = 120 * time.Second

// RenderResult holds the outcome of rendering a single VHS tape file.
type RenderResult struct {
	TapePath  string
	GIFPath   string
	ASCIIPath string
	Success   bool
	Error     string
	Duration  time.Duration
}

// Renderer invokes the VHS CLI to render tape files.
type Renderer struct {
	binaryPath string
}

// NewRenderer returns a new Renderer configured with the given VHS binary path.
//
// Expected: binaryPath is the path to the vhs binary; empty string defaults to "vhs".
// Returns: a non-nil *Renderer ready for use.
// Side effects: none.
func NewRenderer(binaryPath string) *Renderer {
	if binaryPath == "" {
		binaryPath = "vhs"
	}

	return &Renderer{binaryPath: binaryPath}
}

// RenderTape runs the vhs binary against tapePath and returns the result.
//
// Expected: tapePath points to a readable .tape file; a timeout of 0 uses the default of 120s.
// Returns: RenderResult with Success=true on success; a non-nil error if vhs is missing, the tape is unreadable, or rendering fails.
// Side effects: spawns a child vhs process and kills it on timeout.
func (r *Renderer) RenderTape(tapePath string, timeout time.Duration) (RenderResult, error) {
	if timeout <= 0 {
		timeout = defaultRenderTimeout
	}

	result := RenderResult{TapePath: tapePath}

	if _, err := exec.LookPath(r.binaryPath); err != nil {
		return result, fmt.Errorf("%s binary not found in PATH: %w", r.binaryPath, err)
	}

	gifPath, asciiPath, err := parseOutputPaths(tapePath)
	if err != nil {
		return result, fmt.Errorf("reading tape file %q: %w", tapePath, err)
	}

	result.GIFPath = gifPath
	result.ASCIIPath = asciiPath

	start := time.Now()
	renderErr := runVHS(r.binaryPath, tapePath, timeout)
	result.Duration = time.Since(start)

	if renderErr != nil {
		result.Error = renderErr.Error()
		return result, renderErr
	}

	result.Success = true

	return result, nil
}

// RenderAll renders every .tape file found recursively under tapeDir.
//
// Expected: tapeDir is a readable directory; timeout applies per-tape invocation.
// Returns: slice of RenderResult (one per tape); a non-nil error only for directory-level failures.
// Side effects: spawns child vhs processes sequentially, one per tape file.
func (r *Renderer) RenderAll(tapeDir string, timeout time.Duration) ([]RenderResult, error) {
	if timeout <= 0 {
		timeout = defaultRenderTimeout
	}

	if _, err := exec.LookPath(r.binaryPath); err != nil {
		return nil, fmt.Errorf("%s binary not found in PATH: %w", r.binaryPath, err)
	}

	tapes, err := collectTapeFiles(tapeDir)
	if err != nil {
		return nil, fmt.Errorf("scanning tape directory %q: %w", tapeDir, err)
	}

	results := make([]RenderResult, 0, len(tapes))

	for _, tapePath := range tapes {
		result, renderErr := r.RenderTape(tapePath, timeout)
		if renderErr != nil {
			result.Success = false
			result.Error = renderErr.Error()
		}

		results = append(results, result)
	}

	return results, nil
}

// collectTapeFiles returns all .tape files found recursively under dir.
func collectTapeFiles(dir string) ([]string, error) {
	var tapes []string

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if !d.IsDir() && strings.HasSuffix(path, ".tape") {
			tapes = append(tapes, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return tapes, nil
}

// parseOutputPaths reads a tape file and extracts the first GIF and ASCII output paths; relative paths are resolved to absolute form.
func parseOutputPaths(tapePath string) (gifPath, asciiPath string, err error) {
	tapePath = filepath.Clean(tapePath)

	data, err := os.ReadFile(tapePath)
	if err != nil {
		return "", "", err
	}

	tapeDirAbs, err := filepath.Abs(filepath.Dir(tapePath))
	if err != nil {
		return "", "", fmt.Errorf("resolving tape directory: %w", err)
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "Output ") {
			continue
		}

		outPath := strings.TrimSpace(strings.TrimPrefix(line, "Output "))
		rawOutPath := outPath
		if !filepath.IsAbs(outPath) {
			outPath = filepath.Join(tapeDirAbs, outPath)
		}

		resolved := filepath.Clean(outPath)
		cleanTapeDir := filepath.Clean(tapeDirAbs) + string(os.PathSeparator)
		if !strings.HasPrefix(resolved, cleanTapeDir) && resolved != filepath.Clean(tapeDirAbs) {
			return "", "", fmt.Errorf("output path %q escapes tape directory", rawOutPath)
		}

		switch {
		case strings.HasSuffix(outPath, ".gif") && gifPath == "":
			gifPath = outPath
		case strings.HasSuffix(outPath, ".ascii") && asciiPath == "":
			asciiPath = outPath
		}
	}

	return gifPath, asciiPath, scanner.Err()
}

// runVHS executes the vhs binary with the given tape file path and timeout, returning a descriptive error on failure.
func runVHS(binaryPath, tapePath string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, binaryPath, tapePath)
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		return nil
	}

	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("vhs rendering timed out after %s: %w", timeout, ErrTimeout)
	}

	stderrStr := strings.TrimSpace(stderr.String())
	if stderrStr != "" {
		return fmt.Errorf("vhs exited with error: %s", stderrStr)
	}

	return fmt.Errorf("vhs exited with error: %w", err)
}

// ErrTimeout is returned when a VHS render process exceeds its timeout.
var ErrTimeout = errors.New("timed out")
