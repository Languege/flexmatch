// kafka
// @author LanguageY++2013 2023/5/11 09:33
// @company soulgame
package kafka

import (
	"github.com/Languege/flexmatch/service/match/proto/open"
	"github.com/Shopify/sarama"
	kafka_wrapper "github.com/Languege/flexmatch/common/wrappers/kafka"
	"github.com/google/uuid"
	"encoding/json"
)

type KafkaPublisher struct {
	brokerList []string
	producer sarama.AsyncProducer
}

func NewKafkaPublisher(brokerList []string) *KafkaPublisher {
	p := &KafkaPublisher{
		brokerList: brokerList,
	}

	p.producer = kafka_wrapper.NewAsyncProducer(brokerList)

	return p
}

func(p KafkaPublisher) Name() string {
	return "kafka"
}


func(p *KafkaPublisher)  Send(topic string, ev *open.MatchEvent) error {
	key := ev.MatchId
	if key == "" {
		key = uuid.NewString()
	}
	data, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	message := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(data),
	}

	p.producer.Input() <- message

	return nil
}
