# Pharmacy Claims Application

A robust pharmacy claims processing system built with Go and PostgreSQL, designed to handle prescription claim submissions, reversals, and pharmacy management with comprehensive audit logging.

## 🚀 Quick Start

### Prerequisites
- **Docker** and **Docker Compose** (recommended)
- **Go 1.24+** (for local development only)

### Get Started in 2 Steps
1. **Clone and start**:
   ```bash
   git clone <repository-url>
   cd pharmacy-claims-app
   make run
   ```

2. **Test the API**:
   ```bash
   curl http://localhost:8080/health
   ```

The application will automatically:
- Start PostgreSQL database
- Run database migrations
- Load sample data
- Start the API server on port 8080

## ✨ Features

- **Claim Processing**: Submit and validate prescription claims with NDC, quantity, NPI, and pricing
- **Claim Reversals**: Process reversals with complete audit trails
- **Pharmacy Management**: Load and validate pharmacy data from CSV files
- **Event Logging**: Comprehensive audit logging to JSON files and database
- **Data Validation**: Strict validation of NPIs, NDCs, and business rules
- **Graceful Shutdown**: Proper HTTP server lifecycle management
- **Auto Migrations**: Automated database schema management

## 📚 API Reference

### Endpoints
| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/claim` | Submit a prescription claim |
| `POST` | `/reversal` | Reverse an existing claim |
| `GET` | `/health` | Health check |

### Examples

**Submit a Claim:**
```bash
curl -X POST http://localhost:8080/claim \
  -H "Content-Type: application/json" \
  -d '{
    "ndc": "12345678901",
    "quantity": 30,
    "npi": "1234567890",
    "price": 25.99
  }'
```

**Reverse a Claim:**
```bash
curl -X POST http://localhost:8080/reversal \
  -H "Content-Type: application/json" \
  -d '{"claim_id": "your-claim-id-here"}'
```

## 🛠️ Development

### Available Commands
| Command | Description |
|---------|-------------|
| `make run` | Start development environment with live reload |
| `make test` | Run unit tests |
| `make stop` | Stop all services |
| `make clean` | Clean up containers and volumes |
| `make shell` | Open development shell |
| `make db-shell` | Connect to PostgreSQL |
| `make help` | Show all commands |

### Local Development (without Docker)
```bash
# Start database only
docker-compose up postgres -d

# Set environment variables
export DB_HOST=localhost DB_PORT=5432 DB_USER=pharmacy_user DB_PASSWORD=pharmacy_password DB_NAME=pharmacy_claims

# Run locally
go run cmd/server/main.go
```
## 🏗️ Architecture & Data Models

### Clean Architecture
```
┌─────────────────┐
│    Handlers     │  ← HTTP request handling
├─────────────────┤
│    Services     │  ← Business logic & validation
├─────────────────┤
│   Repository    │  ← Data access layer
├─────────────────┤
│   PostgreSQL    │  ← Data persistence
└─────────────────┘
```

### Data Models
**Claim:**
```json
{
  "id": "uuid",
  "ndc": "12345678901",     // National Drug Code (11 digits)
  "quantity": 30.0,         // Quantity dispensed
  "npi": "1234567890",      // National Provider Identifier (10 digits)
  "price": 25.99,           // Claim amount
  "timestamp": "2025-01-30T12:00:00Z"
}
```

**Reversal:**
```json
{
  "id": "uuid",
  "claim_id": "original-claim-uuid",
  "timestamp": "2025-01-30T12:05:00Z"
}
```

**Pharmacy:**
```json
{
  "id": 1,
  "npi": "1234567890",
  "chain": "health"
}
```

## 🗄️ Database & Configuration

### Database Schema
- **pharmacies**: Store pharmacy information (NPI, chain)
- **claims**: Store prescription claims
- **reversals**: Store claim reversals
- **event_logs**: Audit trail for all operations

### Environment Variables
| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `DB_HOST` | `postgres` | ✅ | Database host |
| `DB_PORT` | `5432` | ✅ | Database port |
| `DB_USER` | `pharmacy_user` | ✅ | Database user |
| `DB_PASSWORD` | `pharmacy_password` | ✅ | Database password |
| `DB_NAME` | `pharmacy_claims` | ✅ | Database name |
| `PORT` | `8080` | ❌ | Application port |
| `LOG_LEVEL` | `info` | ❌ | Logging level (debug, info, warn, error) |
| `GO_ENV` | `production` | ❌ | Environment mode |

### Sample Data
The application automatically loads sample data on startup:
- **Pharmacies**: CSV files in `data/pharmacies/` (format: `chain,npi`)
- **Claims**: JSON files in `data/claims/`
- **Reversals**: JSON files in `data/reverts/`

## 📁 Project Structure

```
pharmacy-claims-app/
├── cmd/server/          # Application entry point
├── internal/
│   ├── core/           # Configuration and logging
│   ├── database/       # Database connection and migrations
│   ├── handlers/       # HTTP handlers
│   ├── models/         # Data models
│   ├── repository/     # Data access layer
│   ├── service/        # Business logic
│   └── utility/        # Helper functions
├── data/               # Sample data files
├── migrations/         # Database migration files
├── tests/             # Unit tests
├── logs/              # Application logs
├── docker-compose.yml # Docker setup
├── Dockerfile         # Container definition
└── Makefile          # Development commands
```

## 🧪 Testing & Deployment

### Running Tests
```bash
make test                    # Run all tests
go test -v -cover ./...     # Run with coverage (local)
```

### Deployment Options

**Development:**
```bash
make run                    # Docker with live reload
```

**Production:**
```bash
docker build -t pharmacy-claims:latest .
docker run -d -p 8080:8080 \
  -e DB_HOST=your-db-host \
  -e DB_PASSWORD=your-password \
  pharmacy-claims:latest
```

## � Troubleshooting

**Database Connection Issues:**
```bash
docker-compose ps           # Check container status
docker-compose logs postgres # View database logs
make clean && make run      # Reset everything
```

**Port Conflicts:**
```bash
lsof -i :8080              # Check what's using port 8080
PORT=8081 make run          # Use different port
```

**Build Errors:**
```bash
make clean                  # Clean containers
go mod tidy                 # Fix Go dependencies
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
