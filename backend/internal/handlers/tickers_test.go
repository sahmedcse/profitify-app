package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"profitify-backend/internal/models"
	"profitify-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockTickerService mocks the TickerService interface
type MockTickerService struct {
	mock.Mock
}

func (m *MockTickerService) GetTicker(ctx context.Context, symbol string) (*models.Ticker, error) {
	args := m.Called(ctx, symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Ticker), args.Error(1)
}

func (m *MockTickerService) GetActiveTickers(ctx context.Context) ([]models.Ticker, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Ticker), args.Error(1)
}

func TestHandler_GetAllTickers(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		mockSetup      func(*MockTickerService)
		expectedStatus int
		expectedBody   map[string]interface{}
		wantErr        bool
	}{
		{
			name: "successful retrieval with tickers",
			mockSetup: func(m *MockTickerService) {
				m.On("GetActiveTickers", mock.Anything).Return([]models.Ticker{
					{
						Ticker: "AAPL",
						Name:   "Apple Inc.",
						Active: 1,
					},
					{
						Ticker: "GOOGL",
						Name:   "Alphabet Inc.",
						Active: 1,
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"count": float64(2),
			},
			wantErr: false,
		},
		{
			name: "empty result",
			mockSetup: func(m *MockTickerService) {
				m.On("GetActiveTickers", mock.Anything).Return([]models.Ticker{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"count": float64(0),
			},
			wantErr: false,
		},
		{
			name: "general service error",
			mockSetup: func(m *MockTickerService) {
				m.On("GetActiveTickers", mock.Anything).Return(
					([]models.Ticker)(nil),
					errors.New("database connection error"),
				)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "Failed to retrieve tickers",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := new(MockTickerService)
			tt.mockSetup(mockService)

			// Create handler with mock
			handler := &Handler{
				ctx:           context.Background(),
				tickerService: mockService,
				log:           zap.NewNop().Sugar(),
			}

			// Create test HTTP request
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/api/tickers", nil)

			// Execute handler
			handler.GetAllTickers(c)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse response body
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			// Check expected fields
			for key, expectedValue := range tt.expectedBody {
				assert.Equal(t, expectedValue, response[key])
			}

			// If we expect tickers, verify they exist and have correct structure
			if !tt.wantErr && tt.expectedStatus == http.StatusOK {
				tickers, ok := response["tickers"]
				assert.True(t, ok, "response should contain 'tickers' field")
				assert.NotNil(t, tickers)
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_GetTicker(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		symbol         string
		mockSetup      func(*MockTickerService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:   "successful ticker retrieval",
			symbol: "AAPL",
			mockSetup: func(m *MockTickerService) {
				m.On("GetTicker", mock.Anything, "AAPL").Return(&models.Ticker{
					Ticker: "AAPL",
					Name:   "Apple Inc.",
					Active: 1,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"ticker": "AAPL",
				"name":   "Apple Inc.",
				"active": float64(1),
			},
		},
		{
			name:   "ticker not found",
			symbol: "INVALID",
			mockSetup: func(m *MockTickerService) {
				m.On("GetTicker", mock.Anything, "INVALID").Return(
					(*models.Ticker)(nil),
					service.ErrTickerNotFound,
				)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "Ticker not found",
			},
		},
		{
			name:   "invalid ticker symbol",
			symbol: "",
			mockSetup: func(m *MockTickerService) {
				m.On("GetTicker", mock.Anything, "").Return(
					(*models.Ticker)(nil),
					service.ErrInvalidTicker,
				)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Invalid ticker symbol",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockTickerService)
			tt.mockSetup(mockService)

			// Note: This test assumes you'll add a GetTicker handler method
			// For now, we're just setting up the test structure
			_ = &Handler{
				ctx:           context.Background(),
				tickerService: mockService,
				log:           zap.NewNop().Sugar(),
			}
			t.Skip("GetTicker handler not yet implemented")
		})
	}
}

// BenchmarkGetAllTickers benchmarks the GetAllTickers handler
func BenchmarkGetAllTickers(b *testing.B) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockTickerService)
	mockService.On("GetActiveTickers", mock.Anything).Return([]models.Ticker{
		{
			Ticker: "AAPL",
			Name:   "Apple Inc.",
			Active: 1,
		},
	}, nil)

	handler := &Handler{
		ctx:           context.Background(),
		tickerService: mockService,
		log:           zap.NewNop().Sugar(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/tickers", nil)
		handler.GetAllTickers(c)
	}
}
