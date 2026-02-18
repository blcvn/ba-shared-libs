package event

import (
	"context"
	"fmt"
	"log"
	"time"

	v32 "github.com/blcvn/backend/services/ba-agent-service/domain/v3.2"
	"github.com/google/uuid"
)

// SimpleEventEmitter is a simple in-memory event emitter
// In production, this would publish to a message queue (Redis, Kafka, etc.)
type SimpleEventEmitter struct {
	handlers []v32.EventHandler
	events   chan *v32.PlannerEvent
}

func NewSimpleEventEmitter() *SimpleEventEmitter {
	return &SimpleEventEmitter{
		handlers: make([]v32.EventHandler, 0),
		events:   make(chan *v32.PlannerEvent, 100), // Buffer size 100
	}
}

func (e *SimpleEventEmitter) RegisterHandler(handler v32.EventHandler) {
	e.handlers = append(e.handlers, handler)
}

// Start begins processing events in a background goroutine
func (e *SimpleEventEmitter) Start(ctx context.Context) {
	go func() {
		log.Println("[EVENT] Starting event processor...")
		for {
			select {
			case <-ctx.Done():
				log.Println("[EVENT] Stopping event processor...")
				return
			case event := <-e.events:
				e.processEvent(event)
			}
		}
	}()
}

func (e *SimpleEventEmitter) processEvent(event *v32.PlannerEvent) {
	log.Printf("[EVENT] Processing event: type=%s, document_id=%s, tier=%s",
		event.Type, event.DocumentID, event.Tier)

	// Dispatch to all registered handlers
	for _, handler := range e.handlers {
		// Recovery for each handler to prevent panic affecting others
		func(h v32.EventHandler) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[EVENT] Panic in handler: %v", r)
				}
			}()
			if err := h.Handle(event); err != nil {
				log.Printf("[EVENT] Handler error: %v", err)
			}
		}(handler)
	}
}

func (e *SimpleEventEmitter) Emit(event *v32.PlannerEvent) error {
	// Generate ID if not set
	if event.ID == "" {
		event.ID = uuid.New().String()
	}

	// Set timestamp if not set
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}

	// Push to channel (non-blocking if full)
	select {
	case e.events <- event:
		log.Printf("[EVENT] Emitted event (queued): type=%s, document_id=%s", event.Type, event.DocumentID)
	default:
		log.Printf("[EVENT] Event queue full, dropping event: type=%s", event.Type)
		return fmt.Errorf("event queue full")
	}

	return nil
}
