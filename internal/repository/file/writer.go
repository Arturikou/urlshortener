package file

import (
	"encoding/json"
	"github.com/Arturikou/urlshortener/internal/models"
	"os"
)

type Producer struct {
	file    *os.File
	encoder *json.Encoder
}

func NewProducer(fileName string) (*Producer, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (p *Producer) WriteEvent(record *models.URLData) error {
	return p.encoder.Encode(record)
}

func (p *Producer) Close() error {
	return p.file.Close()
}
