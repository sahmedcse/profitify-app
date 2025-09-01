package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"profitify-backend/internal/models"
)

// Worker pool configuration
const (
	maxWorkers = 10
	batchSize  = 25
)

type seedJob struct {
	client    *dynamodb.Client
	tableName string
	items     []interface{}
}

func main() {
	ctx := context.Background()

	// Configure AWS SDK with LocalStack endpoint
	endpointURL := os.Getenv("AWS_ENDPOINT_URL")
	if endpointURL == "" {
		endpointURL = "http://localhost:4566"
	}

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
	)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	// Create DynamoDB client with custom endpoint for LocalStack
	client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String(endpointURL)
	})

	// Create tables if they don't exist
	tickersTable := "Tickers"
	stockDataTable := "DailySummary"

	if err := createTickersTable(ctx, client, tickersTable); err != nil {
		log.Fatalf("Failed to create Tickers table: %v", err)
	}

	if err := createDailySummaryTable(ctx, client, stockDataTable); err != nil {
		log.Fatalf("Failed to create DailySummary table: %v", err)
	}

	// Wait for tables to be active
	time.Sleep(2 * time.Second)

	// Seed sample data
	sampleTickers := getSampleTickers()

	fmt.Printf("Seeding %d tickers into DynamoDB...\n", len(sampleTickers))

	// Create worker pool for concurrent processing
	var wg sync.WaitGroup
	jobChan := make(chan seedJob, 100)

	// Start workers
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go worker(ctx, &wg, jobChan)
	}

	// Seed tickers first
	tickerItems := make([]interface{}, 0, len(sampleTickers))
	for _, ticker := range sampleTickers {
		tickerItems = append(tickerItems, ticker)
	}

	// Send ticker batch job
	jobChan <- seedJob{
		client:    client,
		tableName: tickersTable,
		items:     tickerItems,
	}

	// Generate and seed 2 years of daily summary data for each ticker
	fmt.Println("\nGenerating 2 years of daily summary data for each ticker...")

	endDate := time.Now()
	startDate := endDate.AddDate(-2, 0, 0)

	// Process each ticker's daily summary data
	for _, ticker := range sampleTickers {
		stockData := generateDailySummaryData(ticker.Ticker, startDate, endDate)

		// Batch the daily summary data
		for i := 0; i < len(stockData); i += batchSize {
			end := i + batchSize
			if end > len(stockData) {
				end = len(stockData)
			}

			batchItems := make([]interface{}, 0, end-i)
			for j := i; j < end; j++ {
				batchItems = append(batchItems, stockData[j])
			}

			jobChan <- seedJob{
				client:    client,
				tableName: stockDataTable,
				items:     batchItems,
			}
		}
	}

	// Close job channel and wait for workers to finish
	close(jobChan)
	wg.Wait()

	fmt.Println("\nSeed data loaded successfully!")
}

func worker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan seedJob) {
	defer wg.Done()

	for job := range jobs {
		processBatch(ctx, job)
	}
}

func processBatch(ctx context.Context, job seedJob) {
	for _, item := range job.items {
		marshaledItem, err := attributevalue.MarshalMap(item)
		if err != nil {
			log.Printf("Failed to marshal item: %v", err)
			continue
		}

		_, err = job.client.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(job.tableName),
			Item:      marshaledItem,
		})
		if err != nil {
			log.Printf("Failed to insert item into %s: %v", job.tableName, err)
			continue
		}
	}

	// Log progress
	if _, ok := job.items[0].(models.Ticker); ok {
		fmt.Printf("✓ Inserted %d tickers\n", len(job.items))
	} else if stock, ok := job.items[0].(models.DailySummary); ok {
		fmt.Printf("✓ Inserted %d daily summary records for %s\n", len(job.items), stock.Ticker)
	}
}

func createTickersTable(ctx context.Context, client *dynamodb.Client, tableName string) error {
	// Delete table if it exists
	fmt.Printf("Deleting table %s if it exists...\n", tableName)
	_, _ = client.DeleteTable(ctx, &dynamodb.DeleteTableInput{
		TableName: aws.String(tableName),
	})


	// Create table
	fmt.Printf("Creating table %s...\n", tableName)
	_, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("ticker"),
				KeyType:       types.KeyTypeHash,
			},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("ticker"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	})

	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	fmt.Printf("Table %s created successfully\n", tableName)
	return nil
}

func createDailySummaryTable(ctx context.Context, client *dynamodb.Client, tableName string) error {
	// Delete table if it exists
	fmt.Printf("Deleting table %s if it exists...\n", tableName)
	_, err := client.DeleteTable(ctx, &dynamodb.DeleteTableInput{
		TableName: aws.String(tableName),
	})
	if err == nil {
		fmt.Printf("Deleted existing table %s\n", tableName)
		// Wait for table to be deleted
		time.Sleep(2 * time.Second)
	}

	// Create table with composite key (ticker as partition key, timestamp as sort key)
	fmt.Printf("Creating table %s...\n", tableName)
	_, err = client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("ticker"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("timestamp"),
				KeyType:       types.KeyTypeRange,
			},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("ticker"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("timestamp"),
				AttributeType: types.ScalarAttributeTypeN,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	})

	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	fmt.Printf("Table %s created successfully\n", tableName)
	return nil
}

func generateDailySummaryData(ticker string, startDate, endDate time.Time) []models.DailySummary {
	// Set initial price based on ticker (for realistic ranges)
	initialPrices := map[string]float32{
		"AAPL":  150.0,
		"GOOGL": 100.0,
		"MSFT":  250.0,
		"AMZN":  120.0,
		"TSLA":  200.0,
		"META":  300.0,
		"NVDA":  400.0,
		"JPM":   140.0,
		"V":     220.0,
		"WMT":   150.0,
		"DIS":   100.0,
		"NFLX":  350.0,
		"BA":    200.0,
		"KO":    60.0,
		"PFE":   40.0,
	}

	basePrice := initialPrices[ticker]
	if basePrice == 0 {
		basePrice = 100.0
	}

	var dailySummaryData []models.DailySummary
	currentPrice := basePrice

	// Generate data for each trading day (excluding weekends)
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		// Skip weekends
		if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
			continue
		}

		// Generate realistic price movement (±5% daily change)
		changePercent := (rand.Float32() - 0.5) * 0.1
		currentPrice = currentPrice * (1 + changePercent)

		// Generate OHLC data
		open := currentPrice * (1 + (rand.Float32()-0.5)*0.02)
		close := currentPrice

		// Ensure high is highest and low is lowest
		dayRange := currentPrice * 0.03
		high := math.Max(float64(open), float64(close)) + float64(rand.Float32()*dayRange)
		low := math.Min(float64(open), float64(close)) - float64(rand.Float32()*dayRange)

		// Generate volume (between 10M and 100M shares)
		volume := 10000000 + rand.Float32()*90000000

		// Calculate VWAP (simplified - between low and high)
		vwap := float32(low) + rand.Float32()*float32(high-low)

		stockItem := models.DailySummary{
			Ticker:           ticker,
			Open:             open,
			High:             float32(high),
			Low:              float32(low),
			Close:            close,
			Volume:           volume,
			Timestamp:        d.Unix(),
			TransactionCount: int32(volume / 1000),
			OTC:              false,
			VWAP:             vwap,
		}

		dailySummaryData = append(dailySummaryData, stockItem)
	}

	return dailySummaryData
}

func getSampleTickers() []models.Ticker {
	now := time.Now().Unix()

	// Common fields for all tickers
	commonFields := models.Ticker{
		Market:         "stocks",
		Locale:         "us",
		Type:           "CS",
		Active:         1,
		Currency:       "USD",
		LastUpdatedUTC: now,
	}

	// Ticker-specific data
	tickerData := []struct {
		Symbol   string
		Name     string
		Exchange string
		Cik      string
	}{
		{"AAPL", "Apple Inc.", "XNAS", "0000320193"},
		{"GOOGL", "Alphabet Inc. Class A", "XNAS", "0001652044"},
		{"MSFT", "Microsoft Corporation", "XNAS", "0000789019"},
		{"AMZN", "Amazon.com Inc.", "XNAS", "0001018724"},
		{"TSLA", "Tesla Inc.", "XNAS", "0001318605"},
		{"META", "Meta Platforms Inc.", "XNAS", "0001326801"},
		{"NVDA", "NVIDIA Corporation", "XNAS", "0001045810"},
		{"JPM", "JPMorgan Chase & Co.", "XNYS", "0000019617"},
		{"V", "Visa Inc.", "XNYS", "0001403161"},
		{"WMT", "Walmart Inc.", "XNYS", "0000104169"},
		{"DIS", "The Walt Disney Company", "XNYS", "0001744489"},
		{"NFLX", "Netflix Inc.", "XNAS", "0001065280"},
		{"BA", "The Boeing Company", "XNYS", "0000012927"},
		{"KO", "The Coca-Cola Company", "XNYS", "0000021344"},
		{"PFE", "Pfizer Inc.", "XNYS", "0000078003"},
	}

	tickers := make([]models.Ticker, len(tickerData))
	for i, data := range tickerData {
		tickers[i] = commonFields
		tickers[i].Ticker = data.Symbol
		tickers[i].Name = data.Name
		tickers[i].PrimaryExchange = data.Exchange
		tickers[i].Cik = data.Cik
	}

	return tickers
}
