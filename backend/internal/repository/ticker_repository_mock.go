package repository

import (
	"context"
	"profitify-backend/internal/models"
	"sync"
)

// MockTickerRepository is a mock implementation of TickerRepository for testing
type MockTickerRepository struct {
	mu      sync.RWMutex
	tickers map[string]*models.Ticker

	// Function fields for custom behavior in tests
	GetTickerFunc        func(ctx context.Context, symbol string) (*models.Ticker, error)
	GetActiveTickersFunc func(ctx context.Context) ([]models.Ticker, error)

	// Call tracking
	Calls struct {
		GetTicker []struct {
			Ctx    context.Context
			Symbol string
		}
		GetActiveTickers []context.Context
	}
}

// NewMockTickerRepository creates a new mock repository with default implementations
func NewMockTickerRepository() *MockTickerRepository {
	return &MockTickerRepository{
		tickers: make(map[string]*models.Ticker),
	}
}

// GetTicker mock implementation
func (m *MockTickerRepository) GetTicker(ctx context.Context, symbol string) (*models.Ticker, error) {
	m.mu.Lock()
	m.Calls.GetTicker = append(m.Calls.GetTicker, struct {
		Ctx    context.Context
		Symbol string
	}{ctx, symbol})
	m.mu.Unlock()

	if m.GetTickerFunc != nil {
		return m.GetTickerFunc(ctx, symbol)
	}

	// Default implementation
	m.mu.RLock()
	defer m.mu.RUnlock()

	ticker, exists := m.tickers[symbol]
	if !exists {
		return nil, ErrTickerNotFound{Symbol: symbol}
	}
	return ticker, nil
}

// GetActiveTickers mock implementation
func (m *MockTickerRepository) GetActiveTickers(ctx context.Context) ([]models.Ticker, error) {
	m.mu.Lock()
	m.Calls.GetActiveTickers = append(m.Calls.GetActiveTickers, ctx)
	m.mu.Unlock()

	if m.GetActiveTickersFunc != nil {
		return m.GetActiveTickersFunc(ctx)
	}

	// Default implementation
	m.mu.RLock()
	defer m.mu.RUnlock()

	var tickers []models.Ticker
	for _, ticker := range m.tickers {
		if ticker.Active == 1 {
			tickers = append(tickers, *ticker)
		}
	}
	return tickers, nil
}

// Reset clears all calls and data
func (m *MockTickerRepository) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tickers = make(map[string]*models.Ticker)
	m.Calls.GetTicker = nil
	m.Calls.GetActiveTickers = nil
}

// SetTickers sets the initial tickers for testing
func (m *MockTickerRepository) SetTickers(tickers []models.Ticker) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tickers = make(map[string]*models.Ticker)
	for i := range tickers {
		m.tickers[tickers[i].Ticker] = &tickers[i]
	}
}
