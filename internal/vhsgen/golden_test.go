package vhsgen_test

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/boodah-consulting/cukesvhs/internal/vhsgen"
)

var _ = Describe("Golden baseline management", func() {
	var (
		goldenDir string
		asciiSrc  string
		gifSrc    string
	)

	BeforeEach(func() {
		goldenDir = GinkgoT().TempDir()

		srcDir := GinkgoT().TempDir()
		asciiSrc = filepath.Join(srcDir, "output.txt")
		gifSrc = filepath.Join(srcDir, "output.gif")

		Expect(os.WriteFile(asciiSrc, []byte("ascii content"), 0o600)).To(Succeed())
		Expect(os.WriteFile(gifSrc, []byte("gif content"), 0o600)).To(Succeed())
	})

	Describe("SaveBaseline", func() {
		Context("with valid source files", func() {
			It("creates the scenario directory under goldenDir", func() {
				Expect(vhsgen.SaveBaseline(goldenDir, "My Scenario", asciiSrc, gifSrc)).To(Succeed())

				expectedDir := filepath.Join(goldenDir, "my-scenario")
				_, err := os.Stat(expectedDir)
				Expect(err).NotTo(HaveOccurred())
			})

			It("writes baseline.txt with the ASCII content", func() {
				Expect(vhsgen.SaveBaseline(goldenDir, "My Scenario", asciiSrc, gifSrc)).To(Succeed())

				data, err := os.ReadFile(filepath.Join(goldenDir, "my-scenario", "baseline.txt"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(data)).To(Equal("ascii content"))
			})

			It("writes baseline.gif with the GIF content", func() {
				Expect(vhsgen.SaveBaseline(goldenDir, "My Scenario", asciiSrc, gifSrc)).To(Succeed())

				data, err := os.ReadFile(filepath.Join(goldenDir, "my-scenario", "baseline.gif"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(data)).To(Equal("gif content"))
			})

			It("slugifies the scenario name to form the directory path", func() {
				Expect(vhsgen.SaveBaseline(goldenDir, "Hello World! 123", asciiSrc, gifSrc)).To(Succeed())

				expectedDir := filepath.Join(goldenDir, "hello-world-123")
				_, err := os.Stat(expectedDir)
				Expect(err).NotTo(HaveOccurred())
			})

			It("handles scenario names with underscores", func() {
				Expect(vhsgen.SaveBaseline(goldenDir, "capture_event", asciiSrc, gifSrc)).To(Succeed())

				expectedDir := filepath.Join(goldenDir, "capture-event")
				_, err := os.Stat(expectedDir)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the ASCII source does not exist", func() {
			It("returns an error", func() {
				err := vhsgen.SaveBaseline(goldenDir, "Missing", "/nonexistent/output.txt", gifSrc)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("ASCII baseline"))
			})
		})

		Context("when the GIF source does not exist", func() {
			It("returns an error", func() {
				err := vhsgen.SaveBaseline(goldenDir, "Missing", asciiSrc, "/nonexistent/output.gif")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("GIF baseline"))
			})
		})

		Context("when the golden dir cannot be created", func() {
			It("returns an error mentioning the baseline dir", func() {
				err := vhsgen.SaveBaseline("/invalid/nonexistent/cannot/create", "Scenario", asciiSrc, gifSrc)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("creating baseline dir"))
			})
		})

		Context("when the baseline dir is read-only", func() {
			It("returns an error when the destination file cannot be created", func() {
				readOnlyParent := GinkgoT().TempDir()
				targetDir := filepath.Join(readOnlyParent, "locked-scenario")
				Expect(os.MkdirAll(targetDir, 0o755)).To(Succeed())
				Expect(os.Chmod(targetDir, 0o000)).To(Succeed())
				defer os.Chmod(targetDir, 0o755) //nolint:errcheck

				err := vhsgen.SaveBaseline(readOnlyParent, "locked scenario", asciiSrc, gifSrc)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("GetBaseline", func() {
		Context("when a baseline exists for the scenario", func() {
			BeforeEach(func() {
				Expect(vhsgen.SaveBaseline(goldenDir, "My Scenario", asciiSrc, gifSrc)).To(Succeed())
			})

			It("returns the correct ASCII path", func() {
				ascii, exists, err := vhsgen.GetBaseline(goldenDir, "My Scenario")
				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeTrue())
				Expect(ascii).To(Equal(filepath.Join(goldenDir, "my-scenario", "baseline.txt")))
			})

			It("returns the correct GIF path via ListBaselines", func() {
				_, exists, err := vhsgen.GetBaseline(goldenDir, "My Scenario")
				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeTrue())
				baselines, listErr := vhsgen.ListBaselines(goldenDir)
				Expect(listErr).NotTo(HaveOccurred())
				Expect(baselines).To(HaveLen(1))
				Expect(baselines[0].GIFPath).To(Equal(filepath.Join(goldenDir, "my-scenario", "baseline.gif")))
			})

			It("returns exists=true", func() {
				_, exists, err := vhsgen.GetBaseline(goldenDir, "My Scenario")
				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeTrue())
			})
		})

		Context("when no baseline exists for the scenario", func() {
			It("returns exists=false and no error", func() {
				_, exists, err := vhsgen.GetBaseline(goldenDir, "Nonexistent Scenario")
				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeFalse())
			})

			It("returns empty ASCII path", func() {
				ascii, _, _ := vhsgen.GetBaseline(goldenDir, "Nonexistent Scenario")
				Expect(ascii).To(BeEmpty())
			})
		})

		Context("when the scenario directory exists but baseline.txt is missing", func() {
			BeforeEach(func() {
				dir := filepath.Join(goldenDir, "partial-scenario")
				Expect(os.MkdirAll(dir, 0o755)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(dir, "baseline.gif"), []byte("gif"), 0o600)).To(Succeed())
			})

			It("returns exists=false and no error", func() {
				_, exists, err := vhsgen.GetBaseline(goldenDir, "partial scenario")
				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeFalse())
			})
		})

		Context("when the scenario directory exists but baseline.gif is missing", func() {
			BeforeEach(func() {
				dir := filepath.Join(goldenDir, "partial-gif")
				Expect(os.MkdirAll(dir, 0o755)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(dir, "baseline.txt"), []byte("ascii"), 0o600)).To(Succeed())
			})

			It("returns exists=false and no error", func() {
				_, exists, err := vhsgen.GetBaseline(goldenDir, "partial gif")
				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeFalse())
			})
		})
	})

	Describe("UpdateBaseline", func() {
		Context("when a baseline already exists", func() {
			var newASCII, newGIF string

			BeforeEach(func() {
				Expect(vhsgen.SaveBaseline(goldenDir, "Update Me", asciiSrc, gifSrc)).To(Succeed())

				newDir := GinkgoT().TempDir()
				newASCII = filepath.Join(newDir, "new.txt")
				newGIF = filepath.Join(newDir, "new.gif")
				Expect(os.WriteFile(newASCII, []byte("updated ascii"), 0o600)).To(Succeed())
				Expect(os.WriteFile(newGIF, []byte("updated gif"), 0o600)).To(Succeed())
			})

			It("overwrites the ASCII baseline with new content", func() {
				Expect(vhsgen.UpdateBaseline(goldenDir, "Update Me", newASCII, newGIF)).To(Succeed())

				data, err := os.ReadFile(filepath.Join(goldenDir, "update-me", "baseline.txt"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(data)).To(Equal("updated ascii"))
			})

			It("overwrites the GIF baseline with new content", func() {
				Expect(vhsgen.UpdateBaseline(goldenDir, "Update Me", newASCII, newGIF)).To(Succeed())

				data, err := os.ReadFile(filepath.Join(goldenDir, "update-me", "baseline.gif"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(data)).To(Equal("updated gif"))
			})
		})

		Context("when no baseline exists yet", func() {
			It("creates the baseline from scratch", func() {
				Expect(vhsgen.UpdateBaseline(goldenDir, "Brand New", asciiSrc, gifSrc)).To(Succeed())

				_, exists, err := vhsgen.GetBaseline(goldenDir, "Brand New")
				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeTrue())
			})
		})

		Context("when the source file is missing", func() {
			It("returns an error", func() {
				err := vhsgen.UpdateBaseline(goldenDir, "Fail", "/nonexistent/file.txt", gifSrc)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("ListBaselines", func() {
		Context("when the golden dir is empty", func() {
			It("returns an empty slice and no error", func() {
				results, err := vhsgen.ListBaselines(goldenDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(BeEmpty())
			})
		})

		Context("when the golden dir does not exist", func() {
			It("returns an empty slice and no error", func() {
				results, err := vhsgen.ListBaselines(filepath.Join(goldenDir, "nonexistent"))
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(BeEmpty())
			})
		})

		Context("when the golden dir is not readable", func() {
			It("returns an error mentioning the golden dir", func() {
				lockedDir := GinkgoT().TempDir()
				Expect(os.Chmod(lockedDir, 0o000)).To(Succeed())
				defer os.Chmod(lockedDir, 0o755) //nolint:errcheck

				_, err := vhsgen.ListBaselines(lockedDir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("reading golden dir"))
			})
		})

		Context("when baselines exist", func() {
			BeforeEach(func() {
				Expect(vhsgen.SaveBaseline(goldenDir, "Scenario Alpha", asciiSrc, gifSrc)).To(Succeed())
				Expect(vhsgen.SaveBaseline(goldenDir, "Scenario Beta", asciiSrc, gifSrc)).To(Succeed())
			})

			It("returns one entry per saved baseline", func() {
				results, err := vhsgen.ListBaselines(goldenDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(HaveLen(2))
			})

			It("populates the Scenario field with the slug", func() {
				results, err := vhsgen.ListBaselines(goldenDir)
				Expect(err).NotTo(HaveOccurred())

				scenarios := make([]string, 0, len(results))
				for _, r := range results {
					scenarios = append(scenarios, r.Scenario)
				}

				Expect(scenarios).To(ConsistOf("scenario-alpha", "scenario-beta"))
			})

			It("populates ASCIIPath pointing to baseline.txt", func() {
				results, err := vhsgen.ListBaselines(goldenDir)
				Expect(err).NotTo(HaveOccurred())

				for _, r := range results {
					Expect(r.ASCIIPath).To(HaveSuffix("baseline.txt"))
				}
			})

			It("populates GIFPath pointing to baseline.gif", func() {
				results, err := vhsgen.ListBaselines(goldenDir)
				Expect(err).NotTo(HaveOccurred())

				for _, r := range results {
					Expect(r.GIFPath).To(HaveSuffix("baseline.gif"))
				}
			})

			It("populates a non-zero ModTime", func() {
				results, err := vhsgen.ListBaselines(goldenDir)
				Expect(err).NotTo(HaveOccurred())

				for _, r := range results {
					Expect(r.ModTime).NotTo(Equal(time.Time{}))
				}
			})
		})

		Context("when a dir exists but has no baseline files", func() {
			BeforeEach(func() {
				emptyDir := filepath.Join(goldenDir, "empty-scenario")
				Expect(os.MkdirAll(emptyDir, 0o755)).To(Succeed())
			})

			It("omits the incomplete entry", func() {
				results, err := vhsgen.ListBaselines(goldenDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(BeEmpty())
			})
		})

		Context("when a dir has only baseline.txt but not baseline.gif", func() {
			BeforeEach(func() {
				dir := filepath.Join(goldenDir, "partial-only")
				Expect(os.MkdirAll(dir, 0o755)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(dir, "baseline.txt"), []byte("ascii"), 0o600)).To(Succeed())
			})

			It("omits the partial entry", func() {
				results, err := vhsgen.ListBaselines(goldenDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(BeEmpty())
			})
		})

		Context("when a scenario dir is not executable (stat fails for ASCII)", func() {
			It("returns an error", func() {
				lockedSubdir := filepath.Join(goldenDir, "locked-sub")
				Expect(os.MkdirAll(lockedSubdir, 0o755)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(lockedSubdir, "baseline.txt"), []byte("ascii"), 0o600)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(lockedSubdir, "baseline.gif"), []byte("gif"), 0o600)).To(Succeed())
				Expect(os.Chmod(lockedSubdir, 0o000)).To(Succeed())
				defer os.Chmod(lockedSubdir, 0o755) //nolint:errcheck

				_, err := vhsgen.ListBaselines(goldenDir)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("BaselineInfo type", func() {
		It("holds Scenario, ASCIIPath, GIFPath, and ModTime fields", func() {
			info := vhsgen.BaselineInfo{
				Scenario:  "my-scenario",
				ASCIIPath: "/golden/my-scenario/baseline.txt",
				GIFPath:   "/golden/my-scenario/baseline.gif",
				ModTime:   time.Now(),
			}

			Expect(info.Scenario).To(Equal("my-scenario"))
			Expect(info.ASCIIPath).To(Equal("/golden/my-scenario/baseline.txt"))
			Expect(info.GIFPath).To(Equal("/golden/my-scenario/baseline.gif"))
			Expect(info.ModTime).NotTo(Equal(time.Time{}))
		})
	})
})
