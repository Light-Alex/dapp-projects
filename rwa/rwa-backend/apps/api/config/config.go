package config

import (
	"github.com/AnchoredLabs/rwa-backend/libs/core/bootstrap"
	"github.com/AnchoredLabs/rwa-backend/libs/core/evm_helper"
	"github.com/AnchoredLabs/rwa-backend/libs/core/redis_cache"
	"github.com/AnchoredLabs/rwa-backend/libs/core/trade"
	"github.com/AnchoredLabs/rwa-backend/libs/core/web"
	"github.com/AnchoredLabs/rwa-backend/libs/database"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
)

type Config struct {
	AppName    string                   `json:"appName" yaml:"appName"`
	Server     *web.ServerConfig        `json:"server" yaml:"server"`
	RpcInfo    evm_helper.RpcInfoMap    `json:"rpcInfo" yaml:"rpcInfo"`
	Redis      *redis_cache.RedisConfig `json:"redis" yaml:"redis"`
	Logger     *log.Conf                `json:"logger" yaml:"logger"`
	GrpcClient *GrpcClient              `json:"grpcClient" yaml:"grpcClient"`
	FrontendTx *FrontendTxConfig        `json:"frontendTx" yaml:"frontendTx"`
	Db         *database.DbConf         `json:"db" yaml:"db"`
	Alpaca     *trade.AlpacaConfig      `json:"alpaca" yaml:"alpaca"`
}

type GrpcClient struct {
	IndexerRpc  map[uint64]string `json:"indexerRpc" yaml:"indexerRpc"`
	SnapshotRpc map[uint64]string `json:"snapshotRpc" yaml:"snapshotRpc"`
	TxService   map[uint64]string `json:"txService" yaml:"txService"`
}

type FrontendTxConfig struct {
	ProxyWallet *ProxyWalletConfig `json:"proxyWallet" yaml:"proxyWallet"`
	JWT         *JWTConfig         `json:"jwt" yaml:"jwt"`
}

type ProxyWalletConfig struct {
	Mnemonic       string `json:"mnemonic" yaml:"mnemonic"`
	DerivationPath string `json:"derivationPath" yaml:"derivationPath"`
}

type JWTConfig struct {
	Secret          string `json:"secret" yaml:"secret"`
	ExpirationHours int    `json:"expirationHours" yaml:"expirationHours"`
}

func NewConfig(configFile string) (*Config, error) {
	return bootstrap.LoadConfig[Config](configFile)
}
