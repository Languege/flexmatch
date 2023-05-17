// kafka
// @author LanguageY++2013 2023/5/11 10:53
// @company soulgame
package kafka

import (
	"time"
	"log"
	"os"
	"os/signal"
	"syscall"
	"github.com/Shopify/sarama"
	"context"
	"github.com/Languege/flexmatch/common/logger"
)

func RegisterConsumerFromBeginning(consumer sarama.ConsumerGroupHandler, brokerList []string, topics []string, group string) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_0_0_0
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.MaxProcessingTime = time.Second

	ctx, cancel := context.WithCancel(context.Background())
	client, err := sarama.NewConsumerGroup(brokerList, group, config)

	if err != nil {
		logger.Panicf("[kafka]Error creating consumer group client: %v", err)
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
	}()
}

