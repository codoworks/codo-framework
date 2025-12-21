package app

import (
	"context"
	"fmt"

	"github.com/codoworks/codo-framework/core/config"
	"golang.org/x/sync/errgroup"
)

// workerDaemonApp implements DaemonApp for background worker processes
type workerDaemonApp struct {
	*foundation
	workers []Worker
	mode    AppMode
}

// RegisterWorker registers a background worker
func (d *workerDaemonApp) RegisterWorker(w Worker) {
	d.workers = append(d.workers, w)
}

// Start starts all registered workers concurrently
func (d *workerDaemonApp) Start(ctx context.Context) error {
	if len(d.workers) == 0 {
		return fmt.Errorf("no workers registered")
	}

	g, ctx := errgroup.WithContext(ctx)

	for _, worker := range d.workers {
		w := worker
		g.Go(func() error {
			log := getOrCreateLogger()
			log.Infof("[Daemon] Starting worker: %s", w.Name())
			return w.Start(ctx)
		})
	}

	return g.Wait()
}

// Stop stops all workers gracefully
func (d *workerDaemonApp) Stop(ctx context.Context) error {
	log := getOrCreateLogger()
	for _, worker := range d.workers {
		log.Infof("[Daemon] Stopping worker: %s", worker.Name())
		if err := worker.Stop(ctx); err != nil {
			return fmt.Errorf("failed to stop worker %s: %w", worker.Name(), err)
		}
	}
	return nil
}

// Shutdown stops all workers and cleans up clients
func (d *workerDaemonApp) Shutdown(ctx context.Context) error {
	if err := d.Stop(ctx); err != nil {
		return err
	}
	return d.foundation.Shutdown(ctx)
}

// Mode returns the bootstrap mode
func (d *workerDaemonApp) Mode() AppMode {
	return d.mode
}

// bootstrapWorkerDaemon creates a daemon app for background workers
func bootstrapWorkerDaemon(cfg *config.Config, opts BootstrapOptions) (BaseApp, error) {
	// Validate worker registrar
	if opts.WorkerRegistrar == nil {
		return nil, fmt.Errorf("WorkerRegistrar is required for WorkerDaemon mode")
	}

	// Phase 1: Initialize foundation
	foundation, err := initFoundation(cfg, opts)
	if err != nil {
		return nil, fmt.Errorf("foundation init: %w", err)
	}

	// Phase 2: Run migrations
	if err := runMigrations(foundation, opts.MigrationAdder); err != nil {
		return nil, fmt.Errorf("migrations: %w", err)
	}

	// Phase 3: Create daemon (no HTTP)
	daemon := &workerDaemonApp{
		foundation: foundation,
		workers:    make([]Worker, 0),
		mode:       WorkerDaemon,
	}

	// Register workers
	if err := opts.WorkerRegistrar(daemon); err != nil {
		return nil, fmt.Errorf("worker registration: %w", err)
	}

	return daemon, nil
}
