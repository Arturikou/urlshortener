// Package audit provides an Audit observer that writes the event to the file
package audit

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"

	"go.uber.org/zap"
)

type FileObserver struct {
	file   *os.File
	writer *bufio.Writer
	mu     sync.Mutex
	logger *zap.SugaredLogger
}

// NewFileObserver creates a new File Audit observer
func NewFileObserver(
	fileName string,
	logger *zap.SugaredLogger,
) (*FileObserver, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	writer := bufio.NewWriter(file)
	return &FileObserver{
		file:   file,
		writer: writer,
		logger: logger,
	}, nil
}

// Notify writes the event to the file
func (p *FileObserver) Notify(event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		p.logger.Errorw("failed to marshal event", "error", err)
		return
	}

	data = append(data, '\n')

	p.mu.Lock()
	defer p.mu.Unlock()

	if _, err = p.writer.Write(data); err != nil {
		p.logger.Errorw("error writing to buffer", "error", err)
		return
	}

	if err = p.writer.Flush(); err != nil {
		p.logger.Errorw("error flushing buffer", "error", err)
	}
}
