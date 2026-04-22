package config

import (
	"github.com/AnchoredLabs/rwa-backend/libs/core/bootstrap"
	"github.com/AnchoredLabs/rwa-backend/libs/core/kafka_help"
	"github.com/AnchoredLabs/rwa-backend/libs/core/redis_cache"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
)

type Config struct {
	AppName string                   `json:"appName" yaml:"appName"`
	Server  *ServerConfig            `json:"server" yaml:"server"`
	Redis   *redis_cache.RedisConfig `json:"redis" yaml:"redis"`
	Kafka   *kafka_help.KafkaConfig  `json:"kafka" yaml:"kafka"`
	Logger  *log.Conf                `json:"logger" yaml:"logger"`
}

type ServerConfig struct {
	Port     int    `json:"port" yaml:"port"`
	BasePath string `json:"basePath" yaml:"basePath"`
}

func NewConfig(configFile string) (*Config, error) {
	return bootstrap.LoadConfig[Config](configFile)
}
