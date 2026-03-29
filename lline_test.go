package lline

import (
	"bufio"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestGetFilesReturnsMatchingFilesFromNestedDirectories(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	goFile := writeTestFile(t, root, "main.go", []string{"package main"})
	nestedGoFile := writeTestFile(t, root, filepath.Join("internal", "worker.go"),
		[]string{"package internal", "", "func Work() {}"})
	nestedTxtFile := writeTestFile(t, root, filepath.Join("internal", "notes.txt"), []string{"note"})
	rootTxtFile := writeTestFile(t, root, filepath.Join("internal", "worker.txt"), []string{"ignore"})
	writeTestFile(t, root, "README.md", []string{"# readme"})

	files, err := getFiles(context.Background(), root, nil, []string{".go", "txt"})
	if err != nil {
		t.Fatalf("getFiles returned error: %v", err)
	}

	got := make(map[string]struct{}, len(files))
	for _, file := range files {
		got[file] = struct{}{}
	}

	want := map[string]struct{}{
		goFile:        {},
		nestedGoFile:  {},
		nestedTxtFile: {},
		rootTxtFile:   {},
	}

	if len(got) != len(want) {
		t.Fatalf("unexpected file count: got %d, want %d", len(got), len(want))
	}

	for file := range want {
		if _, ok := got[file]; !ok {
			t.Fatalf("expected file %q to be returned", file)
		}
	}
}

func TestCountLinesCountsOnlyMatchingExtension(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTestFile(t, root, "main.go", []string{"package main", "", "func main() {}"})
	writeTestFile(t, root, filepath.Join("pkg", "counter.go"),
		[]string{"package pkg", "", "func Count() {}", "func Next() {}"})
	writeTestFile(t, root, "notes.txt", []string{"ignored", "lines"})

	count, err := CountLines(context.Background(), root, nil, []string{"go"}, 2, 10)
	if err != nil {
		t.Fatalf("CountLines returned error: %v", err)
	}

	const want uint64 = 7
	if count != want {
		t.Fatalf("unexpected line count: got %d, want %d", count, want)
	}
}

func TestCountLinesReturnsZeroWhenContextCanceled(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTestFile(t, root, "001.go", []string{"package main", "func one() {}"})
	writeTestFile(t, root, "002.go", []string{"package main", "func two() {}", "func three() {}"})
	writeLargeTestFile(t, filepath.Join(root, "003.go"), 100_000, 512)
	writeTestFile(t, root, "ignored.txt", []string{"ignore"})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	count, err := CountLines(ctx, root, nil, []string{"go"}, 1, 10)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled error, got %v", err)
	}

	const want uint64 = 0
	if count != want {
		t.Fatalf("unexpected line count after cancellation: got %d, want %d", count, want)
	}
}

func TestGetFilesSkipsExcludedPaths(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	includedFile := writeTestFile(t, root, filepath.Join("pkg", "main.go"), []string{"package pkg"})
	writeTestFile(t, root, filepath.Join("vendor", "ignored.go"), []string{"package vendor"})

	files, err := getFiles(context.Background(), root, []string{"vendor"}, []string{"go"})
	if err != nil {
		t.Fatalf("getFiles returned error: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("unexpected file count: got %d, want 1", len(files))
	}

	if files[0] != includedFile {
		t.Fatalf("unexpected included file: got %q, want %q", files[0], includedFile)
	}
}

func TestLineCountHandlesLongLines(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "long.go")
	writeLargeTestFile(t, path, 2, 2*1024*1024)

	count, err := lineCount(context.Background(), path, 10)
	if err != nil {
		t.Fatalf("lineCount returned error: %v", err)
	}

	const want uint64 = 2
	if count != want {
		t.Fatalf("unexpected long-line count: got %d, want %d", count, want)
	}
}

func TestLineCountFailsWhenBufferTooSmall(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "too-long.go")
	writeLargeTestFile(t, path, 1, 2*1024*1024)

	_, err := lineCount(context.Background(), path, 1)
	if err == nil {
		t.Fatal("expected lineCount to fail when buffer is too small")
	}
}

func writeTestFile(t *testing.T, root, relativePath string, lines []string) string {
	t.Helper()

	path := filepath.Join(root, relativePath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create directory for %q: %v", path, err)
	}

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create file %q: %v", path, err)
	}
	defer func() {
		_ = file.Close()
	}()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			t.Fatalf("write file %q: %v", path, err)
		}
	}
	if err := writer.Flush(); err != nil {
		t.Fatalf("flush file %q: %v", path, err)
	}

	return path
}

func writeLargeTestFile(t *testing.T, path string, lineCount, lineWidth int) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create directory for %q: %v", path, err)
	}

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create file %q: %v", path, err)
	}
	defer func() {
		_ = file.Close()
	}()

	writer := bufio.NewWriterSize(file, 1<<20)
	line := make([]byte, lineWidth)
	for i := range line {
		line[i] = 'a'
	}

	for i := 0; i < lineCount; i++ {
		if _, err := writer.Write(line); err != nil {
			t.Fatalf("write file %q: %v", path, err)
		}
		if err := writer.WriteByte('\n'); err != nil {
			t.Fatalf("write newline to %q: %v", path, err)
		}
	}

	if err := writer.Flush(); err != nil {
		t.Fatalf("flush file %q: %v", path, err)
	}
}
