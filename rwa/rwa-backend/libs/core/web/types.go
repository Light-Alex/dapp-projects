package web

type ServerConfig struct {
	Port                       int      `json:"port" yaml:"port"`
	BasePath                   string   `json:"basePath" yaml:"basePath"`
	Env                        string   `json:"env" yaml:"env"`
	EnableSignCheck            bool     `json:"enableSignCheck" yaml:"enableSignCheck"`
	EnableMarketMakerSignCheck bool     `json:"enableMarketMakerSignCheck" yaml:"enableMarketMakerSignCheck"`
	ApiKeys                    []string `json:"apiKeys" yaml:"apiKeys"`
	GinMode                    string   `json:"GinMode" yaml:"ginMode"`
}
