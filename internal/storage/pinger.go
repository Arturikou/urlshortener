package storage

import "context"

//go:generate mockery
type Pinger interface {
	Ping(ctx context.Context) error
}
