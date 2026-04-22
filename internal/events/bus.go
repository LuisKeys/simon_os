package events

import "sync"

// EventBus is a simple in-process pub/sub bus.
type EventBus interface {
	Publish(event Event)
	Subscribe() <-chan Event
}

type Bus struct {
	mu          sync.RWMutex
	subscribers []chan Event
	bufferSize  int
}

func NewBus(bufferSize int) *Bus {
	return &Bus{bufferSize: bufferSize}
}

func (b *Bus) Publish(event Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, subscriber := range b.subscribers {
		select {
		case subscriber <- event:
		default:
		}
	}
}

func (b *Bus) Subscribe() <-chan Event {
	ch := make(chan Event, b.bufferSize)
	b.mu.Lock()
	b.subscribers = append(b.subscribers, ch)
	b.mu.Unlock()
	return ch
}
