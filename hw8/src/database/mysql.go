package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

// InitDB initializes the MySQL database connection with connection pooling
func InitDB() error {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// Validate environment variables
	if dbHost == "" || dbPort == "" || dbUser == "" || dbPassword == "" || dbName == "" {
		return fmt.Errorf("missing required database environment variables")
	}

	// Connection string: user:password@tcp(host:port)/database?parseTime=true
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}

	// Connection pool settings for performance (per HW8 requirements)
	DB.SetMaxOpenConns(25)                 // Maximum open connections
	DB.SetMaxIdleConns(5)                  // Maximum idle connections
	DB.SetConnMaxLifetime(5 * time.Minute) // Connection lifetime

	// Verify connection with retry logic
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		err = DB.Ping()
		if err == nil {
			log.Println("✅ Successfully connected to MySQL database")
			return nil
		}
		log.Printf("⚠️  Database connection attempt %d/%d failed: %v", i+1, maxRetries, err)
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
}

// InitSchema creates the database tables if they don't exist
// This eliminates the need for separate SQL files
func InitSchema() error {
	log.Println("📋 Initializing database schema...")

	// Create shopping_carts table
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS shopping_carts (
			cart_id INT AUTO_INCREMENT PRIMARY KEY,
			customer_id INT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_customer_id (customer_id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create shopping_carts table: %w", err)
	}
	log.Println("✅ shopping_carts table ready")

	// Create cart_items table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS cart_items (
			cart_id INT NOT NULL,
			product_id INT NOT NULL,
			quantity INT NOT NULL DEFAULT 1,
			added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (cart_id, product_id),
			FOREIGN KEY (cart_id) REFERENCES shopping_carts(cart_id) ON DELETE CASCADE,
			INDEX idx_cart_id (cart_id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create cart_items table: %w", err)
	}
	log.Println("✅ cart_items table ready")

	log.Println("✅ Database schema initialization complete")
	return nil
}

// CloseDB closes the database connection
func CloseDB() {
	if DB != nil {
		DB.Close()
		log.Println("Database connection closed")
	}
}
