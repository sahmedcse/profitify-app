package models

import (
	"fmt"
)

// Ticker represents a stock ticker entity
type Ticker struct {
	Ticker          string `json:"ticker" dynamodbav:"ticker"`
	Name            string `json:"name" dynamodbav:"name"`
	Market          string `json:"market" dynamodbav:"market"`
	Locale          string `json:"locale" dynamodbav:"locale"`
	PrimaryExchange string `json:"primaryExchange,omitempty" dynamodbav:"primaryExchange,omitempty"`
	ShareClassFigi  string `json:"shareClassFigi,omitempty" dynamodbav:"shareClassFigi,omitempty"`
	Type            string `json:"type,omitempty" dynamodbav:"type,omitempty"`
	Active          int32  `json:"active,omitempty" dynamodbav:"active,omitempty"`
	Cik             string `json:"cik,omitempty" dynamodbav:"cik,omitempty"`
	CompositeFigi   string `json:"compositeFigi,omitempty" dynamodbav:"compositeFigi,omitempty"`
	Currency        string `json:"currency,omitempty" dynamodbav:"currency,omitempty"`
	DelistedUTC     int64  `json:"delistedUTC,omitempty" dynamodbav:"delistedUTC,omitempty"`
	LastUpdatedUTC  int64  `json:"lastUpdatedUTC,omitempty" dynamodbav:"lastUpdatedUTC,omitempty"`
}

// Validate checks if the ticker data is valid
func (t *Ticker) Validate() error {
	// Required fields
	if t.Ticker == "" {
		return fmt.Errorf("ticker symbol is required")
	}

	if t.Name == "" {
		return fmt.Errorf("ticker name is required")
	}

	if t.Market == "" {
		return fmt.Errorf("market is required")
	}

	if t.Locale == "" {
		return fmt.Errorf("locale is required")
	}

	// Validate active status (should be 0 or 1)
	if t.Active != 0 && t.Active != 1 {
		return fmt.Errorf("active status must be 0 or 1, got: %d", t.Active)
	}

	// Validate timestamps
	if t.LastUpdatedUTC < 0 {
		return fmt.Errorf("lastUpdatedUTC cannot be negative")
	}

	if t.DelistedUTC < 0 {
		return fmt.Errorf("delistedUTC cannot be negative")
	}

	return nil
}
