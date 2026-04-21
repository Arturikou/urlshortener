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
	file    *os.File
	writer  *bufio.Writer
	encoder *json.Encoder
	mu      sync.Mutex
	logger  *zap.SugaredLogger
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
		file:    file,
		writer:  writer,
		encoder: json.NewEncoder(writer),
		logger:  logger,
	}, nil
}

// Notify writes the event to the file
func (p *FileObserver) Notify(event Event) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.encoder.Encode(event); err != nil {
		p.logger.Errorw("error encoding event", "error", err)
	}

	if err := p.writer.Flush(); err != nil {
		p.logger.Errorw("error flushing writer", "error", err)
	}
}
