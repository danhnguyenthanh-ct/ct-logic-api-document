package rabbitmq

import (
	"github.com/carousell/ct-go/pkg/logger"
	rabbitmq "github.com/carousell/ct-go/pkg/rabbit/v2"
	"github.com/ct-logic-standard/config"
)

type Producer interface {
	PublishV2(exchangeConf rabbitmq.RabbitExchangeConfig, routingKey string, message interface{}) error
}

type rabbitMQImpl struct {
	log            *logger.Logger
	rabbitMQClient rabbitmq.RabbitClient
}

func NewRabbitMQProducer(conf *config.Config) Producer {
	log := logger.MustNamed("rabbitmq")
	rabbitCl, err := rabbitmq.NewWithOption(rabbitmq.RabbitConfig{
		Host:     conf.Rabbitmq.Host,
		Port:     conf.Rabbitmq.Port,
		UserName: conf.Rabbitmq.UserName,
		Password: conf.Rabbitmq.Password,
	}, log.Unwrap())
	if err != nil {
		log.Errorf("failed to create rabbitmq client: %v", err)
	}
	return &rabbitMQImpl{
		log:            log,
		rabbitMQClient: rabbitCl,
	}
}

func (r *rabbitMQImpl) PublishV2(exchangeConf rabbitmq.RabbitExchangeConfig, routingKey string, message interface{}) error {
	err := r.rabbitMQClient.PublishV2(exchangeConf, routingKey, message)
	if err != nil {
		return err
	}
	return nil
}
