package data

import (
	"fmt"
	"os"
)

// PostgreSQLConfig contains the configuration for a PostgreSQL database.
type PostgreSQLConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

// NewPostgreSQLConfig creates a new PostgreSQLConfig with values from environment variables,
// or defaults if the environment variables are not set. These defaults are suitable for a
// PostgreSQL database running in a Docker container.
//
// The default values are:
//
//	Host:     localhost
//	Port:     5432
//	User:     postgres
//	Password: password
//	Database: postgres
//
// If you are using a different setup, you can set the following environment variables:
//
//	DB_HOST
//	DB_PORT
//	DB_USER
//	DB_PASSWORD
//	DB_NAME
func NewPostgreSQLConfig() *PostgreSQLConfig {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "password"
	}
	database := os.Getenv("DB_NAME")
	if database == "" {
		database = "postgres"
	}

	return &PostgreSQLConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: database,
	}
}

// Dsn returns the data source name (DSN) for the PostgreSQL database.
func (c *PostgreSQLConfig) Dsn() string {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.Database)

	return dsn
}
