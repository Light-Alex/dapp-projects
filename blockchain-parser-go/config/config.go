package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Blockchain
	RPCURL         string
	ChainID        int64
	ProjectAddress string
	USDTAddress    string
	PrivateKey     string

	// Database
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string

	// Redis
	RedisHost     string
	RedisPort     int
	RedisPassword string

	// Application
	ScanInterval       int
	ConfirmationBlocks int
}

func LoadConfig() *Config {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	return &Config{
		RPCURL:         getEnv("BSC_RPC_URL", "https://bsc-dataseed.binance.org/"),
		ChainID:        int64(getEnvAsInt("CHAIN_ID", 56)),
		ProjectAddress: getEnv("PROJECT_ADDRESS", ""),
		USDTAddress:    getEnv("USDT_ADDRESS", "0x55d398326f99059fF775485246999027B3197955"),
		PrivateKey:     getEnv("SEPOLIA_PRIVATE_KEY", ""),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnvAsInt("DB_PORT", 5432),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "blockchain_parser"),

		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnvAsInt("REDIS_PORT", 6379),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

		ScanInterval:       getEnvAsInt("SCAN_INTERVAL", 3000),
		ConfirmationBlocks: getEnvAsInt("CONFIRMATION_BLOCKS", 6),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
