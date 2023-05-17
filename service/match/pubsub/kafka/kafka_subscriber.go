// kafka
// @author LanguageY++2013 2023/5/11 09:54
// @company soulgame
package kafka

import (
	"encoding/json"
	"fmt"
	"github.com/Languege/flexmatch/common/logger"
	"github.com/Languege/flexmatch/service/match/proto/open"
	"github.com/Shopify/sarama"
	"go.uber.org/zap"
	kafka_wrapper "github.com/Languege/flexmatch/common/wrappers/kafka"
	"github.com/Languege/flexmatch/service/match/pubsub"
)
type eventHandler func(ev *open.MatchEvent) error

type KafkaSubscriber struct {
	brokerList    []string
	topics        []string
	group         string
	pubsub.EventHandlers
}

func NewKafkaSubscriber(brokerList []string, topics []string, group string) (*KafkaSubscriber, error) {
	s := &KafkaSubscriber{
		brokerList:    brokerList,
		topics:        topics,
		group:         group,
		EventHandlers: pubsub.EventHandlers{},
	}

	return s, nil
}



func (s *KafkaSubscriber) Setup(ses sarama.ConsumerGroupSession) error {
	logger.Infof("KafkaSubscriber setup, %s starting...", ses.MemberID())

	return nil
}

func (s *KafkaSubscriber) Cleanup(ses sarama.ConsumerGroupSession) error {
	logger.Info("KafkaSubscriber cleanup, %s stopping...", ses.MemberID())

	return nil
}

func (s *KafkaSubscriber) ConsumeClaim(ses sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) (retErr error) {
	//recover
	defer func() {
		if r := recover(); r != nil {
			retErr = fmt.Errorf("%v", r)
			logger.Errorf("KafkaSubscribe ConsumerClaim Err %s", retErr)
			return
		}
	}()

	logger.Infof("start consumer from topic %s, partitions %d, initialOffset %d", claim.Topic(), claim.Partition(), claim.InitialOffset())
	ch := claim.Messages()
	for {
		msg, ok := <-ch
		if !ok {
			return fmt.Errorf("stop consumer due to topic %s partition %d messages chan closed", claim.Topic(), claim.Partition())
		}

		ev := &open.MatchEvent{}
		err := json.Unmarshal(msg.Value, ev)
		if err != nil {
			logger.Errorf("json unmarshal  '%s' err %s", string(msg.Value), err)
			ses.MarkMessage(msg, "")
			continue
		}

		if err = s.Receive(ev); err != nil {
			logger.Errorw(fmt.Sprintf("handle match event err %s ", err), zap.Any("ev", ev))
		}

		ses.MarkMessage(msg, "")
	}
}

func (s KafkaSubscriber) Name() string {
	return "kafka"
}

func(s *KafkaSubscriber) Start() {
	kafka_wrapper.RegisterConsumerFromBeginning(s, s.brokerList, s.topics, s.group)
}
