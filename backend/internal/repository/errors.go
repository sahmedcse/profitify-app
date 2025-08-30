package repository

import "fmt"

// ErrTickerNotFound is returned when a ticker is not found in the repository
type ErrTickerNotFound struct {
	Symbol string
}

func (e ErrTickerNotFound) Error() string {
	return fmt.Sprintf("ticker not found: %s", e.Symbol)
}

// ErrInvalidTicker is returned when ticker data is invalid
type ErrInvalidTicker struct {
	Reason string
}

func (e ErrInvalidTicker) Error() string {
	return fmt.Sprintf("invalid ticker: %s", e.Reason)
}
