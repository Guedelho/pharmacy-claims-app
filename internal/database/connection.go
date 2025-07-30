package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type Connection struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type DB struct {
	*sql.DB
}

func NewConnection(connInfo Connection) (*DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		connInfo.Host, connInfo.Port, connInfo.User, connInfo.Password, connInfo.DBName, connInfo.SSLMode)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database")
	return &DB{db}, nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}

func (db *DB) Health() error {
	return db.Ping()
}

func (db *DB) BeginTx() (*sql.Tx, error) {
	return db.Begin()
}

func (db *DB) ExecuteInTransaction(fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}

func WaitForConnection(connInfo Connection, maxRetries int, retryInterval time.Duration) error {
	log.Println("Waiting for database to be ready...")

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		connInfo.Host, connInfo.Port, connInfo.User, connInfo.Password, connInfo.DBName, connInfo.SSLMode)

	for i := 0; i < maxRetries; i++ {
		db, err := sql.Open("postgres", psqlInfo)
		if err != nil {
			if i < maxRetries-1 {
				log.Printf("Database not ready (attempt %d/%d), retrying in %v...", i+1, maxRetries, retryInterval)
				time.Sleep(retryInterval)
				continue
			}
			return fmt.Errorf("database connection failed after %d attempts: %w", maxRetries, err)
		}

		if err := db.Ping(); err != nil {
			db.Close()
			if i < maxRetries-1 {
				log.Printf("Database not ready (attempt %d/%d), retrying in %v...", i+1, maxRetries, retryInterval)
				time.Sleep(retryInterval)
				continue
			}
			return fmt.Errorf("database ping failed after %d attempts: %w", maxRetries, err)
		}

		db.Close()
		log.Println("Database is ready")
		return nil
	}

	return fmt.Errorf("database connection failed after %d attempts", maxRetries)
}
