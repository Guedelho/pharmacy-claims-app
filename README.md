# Pharmacy Claims Application

A robust pharmacy claims processing system built with Go and PostgreSQL, designed to handle prescription claim submissions, reversals, and pharmacy management with comprehensive audit logging.

## 🚀 Features

- **Claim Submission**: Accept and validate prescription claims with NDC, quantity, NPI, and pricing
- **Claim Reversal**: Process claim reversals with proper audit trails
- **Pharmacy Management**: Load and validate pharmacy data from CSV files
- **Event Logging**: Comprehensive audit logging to both database and JSON files
- **Data Validation**: Strict validation of NPIs, NDCs, and business rules
- **Graceful Shutdown**: Proper HTTP server lifecycle management
- **Database Migrations**: Automated database schema management

## 🏗️ Architecture

The application follows a clean three-layer architecture:

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

## 📋 Prerequisites

- **Docker** and **Docker Compose** (recommended)
- **Go 1.24+** (for local development)
- **PostgreSQL 15+** (if running without Docker)

## 🚀 Quick Start

### Option 1: Docker (Recommended)

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd pharmacy-claims-app
   ```

2. **Start the application**:
   ```bash
   make run
   ```
   This will:
   - Build the Go application
   - Start PostgreSQL database
   - Run database migrations
   - Load sample data (pharmacies, claims, reversals)
   - Start the API server on port 8080

3. **Verify it's running**:
   ```bash
   curl http://localhost:8080/health
   ```

### Option 2: Local Development

1. **Set up the database**:
   ```bash
   # Start only PostgreSQL
   docker-compose up postgres -d
   ```

2. **Set environment variables**:
   ```bash
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_USER=pharmacy_user
   export DB_PASSWORD=pharmacy_password
   export DB_NAME=pharmacy_claims
   export PORT=8080
   ```

3. **Run the application**:
   ```bash
   go run cmd/server/main.go
   ```

## 🧪 Testing the API

### Submit a Claim
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

### Reverse a Claim
```bash
curl -X POST http://localhost:8080/reversal \
  -H "Content-Type: application/json" \
  -d '{
    "claim_id": "your-claim-id-here"
  }'
```

### Health Check
```bash
curl http://localhost:8080/health
```

## 📚 API Reference

### Endpoints

| Method | Endpoint | Description | Request Body |
|--------|----------|-------------|--------------|
| `POST` | `/claim` | Submit a new prescription claim | Claim object |
| `POST` | `/reversal` | Reverse an existing claim | Reversal object |
| `GET` | `/health` | Health check endpoint | None |

### Request/Response Examples

#### Claim Submission
**Request:**
```json
{
  "ndc": "12345678901",
  "quantity": 30.0,
  "npi": "1234567890",
  "price": 25.99
}
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "ndc": "12345678901",
  "quantity": 30.0,
  "npi": "1234567890",
  "price": 25.99,
  "timestamp": "2025-01-30T12:00:00Z"
}
```

#### Claim Reversal
**Request:**
```json
{
  "claim_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Response:**
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "claim_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2025-01-30T12:05:00Z"
}
```

## 🛠️ Development Commands

| Command | Description |
|---------|-------------|
| `make run` | Start development environment with live reload |
| `make test` | Run unit tests |
| `make clean` | Clean up containers and volumes |
| `make stop` | Stop all services |
| `make shell` | Open development shell with Go tools |
| `make db-shell` | Connect to PostgreSQL shell |
| `make help` | Show all available commands |

## 📊 Data Models

### Claim
```go
type Claim struct {
    ID        uuid.UUID  `json:"id"`
    NDC       string     `json:"ndc"`        // National Drug Code (11 digits)
    Quantity  float64    `json:"quantity"`   // Quantity dispensed
    NPI       string     `json:"npi"`        // National Provider Identifier (10 digits)
    Price     float64    `json:"price"`      // Claim amount
    Timestamp time.Time  `json:"timestamp"`  // Submission time
}
```

### Reversal
```go
type Reversal struct {
    ID        uuid.UUID  `json:"id"`
    ClaimID   uuid.UUID  `json:"claim_id"`   // Reference to original claim
    Timestamp time.Time  `json:"timestamp"`  // Reversal time
}
```

### Pharmacy
```go
type Pharmacy struct {
    ID    int    `json:"id"`
    NPI   string `json:"npi"`    // National Provider Identifier
    Chain string `json:"chain"`  // Pharmacy chain name
}
```

## 🗄️ Database Schema

### Tables
- **pharmacies**: Store pharmacy information (NPI, chain)
- **claims**: Store prescription claims
- **reversals**: Store claim reversals
- **event_logs**: Audit trail for all operations

### Migrations
Database migrations are automatically applied on startup and stored in `/migrations`:
- `000001_init.up.sql`: Creates initial schema
- `000001_init.down.sql`: Rollback script

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
│   ├── claims/         # Sample claim files (JSON)
│   ├── pharmacies/     # Pharmacy data (CSV)
│   └── reverts/        # Sample reversal files
├── migrations/         # Database migration files
├── tests/             # Unit tests
├── logs/              # Application logs
├── docker-compose.yml # Docker setup
├── Dockerfile         # Container definition
└── Makefile          # Development commands
```
## 🔧 Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DB_HOST` | Database host | `postgres` | ✅ |
| `DB_PORT` | Database port | `5432` | ✅ |
| `DB_USER` | Database user | `pharmacy_user` | ✅ |
| `DB_PASSWORD` | Database password | `pharmacy_password` | ✅ |
| `DB_NAME` | Database name | `pharmacy_claims` | ✅ |
| `PORT` | Application port | `8080` | ❌ |
| `MIGRATIONS_DIR` | Migrations directory | `./migrations` | ❌ |
| `DATA_DIR` | Sample data directory | `./data` | ❌ |
| `LOG_DIR` | Log files directory | `./logs` | ❌ |
| `LOG_LEVEL` | Logging level | `info` | ❌ |
| `GO_ENV` | Environment mode | `production` | ❌ |

### Configuration Files

The application uses the following configuration approach:
1. Environment variables (highest priority)
2. Default values in `internal/core/config.go`
3. Docker Compose environment (for containerized deployment)

## 🧪 Testing

### Running Tests
```bash
# Run all tests
make test

# Run tests with coverage (local development)
go test -v -cover ./...

# Run specific test file
go test -v ./tests/handlers_test.go
```

### Test Coverage
The application includes comprehensive unit tests for:
- HTTP handlers (`tests/handlers_test.go`)
- Business services (`tests/service_test.go`)
- Data validation and error handling
- Database operations

## 🐳 Docker Configuration

### Development Environment
```bash
# Start with live reload (mounts source code)
make run

# View logs
docker-compose logs -f app

# Access container shell
make shell
```

### Production Deployment
```bash
# Build production image
docker build --target production -t pharmacy-claims:latest .

# Run with production settings
docker run -d \
  -p 8080:8080 \
  -e DB_HOST=your-db-host \
  -e DB_PASSWORD=your-secure-password \
  pharmacy-claims:latest
```

## 📋 Data Loading

### Sample Data
The application automatically loads sample data on startup:

1. **Pharmacies** (`data/pharmacies/*.csv`):
   - Format: `chain,npi`
   - Example: `health,1234567890`

2. **Claims** (`data/claims/*.json`):
   - Historical claim data for testing
   - Automatically validates against pharmacy NPIs

3. **Reversals** (`data/reverts/*`):
   - Sample reversal transactions

### Custom Data Loading
To load your own data:
1. Place CSV files in `data/pharmacies/`
2. Place JSON claim files in `data/claims/`
3. Restart the application or use the loader service

## 🔍 Monitoring & Logging

### Application Logs
- **Console**: Structured logging to stdout/stderr
- **Files**: JSON event logs in `/logs` directory
- **Database**: Audit trail in `event_logs` table

### Log Levels
- `debug`: Detailed debugging information
- `info`: General operational messages
- `warn`: Warning conditions
- `error`: Error conditions

### Health Monitoring
```bash
# Check application health
curl http://localhost:8080/health

# Monitor logs
make logs

# Database connection check
make db-shell
```

## 🚀 Deployment

### Local Development
1. Use `make run` for development with live reload
2. Modify code and see changes automatically
3. Use `make test` to run tests

### Staging/Production
1. Build production Docker image
2. Configure environment variables
3. Set up PostgreSQL database
4. Run migrations
5. Deploy container with proper resource limits

### Best Practices
- Always run migrations before deploying new versions
- Monitor database connections and performance
- Set appropriate timeouts for HTTP operations
- Use proper logging levels in production
- Implement health checks in your orchestration platform

## 🤝 Contributing

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Make your changes**: Follow Go conventions and add tests
4. **Run tests**: `make test`
5. **Commit changes**: `git commit -m 'Add amazing feature'`
6. **Push to branch**: `git push origin feature/amazing-feature`
7. **Open a Pull Request**

### Development Guidelines
- Follow Go naming conventions
- Add unit tests for new functionality
- Update documentation for API changes
- Use meaningful commit messages
- Ensure all tests pass before submitting PR

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Troubleshooting

### Common Issues

**Database Connection Errors**
```bash
# Check if PostgreSQL is running
docker-compose ps

# View database logs
docker-compose logs postgres

# Restart database
docker-compose restart postgres
```

**Port Conflicts**
```bash
# Check what's using port 8080
lsof -i :8080

# Use different port
PORT=8081 make run
```

**Migration Failures**
```bash
# Check migration status
make db-shell
# Then run: SELECT * FROM schema_migrations;

# Reset database
make clean && make run
```

**Build Errors**
```bash
# Clean and rebuild
make clean
make run

# Check Go module issues
go mod tidy
go mod verify
```

### Getting Help
- Check the logs: `make logs`
- Verify environment variables are set correctly
- Ensure Docker and Docker Compose are properly installed
- Review the database schema in `/migrations`
