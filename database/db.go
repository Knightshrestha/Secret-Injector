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

func OpenDatabase() CustomDB {
	dbPath, err := getDBPath()
	if err != nil {
		log.Fatal("Cannot get database path:", err)
	}

	// Write connection DSN
	pragmasWrite := "?_busy_timeout=5000&_synchronous=NORMAL&cache_size=-2000&_txlock=immediate&_timeout=5000&_foreign_keys=1"

	// Read connection DSN
	pragmasRead := "?mode=ro&_query_only=1&cache_size=-2000&_busy_timeout=5000&_foreign_keys=1"

	// Write connection (1 connection max)
	dbWrite, err := sql.Open("sqlite", dbPath+pragmasWrite)
	if err != nil {
		log.Fatal("Cannot open write database:", err)
	}
	dbWrite.SetMaxOpenConns(1)
	dbWrite.SetMaxIdleConns(1)
	dbWrite.SetConnMaxLifetime(time.Hour)

	if _, err := dbWrite.Exec("PRAGMA journal_mode=WAL"); err != nil {
		log.Fatal("Failed to enable WAL mode:", err)
	}

	// Verify WAL mode is enabled
	var mode string
	if err := dbWrite.QueryRow("PRAGMA journal_mode").Scan(&mode); err != nil {
		log.Fatal("Failed to check WAL mode:", err)
	}
	if mode != "wal" {
		log.Fatal("WAL mode not enabled, got:", mode)
	}
	fmt.Println("✓ WAL mode enabled:", mode)

	// Create tables ONLY on write connection
	ctx := context.Background()
	if _, err := dbWrite.ExecContext(ctx, ddl); err != nil {
		log.Fatal("Failed to create tables:", err)
	}

	// Read connection (25-50 connections for production), Open AFTER creating tables
	dbRead, err := sql.Open("sqlite", dbPath+pragmasRead)
	if err != nil {
		dbWrite.Close()
		log.Fatal("Cannot open read database:", err)
	}

	dbRead.SetMaxOpenConns(25)
	dbRead.SetMaxIdleConns(10)

	dbRead.SetConnMaxLifetime(time.Hour)

	readQueries := generated.New(dbRead)
	writeQueries := generated.New(dbWrite)

	return CustomDB{
		ReadDB:       dbRead,
		WriteDB:      dbWrite,
		ReadQueries:  readQueries,
		WriteQueries: writeQueries,
	}
}

// CloseDatabase properly closes both database connections and checkpoints WAL
func CloseDatabase(db CustomDB) error {
	if _, err := db.WriteDB.Exec(`PRAGMA wal_checkpoint(TRUNCATE);`); err != nil {
		log.Println("Warning: Failed to checkpoint WAL:", err)
	}

	// Close read connection first
	if err := db.ReadDB.Close(); err != nil {
		log.Println("Warning: Failed to close read database:", err)
	}

	// Close write connection
	if err := db.WriteDB.Close(); err != nil {
		return fmt.Errorf("failed to close write database: %w", err)
	}

	log.Println("Database connections closed successfully.")
	return nil
}
