package config

import (
	"log/slog"
	"os"
)

type Config struct {
	Address string
	DB      DB
}

type DB struct {
	User     string
	Password string
	Host     string
	Port     string
	Name     string
	Conn     string
	JDBC     string
}

func Load() *Config {
	var cfg Config
	readEnv(&cfg)
	return &cfg
}

func readEnv(cfg *Config) {
	var exists bool
	cfg.DB.User, exists = os.LookupEnv("POSTGRES_USERNAME")
	if !exists {
		slog.Error(`can't find "POSTGRES_USERNAME" env`)
	}
	cfg.DB.Password, exists = os.LookupEnv("POSTGRES_PASSWORD")
	if !exists {
		slog.Error(`can't find "POSTGRES_PASSWORD" env`)
	}
	cfg.DB.Host, exists = os.LookupEnv("POSTGRES_HOST")
	if !exists {
		slog.Error(`can't find "POSTGRES_HOST" env`)
	}
	cfg.DB.Port, exists = os.LookupEnv("POSTGRES_PORT")
	if !exists {
		slog.Error(`can't find "POSTGRES_PORT" env`)
	}
	cfg.DB.Name, exists = os.LookupEnv("POSTGRES_DATABASE")
	if !exists {
		slog.Error(`can't find "POSTGRES_DATABASE" env`)
	}
}
