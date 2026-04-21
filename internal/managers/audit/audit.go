// Package audit provides an Audit observer that writes the event to the file
package audit

import "github.com/google/uuid"

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
	observers []Observer
}

// NewManager creates and initializes a new Manager with an empty list of observers.
func NewManager() *Manager {
	return &Manager{
		observers: make([]Observer, 0),
	}
}

// Register adds the given observer to the Manager's list of observers.
func (m *Manager) Register(observer Observer) {
	m.observers = append(m.observers, observer)
}

// NotifyAll notifies all registered observers asynchronously of the given event.
func (m *Manager) NotifyAll(event Event) {
	for _, observer := range m.observers {
		go observer.Notify(event)
	}
}
