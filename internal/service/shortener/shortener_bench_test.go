package shortener

import (
	"context"
	"fmt"
	"testing"

	"github.com/Arturikou/urlshortener/internal/middleware"
	"github.com/Arturikou/urlshortener/internal/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func newBenchShortener() *Shortener {
	return New(&stubURLRepo{}, &stubUserURLRepo{}, &stubTransactor{}, &stubAuditManager{}, zap.NewNop().Sugar())
}

func BenchmarkGenerateAlias(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = generateAlias()
	}
}

func BenchmarkAddURL(b *testing.B) {
	s := newBenchShortener()
	ctx := context.WithValue(context.Background(), middleware.ContextKeyUserID, uuid.New())
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.AddURL(ctx, fmt.Sprintf("https://practicum.yandex.com/%d", i))
	}
}

func BenchmarkGetURL(b *testing.B) {
	s := newBenchShortener()
	ctx := context.WithValue(context.Background(), middleware.ContextKeyUserID, uuid.New())
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.GetURL(ctx, "EwHXdJfB")
	}
}

func BenchmarkAddURLs(b *testing.B) {
	svc := newBenchShortener()
	ctx := context.WithValue(context.Background(), middleware.ContextKeyUserID, uuid.New())

	const batchSize = 100
	batches := make([]models.ShortenBatch, batchSize)
	for j := range batches {
		batches[j] = models.ShortenBatch{
			CorrelationID: fmt.Sprintf("id-%d", j),
			OriginalURL:   fmt.Sprintf("https://practicum.yandex.com/%d", j),
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := range batches {
			batches[j].Alias = ""
		}
		_, _ = svc.AddURLs(ctx, batches)
	}
}
