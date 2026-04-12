package audit

import "github.com/google/uuid"

type Event struct {
	Timestamp int64      `json:"ts"`
	Action    string     `json:"action"`
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	URL       string     `json:"url"`
}

type Observer interface {
	Notify(event Event)
}

type Manager struct {
	observers []Observer
}

func NewManager() *Manager {
	return &Manager{
		observers: make([]Observer, 0),
	}
}

func (m *Manager) Register(observer Observer) {
	m.observers = append(m.observers, observer)
}

func (m *Manager) NotifyAll(event Event) {
	for _, observer := range m.observers {
		go observer.Notify(event)
	}
}
