package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config ...
type Config struct {
	Port     string
	Database DatabaseConfig
	JWT      JWTConfig
}

// DatabaseConfig ...
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

// JWTConfig ...
type JWTConfig struct {
	SecretKey string
	ExpiresIn int
}

// Load ...
func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	expiresInStr := getEnv("JWT_EXPIRES_IN", "900")
	expiresIn, err := strconv.Atoi(expiresInStr)
	if err != nil {
		log.Fatalf("Error converting JWT_EXPIRES_IN to int: %v", err)
	}

	return &Config{
		Port: getEnv("PORT", "8080"),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			Name:     getEnv("DB_NAME", "auth"),
		},
		JWT: JWTConfig{
			SecretKey: getEnv("JWT_SECRET", "secret"),
			ExpiresIn: expiresIn,
		},
	}
}

func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Connect подключается к базе данных на основе конфигурации
func (dbConfig *DatabaseConfig) Connect() (*sql.DB, error) {
	// Строка подключения
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.Name)

	// Подключение к базе данных
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть соединение с базой данных: %w", err)
	}

	// Проверка соединения
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("не удалось проверить соединение с базой данных: %w", err)
	}

	log.Println("Успешное подключение к базе данных")
	return db, nil
}
