package filestore

import (
	"encoding/json"
	"github.com/Arturikou/urlshortener/internal/models"
	"os"
)

type Consumer struct {
	file    *os.File
	decoder *json.Decoder
}

func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (c *Consumer) ReadEvent() (*models.URLData, error) {
	record := &models.URLData{}
	if err := c.decoder.Decode(record); err != nil {
		return nil, err
	}
	return record, nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}
