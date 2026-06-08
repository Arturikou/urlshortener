package shortener

import (
	"context"
	"fmt"

	"github.com/Arturikou/urlshortener/internal/models"
	"golang.org/x/sync/errgroup"
)

func (s *Shortener) GetStats(ctx context.Context) (models.Stats, error) {
	g, gCtx := errgroup.WithContext(ctx)

	var usersCount int64
	var urlsCount int64

	g.Go(func() error {
		var err error
		usersCount, err = s.userURLRepo.CountUsers(gCtx)
		if err != nil {
			return fmt.Errorf("can't get users count: %w", err)
		}

		return nil
	})

	g.Go(func() error {
		var err error
		urlsCount, err = s.urlRepo.CountUrls(gCtx)
		if err != nil {
			return fmt.Errorf("can't get urls count: %w", err)
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		s.logger.Errorf("get stats failed: %v", err)
		return models.Stats{}, err
	}

	return models.Stats{
		URLsCount: urlsCount,
		UserCount: usersCount,
	}, nil
}
