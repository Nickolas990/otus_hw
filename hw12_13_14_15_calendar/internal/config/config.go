package config

type Config struct {
	HTTPServer struct {
		Host string `mapstructure:"host"`
		Port string `mapstructure:"port"`
	} `mapstructure:"http_server"`
	Logger struct {
		Level string `mapstructure:"level"`
	} `mapstructure:"logger"`
	StorageType string `mapstructure:"storage_type"`
	DBParams    struct {
		Host     string `mapstructure:"host"`
		Port     string `mapstructure:"port"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		Database string `mapstructure:"database"`
	} `mapstructure:"db_params"`
}
