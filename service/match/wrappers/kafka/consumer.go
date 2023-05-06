package kafka_wrapper

import (
	"github.com/spf13/viper"
	"log"
	"context"
	"github.com/Shopify/sarama"
	"strings"
	"time"
	"os"
	"os/signal"
	"syscall"
)

/**
 *@author LanguageY++2013
 *2020/9/19 3:19 PM
 **/
func RegisterConsumer(consumer sarama.ConsumerGroupHandler, topics []string, group string) {
	if prefix := viper.GetString("kafka.prefix"); prefix  != ""{
		for i := range topics {
			topics[i] = prefix + topics[i]
		}
		group = prefix + group
	}

	if viper.GetBool("kafka.toLower") {
		for i := range topics {
			topics[i] = strings.ToLower(topics[i])
		}
		group = strings.ToLower(group)
	}

	config := sarama.NewConfig()

	ctx, cancel := context.WithCancel(context.Background())
	client, err := sarama.NewConsumerGroup(viper.GetStringSlice("kafka.bootstrapServers"), group, config)

	if err != nil {
		log.Panicf("[kafka]Error creating consumer group client: %v", err)
	}

	go func() {
		for {
			if err := client.Consume(ctx, topics, consumer); err != nil {
				log.Printf("[kafka]Error from consumer: %v  topics:%v  group:%s \n", err, topics, group)
				return
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
		}
	}()

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		<-sigterm

		cancel()
		err = client.Close()
		log.Printf("[kafka]closing comsumer %s err %v \n", group, err)
		time.Sleep(time.Second *3)
		os.Exit(0)
	}()
}

func RegisterConsumerFromBeginning(consumer sarama.ConsumerGroupHandler, topics []string, group string) {
	if prefix := viper.GetString("kafka.prefix"); prefix  != ""{
		for i := range topics {
			topics[i] = prefix + topics[i]
		}
		group = prefix + group
	}
	if viper.GetBool("kafka.toLower") {
		for i := range topics {
			topics[i] = strings.ToLower(topics[i])
		}
		group = strings.ToLower(group)
	}
	config := sarama.NewConfig()
	//config.Version = sarama.V2_0_0_0
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.MaxProcessingTime = time.Second

	ctx, cancel := context.WithCancel(context.Background())
	client, err := sarama.NewConsumerGroup(viper.GetStringSlice("kafka.bootstrapServers"), group, config)

	if err != nil {
		log.Panicf("[kafka]Error creating consumer group client: %v", err)
	}

	go func() {
		for {
			if err := client.Consume(ctx, topics, consumer); err != nil {
				log.Printf("[kafka]Error from consumer: %v  topics:%v  group:%s \n", err, topics, group)
				return
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
		}
	}()

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		<-sigterm

		cancel()
		err = client.Close()
		log.Printf("[kafka]closing comsumer %s err %v \n", group, err)
		time.Sleep(time.Second *3)
		os.Exit(0)
	}()
}


type SimpleConsumer struct {
	handler 	func(message *sarama.ConsumerMessage) error
	name 		string
}

func (consumer *SimpleConsumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (consumer *SimpleConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (consumer *SimpleConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	for message := range claim.Messages() {
		if consumer.handler(message) == nil {
			session.MarkMessage(message, "")
		}
	}

	return nil
}

//简单消费者
func NewSimpleConsumer(name string, handler func(message *sarama.ConsumerMessage) error) *SimpleConsumer {
	return &SimpleConsumer{
		name:    name,
		handler: handler,
	}
}


func RegisterSimpleConsumer(topic, consumerGroup string, handler func(message *sarama.ConsumerMessage) error) {
	consumer := NewSimpleConsumer(topic, handler)

	RegisterConsumer(consumer, []string{topic}, consumerGroup)
}

func RegisterSimpleConsumerFromBeginning(topic, consumerGroup string, handler func(message *sarama.ConsumerMessage) error) {
	consumer := NewSimpleConsumer(topic, handler)

	RegisterConsumerFromBeginning(consumer, []string{topic}, consumerGroup)
}

