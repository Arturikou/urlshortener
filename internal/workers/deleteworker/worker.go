// Package deleteworker provides a worker for deleting URLs.
package deleteworker

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	chanSize      = 1024
	flushLimit    = 512
	flushInterval = 10 * time.Second
)

type deleteTask struct {
	userID  uuid.UUID
	aliases []string
}

//go:generate mockery
type Deleter interface {
	DeleteUserURLs(ctx context.Context, userID uuid.UUID, aliases []string) error
}

type Worker struct {
	deleter  Deleter
	deleteCh chan deleteTask
	logger   *zap.SugaredLogger
}

func New(deleter Deleter, logger *zap.SugaredLogger) *Worker {
	return &Worker{
		deleter:  deleter,
		deleteCh: make(chan deleteTask, chanSize),
		logger:   logger,
	}
}

// EnqueueDelete queues a delete task with the specified user ID and list of aliases into the worker's delete channel.
func (w *Worker) EnqueueDelete(userID uuid.UUID, aliases []string) {
	w.deleteCh <- deleteTask{userID: userID, aliases: aliases}
}

// Run starts the worker's loop to process delete tasks, flushing tasks periodically or when the flush limit is reached.
func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	var tasks []deleteTask

	for {
		select {
		case task := <-w.deleteCh:
			tasks = append(tasks, task)
			if len(w.deleteCh) >= flushLimit {
				for len(w.deleteCh) > 0 {
					tasks = append(tasks, <-w.deleteCh)
				}
				tasks = w.flush(ctx, tasks)
			}
		case <-ticker.C:
			tasks = w.flush(ctx, tasks)
		case <-ctx.Done():
			w.flush(ctx, tasks)
			return
		}
	}
}

// flush deletes URLs for the specified user IDs and aliases, returning the remaining tasks.
func (w *Worker) flush(ctx context.Context, tasks []deleteTask) []deleteTask {
	if len(tasks) == 0 {
		return tasks
	}

	userToAliases := make(map[uuid.UUID][]string)
	for _, t := range tasks {
		userToAliases[t.userID] = append(userToAliases[t.userID], t.aliases...)
	}

	for userID, aliases := range userToAliases {
		if err := w.deleter.DeleteUserURLs(ctx, userID, aliases); err != nil {
			w.logger.Errorf("failed to delete urls: %v", err)
		}
	}

	return tasks[:0]
}
