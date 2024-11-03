package kafka

import (
	"context"
	"github.com/carousell/ct-go/pkg/kafka"
	"github.com/carousell/ct-go/pkg/logger"
	"github.com/ct-logic-standard/config"
	"strings"
)

type Producer interface {
	Publish(ctx context.Context, key, topic string, message kafka.Message) error
}

type kafkaImpl struct {
	log        *logger.Logger
	conf       *config.Config
	publishers map[string]kafka.Publisher
}

func NewKafkaProducer(conf *config.Config) Producer {
	log := logger.MustNamed("kafka")
	var pubOpts = kafka.PublisherOptions{
		Brokers: strings.Split(conf.Kafka.Brokers, ","),
		Topic:   conf.Kafka.TopicAds,
	}
	pub, err := kafka.NewPublisher(pubOpts)
	if err != nil {
		log.Errorf("failed to NewPublisher: %v", err)
	}
	return &kafkaImpl{
		log:  log,
		conf: conf,
		publishers: map[string]kafka.Publisher{
			conf.Kafka.TopicAds: pub,
		},
	}
}

func (k *kafkaImpl) newPublisher(kafkaURL, topic string) error {
	var pubOpts = kafka.PublisherOptions{
		Brokers: strings.Split(kafkaURL, ","),
		Topic:   topic,
	}
	pub, err := kafka.NewPublisher(pubOpts)
	if err != nil {
		k.log.Errorf("failed to NewPublisher: %v", err)
		return err
	}
	k.publishers[topic] = pub
	return nil
}

func (k *kafkaImpl) Publish(ctx context.Context, key, topic string, message kafka.Message) error {
	publisher, found := k.publishers[topic]
	if !found {
		if err := k.newPublisher(k.conf.Kafka.Brokers, k.conf.Kafka.TopicAds); err != nil {
			k.log.Errorf("failed to create publisher: %v", err)
			return err
		}
	}
	err := publisher.Publish(ctx, key, message)
	if err != nil {
		k.log.Errorf("failed to produce message: %v", err)
		return err
	}
	return nil
}
