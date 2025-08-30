package repository

import (
	"context"
	"fmt"
	"profitify-backend/internal/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// TickerRepository defines the interface for ticker data operations
type TickerRepository interface {
	GetTicker(ctx context.Context, symbol string) (*models.Ticker, error)
	GetActiveTickers(ctx context.Context) ([]models.Ticker, error)
}

// tickerRepository implements TickerRepository using DynamoDB
type tickerRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewTickerRepository creates a new DynamoDB-backed ticker repository
func NewTickerRepository(client *dynamodb.Client) TickerRepository {
	tableName := "stocks-data"
	return &tickerRepository{
		client:    client,
		tableName: tableName,
	}
}

// GetTicker retrieves a single ticker by symbol
func (r *tickerRepository) GetTicker(ctx context.Context, symbol string) (*models.Ticker, error) {
	// Build the key condition expression
	keyCond := expression.Key("ticker").Equal(expression.Value(symbol))

	// Build the expression
	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build expression: %w", err)
	}

	// Query the table
	result, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(r.tableName),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		Limit:                     aws.Int32(1),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to query ticker %s: %w", symbol, err)
	}

	if len(result.Items) == 0 {
		return nil, ErrTickerNotFound{Symbol: symbol}
	}

	var ticker models.Ticker
	err = attributevalue.UnmarshalMap(result.Items[0], &ticker)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal ticker: %w", err)
	}

	return &ticker, nil
}

// GetActiveTickers retrieves all active tickers
func (r *tickerRepository) GetActiveTickers(ctx context.Context) ([]models.Ticker, error) {
	// Build filter expression for active tickers
	filt := expression.Name("active").Equal(expression.Value(1))
	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build expression: %w", err)
	}

	var tickers []models.Ticker
	var lastEvaluatedKey map[string]types.AttributeValue

	for {
		input := &dynamodb.ScanInput{
			TableName:                 aws.String(r.tableName),
			FilterExpression:          expr.Filter(),
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
			Limit:                     aws.Int32(100),
		}

		if lastEvaluatedKey != nil {
			input.ExclusiveStartKey = lastEvaluatedKey
		}

		result, err := r.client.Scan(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to scan active tickers: %w", err)
		}

		var batch []models.Ticker
		err = attributevalue.UnmarshalListOfMaps(result.Items, &batch)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal tickers: %w", err)
		}

		tickers = append(tickers, batch...)

		if result.LastEvaluatedKey == nil {
			break
		}
		lastEvaluatedKey = result.LastEvaluatedKey
	}

	return tickers, nil
}
