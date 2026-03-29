package lline

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func getFiles(ctx context.Context, root string, exclude, extensions []string) ([]string, error) {
	normalizedExtensions := make([]string, 0, len(extensions))
	for _, extension := range extensions {
		normalizedExtensions = append(normalizedExtensions, strings.Trim(extension, "."))
	}
	var paths []string

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}

		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		for _, ex := range exclude {
			if strings.Contains(path, ex) {
				return nil
			}
		}

		ext := strings.Trim(filepath.Ext(d.Name()), ".")
		if !slices.Contains(normalizedExtensions, ext) {
			return nil
		}

		paths = append(paths, path)
		return nil
	})

	return paths, err
}

func lineCount(ctx context.Context, path string, bufSize int) (uint64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, bufSize*1024*1024)
	var count uint64

	for scanner.Scan() {
		if ctx.Err() != nil {
			return count, ctx.Err()
		}

		count++
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}

	return count, nil
}
