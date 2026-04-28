// Package audit provides an Audit observer that writes the event to the file
package audit

import (
	"context"
	"slices"
	"sync"

	"github.com/google/uuid"
)

const (
	workerCount = 10
	queueSize   = 1000
)

// Event - event to be sent to observers
type Event struct {
	Timestamp int64     `json:"ts"`
	Action    string    `json:"action"`
	UserID    uuid.UUID `json:"user_id"`
	URL       string    `json:"url"`
}

type Observer interface {
	Notify(event Event)
}

// Manager manages a collection of observers and facilitates notifying them of events.
type Manager struct {
	mu        sync.RWMutex
	observers []Observer
	queue     chan func()
}

// NewManager creates and initializes a new Manager with a worker pool.
func NewManager(ctx context.Context) *Manager {
	m := &Manager{
		observers: make([]Observer, 0),
		queue:     make(chan func(), queueSize),
	}
	go func() {
		<-ctx.Done()
		close(m.queue)
	}()
	for range workerCount {
		go m.worker()
	}
	return m
}

func (m *Manager) worker() {
	for task := range m.queue {
		task()
	}
}

// Register adds the given observer to the Manager's list of observers.
func (m *Manager) Register(observer Observer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.observers = append(m.observers, observer)
}

// Remove observer from the Manager's list of observers.
func (m *Manager) Remove(observer Observer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, o := range m.observers {
		if o == observer {
			m.observers = slices.Delete(m.observers, i, i+1)
			return
		}
	}
}

// NotifyAll notifies all registered observers asynchronously of the given event.
func (m *Manager) NotifyAll(ctx context.Context, event Event) {
	m.mu.RLock()
	observers := make([]Observer, len(m.observers))
	copy(observers, m.observers)
	m.mu.RUnlock()

	for _, observer := range observers {
		select {
		case m.queue <- func() {
			observer.Notify(event)
		}:
		case <-ctx.Done():
			return
		}
	}
}
