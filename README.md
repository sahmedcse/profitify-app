# Profitify App

A full-stack application for stock data analysis and management, featuring a Go backend, React frontend, and DynamoDB for data storage.

## 🏗️ Architecture

- **Backend**: Go with Gin framework
- **Frontend**: React with TypeScript
- **Database**: DynamoDB

## 🚀 Quick Start

### Prerequisites

- Docker and Docker Compose installed
- Ports 3000, 8080, and 4566 available

### Service Management

```bash
# Start all services
docker-compose up -d

# Run DB init
docker-compose --profile tools run --rm dynamodb-init

# Stop all services
docker-compose down

# Stop and remove volumes
docker-compose down -v
```
