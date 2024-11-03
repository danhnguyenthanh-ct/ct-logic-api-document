package controller

import (
	"context"
	"github.com/carousell/ct-go/pkg/logger"
	rabbitmq "github.com/carousell/ct-go/pkg/rabbit/v2"
	"github.com/ct-logic-standard/config"
	"github.com/ct-logic-standard/internal/usecase"
	"go.uber.org/fx"
)

type RabbitMQWorker struct {
	ctx          context.Context
	log          *logger.Logger
	useCase      usecase.AdListingUC
	rabbitClient rabbitmq.RabbitClient
}

func NewRabbitMQWorker(
	lc fx.Lifecycle,
	conf *config.Config,
	useCase usecase.AdListingUC,
) *RabbitMQWorker {
	log := logger.MustNamed("rabbitmq_worker")
	rabbitCli, err := rabbitmq.NewWithOption(rabbitmq.RabbitConfig{
		Host:     conf.Rabbitmq.Host,
		Port:     conf.Rabbitmq.Port,
		UserName: conf.Rabbitmq.UserName,
		Password: conf.Rabbitmq.Password,
	}, log.Unwrap())
	if err != nil {
		log.Fatal("failed to create rabbitmq client")
	}

	worker := &RabbitMQWorker{
		log:          log,
		useCase:      useCase,
		rabbitClient: rabbitCli,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				var exchangeConfig = rabbitmq.RabbitExchangeConfig{
					Name:       conf.Rabbitmq.RoutesAchillesAcceptedAd.Exchange,
					Kind:       "topic",
					Durable:    true,
					AutoDelete: false,
					Internal:   false,
					NoWait:     false,
					Args:       nil,
				}
				var queueConfig = rabbitmq.RabbitQueueConfig{
					Name:       conf.Rabbitmq.RoutesAchillesAcceptedAd.QueueName,
					Durable:    false,
					AutoDelete: false,
					Exclusive:  false,
					NoWait:     false,
					Args:       nil,
				}
				err = worker.Run(conf.Rabbitmq.RoutesAchillesAcceptedAd.RoutingKey,
					true,
					exchangeConfig,
					queueConfig,
					worker.Handler)
				if err != nil {
					log.Fatalf("failed to run kafka worker: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			err := worker.rabbitClient.Close()
			if err != nil {
				log.Fatalf("failed to stop rabbitmq worker: %v", err)
			}
			return err
		},
	})
	return worker
}

func (r *RabbitMQWorker) Run(routingKeys []string, autoACK bool, exchangeConf rabbitmq.RabbitExchangeConfig, queueConf rabbitmq.RabbitQueueConfig, handlerFunc rabbitmq.FuncResultV2) error {
	return r.rabbitClient.ConsumeV2(exchangeConf, routingKeys, autoACK, queueConf, handlerFunc)
}
