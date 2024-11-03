package controller

import (
	"context"
	"github.com/carousell/ct-go/pkg/kafka"
	"github.com/carousell/ct-go/pkg/kafka/delaycalculator"
	"github.com/carousell/ct-go/pkg/logger"
	"github.com/ct-logic-standard/config"
	"github.com/ct-logic-standard/internal/usecase"
	"github.com/ct-logic-standard/pkg/codec"
	"go.uber.org/fx"
	"strings"
)

type KafkaWorker struct {
	ctx         context.Context
	log         *logger.Logger
	useCase     usecase.AdListingUC
	codec       *codec.Codec
	retryWorker kafka.RetryWorker
}

func (w *KafkaWorker) Run() error {
	err := w.retryWorker.Start(w.ctx)
	return err
}

func (w *KafkaWorker) Close() {
	w.retryWorker.Close()
}

func createRetryWorker(conf *config.Config, handlerFunc kafka.HandlerFunc) (kafka.RetryWorker, error) {
	confRetry := kafka.RetryWorkerConfig{
		MaxRetry:                 3,
		Brokers:                  strings.Split(conf.Kafka.Brokers, ","),
		Topic:                    conf.Kafka.TopicAds,
		GroupId:                  conf.Kafka.ConsumerGroup,
		HandlerFunc:              handlerFunc,
		AllowCreateTopicManually: true,
		DelayCalculator:          delaycalculator.NewLinearDelayCalculator(conf.Kafka.RetryInterval),
	}
	return kafka.NewRetryWorker(confRetry)
}

func NewKafkaWorker(
	lc fx.Lifecycle,
	conf *config.Config,
	useCase usecase.AdListingUC,
) *KafkaWorker {
	log := logger.MustNamed("kafka_worker")

	c, err := codec.NewCodec(conf.Client.SchemaRegistryDomain)
	if err != nil {
		log.Fatalf("failed to create codec: %v", err)
	}
	w := &KafkaWorker{
		ctx:     context.Background(),
		log:     log,
		useCase: useCase,
		codec:   c,
	}

	consumer, err := createRetryWorker(conf, w.Handler)
	if err != nil {
		log.Fatalf("failed to create worker: %v", err)
	}

	w.retryWorker = consumer

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				err := w.Run()
				if err != nil {
					log.Fatalf("failed to run kafka worker: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			w.Close()
			return nil
		},
	})

	return w
}
