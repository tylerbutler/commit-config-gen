package main

import (
	"bytes"
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
				Usage: "Generate config files from commit-types.json",
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
					&cli.StringSliceFlag{
						Name:    "generators",
						Aliases: []string{"g"},
						Usage:   "generators to run (default: all). Use --generators to list available generators",
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
					&cli.StringSliceFlag{
						Name:    "generators",
						Aliases: []string{"g"},
						Usage:   "generators to check (default: all present files)",
					},
				},
				Action: runCheck,
			},
			{
				Name:  "list",
				Usage: "List available generators",
				Action: func(c *cli.Context) error {
					for _, g := range generator.All() {
						fmt.Printf("%-25s %s\n", g.Name(), g.FileName())
					}
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func selectedGenerators(names []string) ([]generator.Generator, error) {
	if len(names) == 0 {
		return generator.All(), nil
	}
	var gens []generator.Generator
	for _, name := range names {
		g, err := generator.Get(name)
		if err != nil {
			return nil, fmt.Errorf("%w (available: %s)", err, strings.Join(generator.Names(), ", "))
		}
		gens = append(gens, g)
	}
	return gens, nil
}

func runGenerate(c *cli.Context) error {
	configPath := c.String("config")
	outputDir := c.String("output")
	dryRun := c.Bool("dry-run")

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	gens, err := selectedGenerators(c.StringSlice("generators"))
	if err != nil {
		return err
	}

	for _, gen := range gens {
		filePath := filepath.Join(outputDir, gen.FileName())

		// Read existing file for merge
		var existing []byte
		if data, err := os.ReadFile(filePath); err == nil {
			existing = data
		}

		output, err := gen.Generate(cfg, existing)
		if err != nil {
			return fmt.Errorf("failed to generate %s: %w", gen.FileName(), err)
		}

		if dryRun {
			fmt.Printf("=== %s ===\n", gen.FileName())
			fmt.Println(string(output))
			continue
		}

		if err := os.WriteFile(filePath, output, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", gen.FileName(), err)
		}
		fmt.Printf("Wrote %s\n", filePath)
	}

	return nil
}

func runCheck(c *cli.Context) error {
	configPath := c.String("config")
	dir := c.String("dir")

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	gens, err := selectedGenerators(c.StringSlice("generators"))
	if err != nil {
		return err
	}

	var errs []string

	for _, gen := range gens {
		filePath := filepath.Join(dir, gen.FileName())

		actual, err := os.ReadFile(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				continue // skip files that don't exist
			}
			return fmt.Errorf("failed to read %s: %w", gen.FileName(), err)
		}

		expected, err := gen.Generate(cfg, actual)
		if err != nil {
			return fmt.Errorf("failed to generate %s: %w", gen.FileName(), err)
		}

		if !bytes.Equal(bytes.TrimSpace(expected), bytes.TrimSpace(actual)) {
			errs = append(errs, fmt.Sprintf("%s is out of sync with commit-types.json", gen.FileName()))
		}
	}

	if len(errs) > 0 {
		fmt.Println("Config sync check failed:")
		for _, e := range errs {
			fmt.Printf("  - %s\n", e)
		}
		fmt.Println("\nRun 'commit-config-gen generate' to fix")
		os.Exit(1)
	}

	fmt.Println("All configs are in sync with commit-types.json")
	return nil
}
