package filestore

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/Arturikou/urlshortener/internal/models"
)

type Producer struct {
	file    *os.File
	writer  *bufio.Writer
	encoder *json.Encoder
}

func NewProducer(fileName string) (*Producer, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	writer := bufio.NewWriter(file)
	return &Producer{
		file:    file,
		writer:  writer,
		encoder: json.NewEncoder(writer),
	}, nil
}

func (p *Producer) WriteEvent(record *models.URLData) error {
	if err := p.encoder.Encode(record); err != nil {
		return err
	}

	return p.writer.Flush()
}

func (p *Producer) WriteBatch(records []models.ShortenBatch) error {
	for _, rec := range records {
		event := models.URLData{
			OriginalURL: rec.OriginalURL,
			Alias:       rec.Alias,
		}
		if err := p.encoder.Encode(event); err != nil {
			return err
		}
	}

	return p.writer.Flush()
}

func (p *Producer) Close() error {
	return p.file.Close()
}
