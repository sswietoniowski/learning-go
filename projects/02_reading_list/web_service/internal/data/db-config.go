package data

import (
	"fmt"
	"os"
)

type DbConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

func NewConfig() DbConfig {
	return DbConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Database: os.Getenv("DB_NAME"),
	}
}

func (c DbConfig) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", c.Host, c.Port, c.User, c.Password, c.Database)
}
