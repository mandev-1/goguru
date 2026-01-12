package config

import "os"

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	SMTP     SMTPConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type SMTPConfig struct {
	Host string
	Port string
	User string
	Pass string
	From string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     os.Getenv("DB_NAME"),
		},
		SMTP: SMTPConfig{
			Host: getEnv("SMTP_HOST", "mailhog"),
			Port: getEnv("SMTP_PORT", "1025"),
			User: os.Getenv("SMTP_USER"),
			Pass: os.Getenv("SMTP_PASS"),
			From: getEnv("SMTP_FROM", "camagru@localhost"),
		},
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
