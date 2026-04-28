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

// --- Benchmarks ---

func newBenchShortener(b *testing.B) *Shortener {
	b.Helper()
	return New(&stubURLRepo{}, &stubUserURLRepo{}, &stubTransactor{}, &stubAuditManager{}, zap.NewNop().Sugar())
}

func BenchmarkGenerateAlias(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = generateAlias()
	}
}

func BenchmarkAddURL(b *testing.B) {
	s := newBenchShortener(b)
	ctx := context.WithValue(context.Background(), middleware.ContextKeyUserID, uuid.New())
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = s.AddURL(ctx, "https://practicum.yandex.com")
	}
}

func BenchmarkGetURL(b *testing.B) {
	s := newBenchShortener(b)
	ctx := context.WithValue(context.Background(), middleware.ContextKeyUserID, uuid.New())
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = s.GetURL(ctx, "EwHXdJfB")
	}
}

func BenchmarkAddURLs(b *testing.B) {
	svc := newBenchShortener(b)
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
	for b.Loop() {
		b.StopTimer()
		for j := range batches {
			batches[j].Alias = ""
		}
		b.StartTimer()
		_, _ = svc.AddURLs(ctx, batches)
	}
}
