package main

import (
	"context"
	"embed"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"word-search-in-files/pkg"
)

//go:embed examples/*
var dir embed.FS

func main() {
	srv := &http.Server{
		Addr: ":5000",
	}
	searcher := &pkg.Searcher{FS: dir}

	srv.Handler = searcher.RequestHandler()

	ctx := getContext()
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return Serve(srv, ctx)
	})

	err := eg.Wait()
	if err != nil {
		panic(err)
	}

}

func Serve(srv *http.Server, ctx context.Context) error {
	errChan := make(chan error, 1)
	defer close(errChan)
	go func() {
		errChan <- srv.ListenAndServe()
	}()

	select {
	case err := <-errChan:
		return err

	case <-ctx.Done():
		if err := srv.Shutdown(ctx); err != nil {
			return err
		}
		return nil
	}
}

func getContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-c
		slog.InfoContext(ctx, "%s", sig)
		cancel()
	}()

	return ctx
}
