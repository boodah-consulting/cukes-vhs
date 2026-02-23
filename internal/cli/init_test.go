package cli

import (
	"bytes"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("init subcommand", func() {
	Describe("runInit", func() {
		var tmpDir string

		BeforeEach(func() {
			tmpDir = GinkgoT().TempDir()
		})

		Context("with default output path", func() {
			It("creates a config file at the expected location", func() {
				opts := &initOptions{
					outputDir: tmpDir,
					force:     false,
				}
				var out bytes.Buffer

				err := runInit(opts, &out)
				Expect(err).NotTo(HaveOccurred())

				configPath := filepath.Join(tmpDir, "config.tape")
				_, statErr := os.Stat(configPath)
				Expect(statErr).NotTo(HaveOccurred())
			})

			It("writes the embedded default config content", func() {
				opts := &initOptions{
					outputDir: tmpDir,
					force:     false,
				}
				var out bytes.Buffer

				err := runInit(opts, &out)
				Expect(err).NotTo(HaveOccurred())

				configPath := filepath.Join(tmpDir, "config.tape")
				data, readErr := os.ReadFile(configPath)
				Expect(readErr).NotTo(HaveOccurred())
				Expect(string(data)).NotTo(BeEmpty())
				Expect(string(data)).To(ContainSubstring("Set Shell"))
			})

			It("prints confirmation messages to the writer", func() {
				opts := &initOptions{
					outputDir: tmpDir,
					force:     false,
				}
				var out bytes.Buffer

				err := runInit(opts, &out)
				Expect(err).NotTo(HaveOccurred())

				output := out.String()
				Expect(output).To(ContainSubstring("Created config file:"))
				Expect(output).To(ContainSubstring("customise"))
				Expect(output).To(ContainSubstring("--config-source"))
			})
		})

		Context("when config file already exists", func() {
			var configPath string

			BeforeEach(func() {
				configPath = filepath.Join(tmpDir, "config.tape")
				Expect(os.WriteFile(configPath, []byte("existing content"), 0o644)).To(Succeed())
			})

			Context("without --force flag", func() {
				It("returns an error indicating the file already exists", func() {
					opts := &initOptions{
						outputDir: tmpDir,
						force:     false,
					}
					var out bytes.Buffer

					err := runInit(opts, &out)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("config file already exists"))
					Expect(err.Error()).To(ContainSubstring("--force"))
				})

				It("does not overwrite the existing file", func() {
					opts := &initOptions{
						outputDir: tmpDir,
						force:     false,
					}
					var out bytes.Buffer

					_ = runInit(opts, &out)

					data, readErr := os.ReadFile(configPath)
					Expect(readErr).NotTo(HaveOccurred())
					Expect(string(data)).To(Equal("existing content"))
				})
			})

			Context("with --force flag", func() {
				It("overwrites the existing config file", func() {
					opts := &initOptions{
						outputDir: tmpDir,
						force:     true,
					}
					var out bytes.Buffer

					err := runInit(opts, &out)
					Expect(err).NotTo(HaveOccurred())

					data, readErr := os.ReadFile(configPath)
					Expect(readErr).NotTo(HaveOccurred())
					Expect(string(data)).NotTo(Equal("existing content"))
					Expect(string(data)).To(ContainSubstring("Set Shell"))
				})

				It("prints confirmation messages", func() {
					opts := &initOptions{
						outputDir: tmpDir,
						force:     true,
					}
					var out bytes.Buffer

					err := runInit(opts, &out)
					Expect(err).NotTo(HaveOccurred())
					Expect(out.String()).To(ContainSubstring("Created config file:"))
				})
			})
		})

		Context("when directory creation fails", func() {
			It("returns an error when the parent is read-only", func() {
			skipIfWindows()
				readOnlyDir := filepath.Join(tmpDir, "readonly")
				Expect(os.MkdirAll(readOnlyDir, 0o755)).To(Succeed())
				Expect(os.Chmod(readOnlyDir, 0o000)).To(Succeed())
				defer os.Chmod(readOnlyDir, 0o755) //nolint:errcheck

				opts := &initOptions{
					outputDir: filepath.Join(readOnlyDir, "nested", "config"),
					force:     false,
				}
				var out bytes.Buffer

				err := runInit(opts, &out)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("creating output directory"))
			})
		})

		Context("when output directory does not exist", func() {
			It("creates the directory and the config file", func() {
				nestedDir := filepath.Join(tmpDir, "nested", "output")
				opts := &initOptions{
					outputDir: nestedDir,
					force:     false,
				}
				var out bytes.Buffer

				err := runInit(opts, &out)
				Expect(err).NotTo(HaveOccurred())

				configPath := filepath.Join(nestedDir, "config.tape")
				_, statErr := os.Stat(configPath)
				Expect(statErr).NotTo(HaveOccurred())
			})
		})
	})

	Describe("parseInitFlags", func() {
		Context("with --force flag", func() {
			It("sets force to true", func() {
				var errOut bytes.Buffer
				opts, err := parseInitFlags([]string{"--force"}, &errOut)
				Expect(err).NotTo(HaveOccurred())
				Expect(opts.force).To(BeTrue())
			})
		})

		Context("with --output flag", func() {
			It("sets the output directory", func() {
				var errOut bytes.Buffer
				opts, err := parseInitFlags([]string{"--output", "/custom/path"}, &errOut)
				Expect(err).NotTo(HaveOccurred())
				Expect(opts.outputDir).To(Equal("/custom/path"))
			})
		})

		Context("with no flags", func() {
			It("uses the default output directory", func() {
				var errOut bytes.Buffer
				opts, err := parseInitFlags([]string{}, &errOut)
				Expect(err).NotTo(HaveOccurred())
				Expect(opts.outputDir).To(Equal("config/"))
			})

			It("sets force to false by default", func() {
				var errOut bytes.Buffer
				opts, err := parseInitFlags([]string{}, &errOut)
				Expect(err).NotTo(HaveOccurred())
				Expect(opts.force).To(BeFalse())
			})
		})

		Context("with both flags", func() {
			It("parses both --force and --output correctly", func() {
				var errOut bytes.Buffer
				opts, err := parseInitFlags([]string{"--force", "--output", "/my/dir"}, &errOut)
				Expect(err).NotTo(HaveOccurred())
				Expect(opts.force).To(BeTrue())
				Expect(opts.outputDir).To(Equal("/my/dir"))
			})
		})

		Context("with unknown flag", func() {
			It("returns an error", func() {
				var errOut bytes.Buffer
				_, err := parseInitFlags([]string{"--unknown-flag"}, &errOut)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
