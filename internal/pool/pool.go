package pool

import "sync"

type Resettable interface {
	Reset()
}

type Pool[T Resettable] struct {
	p sync.Pool
}

func New[T Resettable](newFn func() T) *Pool[T] {
	return &Pool[T]{
		p: sync.Pool{
			New: func() any { return newFn() },
		},
	}
}

func (p *Pool[T]) Get() T {
	return p.p.Get().(T)
}

func (p *Pool[T]) Put(v T) {
	v.Reset()
	p.p.Put(v)
}
