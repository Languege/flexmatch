// event_subscribers
// @author LanguageY++2013 2023/5/6 18:18
// @company soulgame
package event_subscribers

import (
	"github.com/google/uuid"
	"encoding/json"
	kafka_wrapper "github.com/Languege/flexmatch/service/match/wrappers/kafka"
	"github.com/Shopify/sarama"
	"github.com/Languege/flexmatch/service/match/proto/open"
)

var kafkaProducer = kafka_wrapper.NewAsyncProducer()

//KafkaMatchEventSubscribe kafka匹配事件订阅
func KafkaMatchEventSubscribe(topic string, ev *open.MatchEvent) {
	key := ev.MatchId
	if key == "" {
		key = uuid.NewString()
	}
	data, _ := json.Marshal(ev)
	message := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(data),
	}

	kafka_wrapper.MessageAdaptor(message)

	kafkaProducer.Input() <- message
}
