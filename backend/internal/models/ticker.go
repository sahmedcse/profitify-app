package models

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