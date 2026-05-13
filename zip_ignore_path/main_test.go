package main

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestParseFlags(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		cfg, err := parseFlags([]string{"./src"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.sourceDir != "./src" {
			t.Errorf("sourceDir = %q, want %q", cfg.sourceDir, "./src")
		}
		if cfg.outputPath != "output.zip" {
			t.Errorf("outputPath = %q, want %q", cfg.outputPath, "output.zip")
		}
		if !reflect.DeepEqual(cfg.patterns, []string{".svn/"}) {
			t.Errorf("patterns = %v, want %v", cfg.patterns, []string{".svn/"})
		}
	})

	t.Run("custom-output-and-patterns", func(t *testing.T) {
		cfg, err := parseFlags([]string{"-o", "out.zip", "-p", "*.log", "-p", "build/", "./src"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.outputPath != "out.zip" {
			t.Errorf("outputPath = %q, want %q", cfg.outputPath, "out.zip")
		}
		if !reflect.DeepEqual(cfg.patterns, []string{"*.log", "build/"}) {
			t.Errorf("patterns = %v, want %v", cfg.patterns, []string{"*.log", "build/"})
		}
	})

	t.Run("missing-source-dir", func(t *testing.T) {
		_, err := parseFlags([]string{"-o", "out.zip"})
		if err == nil {
			t.Fatal("expected error for missing source dir")
		}
	})
}

func TestCollectFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.txt", "hello")
	writeFile(t, dir, "b.log", "log content")
	writeFile(t, dir, "sub/c.txt", "deep")
	writeFile(t, dir, "sub/d.log", "deep log")

	paths, err := collectFiles(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sort.Strings(paths)

	expected := []string{"a.txt", "b.log", "sub/", "sub/c.txt", "sub/d.log"}
	sort.Strings(expected)

	if !reflect.DeepEqual(paths, expected) {
		t.Errorf("got %v, want %v", paths, expected)
	}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	fullPath := filepath.Join(dir, name)
	os.MkdirAll(filepath.Dir(fullPath), 0755)
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestFilterPaths(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		paths := []string{"a.txt", "b.log", "sub/", "sub/c.txt", "sub/d.log"}
		patterns := []string{"*.log", "sub/"}

		got, err := filterPaths(paths, patterns)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := []string{"a.txt"}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("got %v, want %v", got, expected)
		}
	})

	t.Run("star", func(t *testing.T) {
		paths := []string{"a.txt", "a.go", "b.txt", "sub/c.go"}
		got, err := filterPaths(paths, []string{"*.txt"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []string{"a.go", "sub/c.go"}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("got %v, want %v", got, expected)
		}
	})

	t.Run("question", func(t *testing.T) {
		paths := []string{"temp1.txt", "tempa.txt", "temp12.txt"}
		got, err := filterPaths(paths, []string{"temp?.txt"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []string{"temp12.txt"}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("got %v, want %v", got, expected)
		}
	})

	t.Run("doublestar", func(t *testing.T) {
		paths := []string{"backup", "a/backup", "a/b/backup", "other.txt"}
		got, err := filterPaths(paths, []string{"**/backup"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []string{"other.txt"}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("got %v, want %v", got, expected)
		}
	})

	t.Run("root-anchor", func(t *testing.T) {
		paths := []string{"config.yml", "sub/config.yml"}
		got, err := filterPaths(paths, []string{"/config.yml"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []string{"sub/config.yml"}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("got %v, want %v", got, expected)
		}
	})

	t.Run("dir-suffix", func(t *testing.T) {
		paths := []string{"build", "build/", "src/main.go"}
		got, err := filterPaths(paths, []string{"build/"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []string{"build", "src/main.go"}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("got %v, want %v", got, expected)
		}
	})

	t.Run("charclass", func(t *testing.T) {
		paths := []string{"a.o", "a.a", "a.txt"}
		got, err := filterPaths(paths, []string{"*.[oa]"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []string{"a.txt"}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("got %v, want %v", got, expected)
		}
	})
}

func TestWriteZip(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.txt", "hello")
	writeFile(t, dir, "sub/b.txt", "world")
	os.MkdirAll(filepath.Join(dir, "empty"), 0755)

	outputPath := filepath.Join(t.TempDir(), "test.zip")
	paths := []string{"a.txt", "empty/", "sub/", "sub/b.txt"}

	if err := writeZip(dir, outputPath, paths); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r, err := zip.OpenReader(outputPath)
	if err != nil {
		t.Fatalf("cannot open zip: %v", err)
	}
	defer r.Close()

	names := make([]string, len(r.File))
	contents := make(map[string]string)
	for i, f := range r.File {
		names[i] = f.Name
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("cannot open entry %s: %v", f.Name, err)
		}
		b, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Fatalf("cannot read entry %s: %v", f.Name, err)
		}
		contents[f.Name] = string(b)
	}

	sort.Strings(names)
	expectedNames := []string{"a.txt", "empty/", "sub/", "sub/b.txt"}
	sort.Strings(expectedNames)
	if !reflect.DeepEqual(names, expectedNames) {
		t.Errorf("entries = %v, want %v", names, expectedNames)
	}

	if contents["a.txt"] != "hello" {
		t.Errorf("a.txt content = %q, want %q", contents["a.txt"], "hello")
	}
	if contents["sub/b.txt"] != "world" {
		t.Errorf("sub/b.txt content = %q, want %q", contents["sub/b.txt"], "world")
	}
}

func TestZipSlipPrevention(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "safe.txt", "safe")

	outputPath := filepath.Join(t.TempDir(), "test.zip")
	paths := []string{"safe.txt", "../escape.txt"}

	if err := writeZip(dir, outputPath, paths); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r, err := zip.OpenReader(outputPath)
	if err != nil {
		t.Fatalf("cannot open zip: %v", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name == "../escape.txt" {
			t.Errorf("zip slip entry was not rejected: %s", f.Name)
		}
	}
}