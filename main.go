package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tylerbutler/commit-config-gen/internal/config"
	"github.com/tylerbutler/commit-config-gen/internal/generator"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "commit-config-gen",
		Usage: "Generate commit configs from a single source of truth",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "commit-types.json",
				Usage:   "path to commit-types.json",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "generate",
				Usage: "Generate cliff.toml and .commitlintrc.json",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "dry-run",
						Usage: "print generated content without writing files",
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Value:   ".",
						Usage:   "output directory for generated files",
					},
				},
				Action: runGenerate,
			},
			{
				Name:  "check",
				Usage: "Verify configs are in sync with commit-types.json",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "dir",
						Aliases: []string{"d"},
						Value:   ".",
						Usage:   "directory containing config files to check",
					},
				},
				Action: runCheck,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runGenerate(c *cli.Context) error {
	configPath := c.String("config")
	outputDir := c.String("output")
	dryRun := c.Bool("dry-run")

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Generate cliff.toml
	cliffContent, err := generator.GenerateCliff(cfg)
	if err != nil {
		return fmt.Errorf("failed to generate cliff.toml: %w", err)
	}

	// Generate .commitlintrc.json
	commitlintContent, err := generator.GenerateCommitlint(cfg)
	if err != nil {
		return fmt.Errorf("failed to generate .commitlintrc.json: %w", err)
	}

	if dryRun {
		fmt.Println("=== cliff.toml ===")
		fmt.Println(cliffContent)
		fmt.Println("\n=== .commitlintrc.json ===")
		fmt.Println(commitlintContent)
		return nil
	}

	// Write files
	cliffPath := filepath.Join(outputDir, "cliff.toml")
	if err := os.WriteFile(cliffPath, []byte(cliffContent), 0644); err != nil {
		return fmt.Errorf("failed to write cliff.toml: %w", err)
	}
	fmt.Printf("Wrote %s\n", cliffPath)

	commitlintPath := filepath.Join(outputDir, ".commitlintrc.json")
	if err := os.WriteFile(commitlintPath, []byte(commitlintContent), 0644); err != nil {
		return fmt.Errorf("failed to write .commitlintrc.json: %w", err)
	}
	fmt.Printf("Wrote %s\n", commitlintPath)

	return nil
}

func runCheck(c *cli.Context) error {
	configPath := c.String("config")
	dir := c.String("dir")

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var errors []string

	// Check cliff.toml
	cliffPath := filepath.Join(dir, "cliff.toml")
	if _, err := os.Stat(cliffPath); err == nil {
		expected, err := generator.GenerateCliff(cfg)
		if err != nil {
			return fmt.Errorf("failed to generate cliff.toml: %w", err)
		}

		actual, err := os.ReadFile(cliffPath)
		if err != nil {
			return fmt.Errorf("failed to read cliff.toml: %w", err)
		}

		if normalize(expected) != normalize(string(actual)) {
			errors = append(errors, "cliff.toml is out of sync with commit-types.json")
		}
	}

	// Check .commitlintrc.json
	commitlintPath := filepath.Join(dir, ".commitlintrc.json")
	if _, err := os.Stat(commitlintPath); err == nil {
		expected, err := generator.GenerateCommitlint(cfg)
		if err != nil {
			return fmt.Errorf("failed to generate .commitlintrc.json: %w", err)
		}

		actual, err := os.ReadFile(commitlintPath)
		if err != nil {
			return fmt.Errorf("failed to read .commitlintrc.json: %w", err)
		}

		if normalize(expected) != normalize(string(actual)) {
			errors = append(errors, ".commitlintrc.json is out of sync with commit-types.json")
		}
	}

	if len(errors) > 0 {
		fmt.Println("Config sync check failed:")
		for _, e := range errors {
			fmt.Printf("  - %s\n", e)
		}
		fmt.Println("\nRun 'commit-config-gen generate' to fix")
		os.Exit(1)
	}

	fmt.Println("All configs are in sync with commit-types.json")
	return nil
}

func normalize(s string) string {
	return strings.TrimSpace(s)
}
