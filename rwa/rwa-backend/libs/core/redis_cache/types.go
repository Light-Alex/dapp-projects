package redis_cache

type RedisConfig struct {
	Hosts      []string `json:"hosts" yaml:"hosts"`
	Db         int      `json:"db" yaml:"db"`
	MasterName string   `json:"masterName" yaml:"masterName"`
	UserName   string   `json:"userName" yaml:"userName"`
	Password   string   `json:"pass" yaml:"password"`
}
