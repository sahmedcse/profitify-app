package models

import (
	"fmt"
)

// DailySummary represents daily aggregated stock data for a ticker
type DailySummary struct {
	Ticker           string  `json:"ticker" dynamodbav:"ticker"`
	Close            float32 `json:"close" dynamodbav:"close"`
	High             float32 `json:"high" dynamodbav:"high"`
	Low              float32 `json:"low" dynamodbav:"low"`
	Open             float32 `json:"open" dynamodbav:"open"`
	Volume           float32 `json:"volume" dynamodbav:"volume"`
	Timestamp        int64   `json:"timestamp" dynamodbav:"timestamp"`
	TransactionCount int32   `json:"transactionCount,omitempty" dynamodbav:"transactionCount,omitempty"`
	OTC              bool    `json:"otc,omitempty" dynamodbav:"otc,omitempty"`
	VWAP             float32 `json:"vwap,omitempty" dynamodbav:"vwap,omitempty"`
}

// Validate checks if the stock data is valid
func (d *DailySummary) Validate() error {
	if d.Ticker == "" {
		return fmt.Errorf("ticker is required")
	}

	if d.Timestamp <= 0 {
		return fmt.Errorf("timestamp must be positive")
	}

	if d.High < d.Low {
		return fmt.Errorf("high price cannot be less than low price")
	}

	if d.Open <= 0 || d.Close <= 0 || d.High <= 0 || d.Low <= 0 {
		return fmt.Errorf("prices must be positive")
	}

	if d.Volume < 0 {
		return fmt.Errorf("volume cannot be negative")
	}

	if d.VWAP != 0 && (d.VWAP < d.Low || d.VWAP > d.High) {
		return fmt.Errorf("VWAP must be between low and high prices")
	}

	return nil
}
