package config

import (
	"time"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	App struct {
		Name         string        `env:"APP_NAME" envDefault:"ct-logic-api-document"`
		GRPCAddr     string        `env:"GRPC_ADDR" envDefault:"localhost:9090"`
		HTTPAddr     string        `env:"HTTP_ADDR" envDefault:"localhost:8080"`
		StartTimeout time.Duration `env:"APP_START_TIMEOUT" envDefault:"1m"`
		StopTimeout  time.Duration `env:"APP_STOP_TIMEOUT" envDefault:"1m"`
	}
	Mongo struct {
		ConnectionString string `env:"MONGO_CONNECTION_STRING" envDefault:"mongodb://bogus:bogus@localhost:27017/"` // vault
		PoolSize         uint64 `env:"MONGO_POOL_SIZE" envDefault:"20"`
		DBName           string `env:"MONGO_DB_NAME" envDefault:"ct_api_document"`
		Debug            bool   `env:"MONGO_DEBUG" envDefault:"false"`
	}
}

func Load() (*Config, error) {
	var conf Config
	if err := env.Parse(&conf); err != nil {
		return nil, err
	}
	return &conf, nil
}

func MustLoad() *Config {
	conf, err := Load()
	if err != nil {
		panic(err)
	}
	return conf
}
