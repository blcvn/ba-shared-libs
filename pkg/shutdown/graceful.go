package shutdown

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// GracefulShutdown manages graceful shutdown of services
type GracefulShutdown struct {
	timeout  time.Duration
	handlers []ShutdownHandler
	mu       sync.Mutex
}

// ShutdownHandler is a function that performs cleanup
type ShutdownHandler func(ctx context.Context) error

// NewGracefulShutdown creates a new graceful shutdown manager
func NewGracefulShutdown(timeout time.Duration) *GracefulShutdown {
	return &GracefulShutdown{
		timeout:  timeout,
		handlers: make([]ShutdownHandler, 0),
	}
}

// Register registers a shutdown handler
func (gs *GracefulShutdown) Register(handler ShutdownHandler) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.handlers = append(gs.handlers, handler)
}

// Wait waits for shutdown signal and executes handlers
func (gs *GracefulShutdown) Wait() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	sig := <-sigChan
	log.Printf("[Shutdown] Received signal: %v. Starting graceful shutdown...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), gs.timeout)
	defer cancel()

	gs.executeHandlers(ctx)

	log.Println("[Shutdown] Graceful shutdown complete")
}

func (gs *GracefulShutdown) executeHandlers(ctx context.Context) {
	gs.mu.Lock()
	handlers := make([]ShutdownHandler, len(gs.handlers))
	copy(handlers, gs.handlers)
	gs.mu.Unlock()

	// Execute handlers in reverse order (LIFO)
	for i := len(handlers) - 1; i >= 0; i-- {
		handler := handlers[i]

		if err := handler(ctx); err != nil {
			log.Printf("[Shutdown] Handler %d failed: %v", i, err)
		} else {
			log.Printf("[Shutdown] Handler %d completed successfully", i)
		}

		// Check if context is done
		select {
		case <-ctx.Done():
			log.Printf("[Shutdown] Timeout reached, forcing shutdown")
			return
		default:
		}
	}
}

// WaitWithCleanup is a helper that waits for shutdown and runs cleanup
func WaitWithCleanup(timeout time.Duration, cleanup func(context.Context) error) {
	gs := NewGracefulShutdown(timeout)
	gs.Register(cleanup)
	gs.Wait()
}
