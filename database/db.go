package database

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Knightshrestha/Secret-Injector/database/generated"
	_ "modernc.org/sqlite"
)

//go:embed src/schema.sql
var ddl string

type CustomDB struct {
	ReadDB  *sql.DB
	WriteDB *sql.DB

	ReadQueries  *generated.Queries
	WriteQueries *generated.Queries
}

func getDBPath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	exeDir := filepath.Dir(exePath)
	dataDir := filepath.Join(exeDir, "si_data")

	// Create folder if missing
	if err := os.MkdirAll(dataDir, os.ModePerm); err != nil {
		return "", err
	}

	dbPath := filepath.Join(dataDir, "secrets.db")
	return dbPath, nil
}

type DB_Struct struct {
	DB      *sql.DB
	Queries *generated.Queries
}

func SetupDatabase() error {
	dbPath, err := getDBPath()
	if err != nil {
		return fmt.Errorf("cannot get database path: %w", err)
	}

	// Open temporary connection for setup
	database, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("cannot open database for setup: %w", err)
	}
	defer database.Close()

	// Create tables
	ctx := context.Background()
	if _, err := database.ExecContext(ctx, ddl); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

// OpenWriteDatabase opens a single write connection with WAL mode
func OpenWriteDatabase() (DB_Struct, error) {
	dbPath, err := getDBPath()
	if err != nil {
		return DB_Struct{}, fmt.Errorf("cannot get database path: %w", err)
	}

	// Write connection DSN
	pragmasWrite := "?_busy_timeout=5000&_synchronous=NORMAL&cache_size=-2000&_txlock=immediate&_timeout=5000&_foreign_keys=1"

	// Write connection (1 connection max)
	dbWrite, err := sql.Open("sqlite", dbPath+pragmasWrite)
	if err != nil {
		return DB_Struct{}, fmt.Errorf("cannot open write database: %w", err)
	}

	dbWrite.SetMaxOpenConns(1)
	dbWrite.SetMaxIdleConns(1)
	dbWrite.SetConnMaxLifetime(time.Hour)

	// Enable WAL mode
	if _, err := dbWrite.Exec("PRAGMA journal_mode=WAL"); err != nil {
		dbWrite.Close()
		return DB_Struct{}, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Verify WAL mode is enabled
	var mode string
	if err := dbWrite.QueryRow("PRAGMA journal_mode").Scan(&mode); err != nil {
		dbWrite.Close()
		return DB_Struct{}, fmt.Errorf("failed to check WAL mode: %w", err)
	}
	if mode != "wal" {
		dbWrite.Close()
		return DB_Struct{}, fmt.Errorf("WAL mode not enabled, got: %s", mode)
	}
	fmt.Println("✓ WAL mode enabled:", mode)

	return DB_Struct{
		DB:      dbWrite,
		Queries: generated.New(dbWrite),
	}, nil
}

// OpenReadDatabase opens multiple read-only connections
func OpenReadDatabase() (DB_Struct, error) {
	dbPath, err := getDBPath()
	if err != nil {
		return DB_Struct{}, fmt.Errorf("cannot get database path: %w", err)
	}

	// Read connection DSN
	pragmasRead := "?mode=ro&_query_only=1&cache_size=-2000&_busy_timeout=5000&_foreign_keys=1"

	// Read connection (25-50 connections for production)
	dbRead, err := sql.Open("sqlite", dbPath+pragmasRead)
	if err != nil {
		return DB_Struct{}, fmt.Errorf("cannot open read database: %w", err)
	}

	dbRead.SetMaxOpenConns(25)
	dbRead.SetMaxIdleConns(10)
	dbRead.SetConnMaxLifetime(time.Hour)

	return DB_Struct{
		DB:      dbRead,
		Queries: generated.New(dbRead),
	}, nil
}

// OpenDatabase opens both read and write database connections
func OpenDatabase() CustomDB {
	// Setup database schema first (if needed)
	if err := SetupDatabase(); err != nil {
		log.Fatal("Failed to setup database:", err)
	}
	
	// Open write database
	writableDatabase, err := OpenWriteDatabase()
	if err != nil {
		log.Fatal("Failed to open write database:", err)
	}

	// Open read database
	readOnlyDatabase, err := OpenReadDatabase()
	if err != nil {
		writableDatabase.DB.Close()
		log.Fatal("Failed to open read database:", err)
	}

	return CustomDB{
		ReadDB:       readOnlyDatabase.DB,
		WriteDB:      writableDatabase.DB,
		ReadQueries:  readOnlyDatabase.Queries,
		WriteQueries: writableDatabase.Queries,
	}
}

func CloseWriteDatabase(dbWrite *sql.DB) error {
	if dbWrite == nil {
		return nil
	}

	// Checkpoint WAL before closing
	if _, err := dbWrite.Exec(`PRAGMA wal_checkpoint(TRUNCATE);`); err != nil {
		log.Println("Warning: Failed to checkpoint WAL:", err)
	}

	// Close write connection
	if err := dbWrite.Close(); err != nil {
		return fmt.Errorf("failed to close write database: %w", err)
	}

	log.Println("Write database connection closed successfully.")
	return nil
}

func CloseReadDatabase(dbRead *sql.DB) error {
	if dbRead == nil {
		return nil
	}

	if err := dbRead.Close(); err != nil {
		return fmt.Errorf("failed to close read database: %w", err)
	}

	return nil
}

func CloseDatabase(db CustomDB) error {
	// Close read connection first
	if err := CloseReadDatabase(db.ReadDB); err != nil {
		log.Println("Warning:", err)
	}

	// Close write connection (with WAL checkpoint)
	if err := CloseWriteDatabase(db.WriteDB); err != nil {
		return err
	}

	return nil
}
