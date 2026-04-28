// Package compress implements gzip compression for http.ResponseWriter.
package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

var gzipPool = sync.Pool{
	New: func() any {
		gz, _ := gzip.NewWriterLevel(io.Discard, gzip.DefaultCompression)
		return gz
	},
}

var compressibleTypes = []string{
	"application/json",
	"text/html",
}

type CompressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func NewCompressWriter(w http.ResponseWriter) *CompressWriter {
	return &CompressWriter{
		w: w,
	}
}

func (c *CompressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *CompressWriter) Write(p []byte) (int, error) {
	if c.zw == nil && c.isCompressible() {
		c.w.Header().Set("Content-Encoding", "gzip")
		c.w.Header().Del("Content-Length")
		gz := gzipPool.Get().(*gzip.Writer)
		gz.Reset(c.w)
		c.zw = gz
	}

	if c.zw != nil {
		return c.zw.Write(p)
	}

	return c.w.Write(p)
}

func (c *CompressWriter) WriteHeader(statusCode int) {
	if c.isCompressible() {
		c.w.Header().Set("Content-Encoding", "gzip")
		c.w.Header().Del("Content-Length")
	}
	c.w.WriteHeader(statusCode)
}

func (c *CompressWriter) Close() error {
	if c.zw != nil {
		err := c.zw.Close()
		gzipPool.Put(c.zw)
		c.zw = nil
		return err
	}
	return nil
}

func (c *CompressWriter) isCompressible() bool {
	contentType := c.w.Header().Get("Content-Type")
	for _, t := range compressibleTypes {
		if strings.Contains(contentType, t) {
			return true
		}
	}
	return false
}
