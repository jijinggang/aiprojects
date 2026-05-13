package main

import (
	"archive/zip"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	gitignore "github.com/sabhiram/go-gitignore"
)

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(value string) error {
	*m = append(*m, value)
	return nil
}

type config struct {
	sourceDir  string
	outputPath string
	patterns   []string
}

func parseFlags(args []string) (config, error) {
	var cfg config
	fs := flag.NewFlagSet("ziptool", flag.ContinueOnError)
	outputPath := fs.String("o", "output.zip", "output zip file path")
	var patterns multiFlag
	fs.Var(&patterns, "p", "gitignore-style exclusion patterns (repeatable), default is .svn/")

	if err := fs.Parse(args); err != nil {
		return cfg, err
	}

	if len(patterns) == 0 {
		patterns = multiFlag{".svn/"}
	}
	cfg.patterns = []string(patterns)

	if fs.NArg() == 0 {
		return cfg, errors.New("source directory is required")
	}
	cfg.sourceDir = fs.Arg(0)
	cfg.outputPath = *outputPath
	return cfg, nil
}

func collectFiles(sourceDir string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}
		relPath = filepath.ToSlash(relPath)
		if d.IsDir() {
			relPath += "/"
		}
		paths = append(paths, relPath)
		return nil
	})
	return paths, err
}

func filterPaths(paths []string, patterns []string) ([]string, error) {
	if len(patterns) == 0 {
		return paths, nil
	}
	processed := make([]string, len(patterns))
	for i, p := range patterns {
		processed[i] = strings.ReplaceAll(p, "?", "[^/]")
	}
	gi := gitignore.CompileIgnoreLines(processed...)
	var filtered []string
	for _, p := range paths {
		if !gi.MatchesPath(p) {
			filtered = append(filtered, p)
		}
	}
	return filtered, nil
}

func writeZip(sourceDir, outputPath string, paths []string) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("cannot create output file: %w", err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	for _, p := range paths {
		if err := writeZipEntry(zw, sourceDir, p); err != nil {
			zw.Close()
			return err
		}
	}
	return zw.Close()
}

func writeZipEntry(zw *zip.Writer, sourceDir, entryPath string) error {
	for _, part := range strings.Split(filepath.ToSlash(entryPath), "/") {
		if part == ".." {
			fmt.Fprintf(os.Stderr, "warning: skipping path with '..': %s\n", entryPath)
			return nil
		}
	}

	entryPath = filepath.ToSlash(entryPath)

	if strings.HasSuffix(entryPath, "/") {
		_, err := zw.Create(entryPath)
		return err
	}

	sf, err := os.Open(filepath.Join(sourceDir, entryPath))
	if err != nil {
		return fmt.Errorf("cannot open %s: %w", entryPath, err)
	}
	defer sf.Close()

	w, err := zw.Create(entryPath)
	if err != nil {
		return fmt.Errorf("cannot create zip entry %s: %w", entryPath, err)
	}
	if _, err := io.Copy(w, sf); err != nil {
		return fmt.Errorf("cannot write %s: %w", entryPath, err)
	}
	return nil
}

func main() {
	cfg, err := parseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	paths, err := collectFiles(cfg.sourceDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	filtered, err := filterPaths(paths, cfg.patterns)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	if err := writeZip(cfg.sourceDir, cfg.outputPath, filtered); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	fmt.Println("Created:", cfg.outputPath)
}
