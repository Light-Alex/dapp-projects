package utils

import (
	"os"
	"strconv"
)

// GetEnv 获取环境变量，如果不存在则返回默认值
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvInt 获取整数类型的环境变量
func GetEnvInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// GetEnvInt64 获取 int64 类型的环境变量
func GetEnvInt64(key string, defaultValue int64) int64 {
	valueStr := os.Getenv(key)
	if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		return value
	}
	return defaultValue
}

// GetEnvUint64 获取 uint64 类型的环境变量
func GetEnvUint64(key string, defaultValue uint64) uint64 {
	valueStr := os.Getenv(key)
	if value, err := strconv.ParseUint(valueStr, 10, 64); err == nil {
		return value
	}
	return defaultValue
}

// GetEnvBool 获取布尔类型的环境变量
func GetEnvBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	switch valueStr {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return defaultValue
	}
}
