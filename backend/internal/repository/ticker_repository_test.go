package repository_test

import (
	"context"
	"testing"

	"profitify-backend/internal/models"
	"profitify-backend/internal/repository"
)


func TestMockTickerRepository_GetTicker(t *testing.T) {
	tests := []struct {
		name       string
		symbol     string
		setupMock  func(*repository.MockTickerRepository)
		wantTicker *models.Ticker
		wantErr    error
	}{
		{
			name:   "returns ticker when exists",
			symbol: "AAPL",
			setupMock: func(m *repository.MockTickerRepository) {
				m.SetTickers([]models.Ticker{
					{Ticker: "AAPL", Name: "Apple Inc.", Market: "NASDAQ"},
					{Ticker: "GOOGL", Name: "Alphabet Inc.", Market: "NASDAQ"},
				})
			},
			wantTicker: &models.Ticker{Ticker: "AAPL", Name: "Apple Inc.", Market: "NASDAQ"},
			wantErr:    nil,
		},
		{
			name:   "returns error when ticker not found",
			symbol: "INVALID",
			setupMock: func(m *repository.MockTickerRepository) {
				m.SetTickers([]models.Ticker{
					{Ticker: "AAPL", Name: "Apple Inc.", Market: "NASDAQ"},
				})
			},
			wantTicker: nil,
			wantErr:    repository.ErrTickerNotFound{Symbol: "INVALID"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock repository
			mockRepo := repository.NewMockTickerRepository()
			tt.setupMock(mockRepo)

			// Execute
			ctx := context.Background()
			ticker, err := mockRepo.GetTicker(ctx, tt.symbol)

			// Assert error
			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Errorf("GetTicker() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("GetTicker() unexpected error: %v", err)
			}

			// Assert ticker
			if tt.wantTicker != nil {
				if ticker == nil {
					t.Errorf("GetTicker() returned nil, want %+v", tt.wantTicker)
				} else if ticker.Ticker != tt.wantTicker.Ticker || ticker.Name != tt.wantTicker.Name {
					t.Errorf("GetTicker() = %+v, want %+v", ticker, tt.wantTicker)
				}
			}

			// Verify the method was called with correct parameters
			if len(mockRepo.Calls.GetTicker) != 1 {
				t.Errorf("GetTicker() was called %d times, expected 1", len(mockRepo.Calls.GetTicker))
			} else if mockRepo.Calls.GetTicker[0].Symbol != tt.symbol {
				t.Errorf("GetTicker() was called with symbol %s, expected %s",
					mockRepo.Calls.GetTicker[0].Symbol, tt.symbol)
			}
		})
	}
}

func TestMockTickerRepository_GetActiveTickers(t *testing.T) {
	tests := []struct {
		name            string
		setupMock       func(*repository.MockTickerRepository)
		wantActiveCount int
		wantTickers     []string // Expected ticker symbols
		wantErr         bool
		errMessage      string
	}{
		{
			name: "returns only active tickers",
			setupMock: func(m *repository.MockTickerRepository) {
				m.SetTickers([]models.Ticker{
					{Ticker: "AAPL", Name: "Apple Inc.", Active: 1},
					{Ticker: "GOOGL", Name: "Alphabet Inc.", Active: 1},
					{Ticker: "DEAD1", Name: "Delisted Company 1", Active: 0},
					{Ticker: "MSFT", Name: "Microsoft Corp.", Active: 1},
					{Ticker: "DEAD2", Name: "Delisted Company 2", Active: 0},
				})
			},
			wantActiveCount: 3,
			wantTickers:     []string{"AAPL", "GOOGL", "MSFT"},
			wantErr:         false,
		},
		{
			name: "returns empty list when no active tickers",
			setupMock: func(m *repository.MockTickerRepository) {
				m.SetTickers([]models.Ticker{
					{Ticker: "DEAD1", Name: "Delisted Company 1", Active: 0},
					{Ticker: "DEAD2", Name: "Delisted Company 2", Active: 0},
					{Ticker: "DEAD3", Name: "Delisted Company 3", Active: 0},
				})
			},
			wantActiveCount: 0,
			wantTickers:     []string{},
			wantErr:         false,
		},
		{
			name: "returns all tickers when all are active",
			setupMock: func(m *repository.MockTickerRepository) {
				m.SetTickers([]models.Ticker{
					{Ticker: "AAPL", Name: "Apple Inc.", Active: 1},
					{Ticker: "GOOGL", Name: "Alphabet Inc.", Active: 1},
					{Ticker: "MSFT", Name: "Microsoft Corp.", Active: 1},
					{Ticker: "AMZN", Name: "Amazon Inc.", Active: 1},
				})
			},
			wantActiveCount: 4,
			wantTickers:     []string{"AAPL", "GOOGL", "MSFT", "AMZN"},
			wantErr:         false,
		},
		{
			name: "returns empty list when repository is empty",
			setupMock: func(m *repository.MockTickerRepository) {
				// Don't set any tickers
			},
			wantActiveCount: 0,
			wantTickers:     []string{},
			wantErr:         false,
		},
		{
			name: "returns error during db errors",
			setupMock: func(m *repository.MockTickerRepository) {
				m.GetActiveTickersFunc = func(ctx context.Context) ([]models.Ticker, error) {
					return nil, repository.ErrInvalidTicker{Reason: "database connection failed"}
				}
			},
			wantActiveCount: 0,
			wantTickers:     nil,
			wantErr:         true,
			errMessage:      "invalid ticker: database connection failed",
		},
		{
			name: "handles mixed active values correctly",
			setupMock: func(m *repository.MockTickerRepository) {
				m.SetTickers([]models.Ticker{
					{Ticker: "AAPL", Name: "Apple Inc.", Active: 1},
					{Ticker: "GOOGL", Name: "Alphabet Inc.", Active: 1},
					{Ticker: "ZERO", Name: "Zero Active", Active: 0},
				})
			},
			wantActiveCount: 2,
			wantTickers:     []string{"AAPL", "GOOGL"},
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock repository
			mockRepo := repository.NewMockTickerRepository()
			tt.setupMock(mockRepo)

			// Execute
			ctx := context.Background()
			activeTickers, err := mockRepo.GetActiveTickers(ctx)

			// Assert error
			if (err != nil) != tt.wantErr {
				t.Errorf("GetActiveTickers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if tt.errMessage != "" && err.Error() != tt.errMessage {
					t.Errorf("GetActiveTickers() error message = %v, want %v", err.Error(), tt.errMessage)
				}
				return
			}

			// Assert count
			if len(activeTickers) != tt.wantActiveCount {
				t.Errorf("GetActiveTickers() returned %d tickers, want %d", len(activeTickers), tt.wantActiveCount)
			}

			// Assert all returned tickers are active
			for _, ticker := range activeTickers {
				if ticker.Active != 1 {
					t.Errorf("GetActiveTickers() returned inactive ticker: %s with Active=%d", ticker.Ticker, ticker.Active)
				}
			}

			// Assert expected ticker symbols are present
			if tt.wantTickers != nil {
				returnedSymbols := make(map[string]bool)
				for _, ticker := range activeTickers {
					returnedSymbols[ticker.Ticker] = true
				}

				for _, expectedSymbol := range tt.wantTickers {
					if !returnedSymbols[expectedSymbol] {
						t.Errorf("GetActiveTickers() missing expected ticker: %s", expectedSymbol)
					}
				}

				// Check for unexpected tickers
				for symbol := range returnedSymbols {
					found := false
					for _, expectedSymbol := range tt.wantTickers {
						if symbol == expectedSymbol {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("GetActiveTickers() returned unexpected ticker: %s", symbol)
					}
				}
			}

			// Verify the method was called
			if len(mockRepo.Calls.GetActiveTickers) != 1 {
				t.Errorf("GetActiveTickers() was called %d times, expected 1", len(mockRepo.Calls.GetActiveTickers))
			}
		})
	}
}

func TestMockTickerRepository_CallTracking(t *testing.T) {
	// Create mock repository
	mockRepo := repository.NewMockTickerRepository()
	ctx := context.Background()

	// Make several calls
	mockRepo.GetTicker(ctx, "AAPL")
	mockRepo.GetTicker(ctx, "GOOGL")
	mockRepo.GetActiveTickers(ctx)

	// Verify call counts
	if len(mockRepo.Calls.GetTicker) != 2 {
		t.Errorf("GetTicker called %d times, expected 2", len(mockRepo.Calls.GetTicker))
	}

	if len(mockRepo.Calls.GetActiveTickers) != 1 {
		t.Errorf("GetActiveTickers called %d times, expected 1", len(mockRepo.Calls.GetActiveTickers))
	}

	// Verify call parameters
	symbols := []string{
		mockRepo.Calls.GetTicker[0].Symbol,
		mockRepo.Calls.GetTicker[1].Symbol,
	}
	expectedSymbols := []string{"AAPL", "GOOGL"}
	for i, symbol := range symbols {
		if symbol != expectedSymbols[i] {
			t.Errorf("GetTicker call %d was for symbol %s, expected %s", i, symbol, expectedSymbols[i])
		}
	}

	// Reset and verify
	mockRepo.Reset()
}