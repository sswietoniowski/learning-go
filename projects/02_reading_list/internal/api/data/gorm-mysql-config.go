package data

import (
	"fmt"
	"os"
)

// GormMySQLConfig contains the configuration for a MySQL database.
type GormMySQLConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

// NewGormMySQLConfig creates a new MySQLConfig with values from environment variables,
// or defaults if the environment variables are not set. These defaults are suitable for a
// MySQL database running in a Docker container.
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
func NewGormMySQLConfig() *GormMySQLConfig {
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

	return &GormMySQLConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: database,
	}
}

// Dsn returns the data source name (DSN) for the PostgreSQL database.
func (c *GormMySQLConfig) Dsn() string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.Database)

	return dsn
}
