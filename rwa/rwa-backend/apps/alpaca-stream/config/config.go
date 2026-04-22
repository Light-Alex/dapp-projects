package config

import (
	"github.com/AnchoredLabs/rwa-backend/libs/core/bootstrap"
	"github.com/AnchoredLabs/rwa-backend/libs/core/evm_helper"
	"github.com/AnchoredLabs/rwa-backend/libs/core/kafka_help"
	"github.com/AnchoredLabs/rwa-backend/libs/core/redis_cache"
	"github.com/AnchoredLabs/rwa-backend/libs/database"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
)

type Config struct {
	AppName string                   `json:"appName" yaml:"appName"`
	Logger  *log.Conf                `json:"logger" yaml:"logger"`
	Alpaca  *AlpacaWebSocketConfig   `json:"alpaca" yaml:"alpaca"`
	Db      *database.DbConf         `json:"db" yaml:"db"`
	Redis   *redis_cache.RedisConfig `json:"redis" yaml:"redis"`
	Kafka   *kafka_help.KafkaConfig  `json:"kafka" yaml:"kafka"`
	RpcInfo evm_helper.RpcInfoMap    `json:"rpcInfo" yaml:"rpcInfo"`
	Chain   *ChainConfig             `json:"chain" yaml:"chain"`
	Backend *BackendConfig           `json:"backend" yaml:"backend"`
}

// ChainConfig contains blockchain-related configuration
type ChainConfig struct {
	ChainId    uint64 `json:"chainId" yaml:"chainId"`
	PocAddress string `json:"pocAddress" yaml:"pocAddress"`
}

// BackendConfig contains the backend wallet private key for signing transactions
type BackendConfig struct {
	PrivateKey string `json:"privateKey" yaml:"privateKey"`
}

// AlpacaWebSocketConfig contains WebSocket configuration
type AlpacaWebSocketConfig struct {
	APIKey           string   `json:"api_key" yaml:"api_key"`
	APISecret        string   `json:"api_secret" yaml:"api_secret"`
	WSURL            string   `json:"ws_url" yaml:"ws_url"`
	WSDataURL        string   `json:"ws_data_url" yaml:"ws_data_url"`
	EnableMarketData bool     `json:"enable_market_data" yaml:"enable_market_data"`
	Symbols          []string `json:"symbols" yaml:"symbols"`
}

func NewConfig(configFile string) (*Config, error) {
	return bootstrap.LoadConfig[Config](configFile)
}
