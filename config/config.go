package config

import (
	"time"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	App struct {
		Name         string        `env:"APP_NAME" envDefault:"ct-core-chat-exp"`
		GRPCAddr     string        `env:"GRPC_ADDR" envDefault:"localhost:9090"`
		HTTPAddr     string        `env:"HTTP_ADDR" envDefault:"localhost:8080"`
		StartTimeout time.Duration `env:"APP_START_TIMEOUT" envDefault:"1m"`
		StopTimeout  time.Duration `env:"APP_STOP_TIMEOUT" envDefault:"1m"`
	}
	Client struct {
		AdListingDomain      string `env:"AD_LISTING_DOMAIN" envDefault:"https://gateway.chotot.org/v2/public/ad-listing"`
		SchemaRegistryDomain string `env:"SCHEMA_REGISTRY_DOMAIN" envDefault:"http://schema.chotot.org"`
	}
	Redis struct {
		Addr     string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
		DB       int    `env:"REDIS_DB" envDefault:"0"`
		Username string `env:"REDIS_USERNAME" envDefault:""`
		Password string `env:"REDIS_PASSWORD" envDefault:""`
	}
	Kafka struct {
		Brokers       string        `env:"KAFKA_BROKERS" envDefault:"10.60.3.120,10.60.3.121,10.60.3.122"`
		ConsumerGroup string        `env:"KAFKA_CONSUMER_GROUP" envDefault:"worker_go_standard"`
		TopicAds      string        `env:"KAFKA_TOPIC_ADS" envDefault:"blocketdb.public.ads"`
		RetryInterval time.Duration `env:"KAFKA_RETRY_INTERNAL" envDefault:"5m"`
	}
	Rabbitmq struct {
		Host                     string `env:"RABBITMQ_HOST" envDefault:"10.60.7.124"`
		Port                     int    `env:"RABBITMQ_PORT" envDefault:"5672"`
		UserName                 string `env:"RABBITMQ_USERNAME" envDefault:"admin"`
		Password                 string `env:"RABBITMQ_PASSWORD" envDefault:"ctadmin"`
		RoutesAchillesAcceptedAd struct {
			Exchange   string   `env:"ROUTES_ACHILLES_ACCEPTED_AD_EXCHANGE" envDefault:"ad.event"`
			RoutingKey []string `env:"ROUTES_ACHILLES_ACCEPTED_AD_ROUTING_KEY" envDefault:"ad.accepted"`
			QueueName  string   `env:"ROUTES_ACHILLES_ACCEPTED_AD_QUEUE_NAME" envDefault:"ad.accepted.achilles_accepted_ad"`
		}
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
