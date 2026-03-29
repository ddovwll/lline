package lline

import (
	"context"
	"sync/atomic"

	"golang.org/x/sync/errgroup"
)

func CountLines(ctx context.Context, root string, exclude, extensions []string, workers, bufSize int) (uint64, error) {
	var count atomic.Uint64
	eg, ctxEg := errgroup.WithContext(ctx)
	paths, err := getFiles(ctx, root, exclude, extensions)
	if err != nil {
		return 0, err
	}
	eg.SetLimit(workers)

	for _, path := range paths {
		if ctx.Err() != nil {
			break
		}

		eg.Go(func() error {
			lines, err := lineCount(ctxEg, path, bufSize)
			if err != nil {
				return err
			}

			count.Add(lines)
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return 0, err
	}

	return count.Load(), nil
}
