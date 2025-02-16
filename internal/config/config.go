package config

import (
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Config struct {
	Env        string `mapstructure:"env"`
	AuthKey    string `mapstructure:"authKey"`
	HTTPServer `mapstructure:"http_server"`
	DB         `mapstructure:"db"`
}

type HTTPServer struct {
	Address      string        `mapstructure:"address"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}
type DB struct {
	Host         string `mapstructure:"host"`
	Port         string `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"name"`
	DatabaseTest string `mapstructure:"test_db_name"`
}

func MustLoad(cfgName string) (*Config, error) {
	if cfgName != "" {
		viper.SetConfigFile(cfgName)
	} else {
		viper.SetConfigName("config.yaml")
		viper.AddConfigPath("./config")
	}

	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix("SERVICE")
	viper.AutomaticEnv()
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.WithStack(err)
	}

	conf := &Config{}
	if err := viper.Unmarshal(&conf); err != nil {
		return nil, errors.WithStack(err)
	}

	return conf, nil
}
