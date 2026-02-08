package bootstrap

import "github.com/spf13/viper"

type Config struct {
	PublicJWTConfig struct {
		AccessSecretKey  string `mapstructure:"access_secret_key"`
		AccessIssuer     string `mapstructure:"access_issuer"`
		AccessExpire     int    `mapstructure:"access_expire"`
		RefreshSecretKey string `mapstructure:"refresh_secret_key"`
		RefreshIssuer    string `mapstructure:"refresh_issuer"`
		RefreshExpire    int    `mapstructure:"refresh_expire"`
	} `mapstructure:"public-jwt-config"`

	Database struct {
		Host     string `mapstructure:"host"`
		Port     string `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		Name     string `mapstructure:"name"`
	} `mapstructure:"database"`

	ServerConfig struct {
		Name           string `mapstructure:"name"`
		Domain         string `mapstructure:"domain"`
		Host           string `mapstructure:"host"`
		Port           string `mapstructure:"port"`
		Debug          string `mapstructure:"debug"`
		RequestTimeOut int    `mapstructure:"request-timeout"`
	} `mapstructure:"server"`
}

func InitConfig() *Config {
	viper.SetConfigName("config")

	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	config := Config{}
	err = viper.Unmarshal(&config)
	if err != nil {
		panic(err)
	}
	return &config
}
