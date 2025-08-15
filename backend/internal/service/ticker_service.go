package service

import (
	"context"
	"errors"
	"fmt"
	"profitify-backend/internal/models"
	"profitify-backend/internal/repository"

	"go.uber.org/zap"
)

var (
	ErrTickerNotFound    = errors.New("ticker not found")
	ErrInvalidTicker     = errors.New("invalid ticker symbol")
)

type TickerService interface {
	GetTicker(ctx context.Context, symbol string) (*models.Ticker, error)
	GetActiveTickers(ctx context.Context) ([]models.Ticker, error)
}

type tickerService struct {
	repo repository.TickerRepository
	log  *zap.SugaredLogger
}

func NewTickerService(repo repository.TickerRepository, log *zap.SugaredLogger) TickerService {
	return &tickerService{
		repo: repo,
		log:  log,
	}
}

func (s *tickerService) GetTicker(ctx context.Context, symbol string) (*models.Ticker, error) {
	if symbol == "" {
		return nil, ErrInvalidTicker
	}

	s.log.Debugw("fetching ticker", "symbol", symbol)

	ticker, err := s.repo.GetTicker(ctx, symbol)
	if err != nil {
		if errors.Is(err, repository.ErrTickerNotFound{Symbol: symbol}) {
			return nil, ErrTickerNotFound
		}
		s.log.Errorw("failed to get ticker", "symbol", symbol, "error", err)
		return nil, fmt.Errorf("failed to get ticker: %w", err)
	}

	return ticker, nil
}

func (s *tickerService) GetActiveTickers(ctx context.Context) ([]models.Ticker, error) {
	s.log.Debug("fetching active tickers")

	tickers, err := s.repo.GetActiveTickers(ctx)
	if err != nil {
		s.log.Errorw("failed to get active tickers", "error", err)
		return nil, fmt.Errorf("failed to get active tickers: %w", err)
	}

	activeCount := 0
	for _, t := range tickers {
		if t.Active == 1 {
			activeCount++
		}
	}

	s.log.Debugw("fetched active tickers", "total", len(tickers), "active", activeCount)
	return tickers, nil
}