package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DailyAggStockItem struct {
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

type Ticker struct {
    Ticker string `json:"ticker" dynamodbav:"ticker"`
    Name string `json:"name" dynamodbav:"name"`
	Market string `json:"market" dynamodbav:"market"`
	Locale string `json:"locale" dynamodbav:"locale"`
	PrimaryExchange string `json:"primaryExchange,omitempty" dynamodbav:"primaryExchange,omitempty"`
	ShareClassFIGI string `json:"shareClassFIGI,omitempty" dynamodbav:"shareClassFIGI,omitempty"`
	Type string `json:"type,omitempty" dynamodbav:"type,omitempty"`
	Active int32 `json:"active,omitempty" dynamodbav:"active,omitempty"`
	Cik string `json:"cik,omitempty" dynamodbav:"cik,omitempty"`
	CompositeFigi string `json:"compositeFigi,omitempty" dynamodbav:"compositeFigi,omitempty"`
	CurrencyName string `json:"currencyName,omitempty" dynamodbav:"currencyName,omitempty"`
	DelistedUTC int64 `json:"delistedUTC,omitempty" dynamodbav:"delistedUTC,omitempty"`
	LastUpdatedUTC int64 `json:"lastUpdatedUTC,omitempty" dynamodbav:"lastUpdatedUTC,omitempty"`
}


var tickers = []string{
	"AAPL", "GOOGL", "MSFT", "TSLA", "AMZN", "NVDA", "META",
	"JPM", "JNJ", "PG", "UNH", "HD", "BAC", "PFE", "ABBV",
}

func main() {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("Unable to load SDK config: %v", err)
	}

	client := dynamodb.NewFromConfig(cfg)

	if err := ensureTableExists(client, "stocks-data"); err != nil {
		log.Fatalf("Failed to ensure table exists: %v", err)
	}

	if err = ensureTableExists(client, "tickers"); err != nil {
		log.Fatalf("Failed to ensure table exists: %v", err)
	}

	if err := populateStocksDataTable(client); err != nil {
		log.Fatalf("Failed to populate table: %v", err)
	}

	if err := populateTickerTable(client); err != nil {
		log.Fatalf("Failed to populate table: %v", err)
	}

	log.Println("Successfully populated stocks-data table with sample data!")
}

func ensureTableExists(client *dynamodb.Client, tableName string) error {
	_, err := client.DescribeTable(context.TODO(), &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})

	if err == nil {
		log.Printf("Table %s already exists", tableName)
		return nil
	}

	log.Printf("Creating table %s...", tableName)

	tableAttributeDefinitions := getTableAttributeDefinitions(tableName)
	tableKeySchema := getTableKeySchema(tableName)


	_, err = client.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: tableAttributeDefinitions,
		KeySchema: tableKeySchema,
		BillingMode: types.BillingModePayPerRequest,
	})

	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	waiter := dynamodb.NewTableExistsWaiter(client)
	err = waiter.Wait(context.TODO(), &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	}, 30*time.Second)

	if err != nil {
		return fmt.Errorf("table creation timeout: %v", err)
	}

	log.Printf("Table %s created successfully", tableName)
	return nil
}

func getTableAttributeDefinitions(tableName string) []types.AttributeDefinition {
	switch tableName {
		case "tickers":
			return []types.AttributeDefinition{
				{AttributeName: aws.String("ticker"), AttributeType: types.ScalarAttributeTypeS},
			}
		case "stocks-data":
			return []types.AttributeDefinition{
				{AttributeName: aws.String("ticker"), AttributeType: types.ScalarAttributeTypeS},
				{AttributeName: aws.String("timestamp"), AttributeType: types.ScalarAttributeTypeN},
			}
		default:
			return []types.AttributeDefinition{}
	}
}

func getTableKeySchema(tableName string) []types.KeySchemaElement {
	switch tableName {
		case "tickers":
			return []types.KeySchemaElement{
				{AttributeName: aws.String("ticker"), KeyType: types.KeyTypeHash},
			}
		case "stocks-data":
			return []types.KeySchemaElement{
				{AttributeName: aws.String("ticker"), KeyType: types.KeyTypeHash},
				{AttributeName: aws.String("timestamp"), KeyType: types.KeyTypeRange},
			}
		default:
			return []types.KeySchemaElement{}
	}
}

func populateTickerTable(client *dynamodb.Client) error {
	for _, ticker := range tickers {
		item := Ticker{
			Ticker: ticker,
			Name: ticker,
			Market: "NASDAQ",
			Locale: "US",
			PrimaryExchange: "NASDAQ",
			ShareClassFIGI: "BBG000B9XRY4",
			Type: "CS",
			Active: 1,
			Cik: "0001326800",
			CompositeFigi: "BBG000B9XRY4",
			CurrencyName: "USD",
			DelistedUTC: 0,
			LastUpdatedUTC: time.Now().Unix(),
		}
		av, err := attributevalue.MarshalMap(item)
		if err != nil {
			return fmt.Errorf("failed to marshal item: %v", err)
		}
		_, err = client.PutItem(context.TODO(), &dynamodb.PutItemInput{
			TableName: aws.String("tickers"),
			Item: av,
		})
		if err != nil {
			return fmt.Errorf("failed to put item: %v", err)
		}
	}
	return nil
}


func populateStocksDataTable(client *dynamodb.Client) error {
	// Set random seed for reproducible data
	r := rand.New(rand.NewSource(42))
	batchSize := 25
	numWorkers := 10
	totalDays := 5 * 365 * 5 / 7 // Approximately 1300 trading days

	// Start date: 5 years ago
	startDate := time.Now().AddDate(-5, 0, 0)

	// Performance tracking
	startTime := time.Now()
	totalItems := 0
	var mu sync.Mutex

	// Channel to collect all items
	itemChan := make(chan DailyAggStockItem, batchSize*numWorkers)

	// Wait group to track all goroutines
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			worker(client, itemChan, workerID, &mu, &totalItems, batchSize)
		}(i)
	}

	// Generate data for each ticker
	for _, ticker := range tickers {
		log.Printf("Generating data for ticker: %s", ticker)

		// Generate base price for this ticker (different starting prices for variety)
		basePrice := getBasePrice(ticker)
		currentPrice := basePrice

		// Generate daily data
		dayOffset := 0
		for i := 0; i < totalDays; i++ {
			// Calculate date (skip weekends)
			date := startDate.AddDate(0, 0, dayOffset)
			for date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
				dayOffset++
				date = startDate.AddDate(0, 0, dayOffset)
			}

			// Generate realistic stock data
			item := generateDailyAggData(ticker, date, currentPrice, r)

			// Send item to workers via channel
			itemChan <- item

			// Update current price for next day
			currentPrice = item.Close

			// Move to next day
			dayOffset++
		}
	}

	// Close the channel to signal workers to finish
	close(itemChan)

	// Wait for all workers to complete
	log.Println("Waiting for all workers to complete...")
	wg.Wait()

	// Performance summary
	duration := time.Since(startTime)
	rate := float64(totalItems) / duration.Seconds()
	log.Printf("✅ Performance Summary:")
	log.Printf("   Total items processed: %d", totalItems)
	log.Printf("   Total time: %v", duration)
	log.Printf("   Processing rate: %.2f items/second", rate)
	log.Printf("   Workers used: %d", numWorkers)
	log.Printf("   Batch size: %d", batchSize)

	return nil
}

func worker(client *dynamodb.Client, itemChan <-chan DailyAggStockItem, workerID int, mu *sync.Mutex, totalItems *int, batchSize int) {
	var batch []DailyAggStockItem
	batchCount := 0
	workerItems := 0

	for item := range itemChan {
		batch = append(batch, item)
		workerItems++

		// When batch is full, process it
		if len(batch) >= batchSize {
			if err := processBatch(client, batch, workerID, batchCount, "stocks-data"); err != nil {
				log.Printf("Worker %d: Error processing batch %d: %v", workerID, batchCount, err)
			} else {
				// Update total items count
				mu.Lock()
				*totalItems += len(batch)
				mu.Unlock()
			}
			batch = batch[:0] // Reset batch
			batchCount++
		}
	}

	// Process remaining items in the last batch
	if len(batch) > 0 {
		if err := processBatch(client, batch, workerID, batchCount, "stocks-data"); err != nil {
			log.Printf("Worker %d: Error processing final batch %d: %v", workerID, batchCount, err)
		} else {
			// Update total items count
			mu.Lock()
			*totalItems += len(batch)
			mu.Unlock()
		}
	}

	log.Printf("Worker %d: Completed processing %d items in %d batches", workerID, workerItems, batchCount+1)
}

func processBatch(client *dynamodb.Client, batch []DailyAggStockItem, workerID, batchCount int, tableName string) error {
	// Convert items to DynamoDB attribute values
	writeRequests := make([]types.WriteRequest, len(batch))

	for i, item := range batch {
		av, err := attributevalue.MarshalMap(item)
		if err != nil {
			return fmt.Errorf("failed to marshal item: %v", err)
		}

		writeRequests[i] = types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: av,
			},
		}
	}

	// Process batch with retry logic
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		_, err := client.BatchWriteItem(context.TODO(), &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				tableName: writeRequests,
			},
		})

		if err == nil {
			log.Printf("Worker %d: Successfully processed batch %d (%d items)", workerID, batchCount, len(batch))
			return nil
		}

		// If it's the last attempt, return the error
		if attempt == maxRetries-1 {
			return fmt.Errorf("failed to process batch after %d attempts: %v", maxRetries, err)
		}

		// Wait before retry with exponential backoff
		waitTime := time.Duration(1<<attempt) * time.Second
		log.Printf("Worker %d: Batch %d failed, retrying in %v (attempt %d/%d)",
			workerID, batchCount, waitTime, attempt+1, maxRetries)
		time.Sleep(waitTime)
	}

	return nil
}

func getBasePrice(ticker string) float32 {
	// Different starting prices for different tickers
	prices := map[string]float32{
		// Tech Giants
		"AAPL":  175.0, "GOOGL": 140.0, "MSFT": 380.0, "TSLA": 240.0,
		"AMZN":  155.0, "NVDA": 850.0, "META": 485.0,

		// Financial & Healthcare
		"JPM": 195.0, "JNJ": 155.0, "UNH": 520.0, "PFE": 28.0,
		"ABBV": 165.0, "BAC": 37.0, "HD": 380.0, "PG": 155.0,
	}

	if price, exists := prices[ticker]; exists {
		return price
	}

	// Default price for any ticker not in the map
	return 100.0
}

func generateDailyAggData(ticker string, date time.Time, previousClose float32, r *rand.Rand) DailyAggStockItem {
	// Generate realistic price movements
	// Daily change: -5% to +5%
	dailyChange := (r.Float32()-0.5) * 0.1

	// Add some trend and volatility
	trend := r.Float32() * 0.02 // Small trend component
	volatility := r.Float32() * 0.03 // Volatility component

	// Calculate new close price
	changePercent := dailyChange + trend + volatility
	newClose := previousClose * (1 + changePercent)

	// Ensure price doesn't go negative
	if newClose < 1.0 {
		newClose = 1.0
	}

	// Generate open, high, low based on close
	open := previousClose * (1 + (r.Float32()-0.5)*0.02)
	high := max(open, newClose) * (1 + r.Float32()*0.03)
	low := min(open, newClose) * (1 - r.Float32()*0.03)

	// Generate volume (higher for more volatile days)
	volumeMultiplier := 1.0 + abs(changePercent)*10
	baseVolume := float32(1000000 + r.Intn(9000000)) // 1M to 10M base volume
	volume := baseVolume * volumeMultiplier

	// Generate VWAP (Volume Weighted Average Price)
	vwap := (open + high + low + newClose) / 4

	// Generate transaction count
	transactionCount := int32(10000 + r.Intn(90000)) // 10K to 100K transactions

	// Determine if OTC (Over The Counter) - rare
	otc := r.Float32() < 0.01 // 1% chance

	return DailyAggStockItem{
		Ticker:           ticker,
		Close:            newClose,
		High:             high,
		Low:              low,
		Open:             open,
		Volume:           volume,
		Timestamp:        date.Unix(),
		TransactionCount: transactionCount,
		OTC:              otc,
		VWAP:             vwap,
	}
}



func max(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}
