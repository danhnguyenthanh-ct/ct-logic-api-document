package controller

import (
	"encoding/json"
	"github.com/ct-logic-standard/internal/entity"
	"github.com/segmentio/kafka-go"
	"github.com/streadway/amqp"
	"sync"
)

func (w *KafkaWorker) Handler(message kafka.Message) error {
	var ad entity.Ad
	err := w.codec.Decode(message, &ad)
	if err != nil {
		w.log.Error("Handler: failed to decode message", err)
		return err
	}
	w.log.Infof("%#v", ad)
	//TODO implement logic
	data, err := w.useCase.GetAdByListID(w.ctx, ad.ListID)
	w.log.Info(data, err)
	return nil
}

func (r *RabbitMQWorker) Handler(deliveries <-chan amqp.Delivery, wg *sync.WaitGroup, e error) {
	for d := range deliveries {
		go func() {
			var adAccepted entity.Ad
			err := json.Unmarshal(d.Body, &adAccepted)
			if err != nil {
				r.log.Error(err)
				return
			}
			r.log.Info(adAccepted, err)
			//TODO implement logic
			data, err := r.useCase.GetAdByListID(r.ctx, adAccepted.ListID)
			r.log.Info(data, err)
		}()
	}
}
