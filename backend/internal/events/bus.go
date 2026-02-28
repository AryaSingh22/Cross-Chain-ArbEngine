package events

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

// EventType categorizes events flowing through the bus
type EventType string

const (
	EventPriceUpdate EventType = "price.update"
	EventOpportunity EventType = "arb.opportunity"
	EventRelayAlert  EventType = "relay.alert"
	EventPacketStuck EventType = "relay.packet_stuck"
)

// Event represents a message on the internal event bus
type Event struct {
	Type    EventType   `json:"type"`
	Payload interface{} `json:"payload"`
	At      time.Time   `json:"at"`
}

// EventBus provides a goroutine-safe publish/subscribe mechanism
type EventBus struct {
	subscribers map[EventType][]chan Event
	mu          sync.RWMutex
	logger      *zap.Logger
	bufferSize  int
}

// NewEventBus creates a new event bus
func NewEventBus(logger *zap.Logger) *EventBus {
	return &EventBus{
		subscribers: make(map[EventType][]chan Event),
		logger:      logger,
		bufferSize:  256,
	}
}

// Subscribe registers a new subscriber for a given event type
// Returns a channel that will receive events of the given type
func (eb *EventBus) Subscribe(eventType EventType) chan Event {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	ch := make(chan Event, eb.bufferSize)
	eb.subscribers[eventType] = append(eb.subscribers[eventType], ch)
	eb.logger.Debug("new subscriber registered", zap.String("eventType", string(eventType)))
	return ch
}

// Unsubscribe removes a subscriber channel
func (eb *EventBus) Unsubscribe(eventType EventType, ch chan Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	subs := eb.subscribers[eventType]
	for i, sub := range subs {
		if sub == ch {
			eb.subscribers[eventType] = append(subs[:i], subs[i+1:]...)
			close(ch)
			break
		}
	}
}

// Publish sends an event to all subscribers of the given type
func (eb *EventBus) Publish(event Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	event.At = time.Now()
	subs := eb.subscribers[event.Type]

	for _, ch := range subs {
		select {
		case ch <- event:
		default:
			eb.logger.Warn("subscriber channel full, dropping event",
				zap.String("eventType", string(event.Type)))
		}
	}
}

// SubscriberCount returns the number of subscribers for an event type
func (eb *EventBus) SubscriberCount(eventType EventType) int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return len(eb.subscribers[eventType])
}
