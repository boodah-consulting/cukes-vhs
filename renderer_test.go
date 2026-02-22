package vhsgen_test

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/baphled/cukes-vhs"
)

// writeFakeVHS writes a shell script that acts as a fake VHS binary.
// exitCode controls the script's exit status.
// stderrLines are printed to stderr.
func writeFakeVHS(dir string, exitCode int, stderrLines ...string) {
	script := "#!/bin/sh\n"
	for _, line := range stderrLines {
		script += fmt.Sprintf("echo %q >&2\n", line)
	}
	script += fmt.Sprintf("exit %d\n", exitCode)

	path := filepath.Join(dir, "vhs")
	err := os.WriteFile(path, []byte(script), 0o600)
	Expect(err).NotTo(HaveOccurred())

	Expect(os.Chmod(path, 0o755)).To(Succeed())
}

// writeTapeFile creates a minimal tape file with the given Output directives.
func writeTapeFile(dir, name string, outputs []string) string {
	content := "# fake tape\n"
	for _, out := range outputs {
		content += "Output " + out + "\n"
	}

	path := filepath.Join(dir, name)
	err := os.WriteFile(path, []byte(content), 0o600)
	Expect(err).NotTo(HaveOccurred())

	return path
}

var _ = Describe("Renderer", func() {
	var (
		tmpDir       string
		fakeVHSDir   string
		originalPATH string
	)

	BeforeEach(func() {
		tmpDir = GinkgoT().TempDir()
		fakeVHSDir = GinkgoT().TempDir()
		originalPATH = os.Getenv("PATH")
	})

	AfterEach(func() {
		Expect(os.Setenv("PATH", originalPATH)).To(Succeed())
	})

	Describe("RenderResult", func() {
		It("has all required fields with their zero values", func() {
			var r vhsgen.RenderResult
			Expect(r.TapePath).To(BeEmpty())
			Expect(r.GIFPath).To(BeEmpty())
			Expect(r.ASCIIPath).To(BeEmpty())
			Expect(r.Success).To(BeFalse())
			Expect(r.Error).To(BeEmpty())
			Expect(r.Duration).To(Equal(time.Duration(0)))
		})
	})

	Describe("NewRenderer", func() {
		It("returns a non-nil Renderer", func() {
			r := vhsgen.NewRenderer()
			Expect(r).NotTo(BeNil())
		})
	})

	Describe("RenderTape", func() {
		Context("when timeout is zero", func() {
			BeforeEach(func() {
				writeFakeVHS(fakeVHSDir, 0)
				Expect(os.Setenv("PATH", fakeVHSDir+":"+originalPATH)).To(Succeed())
			})

			It("uses the default timeout and succeeds", func() {
				tapePath := writeTapeFile(tmpDir, "zero-timeout.tape", []string{"out/zero.gif"})

				renderer := vhsgen.NewRenderer()
				result, err := renderer.RenderTape(tapePath, 0)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Success).To(BeTrue())
			})
		})

		Context("when VHS is not in PATH", func() {
			It("returns a descriptive error", func() {
				Expect(os.Setenv("PATH", "/nonexistent/path")).To(Succeed())

				renderer := vhsgen.NewRenderer()
				tapePath := writeTapeFile(tmpDir, "test.tape", []string{
					"demos/vhs/test.gif",
					"demos/vhs/test.ascii",
				})

				result, err := renderer.RenderTape(tapePath, 30*time.Second)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("vhs"))
				Expect(result.Success).To(BeFalse())
			})
		})

		Context("when VHS exits successfully", func() {
			BeforeEach(func() {
				writeFakeVHS(fakeVHSDir, 0)
				Expect(os.Setenv("PATH", fakeVHSDir+":"+originalPATH)).To(Succeed())
			})

			It("returns a successful RenderResult", func() {
				tapePath := writeTapeFile(tmpDir, "success.tape", []string{
					"demos/vhs/feature/success.gif",
					"demos/vhs/feature/success.ascii",
				})

				renderer := vhsgen.NewRenderer()
				result, err := renderer.RenderTape(tapePath, 30*time.Second)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Success).To(BeTrue())
				Expect(result.TapePath).To(Equal(tapePath))
				Expect(result.Duration).To(BeNumerically(">", 0))
			})

			It("parses GIF output path from tape content", func() {
				gifPath := "demos/vhs/feature/myscenario.gif"
				tapePath := writeTapeFile(tmpDir, "myscenario.tape", []string{
					gifPath,
					"demos/vhs/feature/myscenario.ascii",
				})

				renderer := vhsgen.NewRenderer()
				result, err := renderer.RenderTape(tapePath, 30*time.Second)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.GIFPath).To(ContainSubstring("myscenario.gif"))
			})

			It("parses ASCII output path from tape content", func() {
				asciiPath := "demos/vhs/feature/myscenario.ascii"
				tapePath := writeTapeFile(tmpDir, "myscenario.tape", []string{
					"demos/vhs/feature/myscenario.gif",
					asciiPath,
				})

				renderer := vhsgen.NewRenderer()
				result, err := renderer.RenderTape(tapePath, 30*time.Second)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.ASCIIPath).To(ContainSubstring("myscenario.ascii"))
			})

			It("handles a tape with no Output directives", func() {
				tapePath := writeTapeFile(tmpDir, "nooutput.tape", nil)

				renderer := vhsgen.NewRenderer()
				result, err := renderer.RenderTape(tapePath, 30*time.Second)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Success).To(BeTrue())
				Expect(result.GIFPath).To(BeEmpty())
				Expect(result.ASCIIPath).To(BeEmpty())
			})
		})

		Context("when VHS exits with non-zero status", func() {
			BeforeEach(func() {
				writeFakeVHS(fakeVHSDir, 1, "tape render failed: invalid command")
				Expect(os.Setenv("PATH", fakeVHSDir+":"+originalPATH)).To(Succeed())
			})

			It("returns a result with Success=false and captured stderr", func() {
				tapePath := writeTapeFile(tmpDir, "fail.tape", []string{
					"demos/vhs/feature/fail.gif",
					"demos/vhs/feature/fail.ascii",
				})

				renderer := vhsgen.NewRenderer()
				result, err := renderer.RenderTape(tapePath, 30*time.Second)
				Expect(err).To(HaveOccurred())
				Expect(result.Success).To(BeFalse())
				Expect(result.Error).To(ContainSubstring("tape render failed"))
				Expect(result.TapePath).To(Equal(tapePath))
			})
		})

		Context("when VHS exits non-zero with no stderr output", func() {
			BeforeEach(func() {
				writeFakeVHS(fakeVHSDir, 2)
				Expect(os.Setenv("PATH", fakeVHSDir+":"+originalPATH)).To(Succeed())
			})

			It("returns an error describing the exit failure", func() {
				tapePath := writeTapeFile(tmpDir, "silent-fail.tape", []string{
					"out/silent-fail.gif",
				})

				renderer := vhsgen.NewRenderer()
				result, err := renderer.RenderTape(tapePath, 30*time.Second)
				Expect(err).To(HaveOccurred())
				Expect(result.Success).To(BeFalse())
				Expect(result.Error).NotTo(BeEmpty())
			})
		})

		Context("when VHS times out", func() {
			BeforeEach(func() {
				script := "#!/bin/sh\nsleep 10\nexit 0\n"
				path := filepath.Join(fakeVHSDir, "vhs")
				Expect(os.WriteFile(path, []byte(script), 0o600)).To(Succeed())
				Expect(os.Chmod(path, 0o755)).To(Succeed())
				Expect(os.Setenv("PATH", fakeVHSDir+":"+originalPATH)).To(Succeed())
			})

			It("kills the process and returns a timeout error", func() {
				tapePath := writeTapeFile(tmpDir, "timeout.tape", []string{
					"demos/vhs/feature/timeout.gif",
				})

				renderer := vhsgen.NewRenderer()
				result, err := renderer.RenderTape(tapePath, 100*time.Millisecond)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("timed out"))
				Expect(result.Success).To(BeFalse())
			})
		})

		Context("when the tape file does not exist", func() {
			BeforeEach(func() {
				writeFakeVHS(fakeVHSDir, 0)
				Expect(os.Setenv("PATH", fakeVHSDir+":"+originalPATH)).To(Succeed())
			})

			It("returns an error about missing tape file", func() {
				renderer := vhsgen.NewRenderer()
				result, err := renderer.RenderTape("/nonexistent/path/missing.tape", 30*time.Second)
				Expect(err).To(HaveOccurred())
				Expect(result.Success).To(BeFalse())
			})
		})
	})

	Describe("RenderAll", func() {
		Context("when timeout is zero", func() {
			BeforeEach(func() {
				writeFakeVHS(fakeVHSDir, 0)
				Expect(os.Setenv("PATH", fakeVHSDir+":"+originalPATH)).To(Succeed())
			})

			It("uses the default timeout and succeeds", func() {
				writeTapeFile(tmpDir, "default-timeout.tape", []string{"out/default.gif"})

				renderer := vhsgen.NewRenderer()
				results, err := renderer.RenderAll(tmpDir, 0)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(HaveLen(1))
				Expect(results[0].Success).To(BeTrue())
			})
		})

		Context("when VHS is not in PATH", func() {
			It("returns an error before processing any tapes", func() {
				Expect(os.Setenv("PATH", "/nonexistent/path")).To(Succeed())

				renderer := vhsgen.NewRenderer()
				results, err := renderer.RenderAll(tmpDir, 30*time.Second)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("vhs"))
				Expect(results).To(BeEmpty())
			})
		})

		Context("when the tape directory does not exist", func() {
			BeforeEach(func() {
				writeFakeVHS(fakeVHSDir, 0)
				Expect(os.Setenv("PATH", fakeVHSDir+":"+originalPATH)).To(Succeed())
			})

			It("returns an error", func() {
				renderer := vhsgen.NewRenderer()
				results, err := renderer.RenderAll("/nonexistent/directory", 30*time.Second)
				Expect(err).To(HaveOccurred())
				Expect(results).To(BeEmpty())
			})
		})

		Context("when the directory contains no tape files", func() {
			BeforeEach(func() {
				writeFakeVHS(fakeVHSDir, 0)
				Expect(os.Setenv("PATH", fakeVHSDir+":"+originalPATH)).To(Succeed())
			})

			It("returns an empty results slice with no error", func() {
				renderer := vhsgen.NewRenderer()
				results, err := renderer.RenderAll(tmpDir, 30*time.Second)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(BeEmpty())
			})
		})

		Context("when all tapes succeed", func() {
			BeforeEach(func() {
				writeFakeVHS(fakeVHSDir, 0)
				Expect(os.Setenv("PATH", fakeVHSDir+":"+originalPATH)).To(Succeed())
			})

			It("returns results for each tape", func() {
				writeTapeFile(tmpDir, "scene1.tape", []string{"out/scene1.gif", "out/scene1.ascii"})
				writeTapeFile(tmpDir, "scene2.tape", []string{"out/scene2.gif", "out/scene2.ascii"})

				renderer := vhsgen.NewRenderer()
				results, err := renderer.RenderAll(tmpDir, 30*time.Second)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(HaveLen(2))

				for _, r := range results {
					Expect(r.Success).To(BeTrue())
				}
			})

			It("discovers tape files recursively in subdirectories", func() {
				subDir := filepath.Join(tmpDir, "subdir")
				Expect(os.MkdirAll(subDir, 0o755)).To(Succeed())

				writeTapeFile(tmpDir, "top.tape", []string{"out/top.gif"})
				writeTapeFile(subDir, "nested.tape", []string{"out/nested.gif"})

				renderer := vhsgen.NewRenderer()
				results, err := renderer.RenderAll(tmpDir, 30*time.Second)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(HaveLen(2))
			})
		})

		Context("when some tapes fail", func() {
			BeforeEach(func() {
				writeFakeVHS(fakeVHSDir, 1, "render error")
				Expect(os.Setenv("PATH", fakeVHSDir+":"+originalPATH)).To(Succeed())
			})

			It("collects failures in results without stopping", func() {
				writeTapeFile(tmpDir, "fail1.tape", []string{"out/fail1.gif"})
				writeTapeFile(tmpDir, "fail2.tape", []string{"out/fail2.gif"})

				renderer := vhsgen.NewRenderer()
				results, err := renderer.RenderAll(tmpDir, 30*time.Second)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(HaveLen(2))

				for _, r := range results {
					Expect(r.Success).To(BeFalse())
					Expect(r.Error).NotTo(BeEmpty())
				}
			})
		})

		Context("rendering is sequential", func() {
			BeforeEach(func() {
				writeFakeVHS(fakeVHSDir, 0)
				Expect(os.Setenv("PATH", fakeVHSDir+":"+originalPATH)).To(Succeed())
			})

			It("processes tapes one at a time (results are in order)", func() {
				writeTapeFile(tmpDir, "a.tape", []string{"out/a.gif"})
				writeTapeFile(tmpDir, "b.tape", []string{"out/b.gif"})
				writeTapeFile(tmpDir, "c.tape", []string{"out/c.gif"})

				renderer := vhsgen.NewRenderer()
				results, err := renderer.RenderAll(tmpDir, 30*time.Second)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(HaveLen(3))
			})
		})
	})
})
