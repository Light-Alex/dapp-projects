package config

import (
	"strings"

	"github.com/spf13/viper"

	logging "github.com/ProjectsTask/EasySwapBase/logger"
	"github.com/ProjectsTask/EasySwapBase/stores/gdb"
)

type Config struct {
	Monitor     *Monitor         `toml:"monitor" mapstructure:"monitor" json:"monitor"`
	Log         *logging.LogConf `toml:"log" mapstructure:"log" json:"log"`
	Kv          *KvConf          `toml:"kv" mapstructure:"kv" json:"kv"`
	DB          *gdb.Config      `toml:"db" mapstructure:"db" json:"db"`
	AnkrCfg     AnkrCfg          `toml:"ankr_cfg" mapstructure:"ankr_cfg" json:"ankr_cfg"`
	ChainCfg    ChainCfg         `toml:"chain_cfg" mapstructure:"chain_cfg" json:"chain_cfg"`
	ContractCfg ContractCfg      `toml:"contract_cfg" mapstructure:"contract_cfg" json:"contract_cfg"`
	ProjectCfg  ProjectCfg       `toml:"project_cfg" mapstructure:"project_cfg" json:"project_cfg"`
}

type ChainCfg struct {
	Name string `toml:"name" mapstructure:"name" json:"name"`
	ID   int64  `toml:"id" mapstructure:"id" json:"id"`
}

type ContractCfg struct {
	EthAddress  string `toml:"eth_address" mapstructure:"eth_address" json:"eth_address"`
	WethAddress string `toml:"weth_address" mapstructure:"weth_address" json:"weth_address"`
	DexAddress  string `toml:"dex_address" mapstructure:"dex_address" json:"dex_address"`
}

type Monitor struct {
	PprofEnable bool  `toml:"pprof_enable" mapstructure:"pprof_enable" json:"pprof_enable"`
	PprofPort   int64 `toml:"pprof_port" mapstructure:"pprof_port" json:"pprof_port"`
}

type AnkrCfg struct {
	ApiKey       string `toml:"api_key" mapstructure:"api_key" json:"api_key"`
	HttpsUrl     string `toml:"https_url" mapstructure:"https_url" json:"https_url"`
	WebsocketUrl string `toml:"websocket_url" mapstructure:"websocket_url" json:"websocket_url"`
	EnableWss    bool   `toml:"enable_wss" mapstructure:"enable_wss" json:"enable_wss"`
}

type ProjectCfg struct {
	Name string `toml:"name" mapstructure:"name" json:"name"`
}

type KvConf struct {
	Redis []*Redis `toml:"redis" json:"redis"`
}

type Redis struct {
	Host string `toml:"host" json:"host"`
	Type string `toml:"type" json:"type"`
	Pass string `toml:"pass" json:"pass"`
}

type LogLevel struct {
	Api      string `toml:"api" json:"api"`
	DataBase string `toml:"db" json:"db"`
	Utils    string `toml:"utils" json:"utils"`
}

// UnmarshalConfig unmarshal conifg file
// @params path: the path of config dir
// UnmarshalConfig 函数用于从指定的 TOML 配置文件中解析配置信息，并将其映射到 Config 结构体中。
// 该函数会读取配置文件，同时支持通过环境变量覆盖配置文件中的值。
// 参数:
// - configFilePath: 配置文件的路径。
// 返回值:
// - *Config: 解析后的配置结构体指针。
// - error: 如果解析过程中出现错误，返回错误信息；否则返回 nil。
func UnmarshalConfig(configFilePath string) (*Config, error) {
	// 设置要读取的配置文件路径
	viper.SetConfigFile(configFilePath)
	// 指定配置文件的类型为 TOML
	viper.SetConfigType("toml")
	// 启用自动从环境变量读取配置的功能
	viper.AutomaticEnv()
	// 设置环境变量的前缀为 "CNFT"，即只有以 "CNFT_" 开头的环境变量会被考虑
	viper.SetEnvPrefix("CNFT")
	// 创建一个字符串替换器，将配置文件中的点号（.）替换为下划线（_），以便环境变量可以使用下划线表示嵌套配置
	replacer := strings.NewReplacer(".", "_")
	// 设置环境变量键的替换器
	viper.SetEnvKeyReplacer(replacer)

	// 读取配置文件，如果出现错误则返回错误信息
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// 定义一个 Config 结构体变量，用于存储解析后的配置信息
	var c Config
	// 将读取的配置信息映射到 Config 结构体中，如果出现错误则返回错误信息
	if err := viper.Unmarshal(&c); err != nil {
		return nil, err
	}

	// 返回解析后的配置结构体指针和 nil 错误信息
	return &c, nil
}

// UnmarshalCmdConfig unmarshal conifg file
// @params path: the path of config dir
// UnmarshalCmdConfig 函数用于从已配置的文件中解析命令行相关的配置信息，并将其映射到 Config 结构体中。
// 该函数假设 viper 已经通过其他方式（如 UnmarshalConfig 函数）配置了配置文件路径和类型。
// 返回值:
// - *Config: 解析后的配置结构体指针。
// - error: 如果解析过程中出现错误，返回错误信息；否则返回 nil。
func UnmarshalCmdConfig() (*Config, error) {
	// 读取已配置的配置文件，如果出现错误则返回错误信息
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// 定义一个 Config 结构体变量，用于存储解析后的配置信息
	var c Config

	// 将读取的配置信息映射到 Config 结构体中，如果出现错误则返回错误信息
	if err := viper.Unmarshal(&c); err != nil {
		return nil, err
	}

	// 返回解析后的配置结构体指针和 nil 错误信息
	return &c, nil
}
