package config

import (
	"strings"

	"github.com/ProjectsTask/EasySwapBase/evm/erc"
	//"github.com/ProjectsTask/EasySwapBase/image"
	logging "github.com/ProjectsTask/EasySwapBase/logger"
	"github.com/ProjectsTask/EasySwapBase/stores/gdb"
	"github.com/spf13/viper"
)

type Config struct {
	Api        `toml:"api" json:"api"`
	ProjectCfg *ProjectCfg     `toml:"project_cfg" mapstructure:"project_cfg" json:"project_cfg"`
	Log        logging.LogConf `toml:"log" json:"log"`
	//ImageCfg       *image.Config     `toml:"image_cfg" mapstructure:"image_cfg" json:"image_cfg"`
	DB             gdb.Config        `toml:"db" json:"db"`
	Kv             *KvConf           `toml:"kv" json:"kv"`
	Evm            *erc.NftErc       `toml:"evm" json:"evm"`
	MetadataParse  *MetadataParse    `toml:"metadata_parse" mapstructure:"metadata_parse" json:"metadata_parse"`
	ChainSupported []*ChainSupported `toml:"chain_supported" mapstructure:"chain_supported" json:"chain_supported"`
}

type ProjectCfg struct {
	Name string `toml:"name" mapstructure:"name" json:"name"`
}

type Api struct {
	Port   string `toml:"port" json:"port"`
	MaxNum int64  `toml:"max_num" json:"max_num"`
}

type KvConf struct {
	Redis []*Redis `toml:"redis" mapstructure:"redis" json:"redis"`
}

type Redis struct {
	MasterName string `toml:"master_name" mapstructure:"master_name" json:"master_name"`
	Host       string `toml:"host" json:"host"`
	Type       string `toml:"type" json:"type"`
	Pass       string `toml:"pass" json:"pass"`
}

type MetadataParse struct {
	NameTags       []string `toml:"name_tags" mapstructure:"name_tags" json:"name_tags"`
	ImageTags      []string `toml:"image_tags" mapstructure:"image_tags" json:"image_tags"`
	AttributesTags []string `toml:"attributes_tags" mapstructure:"attributes_tags" json:"attributes_tags"`
	TraitNameTags  []string `toml:"trait_name_tags" mapstructure:"trait_name_tags" json:"trait_name_tags"`
	TraitValueTags []string `toml:"trait_value_tags" mapstructure:"trait_value_tags" json:"trait_value_tags"`
}

type ChainSupported struct {
	Name     string `toml:"name" mapstructure:"name" json:"name"`
	ChainID  int    `toml:"chain_id" mapstructure:"chain_id" json:"chain_id"`
	Endpoint string `toml:"endpoint" mapstructure:"endpoint" json:"endpoint"`
}

// UnmarshalConfig unmarshal conifg file
// @params path: the path of config dir
// UnmarshalConfig 函数用于从指定的配置文件中读取配置信息，并将其反序列化为 Config 结构体。
//
// 参数：
//     configFilePath string - 配置文件的路径。
//
// 返回值：
//     *Config - 反序列化后的配置对象。
//     error - 如果读取或反序列化配置文件时发生错误，将返回一个错误对象；否则返回 nil。
func UnmarshalConfig(configFilePath string) (*Config, error) {
	// 设置配置文件路径
	viper.SetConfigFile(configFilePath)
	// 设置配置文件类型为toml
	viper.SetConfigType("toml")
	// 自动读取环境变量
	viper.AutomaticEnv()
	// 设置环境变量前缀为CNFT
	viper.SetEnvPrefix("CNFT")
	// 替换环境变量中的.为_
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// 获取默认配置
	config, err := DefaultConfig()
	if err != nil {
		return nil, err
	}

	// 将配置文件内容反序列化为config对象
	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	// 返回配置对象
	return config, nil
}

func DefaultConfig() (*Config, error) {
	return &Config{}, nil
}
