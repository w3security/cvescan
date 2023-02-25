package parallel

import (
	"context"
	"io/fs"

	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"

	dio "github.com/aquasecurity/go-dep-parser/pkg/io"
)

type onFile[T any] func(string, fs.FileInfo, dio.ReadSeekerAt) (T, error)
type onResult[T any] func(T) error

func WalkDir[T any](ctx context.Context, fsys fs.FS, root string, slow bool,
	onFile onFile[T], onResult onResult[T]) error {

	g, ctx := errgroup.WithContext(ctx)
	paths := make(chan string)

	g.Go(func() error {
		defer close(paths)
		err := fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			} else if !d.Type().IsRegular() {
				return nil
			}

			select {
			case paths <- path:
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		})
		if err != nil {
			return xerrors.Errorf("walk error: %w", err)
		}
		return nil
	})

	// Start a fixed number of goroutines to read and digest files.
	c := make(chan T)
	limit := 10
	if slow {
		limit = 1
	}
	for i := 0; i < limit; i++ {
		g.Go(func() error {
			for path := range paths {
				if err := walk(ctx, fsys, path, c, onFile); err != nil {
					return err
				}
			}
			return nil
		})
	}
	go func() {
		_ = g.Wait()
		close(c)
	}()

	for res := range c {
		if err := onResult(res); err != nil {
			return err
		}
	}
	// Check whether any of the goroutines failed. Since g is accumulating the
	// errors, we don't need to send them (or check for them) in the individual
	// results sent on the channel.
	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}

func walk[T any](ctx context.Context, fsys fs.FS, path string, c chan T, onFile onFile[T]) error {
	f, err := fsys.Open(path)
	if err != nil {
		return xerrors.Errorf("file open error: %w", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return xerrors.Errorf("stat error: %w", err)
	}

	rsa, ok := f.(dio.ReadSeekerAt)
	if !ok {
		return xerrors.New("type assertion failed")
	}
	res, err := onFile(path, info, rsa)
	if err != nil {
		return xerrors.Errorf("on file: %w", err)
	}

	select {
	case c <- res:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}
