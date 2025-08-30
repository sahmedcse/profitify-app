package handlers

import (
	"context"
	"fmt"
	"net/http"

	"profitify-backend/internal/repository"
	"profitify-backend/internal/service"
	"profitify-backend/pkg/logger"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	ctx           context.Context
	tickerService service.TickerService
	log           *zap.SugaredLogger
}

func NewHandler(ctx context.Context) (*Handler, error) {
	log := logger.Get()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	db := dynamodb.NewFromConfig(cfg)

	// Create repository and service
	tickerRepo := repository.NewTickerRepository(db)
	tickerService := service.NewTickerService(tickerRepo, log)

	return &Handler{
		ctx:           ctx,
		tickerService: tickerService,
		log:           log,
	}, nil
}

func (h *Handler) GetAllTickers(c *gin.Context) {
	h.log.Info("Getting all tickers")

	tickers, err := h.tickerService.GetActiveTickers(c.Request.Context())

	if err != nil {
		h.log.Errorw("failed to get tickers", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve tickers",
		})
		return
	}

	h.log.Infow("retrieved tickers", "count", len(tickers))

	c.JSON(http.StatusOK, gin.H{
		"tickers": tickers,
		"count":   len(tickers),
	})
}
